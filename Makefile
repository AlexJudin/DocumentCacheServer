# Зависимости
dep:
	go mod tidy
.PHONY: dep

# Запуск тестов
test:
	go test ./test
.PHONY: test

# Запуск всей системы в докер контейнере
dk-start:
	docker compose up -d
.PHONY: dk-start

# Остановка всей системы
dk-stop:
	docker compose down
.PHONY: dk-stop

# Сборка всей системы
dk-build:
	docker compose build --no-cache doc_serv
.PHONY: dk-build

# Help
h:
	@echo "Usage: make [target]"
	@echo "  target is:"
	@echo "      dep - Обновление зависимостей"
	@echo "     test - Запуск всех тестов"
	@echo " dk-start - Запуск служб в докер контейнерах (окружения)"
	@echo "  dk-stop - Остановка запущенных служб (окружения)"
	@echo " dk-build - Сборка приложения"
.PHONY: h
help: h
.PHONY: help
