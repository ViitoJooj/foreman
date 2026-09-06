# foreman

An autonomous multi-agent harness, in Go, that keeps GitHub repositories
maintained on your behalf — a Task Creator researches work, a Coder implements it,
a Tester validates it and a PR Reviewer decides whether it merges. Every agent
talks over a Slack-style channel; an orchestrator drives the task state machine,
guards a token budget and honours a kill switch.

> Product vision and full scope: [`claude/SKILL.md`](claude/SKILL.md).
> Working conventions for this repo: `CLAUDE.local.md` (not checked in).

## Status

Phase 1 skeleton. What works end to end today:

- **HTTP API** (`cmd/api`, Gin) — companies, agents, channels, tasks, channel
  messages (with an SSE stream), and control commands, all under `/api/v1` and
  guarded by an `X-Api-Key` header.
- **CLI** (`cmd/cli`, Cobra) — `foreman`, a pure HTTP client of the API:
  `status`, `company`, `agent`, `channel` (incl. `channel tail`), `task`,
  `message`, and `kill` / `pause` / `resume` / `command`.
- **Orchestrator** (`internal/orchestrator`) — runs in-process with the API when
  `ORCHESTRATOR_ENABLED=true`. Advances every actionable task one step per cycle,
  posts a status message on each transition, applies pending commands (kill /
  pause / resume) and enforces `BUDGET_HARD_CAP_USD`.
- **Runners** — stubs by default; `LLM_RUNNERS_ENABLED=true` switches to
  LLM-backed runners that decide the step outcome and write the channel message.
- **LLM adapters** (`internal/adapters/driven/llm`) — Claude (API key or OAuth
  bearer), OpenAI/Codex and DeepSeek, behind a fallback router.
- **GitHub adapter** (`internal/adapters/driven/github`) — wraps the `gh` CLI.
- **Persistence** — Postgres/Supabase via `pgx` (`internal/adapters/driven/supabase`),
  schema in `migrations/`.

Not wired yet: the autonomous Task Creator loop, the sandbox (ephemeral Docker
per task), the browser adapter, and the TUI dashboard.

## Quick start

See [`docs/GETTING_STARTED.md`](docs/GETTING_STARTED.md) for a full walk-through.
The short version, against a local Postgres:

```sh
docker run -d --name foreman-db -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=foreman -p 5432:5432 postgres:17-alpine
export DATABASE_URL='postgres://postgres:postgres@localhost:5432/foreman?sslmode=disable' API_KEY=dev-key
make migrate-up
make run-api        # in another shell: export FOREMAN_API_URL / FOREMAN_API_KEY, then ./bin/cli status
```

## Architecture

Hexagonal (ports & adapters), single module. Dependencies point inward:
`internal/core` (domain + ports + service use cases) never imports a framework or
an SDK — those live only in `internal/adapters`. `internal/orchestrator` depends
on the core ports plus the standard library.

```
cmd/{api,cli}            entrypoints / dependency wiring
internal/core/domain     pure entities (Task, Company, Agent, Message, Command, Channel)
internal/core/ports      interfaces (repositories, MessageBus, GitHubClient, LLMClient, ...)
internal/core/service    use cases
internal/orchestrator    scheduler, task state machine, budget guard, runners
internal/adapters/driving http (Gin), cli (Cobra)
internal/adapters/driven  supabase, bus (in-process), github, llm
migrations/               versioned SQL (golang-migrate)
testes/                   e2e suite (build tag `e2e`, testcontainers Postgres)
references/agents/        the inter-agent protocol and the four role prompts
```

## Development

```sh
make build         # ./bin/api and ./bin/cli
make test          # go test -race ./...
make test-e2e      # go test -tags=e2e ./testes/...  (needs Docker)
make lint          # golangci-lint (v2)
make fmt           # gofmt + goimports
make docker-build  # deploy/Dockerfile -> foreman:latest
```

Tooling and environment variables: [`docs/TOOLING.md`](docs/TOOLING.md).
