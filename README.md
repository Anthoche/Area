<div align="center">
    <h1>KiKoNect</h1>
    <h3>Action - REAction</h3>
</div>

---

## 📋 Project Description

KiKoNect is an “Area”-style automation platform.  
The application offers the following functionalities (high level user flow):

1. The user registers on the application KiKoNect in order to obtain an account.
2. The registered user then confirms their enrollment on the application before being able to use it.
3. The application then asks the authenticated user to subscribe to Services.
4. Each service offers the following components:
  - type Action
  - type REAction
5. The authenticated user composes a *Konect* by interconnecting an Action to a REAction previously configured.
6. The application triggers *Konects* automatically thanks to hooks.


## ✨ Key Features

- 🔐 Auth: register/login with bcrypt.
- ⚙️ Workflows: manual, webhook, interval, and polling-based triggers.
- 🚀 Execution: executor drains pending jobs and POSTs payloads to targets.
- 🌐 HTTP API with permissive CORS for the web app.
- 🔌 Integrations: Google, GitHub, Discord, Slack, Notion, Weather, Reddit, YouTube, Air Quality, Crypto, NASA, Steam, Trello.
- 📖 Auto‑generated API docs at `/docs/` and service catalog at `/about.json`.
- 📦 Docker Compose stack (Postgres + API + web + mobile build).

## 🛠️ Stack

- Go 1.21+, PostgreSQL, Docker Compose
- React 18 + Vite (web), Flutter (mobile)
- Go testing (`testing`, `sqlmock`)

## 📂 Project Structure

```
Area/
├── CONTRIBUTING.md
├── docker-compose.yml
├── LICENSE
├── README.md
├── backend/                    # Go backend
│   ├── Dockerfile
│   ├── go.mod
│   ├── go.sum
│   ├── resources/
│   │   └── database_scheme.sql
│   └── src/
│       ├── main.go
│       ├── auth/               # auth service, oauth helpers, tests
│       ├── database/           # postgres store + migrations/tests
│       ├── httpapi/            # HTTP handlers, routes, server setup
│       ├── workflows/          # store, triggers, executor, scheduler
│       └── integrations/       # external integrations / adapters
├── frontend/
│   ├── web/                    # React + Vite app
│   │   ├── package.json
│   │   ├── vite.config.js
│   │   └── src/
│   │       ├── App.jsx
│   │       └── components/
│   │
│   └── mobile/
│       └── kikonect/           # Flutter project (android/ios/lib/...)
│           ├── pubspec.yaml
│           ├── lib/            # Contain all the app code
│           │    ├── main.dart  # Entry point of the app
│           │    ├── app.dart   # App-level configurations (themes, routes, ...)
│           │    ├── assets/    # Images, fonts, and other static assets
│           │    ├── screens/   # Different screens of the app
│           │    └─── widgets/  # Reusable UI components
│           ├── test/           # Unit tests for the Flutter
│           └── android/
├── Reports/
│   ├── Defense/
│   └── Meeting/
```

## 🏗️ Architecture

The project is composed of **four main components** deployed with `docker-compose`:

### **API Backend (Go)**
- HTTP server exposed on port `8080` (configurable using `PORT`).
- Main routes: auth (`/login`, `/register`), workflows (`/workflows`, `/hooks/{token}`), OAuth endpoints (`/oauth/*`), docs (`/docs/`), and `about.json`.
- Manages workflow logic: creation, interval scheduling, job queueing, and execution.
- Uses PostgreSQL for persistence (schema: `backend/resources/database_scheme.sql`).

### **PostgreSQL**
- Relational database storing users, workflows, runs, and jobs.
- Configured via environment variables (`POSTGRES_*`).

### **Web Frontend (React + Vite)**
- Development server on `5173` (Vite), production build served via nginx on `8081`.
- Communicates with API through `VITE_API_URL`.
- Handles OAuth initialization on the client (using a `json-init` flow that stores `state` in `localStorage` to avoid cross-port cookie issues).

### **Mobile (Flutter)**
- Built from `frontend/mobile/kikonect` by the `client_mobile` service in Docker Compose, generating an APK.

