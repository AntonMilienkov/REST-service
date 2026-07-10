# REST-service

REST-сервис агрегации данных об онлайн-подписках пользователей.

Тестовое задание "Junior Golang Developer" для Effective Mobile.

## Запуск

```bash
docker compose up --build
```

Поднимутся два сервиса: `postgres` (с healthcheck) и `app` (ждёт готовности БД, сам накатывает миграции при старте). После этого сервис доступен на `http://localhost:8080`.

Swagger-документация: http://localhost:8080/swagger/index.html

Если нужно поменять настройки — скопируй `.env.example` в `.env` и поправь под себя (`docker-compose.yml` эти переменные не читает, там значения заданы явно; `.env` пригодится при локальном запуске через `go run` без Docker).

## Примеры запросов

### Создать подписку

```bash
curl -s -X POST http://localhost:8080/subscriptions -d '{
  "service_name": "Yandex Plus",
  "price": 400,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025"
}'
```

### Получить список всех подписок

```bash
curl -s http://localhost:8080/subscriptions
```

### Получить одну подписку по id

```bash
curl -s http://localhost:8080/subscriptions/<id>
```

### Обновить подписку

```bash
curl -s -X PUT http://localhost:8080/subscriptions/<id> -d '{
  "service_name": "Yandex Plus",
  "price": 500,
  "user_id": "60601fee-2bf1-4721-ae6f-7636e79a0cba",
  "start_date": "07-2025",
  "end_date": "12-2025"
}'
```

### Удалить подписку

```bash
curl -s -X DELETE http://localhost:8080/subscriptions/<id> -o /dev/null -w "%{http_code}\n"
```

### Суммарная стоимость подписок за период

```bash
curl -s "http://localhost:8080/subscriptions/total-cost?period_from=01-2025&period_to=12-2025"

# с фильтром по пользователю и/или названию сервиса
curl -s "http://localhost:8080/subscriptions/total-cost?user_id=60601fee-2bf1-4721-ae6f-7636e79a0cba&service_name=Yandex%20Plus&period_from=01-2025&period_to=12-2025"
```

## Технические решения и допущения

- Роутер — chi: нужны path-параметры (`/subscriptions/{id}`) и готовые middleware для request ID, логирования и recover из коробки.
- Формат дат `start_date`/`end_date` — строго `"MM-YYYY"` (месяц-год, день всегда 01), и в теле запроса/ответа, и в query-параметрах `period_from`/`period_to`.
- Допущение по `GET /subscriptions/total-cost`: суммируются `price` всех подписок, чей интервал `[start_date; end_date или бесконечность)` пересекается с заданным периодом `[period_from; period_to]`. `period_from`/`period_to` обязательны, `user_id`/`service_name` — опциональные фильтры.
- Миграции применяются автоматически при старте `app` (внутри `main.go`, до открытия пула соединений) — отдельного шага/сервиса в `docker-compose.yml` для этого нет. `docker-compose` ждёт готовности `postgres` через healthcheck (`depends_on: condition: service_healthy`) перед стартом `app`.
