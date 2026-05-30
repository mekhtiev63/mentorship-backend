.PHONY: run build test migrate-up compose-up compose-down compose-logs compose-api-up

build:
	go build -o bin/api ./cmd/api
	go build -o bin/migrate ./cmd/migrate

run:
	go run ./cmd/api

test:
	go test ./...

migrate-up:
	go run ./cmd/migrate -direction up

compose-up:
	docker-compose up -d --build

compose-api-up:
	docker-compose up -d --build postgres redis migrate backend

compose-down:
	docker-compose down

compose-logs:
	docker-compose logs -f
