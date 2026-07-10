# REST-service

REST-сервис агрегации данных об онлайн-подписках пользователей.

Тестовое задание "Junior Golang Developer" для Effective Mobile.

## Технические решения

- Роутер — chi: нужны path-параметры (`/subscriptions/{id}`) и готовые middleware для request ID, логирования и recover из коробки.
- Допущение по `GET /subscriptions/total-cost`: суммируются `price` всех подписок, чей интервал `[start_date; end_date или бесконечность)` пересекается с заданным периодом `[period_from; period_to]`. `period_from`/`period_to` обязательны, `user_id`/`service_name` — опциональные фильтры.
- Миграции применяются автоматически при старте `app` (внутри `main.go`, до открытия пула соединений) — отдельного шага/сервиса в `docker-compose.yml` для этого нет. `docker-compose` ждёт готовности `postgres` через healthcheck (`depends_on: condition: service_healthy`) перед стартом `app`.
