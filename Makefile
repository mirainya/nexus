.PHONY: build run dev proto clean test lint migrate console-build console-dev docker-build docker-up docker-down migrate-up migrate-down migrate-create

APP_NAME=nexus
BUILD_DIR=bin

build:
	go build -o $(BUILD_DIR)/$(APP_NAME) ./cmd/server
	go build -o $(BUILD_DIR)/migrate ./cmd/migrate

run: build
	./$(BUILD_DIR)/$(APP_NAME)

dev:
	NEXUS_DEV=1 go run ./cmd/server

proto:
	protoc --go_out=. --go-grpc_out=. api/proto/*.proto

clean:
	rm -rf $(BUILD_DIR)

test:
	go test ./... -v

lint:
	golangci-lint run ./...

migrate-up:
	go run ./cmd/migrate -cmd up -dsn "$(NEXUS_DATABASE_URL)"

migrate-down:
	go run ./cmd/migrate -cmd down -dsn "$(NEXUS_DATABASE_URL)"

migrate-create:
	@read -p "Migration name: " name; \
	touch database/migrations/$$(date +%Y%m%d%H%M%S)_$$name.up.sql; \
	touch database/migrations/$$(date +%Y%m%d%H%M%S)_$$name.down.sql; \
	echo "Created migration: $$name"

docker-build:
	docker build -t $(APP_NAME) .

docker-up:
	docker compose up -d

docker-down:
	docker compose down

console-build:
	cd console && npm run build

console-dev:
	cd console && npm run dev
