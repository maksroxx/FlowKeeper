APP_NAME=flowkeeper
BIN_DIR=bin
CONFIG=config/config.yaml
DB_DIR=data
export CGO_ENABLED=1
BUILD_TAGS=sqlite_icu

.PHONY: all build run clean clean-data test lint seed migrate

all: build

build:
	@echo "Building $(APP_NAME) with static ICU support..."
	@go build -tags "$(BUILD_TAGS)" -o $(BIN_DIR)/$(APP_NAME) ./cmd

run: build
	@echo "Running $(APP_NAME)..."
	@mkdir -p $(DB_DIR)
	@./$(BIN_DIR)/$(APP_NAME) --config $(CONFIG)

clean: clean-data
	@echo "Cleaning binaries..."
	@rm -rf $(BIN_DIR)

clean-data:
	@echo "Cleaning data and test databases..."
	@rm -rf $(DB_DIR)
	@rm -f test_*.db

test:
	@echo "Running tests with static ICU support..."
	@go test -v -cover -race -tags "$(BUILD_TAGS)" ./...

lint:
	@echo "Running golangci-lint..."
	@golangci-lint run ./...

seed:
	@echo "Seeding the database with ICU support..."
	@mkdir -p $(DB_DIR)
	@go run -tags "$(BUILD_TAGS)" ./seeder/seeder.go