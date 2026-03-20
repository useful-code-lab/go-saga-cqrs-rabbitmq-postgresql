package main

import (
	"log"

	"github.com/google/uuid"
	"gitverse.ru/pavlov-systems/distributed-integrity-orchestrator/internal/storage"
	"gitverse.ru/pavlov-systems/distributed-integrity-orchestrator/internal/transport"
)

func main() {
	// Создаем клиент (используем credentials из docker-compose)
	rabbitAddr := "amqp://guest:guest@localhost:5672/"
	client, err := transport.NewRabbitClient(rabbitAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// Инициализация БД (Write Model)
	dbAddr := "postgres://dio_admin:dio_password@localhost:5432/orders_db?sslmode=disable"
	repo, err := storage.NewOrderRepository(dbAddr)
	if err != nil {
		log.Fatal("Ошибка БД:", err)
	}

	// 2. Инициализируем необходимые очереди для Саги
	queues := []string{"order.created", "stock.reserved"}
	for _, q := range queues {
		if err := client.SetupQueue(q); err != nil {
			log.Fatalf("Ошибка создания очереди %s: %v", q, err)
		}
	}

	// 3. ТЕСТ: Создаем заказ в БД
	testID := uuid.New().String()
	err = repo.CreateOrder(testID, "SKU-99", 2)
	if err != nil {
		log.Printf("Ошибка создания заказа: %v", err)
	} else {
		log.Printf("Заказ %s успешно сохранен в БД со статусом PENDING", testID)
	}

	// ТЕСТ: Отправляем сообщение
	testMessage := []byte(`{"order_id": "test-123", "status": "ping"}`)
	err = client.Publish("order.created", testMessage)
	if err != nil {
		log.Printf("Ошибка отправки: %v", err)
	}

	// ТЕСТ: Читаем сообщение
	msgs, _ := client.Consume("order.created")
	go func() {
		for d := range msgs {
			log.Printf("Успешно получено тестовое сообщение: %s", d.Body)
		}
	}()

	status, _ := repo.GetOrderStatus(testID)
	log.Printf("Текущий статус заказа в БД: %s", status)

	// Оставляем программу запущенной
	select {}

}
