.PHONY: build run test lint clean docker-up docker-down migrate-up migrate-down

APP_NAME=life-journaling-api
MIGRATIONS_DIR=migrations
DB_DSN?=postgres://postgres:postgres@localhost:5432/life_journaling?sslmode=disable

## build: Build the API binary
build:
	go build -o bin/$(APP_NAME) ./cmd/api

## run: Run the API locally
run:
	go run ./cmd/api

## test: Run all tests with coverage
test:
	go test -race -coverprofile=coverage.out ./...
	go tool cover -func=coverage.out

## test-coverage: Generate HTML coverage report
test-coverage: test
	go tool cover -html=coverage.out -o coverage.html

## lint: Run go vet
lint:
	go vet ./...

## clean: Remove build artifacts
clean:
	rm -rf bin/ coverage.out coverage.html

## docker-up: Start services with docker compose
docker-up:
	docker compose up -d

## docker-down: Stop services
docker-down:
	docker compose down

## migrate-up: Run database migrations up
migrate-up:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" up

## migrate-down: Run database migrations down
migrate-down:
	migrate -path $(MIGRATIONS_DIR) -database "$(DB_DSN)" down

## migrate-create: Create a new migration (usage: make migrate-create name=create_users)
migrate-create:
	migrate create -ext sql -dir $(MIGRATIONS_DIR) -seq $(name)

## mock-generate: Generate mocks for interfaces
mock-generate:
	mockgen -source=internal/usecase/ports.go -destination=internal/usecase/mocks/mock_ports.go -package=mocks

## help: Show this help
help:
	@grep -E '^## ' Makefile | sed 's/## //' | column -t -s ':'
