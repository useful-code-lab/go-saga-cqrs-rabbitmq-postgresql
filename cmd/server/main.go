package main

import (
	"encoding/json"
	"log"
	"time"

	"github.com/google/uuid"
	"gitverse.ru/pavlov-systems/distributed-integrity-orchestrator/internal/saga"
	"gitverse.ru/pavlov-systems/distributed-integrity-orchestrator/internal/storage"
	"gitverse.ru/pavlov-systems/distributed-integrity-orchestrator/internal/transport"
	"gitverse.ru/pavlov-systems/distributed-integrity-orchestrator/pkg/events"
)

func main() {
	// 1. Инфраструктура
	rabbitAddr := "amqp://guest:guest@localhost:5672/"
	client, _ := transport.NewRabbitClient(rabbitAddr)
	defer client.Close()

	dbAddr := "postgres://dio_admin:dio_password@localhost:5432/orders_db?sslmode=disable"
	repo, _ := storage.NewOrderRepository(dbAddr)

	// 2. Инициализация Оркестратора
	orc := saga.NewOrchestrator(repo, client)

	// 3. Подписка на ответы от Склада (Stock Service)
	stockMsgs, _ := client.Consume("stock.reserved")
	go func() {
		for d := range stockMsgs {
			orc.ProcessStockResponse(d.Body)
		}
	}()

	// 4. Имитация создания нового заказа (Command CQRS)
	orderID := uuid.New().String()
	_ = repo.CreateOrder(orderID, "IPHONE-15", 1)
	log.Printf("[API] Создан заказ %s, статус PENDING", orderID)

	// Отправляем событие в шину для склада
	event := events.OrderCreated{OrderID: orderID, ItemID: "IPHONE-15", Quantity: 1, CreatedAt: time.Now()}
	body, _ := json.Marshal(event)
	_ = client.Publish("order.created", body)

	// Оставляем сервер работать
	log.Println("[SYSTEM] Оркестратор запущен. Ожидание событий...")
	select {}
}
