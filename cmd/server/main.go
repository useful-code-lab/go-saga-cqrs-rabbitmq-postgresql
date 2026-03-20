package main

import (
	"log"

	"gitverse.ru/pavlov-systems/distributed-integrity-orchestrator/internal/transport"
)

func main() {
	// 1. Создаем клиент (используем credentials из docker-compose)
	rabbitAddr := "amqp://guest:guest@localhost:5672/"
	client, err := transport.NewRabbitClient(rabbitAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	// 2. Инициализируем необходимые очереди для Саги
	queues := []string{"order.created", "stock.reserved"}
	for _, q := range queues {
		if err := client.SetupQueue(q); err != nil {
			log.Fatalf("Ошибка создания очереди %s: %v", q, err)
		}
	}

	log.Println("Транспортный слой успешно запущен и готов к работе!")

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

	// Оставляем программу запущенной
	select {}

}
