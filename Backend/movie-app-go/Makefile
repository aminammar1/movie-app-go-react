# --- Configuration Variables ---
APP_NAME := movie-app-go
MAIN_FILE := main.go
DOCKER_IMAGE_NAME := ${APP_NAME}

# --- Phony Targets ---
.PHONY: init run build build-docker build-compose down-compose clean-compose help

# --- GoLang Commands ---

# Go Module Initialization
init:
	@echo "Initializing Go module: ${APP_NAME}"
	go mod init ${APP_NAME}
	@echo "Go module initialized."

# Run Application
run:
	go run ${MAIN_FILE}

# Build Executable
build:
	go build -o ${APP_NAME} ${MAIN_FILE}
	@echo "Build successful: ./${APP_NAME}"

# --- Docker Commands ---

build-docker: build
	@echo "Building Docker image: ${DOCKER_IMAGE_NAME}"
	docker build -t ${DOCKER_IMAGE_NAME} .
	@echo "Docker image built: ${DOCKER_IMAGE_NAME}"

# --- Docker Compose Commands ---

# Build and Start with Compose
build-compose:
	@echo "Building and starting services with Docker Compose..."
	docker-compose up --build

# Stop Compose Services
down-compose:
	@echo "Stopping services defined in Docker Compose..."
	docker-compose down

# Clean Up Compose Resources
clean-compose:
	@echo "Cleaning up all Docker Compose resources (images, volumes, orphans)..."
	docker-compose down --rmi all -v --remove-orphans

# --- Help Command ---
# Display Help
help:
	@echo "------------------------------------"
	@echo "      Makefile Commands for ${APP_NAME}     "
	@echo "------------------------------------"
	@echo "Go Application:"
	@echo "  init            : Initialize a new Go module (${APP_NAME})"
	@echo "  run             : Run the Go application (go run ${MAIN_FILE})"
	@echo "  build           : Build the Go executable (${APP_NAME})"
	@echo ""
	@echo "Docker:"
	@echo "  build-docker    : Build the Docker image for the Go application (${DOCKER_IMAGE_NAME})"
	@echo ""
	@echo "Docker Compose:"
	@echo "  build-compose   : Build and start services using Docker Compose"
	@echo "  down-compose    : Stop services defined in Docker Compose"
	@echo "  clean-compose   : Clean up Docker Compose resources (images, volumes, orphans)"
	@echo ""
	@echo "Other:"
	@echo "  help            : Display this help message"