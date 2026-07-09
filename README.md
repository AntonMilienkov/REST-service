# REST-service

REST-сервис агрегации данных об онлайн-подписках пользователей.

Тестовое задание "Junior Golang Developer" для Effective Mobile.

## Технические решения

- Роутер — chi: нужны path-параметры (`/subscriptions/{id}`) и готовые middleware для request ID, логирования и recover из коробки.
