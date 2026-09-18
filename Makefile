.PHONY: swag run up down logs

swag:
	swag init -g cmd/app/main.go -o docs

run: swag
	go run ./cmd/app

up:
	docker compose up -d --build

down:
	docker compose down

logs:
	docker compose logs -f backend