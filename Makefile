COMPOSE_DIR := deployment/roadmap-local
COMPOSE_PROJECT := roadmap
APP_SERVICE := roadmap
FRONTEND_SERVICE := frontend

DROP_VOLUMES ?= 0
DOWN_VOL_ARGS := $(if $(filter 1,$(DROP_VOLUMES)),-v,)

.PHONY: help from-zero app frontend down build up

help:
	@echo "  make from-zero   — всё с нуля (pg + api + frontend). DROP_VOLUMES=1 — снести тома."
	@echo "  make app         — пересборка только API без кэша и up"
	@echo "  make frontend    — пересборка только frontend без кэша и up"
	@echo "  make down        — остановить всё (DROP_VOLUMES=1 — и тома)"
	@echo "  make build       — собрать api + frontend"
	@echo "  make up          — поднять без сборки"
	@echo ""
	@echo "  UI:  http://localhost:3000"
	@echo "  API: http://localhost:8080"

down:
	cd $(COMPOSE_DIR) && docker compose -p $(COMPOSE_PROJECT) down --rmi local $(DOWN_VOL_ARGS)

build:
	@echo "Сборка $(APP_SERVICE) + $(FRONTEND_SERVICE)..."
	cd $(COMPOSE_DIR) && docker compose --progress=quiet -p $(COMPOSE_PROJECT) build $(APP_SERVICE) $(FRONTEND_SERVICE)
	@echo "Готово."

build-all:
	@echo "Сборка всех образов без кэша..."
	cd $(COMPOSE_DIR) && docker compose --progress=quiet -p $(COMPOSE_PROJECT) build --no-cache
	@echo "Готово."

up:
	cd $(COMPOSE_DIR) && docker compose -p $(COMPOSE_PROJECT) up -d

from-zero: down build-all up

app:
	@echo "Пересборка $(APP_SERVICE) без кэша..."
	cd $(COMPOSE_DIR) && docker compose --progress=quiet -p $(COMPOSE_PROJECT) build --no-cache $(APP_SERVICE)
	@echo "Готово."
	$(MAKE) up

frontend:
	@echo "Пересборка $(FRONTEND_SERVICE) без кэша..."
	cd $(COMPOSE_DIR) && docker compose --progress=quiet -p $(COMPOSE_PROJECT) build --no-cache $(FRONTEND_SERVICE)
	@echo "Готово."
	$(MAKE) up
