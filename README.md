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