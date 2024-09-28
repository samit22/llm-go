GEMINI_FLASH_API_KEY?= $(shell cat ./.env.gemini-flash-api-key)

.PHONY: start-docker
start-docker:
	docker build -t rag-server:latest .
	RAG_SERVER_IMAGE=rag-server:latest GEMINI_FLASH_API_KEY=${GEMINI_FLASH_API_KEY} docker compose up --remove-orphans

.PHONY: start-local
start-local: run

.PHONY: start-vector-db
start-vector-db:
	@echo "Starting the Vector database"
	docker compose -f docker-compose-vector-db.yaml up

.PHONY: run
run: build
	@echo "Running Rag Server"
	GEMINI_FLASH_API_KEY=$(GEMINI_FLASH_API_KEY) ./server

.PHONY: build
build:
	@echo "Building the Go program"
	go mod download
	go build -o server .
