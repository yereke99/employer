# Employer API

CRUD-сервис для сотрудников на **Go** + **Postgres**.  
Есть Swagger-документация.

## Как запустить

1. Сгенерируй Swagger-доки:
   ```
   make swagger
   ```

2. Подними Postgres и приложение:
   ```
   make run-app
   ```

3. Проверка здоровья:
   ```
   curl http://localhost:8080/health
   ```

4. Front-End:  
   https://meily.kz

## Тесты
```
make up-test
```

## Основные команды
- `make up` — поднять только Postgres  
- `make down` — остановить сервисы  
- `make logs` — посмотреть логи приложения  
- `make build-app` — пересобрать образ  
