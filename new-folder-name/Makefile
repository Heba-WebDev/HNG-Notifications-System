.PHONY: help build run test docker-build docker-run clean

build: ## Build the application
	go build -o bin/api-gateway ./cmd/server

run: ## Run the application
	go run ./cmd/server/main.go

test: ## Run tests
	go test -v -race -coverprofile=coverage.out ./...

test-coverage: test ## Run tests with coverage report
	go tool cover -html=coverage.out

docker-build: ## Build Docker image
	docker build -t api-gateway:latest .

docker-run: ## Run Docker container
	docker-compose up -d

docker-stop: ## Stop Docker containers
	docker-compose down

docker-logs: ## View Docker logs
	docker-compose logs -f api-gateway

clean: ## Clean build artifacts
	rm -rf bin/
	rm -f coverage.out
	go clean

deps: ## Download dependencies
	go mod download
	go mod tidy