### Important Execution Notes
- **OAuth**: backend validates callback URL and exchanges code→token; frontend stores OAuth `state` in `localStorage` and posts the `code` back to backend.
- **CORS**: backend uses permissive headers for development; in production origin should be restricted.
- **Environment variables**: although `.env` is loaded automatically, core variables should be explicitly listed in the compose file.
- **Scalability**: executor sends POST payloads to `action_url`; scaling horizontally requires splitting workers and/or using distributed locking or message queues.

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
- `BCRYPT_COST`
- OAuth:
  - `GOOGLE_OAUTH_CLIENT_ID`, `GOOGLE_OAUTH_CLIENT_SECRET`, `GOOGLE_OAUTH_REDIRECT_URI`
  - `GITHUB_OAUTH_CLIENT_ID`, `GITHUB_OAUTH_CLIENT_SECRET`, `GITHUB_OAUTH_REDIRECT_URI`
  - `GITHUB_MOBILE_OAUTH_CLIENT_ID`, `GITHUB_MOBILE_OAUTH_CLIENT_SECRET`, `GITHUB_MOBILE_OAUTH_REDIRECT_URI`
- Bot/API tokens:
  - `DISCORD_BOT_TOKEN`
  - `SLACK_BOT_TOKEN`
  - `NOTION_TOKEN`
  - `STEAM_API_KEY`
  - `NASA_API_KEY`
  - `TRELLO_API_KEY`
  - `TRELLO_TOKEN`

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

**Docs**
- `GET /docs/` — Swagger UI.
- `GET /resources/openapi.json` — OpenAPI JSON.

**About**
- `GET /about.json` — service catalog for the current server, includes client IP and server time.

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
  Trigger types: `interval`, `manual`, `webhook`, `gmail_inbound`, `github_commit`, `github_pull_request`, `github_issue`, `weather_temp`, `weather_report`, `reddit_new_post`, `youtube_new_video`.
- `POST /workflows/{id}/trigger` — enqueue a run with arbitrary JSON payload (202, 404 if missing).
- `POST /hooks/{token}` — trigger a webhook workflow (matches `trigger_config.token`).

**Execution**
- A trigger creates a run + job; the executor drains pending jobs and POSTs the payload to `action_url`.
- Interval workflows are rescheduled via `ClaimDueIntervalWorkflows`.

**OAuth**
- `GET /oauth/google/login`, `GET /oauth/google/callback`
- `GET /oauth/github/login`, `GET /oauth/github/callback`

**Actions (HTTP)**
- Google: `POST /actions/google/email`, `POST /actions/google/calendar`
- GitHub: `POST /actions/github/issue`, `POST /actions/github/pr`
- Discord: `POST /actions/discord/message`, `/embed`, `/message/edit`, `/message/delete`, `/message/react`
- Slack: `POST /actions/slack/message`, `/blocks`, `/message/update`, `/message/delete`, `/message/react`
- Notion: `POST /actions/notion/page`, `/blocks`, `/database`, `/page/update`

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

`frontend/mobile/kikonect`: the `client_mobile` service in Docker Compose builds a release APK and copies it into the web container (`/usr/share/nginx/html/apk/client.apk`).

## 🔧 Useful Commands

- `go fmt ./...` — format Go code.
- `go clean -testcache` — clear Go test cache.

<br>

## 🤝 Contribution

Contributions are welcome! Please follow these guidelines:

- Read the `CONTRIBUTING.md` file for branch, test and PR rules.

- Create a feature branch from `dev`: `git checkout -b feat/your-feature`.

- Run tests and linters before submitting a PR:

- Write clear PR descriptions and link any related issues.

If you're adding a breaking change, please open an issue first to discuss the design.

## 📜 License

This project is provided under the MIT License — see the `LICENSE` file for details.


## 👥 Team

- [Bastien Leroux](https://github.com/bast0u)
- [Anthony El-Achkar](https://github.com/Anthoche)
- [Mariia Semenchenko](https://github.com/mariiasemenchenko)
- [Corto Morrow](https://github.com/NuggetReckt)

*Last update: January 2026*
