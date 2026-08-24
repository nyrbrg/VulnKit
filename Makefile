.PHONY: dev test test-unit test-integration db-up db-down backend frontend up down

# Start PostgreSQL for local dev
db-up:
	docker compose up postgres -d

# Stop PostgreSQL
db-down:
	docker compose stop postgres

test: db-up
	cd backend && go test -p 1 ./...

# Unit tests only (no DB needed)
test-unit:
	cd backend && go test ./internal/compose/...

# Integration tests only (requires running postgres)
test-integration: db-up
	cd backend && TEST_DATABASE_URL=postgres://vulnkit:vulnkit@localhost:5432/vulnkit_test?sslmode=disable \
		go test -p 1 ./internal/presets/... ./internal/api/...

# Run backend
backend:
	cd backend && go run ./cmd/server

# Run frontend
frontend:
	cd frontend && npm run dev

# Full stack via Docker
up:
	docker compose up --build

down:
	docker compose down
