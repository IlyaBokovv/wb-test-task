## Запуск

```powershell
docker compose up -d
go run ./migrations/auto.go
go run ./cmd/order-service
```

Откройте http://localhost:8085.
При нажатии ``Создать заказ (POST)`` создается заказ по умолчанию с измененным uid

Настройки — в [`./.env.example`](/.env.example) (переменные окружения). Таблицы создаются GORM AutoMigrate при запуске ```./migrations/auto.go``` . Контейнер PostgreSQL создаёт БД `orders` и пользователя `orders` с нужными правами.
