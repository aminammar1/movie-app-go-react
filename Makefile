# --- Configuration Variables ---
APP_NAME := movie-app-go
BACKEND_PATH := Backend/movie-app-go
CLIENT_PATH := Client/movie-app-react
MAIN_FILE := main.go
DOCKER_IMAGE_NAME := ${APP_NAME}
GOBIN := $(shell go env GOPATH)/bin

# --- Phony Targets ---
.PHONY: swagger init run build build-docker build-compose down-compose clean-compose client-install client-build help

swagger:
	go install github.com/swaggo/swag/cmd/swag@latest
	cd $(BACKEND_PATH) && $(GOBIN)/swag init -g $(MAIN_FILE) -o docs

# --- GoLang Commands ---

# Go Module Initialization
init:
	@echo "Initializing Go module: ${APP_NAME}"
	cd $(BACKEND_PATH) && go mod init ${APP_NAME}
	@echo "Go module initialized."

# Run Application
run:
	$(MAKE) swagger
	cd $(BACKEND_PATH) && go run ${MAIN_FILE}

# Build Executable
build:
	$(MAKE) swagger
	cd $(BACKEND_PATH) && go build -o ${APP_NAME} ${MAIN_FILE}
	@echo "Build successful: $(BACKEND_PATH)/${APP_NAME}"

# --- Docker Commands ---

build-docker: build
	@echo "Building Docker image: ${DOCKER_IMAGE_NAME}"
	docker build -t ${DOCKER_IMAGE_NAME} $(BACKEND_PATH)
	@echo "Docker image built: ${DOCKER_IMAGE_NAME}"

# --- Docker Compose Commands ---

# Build and Start with Compose
build-compose:
	@echo "Building and starting services with Docker Compose..."
	docker-compose --env-file $(BACKEND_PATH)/.env up --build

# Stop Compose Services
down-compose:
	@echo "Stopping services defined in Docker Compose..."
	docker-compose --env-file $(BACKEND_PATH)/.env down

# Clean Compose Services and Volumes
clean-compose:
	@echo "Stopping and removing services, volumes, images, and orphans..."
	docker-compose --env-file $(BACKEND_PATH)/.env down --volumes --rmi all --remove-orphans

# --- Client Commands ---

client-install:
	cd $(CLIENT_PATH) && yarn install

client-build:
	cd $(CLIENT_PATH) && yarn build

# Help
help:
	@echo "Available targets:"
	@echo "  swagger        - Generate Swagger documentation"
	@echo "  init           - Initialize Go module"
	@echo "  run            - Run the application"
	@echo "  build          - Build the application executable"
	@echo "  build-docker   - Build Docker image"
	@echo "  build-compose  - Build and start services with Docker Compose"
	@echo "  down-compose   - Stop Docker Compose services"
	@echo "  clean-compose  - Stop services and remove volumes"
	@echo "  client-install - Install client dependencies with yarn"
	@echo "  client-build   - Build client bundle with yarn"
	@echo "  help           - Show this help message"
