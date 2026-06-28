# Shush 🤫

A self-hosted secret manager, built with Go (ConnectRPC), PostgreSQL, and Next.js.

The API is defined once in protobuf and served via [ConnectRPC](https://connectrpc.com):
a single handler speaks gRPC, gRPC-Web, and plain HTTP/JSON. The CLI talks gRPC, the
dashboard and CI/SDKs use the HTTP/JSON API — no separate gateway proxy.

## Prerequisites

- [Go 1.25+](https://go.dev/dl/)
- [Docker & Docker Compose](https://docs.docker.com/get-docker/)
- [Buf CLI](https://buf.build/docs/installation) — for protobuf code generation
- [golang-migrate](https://github.com/golang-migrate/migrate) — for database migrations
- [Air](https://github.com/air-verse/air) — for backend hot-reload (`go install github.com/air-verse/air@latest`)

## Quick Start

```bash
# Start Postgres, run migrations, start backend + frontend
make dev
```

Or step by step:

```bash
# 1. Start Postgres
docker compose up -d postgres

# 2. Run database migrations
export DATABASE_URL="postgres://shush:shush@localhost:5432/shush?sslmode=disable"
make migrate-up

# 3. Start backend (with hot-reload)
make dev-backend

# 4. In another terminal, start frontend
make dev-frontend
```

| Service | URL |
|---|---|
| Frontend | http://localhost:3000 |
| API (Connect: gRPC + HTTP/JSON) | http://localhost:8080 |
| PostgreSQL | localhost:5432 |

## Makefile Commands

| Command | Description |
|---|---|
| `make proto` | Generate Go code from `.proto` files |
| `make migrate-up` | Apply all pending database migrations |
| `make migrate-down` | Rollback the last migration |
| `make dev-backend` | Start Go server with hot-reload |
| `make dev-frontend` | Start Next.js dev server |
| `make dev` | Start everything |
| `make test` | Run all Go tests |
| `make build` | Build backend binary to `bin/server` |
| `make build-cli` | Compile the Shush CLI locally to `cli/shush` |

## Project Structure

```
├── proto/               # Protobuf API definitions (source of truth)
├── backend/
│   ├── cmd/server/      # Entry point
│   ├── internal/
│   │   ├── crypto/      # Envelope encryption (AES-256-GCM)
│   │   ├── auth/        # JWT + RBAC
│   │   ├── server/      # Connect service handlers
│   │   └── store/       # PostgreSQL repository layer
│   └── migrations/      # SQL migration files
├── frontend/            # Next.js 14 dashboard
├── cli/                 # Go CLI (planned)
├── buf.yaml             # Buf proto linting config
├── buf.gen.yaml         # Buf code generation config
├── docker-compose.yml   # Local dev services
└── Makefile             # Task runner
```

## Running Tests

```bash
make test
```
