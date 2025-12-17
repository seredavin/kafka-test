.PHONY: help build test coverage lint clean install run docker-build docker-run

# Переменные
BINARY_NAME=kafka-producer-ui
VERSION?=dev
BUILD_DIR=build
DIST_DIR=dist
GO_FILES=$(shell find . -name '*.go' -type f -not -path "./vendor/*")
COVERAGE_FILE=coverage.out

# Цвета для вывода
GREEN=\033[0;32m
NC=\033[0m # No Color

## help: Показать справку
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## build: Собрать бинарный файл
build:
	@echo "$(GREEN)Building $(BINARY_NAME)...$(NC)"
	@mkdir -p $(BUILD_DIR)
	go build -ldflags="-s -w -X main.version=$(VERSION)" -o $(BUILD_DIR)/$(BINARY_NAME) .
	@echo "$(GREEN)Build complete: $(BUILD_DIR)/$(BINARY_NAME)$(NC)"

## build-all: Собрать для всех платформ
build-all:
	@echo "$(GREEN)Building for all platforms...$(NC)"
	@mkdir -p $(DIST_DIR)
	
	# Linux AMD64
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" \
		-o $(DIST_DIR)/$(BINARY_NAME)-linux-amd64 .
	
	# Linux ARM64
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" \
		-o $(DIST_DIR)/$(BINARY_NAME)-linux-arm64 .
	
	# macOS AMD64
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" \
		-o $(DIST_DIR)/$(BINARY_NAME)-darwin-amd64 .
	
	# macOS ARM64
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" \
		-o $(DIST_DIR)/$(BINARY_NAME)-darwin-arm64 .
	
	# Windows AMD64
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" \
		-o $(DIST_DIR)/$(BINARY_NAME)-windows-amd64.exe .
	
	# Windows ARM64
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" \
		-o $(DIST_DIR)/$(BINARY_NAME)-windows-arm64.exe .
	
	@echo "$(GREEN)All builds complete!$(NC)"

## test: Запустить тесты
test:
	@echo "$(GREEN)Running tests...$(NC)"
	go test -v -race ./...

## coverage: Запустить тесты с покрытием
coverage:
	@echo "$(GREEN)Running tests with coverage...$(NC)"
	go test -v -race -coverprofile=$(COVERAGE_FILE) -covermode=atomic ./...
	go tool cover -func=$(COVERAGE_FILE)
	@echo "$(GREEN)Coverage report saved to $(COVERAGE_FILE)$(NC)"

## coverage-html: Показать покрытие в браузере
coverage-html: coverage
	go tool cover -html=$(COVERAGE_FILE)

## lint: Запустить линтер
lint:
	@echo "$(GREEN)Running linter...$(NC)"
	golangci-lint run

## fmt: Форматировать код
fmt:
	@echo "$(GREEN)Formatting code...$(NC)"
	gofmt -s -w $(GO_FILES)
	goimports -w $(GO_FILES)

## vet: Запустить go vet
vet:
	@echo "$(GREEN)Running go vet...$(NC)"
	go vet ./...

## mod-tidy: Очистить зависимости
mod-tidy:
	@echo "$(GREEN)Tidying modules...$(NC)"
	go mod tidy
	go mod verify

## clean: Очистить артефакты сборки
clean:
	@echo "$(GREEN)Cleaning...$(NC)"
	rm -rf $(BUILD_DIR) $(DIST_DIR)
	rm -f $(COVERAGE_FILE) coverage.html
	go clean

## install: Установить в GOPATH/bin
install:
	@echo "$(GREEN)Installing $(BINARY_NAME)...$(NC)"
	go install -ldflags="-s -w -X main.version=$(VERSION)" .

## run: Запустить программу
run: build
	@echo "$(GREEN)Running $(BINARY_NAME)...$(NC)"
	./$(BUILD_DIR)/$(BINARY_NAME)

## dev: Режим разработки (build + run)
dev:
	@echo "$(GREEN)Development mode...$(NC)"
	go run .

## check: Выполнить все проверки (fmt, vet, lint, test)
check: fmt vet lint test
	@echo "$(GREEN)All checks passed!$(NC)"

## deps: Обновить зависимости
deps:
	@echo "$(GREEN)Updating dependencies...$(NC)"
	go get -u ./...
	go mod tidy

## version: Показать версию
version:
	@echo "Version: $(VERSION)"

## docker-build: Собрать Docker образ
docker-build:
	@echo "$(GREEN)Building Docker image...$(NC)"
	docker build -t $(BINARY_NAME):$(VERSION) .

## docker-run: Запустить в Docker
docker-run:
	@echo "$(GREEN)Running in Docker...$(NC)"
	docker run -it --rm $(BINARY_NAME):$(VERSION)

## release-check: Проверить готовность к релизу
release-check: check coverage
	@echo "$(GREEN)Checking release readiness...$(NC)"
	@if [ -n "$$(git status --porcelain)" ]; then \
		echo "Error: Working directory is not clean"; \
		exit 1; \
	fi
	@echo "$(GREEN)Ready for release!$(NC)"

## bench: Запустить бенчмарки
bench:
	@echo "$(GREEN)Running benchmarks...$(NC)"
	go test -bench=. -benchmem ./...

## tool-install: Установить необходимые инструменты
tool-install:
	@echo "$(GREEN)Installing tools...$(NC)"
	go install github.com/golangci/golangci-lint/cmd/golangci-lint@latest
	go install golang.org/x/tools/cmd/goimports@latest

.DEFAULT_GOAL := help

