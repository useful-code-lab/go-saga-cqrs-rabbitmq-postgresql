package transport

import (
	"fmt"
	"log"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// RabbitClient инкапсулирует логику работы с брокером
type RabbitClient struct {
	Conn    *amqp.Connection
	Channel *amqp.Channel
}

// NewRabbitClient — конструктор с логикой переподключения (Resilience)
func NewRabbitClient(connString string) (*RabbitClient, error) {
	var conn *amqp.Connection
	var err error

	// Системный подход: пробуем подключиться несколько раз, если RabbitMQ еще грузится
	for i := 0; i < 5; i++ {
		conn, err = amqp.Dial(connString)
		if err == nil {
			break
		}
		log.Printf("Попытка подключения к RabbitMQ #%d не удалась, ждем...", i+1)
		time.Sleep(5 * time.Second)
	}

	if err != nil {
		return nil, fmt.Errorf("не удалось подключиться к RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		return nil, err
	}

	return &RabbitClient{
		Conn:    conn,
		Channel: ch,
	}, nil
}

// Close корректно закрывает ресурсы при выходе
func (rc *RabbitClient) Close() {
	rc.Channel.Close()
	rc.Conn.Close()
}

// SetupQueue объявляет очередь (идемпотентная операция)
func (rc *RabbitClient) SetupQueue(name string) error {
	_, err := rc.Channel.QueueDeclare(
		name, // имя очереди
		true, // надежная (сохраняется на диске)
		false,
		false,
		false,
		nil,
	)
	return err
}

// Publish отправляет сообщение в очередь
func (rc *RabbitClient) Publish(queueName string, body []byte) error {
	return rc.Channel.Publish(
		"",
		queueName,
		false,
		false,
		amqp.Publishing{
			ContentType:  "application/json",
			Body:         body,
			DeliveryMode: amqp.Persistent, // сообщение выживет при перезагрузке Rabbit
		},
	)
}

// Consume подписывается на сообщения
func (rc *RabbitClient) Consume(queueName string) (<-chan amqp.Delivery, error) {
	return rc.Channel.Consume(
		queueName,
		"",   // consumer tag
		true, // auto-ack (для простоты)
		false,
		false,
		false,
		nil,
	)
}
