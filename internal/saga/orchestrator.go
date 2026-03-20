package saga

import (
	"encoding/json"
	"log"

	"gitverse.ru/pavlov-systems/distributed-integrity-orchestrator/pkg/events"
)

// Описываем, что Оркестратору нужно от БД и Шины
type OrderStorage interface {
	UpdateOrderStatus(id string, status string) error
}

type EventBus interface {
	Publish(queueName string, body []byte) error
}

// Orchestrator — мозг системы
type Orchestrator struct {
	storage OrderStorage
	bus     EventBus
}

func NewOrchestrator(s OrderStorage, b EventBus) *Orchestrator {
	return &Orchestrator{storage: s, bus: b}
}

// ProcessStockResponse — это точка принятия решения (Commit или Rollback)
func (o *Orchestrator) ProcessStockResponse(payload []byte) {
	var msg events.StockReserved
	if err := json.Unmarshal(payload, &msg); err != nil {
		log.Printf("[SAGA] Ошибка парсинга: %v", err)
		return
	}

	if msg.Success {
		// COMMIT: Склад подтвердил наличие
		err := o.storage.UpdateOrderStatus(msg.OrderID, "COMPLETED")
		if err != nil {
			log.Printf("[SAGA] Ошибка фиксации заказа %s: %v", msg.OrderID, err)
			return
		}
		log.Printf("[SAGA] УСПЕХ: Заказ %s переведен в статус COMPLETED", msg.OrderID)
	} else {
		// ROLLBACK: Склад отказал (компенсирующая транзакция)
		err := o.storage.UpdateOrderStatus(msg.OrderID, "REJECTED_BY_STOCK")
		if err != nil {
			log.Printf("[SAGA] Ошибка отката заказа %s: %v", msg.OrderID, err)
			return
		}
		log.Printf("[SAGA] ОТКАТ: Заказ %s отклонен. Причина: %s", msg.OrderID, msg.ErrorMessage)
	}
}
