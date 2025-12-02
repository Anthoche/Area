<div align="center">
    <h1>KiKonect (Area)</h1>
    <h3>Automation backend with web/mobile clients</h3>
</div>

---

## 📋 Project Description

KiKonect is an “Area”-style automation platform.  
Backend in Go + PostgreSQL, web in React/Vite, mobile build in Flutter.  
Workflows can be manual, interval, or webhook-triggered; they enqueue jobs executed via HTTP to an `action_url`.
If you want to extend the platform (new triggers, services, actions, etc.), read the [Contributing Guide](CONTRIBUTING.md).

## ✨ Key Features

- 🔐 Auth: register/login with bcrypt.
- ⚙️ Workflows: create/list, trigger manually or via webhook, interval scheduling.
- 🚀 Execution: executor drains pending jobs and POSTs payloads to targets.
- 🌐 HTTP API with permissive CORS for the web app.
- 📦 Docker Compose stack (Postgres + API + web + mobile build).

## 🛠️ Stack

- Go 1.21+, PostgreSQL, Docker Compose
- React 18 + Vite (web), Flutter (mobile)
- Go testing (`testing`, `sqlmock`)

## 📂 Project Structure

```
Area/
├── backend/                  # Go backend
│   ├── src/
│   │   ├── auth/              # Auth service + store
│   │   ├── database/          # PostgreSQL access layer
│   │   ├── httpapi/           # HTTP handlers (login, register, workflows)
│   │   ├── integrations/      # placeholder integrations
│   │   ├── workflows/         # Store, triggerer, executor, scheduler
│   │   └── main.go            # server entrypoint
│   ├── resources/            # database_scheme.sql
│   ├── Dockerfile
│   ├── go.mod / go.sum
├── frontend/
│   ├── web/                  # React/Vite app
│   └── mobile/               # Flutter project (APK via Docker)
├── docker-compose.yml        # Full stack
└── README.md
```

## 🚀 Quick Start (Docker Compose)

```bash
cp backend/.env .env   # reuse defaults for compose
docker-compose up --build
```

- Backend API: http://localhost:8080  
- Web (nginx): http://localhost:8081  
- PostgreSQL: localhost:5432 (credentials from `.env`)

## 🧰 Backend (Go API)

### Configuration
Env vars (see `backend/.env`):
- `PORT` (default 8080)
- `POSTGRES_HOST`, `POSTGRES_PORT`, `POSTGRES_USER`, `POSTGRES_PASSWORD`, `POSTGRES_DB`, `POSTGRES_SSLMODE`
- `BCRYPT_COST` (optional)

### Run locally (without Docker)
```bash
cd backend
cp .env .env.local
export $(cat .env.local | xargs)   # or use direnv
go run ./src
```
Requires PostgreSQL seeded with `backend/resources/database_scheme.sql`.

### Tests & build
```bash
cd backend
go test ./...
go build ./...
```
CI (GitHub Actions) runs `go build` + `go test` on push/PR.

### API (base: http://localhost:8080)

**Auth**
- `POST /login` — body `{"email","password"}`; 200 user JSON, 401 invalid creds.
- `POST /register` — body `{"email","password","firstname","lastname"}`; 201 on success, 409 if existing.

**Health**
- `GET /healthz` — `{"status":"ok"}`.

**Workflows**
- `GET /workflows` — list.
- `POST /workflows` — create a workflow. Interval example:
  ```json
  {
    "name": "My WF",
    "trigger_type": "interval",
    "action_url": "https://example.com/webhook",
    "trigger_config": { "interval_minutes": 5, "payload": { "foo": "bar" } }
  }
  ```
  Trigger types: `interval`, `manual`, `webhook` (for manual/webhook, `trigger_config` defaults to `{}` if empty).
- `POST /workflows/{id}/trigger` — enqueue a run with arbitrary JSON payload (202, 404 if missing).
- `POST /hooks/{token}` — trigger a webhook workflow (matches `trigger_config.token`).

**Execution**
- A trigger creates a run + job; the executor drains pending jobs and POSTs the payload to `action_url`.
- Interval workflows are rescheduled via `ClaimDueIntervalWorkflows`.

## 🌐 Frontend (React/Vite)

```bash
cd frontend/web
npm install
npm run dev          # http://localhost:5173
```
- API base URL: `VITE_API_URL` (fallback `http(s)://<host>:8080`).
- Routes: `/` (login), `/register`.
- Production: `npm run build` (served by nginx in docker-compose).

## 📱 Mobile (Flutter)

`frontend/mobile`: the `client_mobile` service in Docker Compose builds a release APK and copies it into the web container (`/usr/share/nginx/html/apk/client.apk`).

## 🔧 Useful Commands

- `go fmt ./...` — format Go code.
- `go clean -testcache` — clear Go test cache.

## 👥 Team

- [Bastien Leroux](https://github.com/bast0u)
- [Anthony El-Achkar](https://github.com/Anthoche)
- [Mariia Semenchenko](https://github.com/mariiasemenchenko)
- [Corto Morrow](https://github.com/NuggetReckt)

<br>
*Last update: December 2025*
