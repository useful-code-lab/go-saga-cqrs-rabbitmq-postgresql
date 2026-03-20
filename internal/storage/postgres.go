package storage

import (
	"database/sql"

	_ "github.com/lib/pq"
)

// OrderRepository управляет состоянием заказов (Write Model)
type OrderRepository struct {
	db *sql.DB
}

// NewOrderRepository создает новое подключение к БД
func NewOrderRepository(connStr string) (*OrderRepository, error) {
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, err
	}

	// Проверка связи
	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &OrderRepository{db: db}, nil
}

// CreateOrder сохраняет новый заказ со статусом PENDING
func (r *OrderRepository) CreateOrder(id string, itemID string, qty int) error {
	query := `INSERT INTO orders (id, item_id, quantity, status) VALUES ($1, $2, $3, $4)`
	_, err := r.db.Exec(query, id, itemID, qty, "PENDING")
	return err
}

// UpdateOrderStatus изменяет статус (используется Сагой для Commit/Rollback)
func (r *OrderRepository) UpdateOrderStatus(id string, status string) error {
	query := `UPDATE orders SET status = $1, updated_at = NOW() WHERE id = $2`
	_, err := r.db.Exec(query, status, id)
	return err
}

// GetOrderStatus — простейшая Read Model для проверки
func (r *OrderRepository) GetOrderStatus(id string) (string, error) {
	var status string
	err := r.db.QueryRow("SELECT status FROM orders WHERE id = $1", id).Scan(&status)
	return status, err
}
