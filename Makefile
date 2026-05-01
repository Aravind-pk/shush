.PHONY: proto migrate-up migrate-down dev-backend dev-frontend dev test build

# Generate Go code from .proto files using Buf
proto:
	buf generate

# Run all pending database migrations
migrate-up:
	migrate -path backend/migrations -database "$(DATABASE_URL)" up

# Rollback the last migration
migrate-down:
	migrate -path backend/migrations -database "$(DATABASE_URL)" down 1

# Start Go backend with hot-reload (requires: go install github.com/air-verse/air@latest)
dev-backend:
	cd backend && air

# Start Next.js frontend dev server
dev-frontend:
	cd frontend && npm run dev

# Start everything: Postgres + backend + frontend
dev:
	docker compose up -d postgres
	$(MAKE) migrate-up
	$(MAKE) dev-backend &
	$(MAKE) dev-frontend

# Run all Go tests
test:
	cd backend && go test ./...

# Build backend binary
build:
	cd backend && go build -o ../bin/server ./cmd/server
