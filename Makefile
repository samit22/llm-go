GEMINI_FLASH_API_KEY?=$(shell cat ./.env.gemini-flash-api-key)
RAG_CLIENT?=langchain

start-docker:
	docker build -t rag-server:latest .
	RAG_SERVER_IMAGE=rag-server:latest GEMINI_FLASH_API_KEY=${GEMINI_FLASH_API_KEY} docker compose up --remove-orphans

start-local: run

start-vector-db:
	@echo "Starting the Vector database"
	docker compose -f docker-compose-vector-db.yaml up

run: build
	@echo "Running Rag Server"
	GEMINI_FLASH_API_KEY=$(GEMINI_FLASH_API_KEY) RAG_CLIENT=$(RAG_CLIENT) ./server

.PHONY: build
build:
	@echo "Building the Go program"
	go mod download
	go build -o server .

short-test:
	go test ./... -short
test:
	docker compose -f docker-compose-vector-db.yaml up -d
	go test -race ./... -coverprofile=coverage.txt -covermode=atomic
