## KiKonect Backend Interaction Diagrams

### System Overview
```mermaid
flowchart LR
  subgraph Frontend
    Web[React/Vite]
    Mobile[Flutter APK]
  end

  subgraph Backend
    API[HTTP API]
    AuthSvc[Auth Service]
    WFSvc[Workflow Service]
    Triggerer
    Executor
  end

  subgraph DB[PostgreSQL]
    Users[(users)]
    Workflows[(workflows)]
    Runs[(workflow_runs)]
    Jobs[(jobs)]
  end

  Web -->|/login /register /workflows| API
  Mobile -->|future| API
  API --> AuthSvc
  API --> WFSvc
  AuthSvc --> Users
  WFSvc --> Workflows
  WFSvc --> Triggerer
  Triggerer --> Runs
  Triggerer --> Jobs
  Executor --> Jobs
  Executor -->|POST action_url| External[Action URL]
  Jobs --> Runs
  WFSvc -.scheduler/cron.-> Triggerer
```

### Auth: Register & Login
```mermaid
sequenceDiagram
  participant FE as Frontend
  participant API as /register or /login
  participant Auth as Auth Service
  participant Store as DBStore
  participant DB as PostgreSQL (users)

  FE->>API: POST /register {email,password,firstname,lastname}
  API->>Auth: Register(ctx, email, password, fn, ln)
  Auth->>Auth: HashPassword(password)
  Auth->>Store: Create(user, passwordHash)
  Store->>DB: INSERT INTO users ...
  DB-->>Store: new user id
  Store-->>Auth: user with id
  Auth-->>API: user (json)
  API-->>FE: 201 Created

  FE->>API: POST /login {email,password}
  API->>Auth: Authenticate(ctx, email, password)
  Auth->>Store: GetByEmail(email)
  Store->>DB: SELECT ... FROM users WHERE email=$1
  DB-->>Store: user row + password_hash
  Store-->>Auth: user + hash
  Auth->>Auth: CheckPassword(hash, password)
  Auth-->>API: user
  API-->>FE: 200 OK
```

### Workflow Creation
```mermaid
sequenceDiagram
  participant FE as Frontend
  participant API as POST /workflows
  participant Svc as Workflow Service
  participant Store as Workflow Store
  participant DB as PostgreSQL (workflows)

  FE->>API: POST /workflows {name, trigger_type, action_url, trigger_config}
  API->>Svc: CreateWorkflow(ctx, payload)
  Svc->>Svc: validate trigger_type/config
  Svc->>Store: CreateWorkflow(...)
  Store->>DB: INSERT INTO workflows (...)
  DB-->>Store: id (+ next_run_at if interval)
  Store->>DB: SELECT workflow by id
  DB-->>Store: workflow row
  Store-->>Svc: workflow
  Svc-->>API: workflow
  API-->>FE: 201 Created (workflow json)
```

### Manual Trigger (POST /workflows/{id}/trigger)
```mermaid
sequenceDiagram
  participant FE as Frontend
  participant API as /workflows/{id}/trigger
  participant Svc as Workflow Service
  participant Triggerer
  participant Store as Workflow Store
  participant DB as PostgreSQL

  FE->>API: POST payload
  API->>Svc: Trigger(ctx, workflowID, payload)
  Svc->>Store: GetWorkflow(id)
  Store->>DB: SELECT workflow WHERE id=$1
  DB-->>Store: workflow row
  Store-->>Svc: workflow
  Svc->>Triggerer: EnqueueRun(workflowID, payload)
  Triggerer->>Store: CreateRun(workflowID)
  Store->>DB: INSERT INTO workflow_runs
  DB-->>Store: run id
  Triggerer->>Store: CreateJob(workflowID, runID, payload)
  Store->>DB: INSERT INTO jobs
  DB-->>Store: job id
  Triggerer-->>Svc: run (pending)
  Svc-->>API: run
  API-->>FE: 202 Accepted
```

### Webhook Trigger (POST /hooks/{token})
```mermaid
sequenceDiagram
  participant Ext as External Caller
  participant API as /hooks/{token}
  participant Svc as Workflow Service
  participant Store as Workflow Store
  participant Triggerer
  participant DB as PostgreSQL

  Ext->>API: POST /hooks/{token} {payload}
  API->>Svc: TriggerWebhook(ctx, token, payload)
  Svc->>Store: FindWorkflowByToken(token)
  Store->>DB: SELECT ... WHERE trigger_type='webhook' AND trigger_config->>'token'=token
  DB-->>Store: workflow row
  Store-->>Svc: workflow
  Svc->>Triggerer: EnqueueRun(workflowID, payload)
  Triggerer->>Store: CreateRun / CreateJob
  Store->>DB: INSERT run + job
  Triggerer-->>Svc: run
  Svc-->>API: run
  API-->>Ext: 202 Accepted
```

### Interval Scheduler + Executor Loop
```mermaid
sequenceDiagram
  participant Scheduler as Interval loop (goroutine)
  participant Store as Workflow Store
  participant DB as PostgreSQL
  participant Triggerer
  participant Executor as Executor loop
  participant Jobs as jobs table
  participant Runs as workflow_runs
  participant Action as action_url endpoint

  Scheduler->>Store: ClaimDueIntervalWorkflows(now)
  Store->>DB: SELECT ... WHERE trigger_type='interval' AND next_run_at<=now FOR UPDATE SKIP LOCKED
  DB-->>Store: due workflows
  Store->>DB: UPDATE next_run_at = now + interval_minutes
  loop each due workflow
    Scheduler->>Triggerer: EnqueueRun(wfID, payload from config)
    Triggerer->>Store: CreateRun + CreateJob
    Store->>DB: INSERT run, INSERT job
  end

  loop Executor poll
    Executor->>Store: FetchNextPendingJob()
    Store->>DB: SELECT ... FROM jobs WHERE status='pending' FOR UPDATE SKIP LOCKED LIMIT 1
    alt no job
      DB-->>Store: sql.ErrNoRows
      Executor-->Executor: sleep until next tick
    else job found
      DB-->>Store: job row
      Store->>DB: UPDATE jobs SET status='processing', started_at=now
      Executor->>Store: GetWorkflow(job.WorkflowID)
      Store->>DB: SELECT workflow...
      Executor->>Action: HTTP POST (payload) with timeout
      alt success
        Executor->>Store: MarkJobSuccess(jobID)
        Store->>DB: UPDATE jobs SET status='succeeded', ended_at=now
        Executor->>Store: UpdateRun(runID, status=succeeded, ended_at=now)
        Store->>DB: UPDATE workflow_runs ...
      else failure
        Executor->>Store: MarkJobFailed(jobID, reason)
        Store->>DB: UPDATE jobs SET status='failed', error=reason
        Executor->>Store: UpdateRun(runID, status=failed, error=reason, ended_at=now)
        Store->>DB: UPDATE workflow_runs ...
      end
    end
  end
```
