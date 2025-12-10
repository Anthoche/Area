---------------------------
-- USERS
---------------------------
CREATE TABLE IF NOT EXISTS users (
    id              SERIAL PRIMARY KEY,
    email           VARCHAR(255) UNIQUE NOT NULL,
    firstname       VARCHAR(255) NOT NULL,
    lastname        VARCHAR(255) NOT NULL,
    password_hash   TEXT NOT NULL,
    created_at      TIMESTAMPTZ DEFAULT NOW()
);

---------------------------
-- WORKFLOWS
---------------------------
CREATE TABLE IF NOT EXISTS workflows (
    id               SERIAL PRIMARY KEY,
    user_id          INTEGER REFERENCES users(id) ON DELETE CASCADE,
    name             VARCHAR(255) NOT NULL,
    trigger_type     VARCHAR(64) NOT NULL,
    trigger_config   JSONB NOT NULL DEFAULT '{}'::jsonb,
    action_url       TEXT NOT NULL,
    enabled          BOOLEAN NOT NULL DEFAULT TRUE,
    next_run_at      TIMESTAMPTZ,
    created_at       TIMESTAMPTZ DEFAULT NOW()
);

UPDATE workflows
SET enabled = FALSE
WHERE trigger_type <> 'manual' AND enabled = TRUE;

---------------------------
-- WORKFLOW RUNS
---------------------------
CREATE TABLE IF NOT EXISTS workflow_runs (
    id           SERIAL PRIMARY KEY,
    workflow_id  INTEGER NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    status       VARCHAR(32) NOT NULL DEFAULT 'pending',
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    started_at   TIMESTAMPTZ,
    ended_at     TIMESTAMPTZ,
    error        TEXT
);

---------------------------
-- JOBS
---------------------------
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'job_status') THEN
        CREATE TYPE job_status AS ENUM ('pending', 'processing', 'succeeded', 'failed');
    END IF;
END
$$;

CREATE TABLE IF NOT EXISTS jobs (
    id           SERIAL PRIMARY KEY,
    workflow_id  INTEGER NOT NULL REFERENCES workflows(id) ON DELETE CASCADE,
    run_id       INTEGER NOT NULL REFERENCES workflow_runs(id) ON DELETE CASCADE,
    payload      JSONB,
    status       job_status NOT NULL DEFAULT 'pending',
    error        TEXT,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW(),
    started_at   TIMESTAMPTZ,
    ended_at     TIMESTAMPTZ
);

CREATE INDEX IF NOT EXISTS idx_jobs_status_created_at ON jobs (status, created_at);

---------------------------
-- GOOGLE TOKENS
---------------------------
CREATE TABLE IF NOT EXISTS google_tokens (
    id             SERIAL PRIMARY KEY,
    user_id        INTEGER REFERENCES users(id) ON DELETE CASCADE,
    access_token   TEXT NOT NULL,
    refresh_token  TEXT NOT NULL,
    expiry         TIMESTAMPTZ NOT NULL,
    created_at     TIMESTAMPTZ DEFAULT NOW()
);
