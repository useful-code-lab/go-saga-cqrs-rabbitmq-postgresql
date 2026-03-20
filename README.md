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

aleksey@aleksey:~/distributed-integrity-orchestrator$ go run cmd/server/main.go
2026/03/20 12:19:38 Транспортный слой успешно запущен и готов к работе!
2026/03/20 12:19:38 Успешно получено тестовое сообщение: {"order_id": "test-123", "status": "ping"}


