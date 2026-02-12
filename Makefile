.PHONY: all test build clean lint check coverage

# Переменные
BINARY_NAME=service-boilerplate
MAIN_PACKAGE=./cmd/service-boilerplate
BUILD_DIR=./build
TEST_TIMEOUT=120s

# Отключаем CGO (требуется для кроссплатформенной сборки)
export CGO_ENABLED=0

# Цель по умолчанию - запуск тестов и сборка
all: test build

# Запуск всех тестов
test:
	@echo "==> Running tests..."
	go test -v -race -timeout $(TEST_TIMEOUT) ./...

# Быстрый запуск тестов (без race detector)
test-fast:
	@echo "==> Running tests (fast mode)..."
	go test -v -timeout 60s ./...

# Сборка бинарника (только если тесты прошли)
build: test
	@echo "==> Building binary..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "==> Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Сборка без запуска тестов
build-only:
	@echo "==> Building binary..."
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME) $(MAIN_PACKAGE)
	@echo "==> Build complete: $(BUILD_DIR)/$(BINARY_NAME)"

# Сборка для разных платформ
build-all: test
	@echo "==> Building for multiple platforms..."
	@mkdir -p $(BUILD_DIR)
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PACKAGE)
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-linux-arm64 $(MAIN_PACKAGE)
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w" -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PACKAGE)
	@echo "==> Multi-platform build complete"

# Проверка кода (lint + format + vet)
check:
	@echo "==> Running checks..."
	@echo "==> Formatting code..."
	go fmt ./...
	@echo "==> Running go vet..."
	go vet ./...
	@echo "==> Running tests..."
	go test -timeout $(TEST_TIMEOUT) ./...

# Покрытие кода тестами
coverage:
	@echo "==> Running tests with coverage..."
	go test -coverprofile=coverage.out -timeout $(TEST_TIMEOUT) ./...
	go tool cover -func=coverage.out
	@echo "==> Coverage report generated: coverage.out"

# Просмотр покрытия в HTML
coverage-html: coverage
	go tool cover -html=coverage.out -o coverage.html
	@echo "==> HTML coverage report: coverage.html"

# Очистка
 clean:
	@echo "==> Cleaning..."
	rm -rf $(BUILD_DIR)
	rm -f coverage.out coverage.html
	go clean -cache

# Установка зависимостей
deps:
	@echo "==> Downloading dependencies..."
	go mod download
	go mod tidy

# CI/CD pipeline (полный набор проверок)
ci: deps check test build
	@echo "==> CI pipeline completed successfully!"

# Запуск приложения в dev режиме
run:
	go run $(MAIN_PACKAGE) run

# Помощь
help:
	@echo "Available targets:"
	@echo "  make all          - Run tests and build"
	@echo "  make test         - Run all tests with race detector"
	@echo "  make test-fast    - Run tests without race detector"
	@echo "  make build        - Build binary (requires passing tests)"
	@echo "  make build-only   - Build without running tests"
	@echo "  make build-all    - Build for multiple platforms"
	@echo "  make check        - Run formatting, vet, and tests"
	@echo "  make coverage     - Generate coverage report"
	@echo "  make coverage-html- Generate HTML coverage report"
	@echo "  make clean        - Clean build artifacts"
	@echo "  make deps         - Download dependencies"
	@echo "  make ci           - Full CI pipeline"
	@echo "  make run          - Run in dev mode"
