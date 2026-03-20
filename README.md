# Distributed Integrity Orchestrator (DIO) 🚀

### Референсная реализация паттернов Saga и CQRS на языке Go

Проект демонстрирует решение проблемы **распределенных транзакций** в микросервисной архитектуре. Вместо тяжелых двухфазных коммитов (2PC) используется асинхронная **Saga (Оркестрация)** для обеспечения Eventual Consistency.

---

## 🏗 Архитектура системы

Проект разделен на три логических слоя, взаимодействующих через **RabbitMQ**:

1.  **CQRS (Command Side)**: HTTP POST создает заказ в PostgreSQL со статусом `PENDING`.
2.  **Saga Orchestrator**: «Мозг» системы. Слушает ответы от склада и принимает решение: подтвердить заказ (`COMPLETED`) или выполнить откат (`REJECTED`).
3.  **CQRS (Query Side)**: Мгновенное чтение текущего состояния заказа без блокировки транзакций.


🛠 Управление проектом (Makefile)
Автоматизация типовых операций для быстрого развертывания и тестирования.
Команда	Описание
- `make up`	Развернуть PostgreSQL и RabbitMQ (Docker).
- `make down`	Остановить и удалить инфраструктуру.
- `make db-init`	Инициализировать схему БД (init.sql).
- `make run`	Запустить основной сервис Go.
- `make test-order`	Создать тестовый заказ через curl.
- `make tidy`	Очистить и обновить зависимости Go.

## Демонстрация работы Saga-оркестратора: от HTTP-запроса до завершения транзакции

```
docker-compose up -d
```

Админка RabbitMQ через браузер:
URL: http://localhost:15672
Login: guest
Password: guest

``bash
go run cmd/server/main.go
``

![2 очереди со ставтусами running](/docs/images/rabbitmq.png)

```bash
aleksey@aleksey:~/distributed-integrity-orchestrator$ go run cmd/server/main.go
2026/03/20 12:19:38 Транспортный слой успешно запущен и готов к работе!
2026/03/20 12:19:38 Успешно получено тестовое сообщение: {"order_id": "test-123", "status": "ping"}
```

Создание БД
```bash
aleksey@aleksey:~/distributed-integrity-orchestrator$ docker exec -i dio_db_orders psql -U dio_admin -d orders_db < scripts/init.sql
```

Проверки после выполнения:
```bash
aleksey@aleksey:~/distributed-integrity-orchestrator$ go run cmd/server/main.go
2026/03/20 12:26:30 Заказ d1bfa8e3-214d-49cd-83f2-332faba74533 успешно сохранен в БД со статусом PENDING
2026/03/20 12:26:30 Текущий статус заказа в БД: PENDING
```
```bash
aleksey@aleksey:~/distributed-integrity-orchestrator$ docker exec -it dio_db_orders psql -U dio_admin -d orders_db
psql (15.17)
Type "help" for help.

orders_db=# \dt
          List of relations
 Schema |  Name  | Type  |   Owner   
--------+--------+-------+-----------
 public | orders | table | dio_admin
(1 row)

orders_db=# SELECT * FROM orders;
                  id                  | item_id | quantity | status  |         created_at         |         updated_at         
--------------------------------------+---------+----------+---------+----------------------------+----------------------------
 d1bfa8e3-214d-49cd-83f2-332faba74533 | SKU-99  |        2 | PENDING | 2026-03-20 09:26:30.933481 | 2026-03-20 09:26:30.933481
(1 row)

orders_db=# 
```


Команда сейчас 
```bash
aleksey@aleksey:~/distributed-integrity-orchestrator$ go run cmd/server/main.go
2026/03/20 12:35:48 Заказ b5501cdb-1a62-4d9b-8629-b01569804557 успешно сохранен в БД со статусом PENDING
2026/03/20 12:35:48 Успешно получено тестовое сообщение: {"order_id": "test-123", "status": "ping"}
2026/03/20 12:35:48 Текущий статус заказа в БД: PENDING
```

```bash
aleksey@aleksey:~/distributed-integrity-orchestrator$ go run cmd/server/main.go
2026/03/20 13:02:24 [SYSTEM] API запущен на :8080. Доступны /orders (POST) и /status (GET)
2026/03/20 13:02:26 [SAGA] УСПЕХ: Заказ 2f89b49f-5d18-4215-a823-2953e5476fe2 переведен в статус COMPLETED
2026/03/20 13:02:40 [SAGA] УСПЕХ: Заказ 6983d6d9-25de-4e87-be12-34fb76cf1194 переведен в статус COMPLETED
```

Реализация полного цикла программы
```bash
aleksey@aleksey:~/distributed-integrity-orchestrator$ curl -X POST http://localhost:8080/orders
{"id":"6983d6d9-25de-4e87-be12-34fb76cf1194","status":"PENDING"}
aleksey@aleksey:~/distributed-integrity-orchestrator$ curl -X POST http://localhost:8080/orders
{"id":"c678db0a-1d84-40b9-9bd2-36b927cc8973","status":"PENDING"}
aleksey@aleksey:~/distributed-integrity-orchestrator$ curl -X POST http://localhost:8080/status
{"id":"","status":""}
aleksey@aleksey:~/distributed-integrity-orchestrator$ curl -X POST http://localhost:8080/status?id=c678db0a-1d84-40b9-9bd2-36b927cc8973
{"id":"c678db0a-1d84-40b9-9bd2-36b927cc8973","status":"COMPLETED"}
aleksey@aleksey:~/distributed-integrity-orchestrator$ curl -X POST http://localhost:8080/status?id=6983d6d9-25de-4e87-be12-34fb76cf1194
{"id":"6983d6d9-25de-4e87-be12-34fb76cf1194","status":"COMPLETED"}
aleksey@aleksey:~/distributed-integrity-orchestrator$ 
```

Реализовано корректное завершение
```bash
aleksey@aleksey:~/distributed-integrity-orchestrator$ go run cmd/server/main.go
2026/03/20 13:11:52 [SYSTEM] API запущен на :8080. Доступны /orders (POST) и /status (GET)
2026/03/20 13:11:52 [SYSTEM] API запущен на :8080
2026/03/20 13:12:10 [SAGA] УСПЕХ: Заказ eaece8cb-9059-4c55-89da-01e2da11aa18 переведен в статус COMPLETED
^C2026/03/20 13:13:45 
[SYSTEM] Завершение работы...
2026/03/20 13:13:45 [SYSTEM] Все соединения закрыты. Выход.
```