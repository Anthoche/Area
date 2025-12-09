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

## 🔎 Functionalities

The application offers the following functionalities (high level user flow):

1. The user registers on the application KiKoNect in order to obtain an account.
2. The registered user then confirms their enrollment on the application before being able to use it.
3. The application then asks the authenticated user to subscribe to Services.
4. Each service offers the following components:
  - type Action
  - type REAction
5. The authenticated user composes a KoNect by interconnecting an Action to a REAction previously configured.
6. The application triggers KoNect automatically thanks to hooks.


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
│   └── mobile/                 # Flutter project (android/ios/lib/...)
│       ├── pubspec.yaml
│       └── android/
├── Reports/
│   ├── Defense/
│   └── Meeting/
```

## 🏗️ Architecture

Le projet est composé de 4 briques principales déployées par `docker-compose` :

- **API backend (Go)**
  - Serveur HTTP exposé sur le port `8080` (configurable via `PORT`).
  - Routes principales : auth (`/login`, `/register`), workflows (`/workflows`, `/hooks/{token}`), endpoints d'OAuth (`/auth/*` et `/oauth/*`).
  - Gère la logique des workflows : création, planification (interval), files d'exécution et exécution des jobs.
  - Accède à PostgreSQL pour persistance (schéma : `backend/resources/database_scheme.sql`).

- **PostgreSQL**
  - Base de données relationnelle contenant utilisateurs, workflows, runs et jobs.
  - Configurée via les variables d'environnement (`POSTGRES_*`).

- **Frontend web (React + Vite)**
  - App en développement sur `5173` (Vite) et build servie par le service nginx du compose sur `8081`.
  - Communique avec l'API backend via l'URL configurée (`VITE_API_URL`).
  - Gère l'init OAuth côté client (pattern `json-init` utilisé pour stocker le `state` en `localStorage` afin d'éviter les problèmes de cookie cross-port).

- **Mobile (Flutter)**
  - Projet Flutter construit par le service `client_mobile` qui produit un APK.

Flux importants et remarques opérationnelles
- OAuth: le backend construit/valide l'URL de callback enregistrée et fait l'échange de code -> token ; le frontend lance l'auth via l'endpoint JSON du backend, stocke `state` en `localStorage`, puis poste le `code` au backend pour échange.
- CORS: l'API renvoie des en-têtes permissifs pour permettre l'appel depuis le client web (en environnement dev). En production, verrouiller `Access-Control-Allow-Origin`.
- Variables d'environnement: `docker-compose` lit `.env` via `env_file`, mais il est préférable d'énumérer explicitement les variables critiques dans le bloc `environment:` du service `server` pour assurer leur disponibilité.
- Exécution & scalabilité: l'executor (composant du backend) draine les jobs et POSTe les payloads vers `action_url`. Pour monter en charge, séparer l'executor en workers horizontaux et utiliser une file de messages durable (ou verrou distribué) pour l'attribution des jobs.

Fichiers importants
- `backend/src/` : code Go de l'API.
- `backend/resources/database_scheme.sql` : schéma initial de la base.
- `docker-compose.yml` : orchestration locale (db, server, client_web, client_mobile).


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
