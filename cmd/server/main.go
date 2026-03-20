package main

import (
	"encoding/json"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"
	"gitverse.ru/pavlov-systems/distributed-integrity-orchestrator/internal/saga"
	"gitverse.ru/pavlov-systems/distributed-integrity-orchestrator/internal/storage"
	"gitverse.ru/pavlov-systems/distributed-integrity-orchestrator/internal/transport"
	"gitverse.ru/pavlov-systems/distributed-integrity-orchestrator/pkg/events"
)

func main() {
	// 1. Инициализация зависимостей
	rabbitAddr := "amqp://guest:guest@localhost:5672/"
	client, _ := transport.NewRabbitClient(rabbitAddr)
	defer client.Close()

	dbAddr := "postgres://dio_admin:dio_password@localhost:5432/orders_db?sslmode=disable"
	repo, _ := storage.NewOrderRepository(dbAddr)
	orc := saga.NewOrchestrator(repo, client)

	// 2. Фоновые процессы (Saga Workers)
	// Слушаем ответы от склада
	stockResponses, _ := client.Consume("stock.reserved")
	go func() {
		for d := range stockResponses {
			orc.ProcessStockResponse(d.Body)
		}
	}()

	// Имитатор Склада (для демонстрации цикла)
	orderEvents, _ := client.Consume("order.created")
	go func() {
		for d := range orderEvents {
			var ev events.OrderCreated
			json.Unmarshal(d.Body, &ev)
			time.Sleep(2 * time.Second) // "Работа" склада
			resp, _ := json.Marshal(events.StockReserved{OrderID: ev.OrderID, Success: true})
			client.Publish("stock.reserved", resp)
		}
	}()

	// 3. HTTP Роутинг (CQRS Interfaces)

	// COMMAND: Создать заказ
	http.HandleFunc("/orders", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			return
		}

		orderID := uuid.New().String()
		// Write Model: Сохраняем в БД
		repo.CreateOrder(orderID, "PRO-UNIT", 1)

		// Saga Start: Событие в шину
		ev, _ := json.Marshal(events.OrderCreated{OrderID: orderID, ItemID: "PRO-UNIT", Quantity: 1})
		client.Publish("order.created", ev)

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusAccepted)
		json.NewEncoder(w).Encode(map[string]string{"id": orderID, "status": "PENDING"})
	})

	// QUERY: Проверить статус заказа (Read Model)
	http.HandleFunc("/status", func(w http.ResponseWriter, r *http.Request) {
		id := r.URL.Query().Get("id")
		status, _ := repo.GetOrderStatus(id)
		json.NewEncoder(w).Encode(map[string]string{"id": id, "status": status})
	})

	log.Println("[SYSTEM] API запущен на :8080. Доступны /orders (POST) и /status (GET)")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
