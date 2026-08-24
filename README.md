# VulnKit

VulnKit is a vulnerability lab management platform. It lets you spin up deliberately vulnerable Docker containers (old MySQL/Apache versions, XSS playgrounds, etc.) for hands-on security practice.

- **Backend** (`backend/`) — Go service that orchestrates Docker (by shelling out to the `docker` CLI) and stores lab presets in Postgres.
- **Frontend** (`frontend/`) — SvelteKit dashboard for browsing presets, building custom labs, and managing running containers.
- **Labs** (`labs/`) — a PHP/Apache container with intentionally vulnerable pages, served on port 8888.

## ⚠️ Status: work in progress

This project is under development and **not production-ready**. Some parts are incomplete or known to be rough around the edges. If you hit something broken, please open an issue in this repository describing what you did and what went wrong (include the relevant terminal/browser console output if you can) so it can be tracked and fixed.

## Running the app

### Prerequisites

- [Docker](https://www.docker.com/) (with Docker Compose) — for Postgres and the labs container
- [Go](https://go.dev/) 1.22+
- [Node.js](https://nodejs.org/) 18+ and npm

### Recommended: run backend and frontend on the host

The fully containerized stack (`make up`) can't drive Docker from inside its own container (see [Known issues](#known-issues)), so for actual lab usage, run the backend and frontend directly. The easiest way is the bundled script:

```bash
./start.sh
```

This starts Postgres and the labs container, then runs the backend (:8080) and frontend (:5173) together, streaming both logs to your terminal. Press Ctrl+C to stop the backend and frontend; Postgres and the labs container keep running afterward (stop them with `make db-down`).

Then open **http://localhost:5173** in your browser.

#### Manual startup

If you'd rather run each piece in its own terminal (e.g. to restart just the backend while debugging):

```bash
# 1. Start Postgres (and create the database VulnKit expects)
make db-up

# 2. In one terminal: run the backend (listens on :8080)
make backend

# 3. In another terminal: run the frontend dev server (listens on :5173)
make frontend
```

To stop Postgres afterwards: `make db-down`.

### Alternative: full stack via Docker Compose

```bash
make up    # docker compose up --build
make down  # docker compose down
```

This starts Postgres, the backend, the frontend (on :3000), and the labs container together — useful for a quick look at the UI, but (again) lab start/stop/build operations will fail from this setup.

### Configuration

| Variable | Where | Default | Purpose |
|---|---|---|---|
| `DATABASE_URL` | backend | `postgres://vulnkit:vulnkit@localhost:5432/vulnkit?sslmode=disable` | Main Postgres connection |
| `PROJECT_DIR` | backend | current working directory | Directory the backend uses for `docker compose` operations against the built-in labs stack |
| `NVD_API_KEY` | backend | unset | Optional — speeds up NVD CVE API lookups on the CVE page |
| `VITE_API_URL` | frontend (`.env`) | `http://localhost:8080` | Backend URL the frontend calls |
| `VITE_ENV_NAME` | frontend (`.env`) | — | Optional label shown in the UI |

### Running the tests

```bash
make test-unit         # fast, no DB required
make test               # full suite (starts Postgres, but see note below)
make test-integration   # DB-backed tests only (same note)
```

`make test`/`make test-integration` start Postgres via `make db-up`, but **on a fresh checkout the `vulnkit_test` database doesn't exist yet** — `db-up` only provisions the main `vulnkit` database. Without it, the DB-backed tests silently `SKIP` instead of failing, which can look like a passing suite when it isn't. Create it once, manually:

```bash
make db-up
docker exec vulnkit-postgres-1 psql -U vulnkit -c "CREATE DATABASE vulnkit_test;"
```

Backend-only, from `backend/`:

```bash
go vet ./...
go test -p 1 ./...   # -p 1 avoids a shared-database race between packages' tests
```
