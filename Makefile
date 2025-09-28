APP_NAME=flowkeeper
BIN_DIR=bin
CONFIG=config/config.yaml
DB_DIR=data

.PHONY: all build run clean test lint test

all: build

build:
	@echo "Building $(APP_NAME)..."
	@go build -o $(BIN_DIR)/$(APP_NAME) ./cmd

run: build
	@echo "Running $(APP_NAME)..."
	@mkdir -p $(DB_DIR)
	@./$(BIN_DIR)/$(APP_NAME) --config $(CONFIG)

clean:
	@echo "Cleaning..."
	@rm -rf $(BIN_DIR) $(DB_DIR)

test:
	@echo "Running tests..."
	@go test ./... -v

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

migrate:
	@echo "Running DB migrations..."
	@go run ./cmd/migrate

test:
	@go test -v ./...
