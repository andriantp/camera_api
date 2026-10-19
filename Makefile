PRODUCT = camera
PROJECT = camera_api

COMPOSE = docker/docker-compose.yml
NETWORK = camera_network
FILE    = api

.PHONY: network go swagger-install swagger-version swagger-init dev prod ps logs logs100

network:
	@echo "📁 Preparing network..."
	@if ! docker network inspect $(NETWORK) > /dev/null 2>&1; then \
		echo "Network $(NETWORK) not found, creating..."; \
		docker network create --driver=bridge $(NETWORK); \
	else \
		echo "Network $(NETWORK) already exists."; \
	fi

go:
	go version

swagger-install:
	go get -tool github.com/swaggo/swag/cmd/swag@v1.16.6

swagger-version:
	go tool swag --version

swagger-init:
	go tool swag init

dev:
	ENV=development go run .

prod: network
	@echo "🐳 Starting Deploy"
	ENV=production PRODUCT=$(PRODUCT) FILE=$(FILE) \
		docker compose -p $(PROJECT) -f $(COMPOSE) up --force-recreate -d --build

ps:
	docker compose -p $(PROJECT) -f $(COMPOSE) ps

logs:
	docker compose -p $(PROJECT) -f $(COMPOSE) logs -f camera_api

logs100:
	docker compose -p $(PROJECT) -f $(COMPOSE) logs --tail 100 -f camera_api