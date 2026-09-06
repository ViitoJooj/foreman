# Getting started

Bring the API up and drive it with the `foreman` CLI. Two paths for the
database: a throwaway local Postgres (fastest) or a hosted Supabase project.

See [TOOLING.md](TOOLING.md) for what to install.

## Path A — local Postgres (fastest)

### 1. Postgres

```bash
docker run -d --name foreman-db \
  -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=foreman \
  -p 5432:5432 postgres:17-alpine
```

### 2. Environment + schema

```bash
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/foreman?sslmode=disable'
export API_KEY='dev-key'
make migrate-up          # applies migrations/0001_init
```

### 3. Run the API

```bash
make run-api             # foreground on :8080
```

## Path B — hosted Supabase

Create a project at supabase.com, then from **Settings -> Database** copy the
connection string (URI). Skip the Docker steps above:

```bash
export DATABASE_URL='postgres://postgres:<pass>@db.<ref>.supabase.co:5432/postgres?sslmode=require'
export API_KEY='dev-key'
make migrate-up
make run-api
```

## Drive it with the CLI

In another terminal:

```bash
export FOREMAN_API_URL=http://localhost:8080
export FOREMAN_API_KEY=dev-key
make build               # produces ./bin/api and ./bin/cli
```

> The CLI binary is `./bin/cli` (the root command is named `foreman`).
> `make run-cli ARGS="company list"` works too.

### A full flow

```bash
# 1. a company (a repo the harness maintains)
./bin/cli company create --name "Demo" --slug demo \
  --repo-owner me --repo-name demo-repo
# -> created company <COMPANY_ID> (demo)

# 2. an agent and a channel for that company
./bin/cli agent create   --company <COMPANY_ID> --name "Ada" --role coder
./bin/cli channel create --company <COMPANY_ID> --name general
# -> created channel <CHANNEL_ID> (general)

# 3. a task
./bin/cli task create --company <COMPANY_ID> --title "first task" --risk low
./bin/cli task list   --company <COMPANY_ID>

# 4. a channel message (--from is an agent id from `agent list`)
./bin/cli agent list --company <COMPANY_ID>
./bin/cli message post --channel <CHANNEL_ID> --type status \
  --from <AGENT_ID> --body "hello" --payload '{"k":1}'
```

Add `--json` to any command for the raw API response (handy for scripting with
`jq`).

## Endpoints

All under `/api/v1`, guarded by the `X-Api-Key` header.

| Method & path | CLI |
| --- | --- |
| `POST /companies` | `company create` |
| `GET /companies` | `company list` |
| `GET /companies/:id` | `company get <id>` |
| `POST /agents` | `agent create` |
| `GET /agents?company=<id>` | `agent list --company <id>` |
| `GET /agents/:id` | `agent get <id>` |
| `POST /channels` | `channel create` |
| `GET /channels?company=<id>` | `channel list --company <id>` |
| `GET /channels/:id` | `channel get <id>` |
| `POST /tasks` | `task create` |
| `GET /tasks?company=<id>&state=<state>` | `task list` |
| `POST /channels/:id/messages` | `message post` |
| `POST /commands` | `kill` / `kill --panic` / `pause --agent` / `resume --agent` |
| `GET /commands?status=<status>` | `command list` |
| `GET /commands/:id` | `command get <id>` |

## Orchestrator

The orchestrator runs as a goroutine inside the API process. Enable it with
environment variables before `make run-api`:

```bash
export ORCHESTRATOR_ENABLED=true
export ORCHESTRATOR_INTERVAL=5s       # cycle interval (default 30s)
export BUDGET_HARD_CAP_USD=0          # 0 disables the spend guard
```

Each cycle it: applies pending commands (see below), then advances every task in
an actionable state one step, posting a `status` message to the company channel
on each transition and honouring `pause`/`kill`.

The runners are **stubs** for now — every step reports success, so a `low` risk
task walks `queued -> coding -> testing -> reviewing -> merged` on its own, and a
`medium`/`high` risk task stops at `needs_human`. Real LLM-backed runners come
later.

> A task created via `task create` starts in `created`, which the pipeline does
> not pick up (that is the Task Creator's stage). Until the Task Creator runner
> exists, move it into the queue manually:
> `UPDATE tasks SET state = 'queued' WHERE id = '<TASK_ID>';`

## Control actions (kill switch)

Issuing a command only **records** it as `pending`; the orchestrator (future
work) applies pending commands on its next cycle.

```bash
./bin/cli kill --company <COMPANY_ID> --reason "manual stop"   # graceful stop
./bin/cli kill --panic                                          # also close agent PRs / branches
./bin/cli pause  --agent <AGENT_ID>
./bin/cli resume --agent <AGENT_ID>
./bin/cli command list                                          # audit (defaults to pending)
```

## Run the API in a container instead

```bash
export DATABASE_URL=... API_KEY=...
make docker-up           # builds and runs foreman:latest
make docker-down
```

## Teardown (Path A)

```bash
docker rm -f foreman-db
```
