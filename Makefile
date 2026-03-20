# Переменные
APP_NAME=distributed-integrity-orchestrator

.PHONY: up down run clean test-order

# Запуск инфраструктуры (БД и Rabbit)
up:
	docker-compose up -d

# Остановка всего
down:
	docker-compose down

# Запуск самого приложения
run:
	go run cmd/server/main.go

# Создание таблицы (первичная настройка)
db-init:
	docker exec -i dio_db_orders psql -U dio_admin -d orders_db < scripts/init.sql

# Тестовый запрос через curl (создание заказа)
test-order:
	@curl -X POST http://localhost:8080/orders | jq .

# Очистка зависимостей
tidy:
	go mod tidy
