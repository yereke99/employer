# ==== настройки ====
COMPOSE := docker compose -f build/docker-compose.yml
APP_IMAGE := employer_app
APP_NAME := employer_app
PG_CONTAINER := employee_postgres      # имя контейнера Postgres (см. compose)
COMPOSE_NETWORK := build_default       # сеть compose (для папки build это build_default)

.PHONY: test swagger up down up-test build-app run-app logs health stop-app

# Прогон тестов
test:
	go test ./... -count=1

# Поднять только Postgres
up:
	$(COMPOSE) up -d postgres

# Остановить Postgres + удалить volume + сироты
down:
	$(COMPOSE) down -v --remove-orphans
	- docker rm -f $(APP_NAME) 2>/dev/null || true

# Поднять Postgres, подождать готовности, затем тесты
up-test: up
	@echo ">>> Waiting for Postgres to be ready..."
	@for i in $$(seq 1 30); do \
		if docker exec $(PG_CONTAINER) pg_isready -U postgres -d employee >/dev/null 2>&1 ; then \
			echo "Postgres is ready"; \
			break; \
		fi; \
		echo "waiting... ($$i)"; \
		sleep 1; \
	done
	@echo ">>> Running tests..."
	go test ./... -count=1
	@echo ">>> Tests finished!"

# Собрать образ приложения
build-app:
	docker build -t $(APP_IMAGE) -f build/Dockerfile .

# Запустить приложение + Postgres (приложение подключаем к сети compose!)
run-app: build-app up
	@echo ">>> Waiting for Postgres to be ready..."
	@for i in $$(seq 1 30); do \
		if docker exec $(PG_CONTAINER) pg_isready -U postgres -d employee >/dev/null 2>&1 ; then \
			echo "Postgres is ready"; \
			break; \
		fi; \
		echo "waiting... ($$i)"; \
		sleep 1; \
	done
	- docker rm -f $(APP_NAME) 2>/dev/null || true
	docker run -d --name $(APP_NAME) \
		--network $(COMPOSE_NETWORK) \
		-p 8080:8080 \
		--env-file ./.env \
		-e DB_HOST=postgres \
		$(APP_IMAGE)
	@echo ">>> App is running: http://localhost:8080/health"

# Остановить только приложение
stop-app:
	- docker rm -f $(APP_NAME) 2>/dev/null || true

# Логи приложения
logs:
	docker logs -f $(APP_NAME)

# Быстрый health-check
health:
	@echo "GET /health"
	@curl -s -i http://localhost:8080/health || true
