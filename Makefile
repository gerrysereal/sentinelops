.PHONY: up down logs api web test fmt

up:
	docker compose up --build

down:
	docker compose down -v

logs:
	docker compose logs -f

api:
	cd apps/api && go run ./cmd/api

web:
	cd apps/web && npm run dev

test:
	cd apps/api && go test ./...
	cd apps/web && npm run lint

fmt:
	cd apps/api && gofmt -w .
