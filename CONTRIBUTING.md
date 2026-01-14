<div align="center">
  <h1>Contribute to KiKonect</h1>
  <h3>Practical guide to extend the backend and frontends</h3>
</div>

---

## Table of Contents
- [Prerequisites](#prerequisites)
- [Workflow Overview](#workflow-overview)
- [Adding Backend Features](#adding-backend-features)
  - [New Workflow Trigger](#new-workflow-trigger)
  - [New Outbound Integration (action/reaction)](#new-outbound-integration-actionreaction)
  - [New HTTP Endpoint](#new-http-endpoint)
  - [Database Changes](#database-changes)
  - [Testing](#testing)
- [Frontend Changes (React/Vite)](#frontend-changes-reactvite)
- [Mobile Changes (Flutter)](#mobile-changes-flutter)
  - [New Screens (Pages)](#new-screens-pages)
  - [New Widgets](#new-widgets)
  - [New Services and API Integrations](#new-services-and-api-integrations)
- [Coding Standards](#coding-standards)
- [CI & Local Checks](#ci--local-checks)
- [Pull Request Checklist](#pull-request-checklist)

---

## Prerequisites
- Go 1.21+, Node 18+, Docker (optional for DB and full stack).
- A running PostgreSQL (or use `docker-compose up db server`).
- Environment variables set (see `backend/.env`).

## Workflow Overview
- Auth: handled in `backend/src/auth` via `Service` + `DBStore`.
- Workflows: stored in `backend/src/workflows` with `Store`, `Triggerer`, `Executor`.
- HTTP API: routes in `backend/src/httpapi/server.go` calling services, plus `/about.json` and `/docs/`.
- DB schema: `backend/resources/database_scheme.sql`.
- Frontend web: `frontend/web` (React + Vite), mobile: `frontend/mobile` (Flutter).
- Integrations: `backend/src/integrations/{google,github,discord,slack,notion,weather,reddit,youtube}`.

---

## Adding Backend Features

### New Workflow Trigger
1. **Define trigger type handling** in `workflows/service.go` (validate config, default values).
2. **Persist trigger config** in `workflows/store.go` (e.g., how `interval` uses `next_run_at`).
3. **Dispatch trigger** in `workflows/triggerer.go`, scheduler loop, or a poller (for interval/polling triggers).
4. **Expose API** in `httpapi/server.go` if needed (new endpoint or payload shape).
5. **Test**:
   - Unit tests in `backend/tests/workflows` and `backend/tests/httpapi` as appropriate.
   - Add config parsing tests like `intervalConfigFromJSON`.

### New Outbound Integration (action/reaction)
1. **Implement `OutboundSender`** or a dedicated client under `backend/src/integrations/`.
2. **Wire into executor**: swap `newHTTPSender` usage in `main.go` or make it injectable.
3. **Add configuration** (env vars) and document in README.
4. **Test** with fakes/mocks; avoid network calls in tests.

### New HTTP Endpoint
1. Add handler in `httpapi/server.go` (prefer small helpers).
2. Validate inputs strictly (`DisallowUnknownFields`, `EnsureNoTrailingData`).
3. Call the right service method; avoid DB access directly from handlers.
4. Unit test with `httptest` (see `server_test.go`, `cors_test.go`).

### Database Changes
1. Edit `backend/resources/database_scheme.sql`.
2. If using migrations, add a new migration file (not present yet—could be added).
3. Update store methods in `workflows/store.go` or `database/user.go`.
4. Add tests with `sqlmock` to validate queries.

### Testing
Run from `backend/`:
```bash
go test ./...          # all packages
go test -count=1 ./... # bypass cache if needed
```
Use `sqlmock` for DB-facing code. Avoid network in tests; mock senders/HTTP.

---

## Frontend Changes (React/Vite)
Location: `frontend/web`
1. Install deps: `npm install`.
2. Add routes/components under `src/components` and `src/App.jsx`.
3. API base URL via `VITE_API_URL` (defaults to `http(s)://<host>:8080`).
4. Run dev server: `npm run dev`.
5. Build: `npm run build` (used by nginx image in docker-compose).

## Mobile Changes (Flutter)
Location: `frontend/mobile/kikonect`
1. Standard Flutter workflow (`flutter pub get`, `flutter run`).
2. In Compose, `client_mobile` builds a release APK and copies it into the web container.
3. If you add assets/plugins, ensure the Docker image or CI has the right SDKs.

### New Screens (Pages)
1. Add new screens under `lib/screens` (keep the page layout in the screen file).
2. Wire navigation with `Navigator.push` + `MaterialPageRoute` from the calling screen (or update `lib/app.dart` for a new entry point).
3. Use `Theme.of(context)` and shared widgets from `lib/widgets` to keep styling consistent.
4. Reuse `utils/ui_feedback.dart` for snackbars/errors instead of custom toasts.

### New Widgets
1. Place reusable UI components in `lib/widgets` once they are shared by multiple screens.
2. Keep widget APIs small and explicit (data in, callbacks out); avoid API calls inside widgets.
3. Follow the app theme (`Theme.of(context)` / `AppTheme`) instead of hardcoded colors.
4. Add widget tests under `test/` if the widget has logic or state.

### New Services and API Integrations
1. Add API calls in `lib/services/api_service.dart` or a new file under `lib/services`.
2. Use `flutter_dotenv` for base URLs and `flutter_secure_storage` for tokens; update `.env` and `.env.example` when you add env vars.
3. For new OAuth providers, extend `lib/services/oauth_service.dart` and keep the redirect logic in one place.
4. If a service needs icons/colors, add assets under `lib/assets`, register them in `pubspec.yaml`, and update `ServiceSelectionPage` fallback maps (`_serviceColors`, `_serviceIcons`) as needed.

---

## Coding Standards
- Go: `go fmt ./...` before committing. Keep functions small; prefer interfaces for seams.
- Tests: table-driven when possible; use `httptest` and `sqlmock`; avoid real network/DB in unit tests.
- API: strict JSON decoding (`DisallowUnknownFields`, `EnsureNoTrailingData`).
- Comments: short doc comments above exported functions (already present).

## CI & Local Checks
- CI runs in `.github/workflows/ci.yml`: `go build ./...` then `go test ./...` (working directory `backend`).
- Run the same locally before opening a PR.

## Pull Request Checklist
- Code formatted (`go fmt ./...`).
- Tests added/updated and passing (`go test ./...`).
- README/HOWTOCONTRIBUTE updated if behavior/config changed.
- No hardcoded secrets; env vars documented.
- New endpoints/config validated and error-handled.
