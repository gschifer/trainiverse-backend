.PHONY: all build run test clean tidy lint docker-build docker-run

# Installs useful development tools
TOOLS = \
	golang.org/x/lint/golint@latest \
	honnef.co/go/tools/cmd/staticcheck@latest

install-tools:
	@echo "Installing golint..."
	go install golang.org/x/lint/golint@latest
	@echo "Installing staticcheck..."
	go install honnef.co/go/tools/cmd/staticcheck@latest

tidy-vendor:
	@echo "Tidying modules and vending dependencies..."
	go mod tidy
	go mod vendor

build: tidy-vendor
	@echo "Compiling the project..."
	go build -o bin/app ./cmd

run: build
	@echo "Running the application..."
	./bin/app

test: tidy-vendor
	@echo "Running tests..."
	go test ./... -v

# Cleans build and test artifacts
clean:
	@echo "Cleaning artifacts..."
	go clean -cache -modcache
	rm -rf bin/

# Runs static code analysis tools (linters)
# Need to specify packages for linting too
lint: install-tools
	@echo "Running linters (golint and staticcheck)..."
	golangci-lint run -v
	staticcheck ./...

docker-build:
	@echo "Building Docker image..."
	set -a; source internal/config/.env; set +a; \
	docker build -t "$$DB_NAME" .

docker-run:
	@echo "Running container Docker..."
	set -a; source internal/config/.env; set +a; \
	docker run -p 8080:8080 "$$DB_NAME"

start-postgres:
	@echo "Starting postgresql"
	brew services start postgresql

stop-postgres:
	@echo "Turning off postgresql"
	brew services stop postgresql

run-postgres-container:
	@echo "Starting postgresql in Docker"
	docker-compose up postgres_db

postgres-term:
	@echo "Connecting to PostgreSQL database in docker"
	set -a; source internal/config/.env; set +a; \
	psql -h localhost -p "$$DB_PORT" -U "$$DB_USER" -d "$$DB_NAME"

run-app-docker:
	@echo "Running app"
	docker-compose up app --build

# Default command that executes build and test (great for CI/CD)
all: build test

