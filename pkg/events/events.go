package events

import "time"

// OrderCreated — событие, инициирующее Сагу (Command/Write Model)
type OrderCreated struct {
	OrderID   string    `json:"order_id"`
	ItemID    string    `json:"item_id"`
	Quantity  int       `json:"quantity"`
	CreatedAt time.Time `json:"created_at"`
}

// StockReserved — ответ от сервиса склада (Saga Step Result)
type StockReserved struct {
	OrderID    string `json:"order_id"`
	Success    bool   `json:"success"`
	ErrorMessage string `json:"error_message,omitempty"`
}
