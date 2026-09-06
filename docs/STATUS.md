# Project status

Snapshot at the end of the autonomous build session. `main` has ~60 commits on
top of the initial scaffold; nothing is pushed.

## Green across the board

`make build`, `go vet ./...` (+ `-tags=e2e`), `gofmt`, `make test` (`-race`, 12
packages), `make test-e2e` (repos + full HTTP API + orchestrator against a real
Postgres 17 via testcontainers), `golangci-lint` v2 (+ `e2e` tag, 0 issues), and
`make docker-build` all pass.

## Built and verified

| Area | State |
| --- | --- |
| **Core domain / ports / services** | Company, Agent, Channel, Task, Message, Command entities; repository + `MessageBus` + `GitHubClient` + `BrowserClient` + `LLMClient` ports; use cases for every CRUD path plus `PostMessage`, `TailMessages`, `IssueCommand`. |
| **Persistence** | `pgx` + hand-written SQL against Supabase/Postgres; custom `uuid.UUID` pgtype codec; `migrations/0001_init` (native enums, `updated_at` triggers, deny-all RLS). |
| **HTTP API** (`cmd/api`, Gin) | `/api/v1` for companies, agents, channels, tasks, messages (+ SSE stream), commands; `X-Api-Key` middleware. |
| **CLI** (`cmd/cli`, Cobra) | `foreman` — `status`, `dashboard` (Bubble Tea TUI), `company`, `agent`, `channel` (+ `tail`), `task`, `message`, `kill`/`pause`/`resume`, `command`. Pure HTTP client. |
| **Orchestrator** | In-process with the API (`ORCHESTRATOR_ENABLED`). Pure state machine (`Next`, `AllowedOutcomes`), scheduler cycle, `Budget` guard with token accounting, applies pending kill/pause/resume, posts a status message per transition. Verified live: a low-risk task cascades `queued → merged` with per-role message attribution. |
| **Runners** | `StubRunner` (default) and `LLMRunner` (`LLM_RUNNERS_ENABLED`) that asks an LLM for a state-constrained `{outcome, note}` decision. |
| **LLM adapters** | Claude (API key or OAuth bearer + beta header), OpenAI/Codex and DeepSeek (chat-completions), behind a fallback `Router` configured from env. |
| **GitHub adapter** | `gh` CLI wrapper (branch/PR/merge/close/delete), command runner injectable. |
| **Agent protocol + prompts** | `references/agents/protocol.md` + the four role prompts, incl. the PR Reviewer's adversarial checklist. |

## Not done yet

- **Autonomous Task Creator loop** — research → generate tasks → `created → researching → queued`. Today `task create` puts work straight into `queued`.
- **Sandbox adapter** — ephemeral Docker container per task; no port or impl yet. LLM runners do not run real code / git / tests.
- **Real browser adapter** — `browser.NotImplemented` is the wired placeholder for `BrowserClient`.
- **`foreman auth login`** — the OAuth PKCE / device-code flows for Claude and Codex. Adapters accept a bearer token today; acquiring/refreshing it is manual.
- **`foreman budget`** — the `Budget` lives in the orchestrator goroutine; no endpoint exposes it.
- **Panic mode wiring** — a `panic` command sets the stop flag like `kill`; it does not yet close agent PRs / delete branches.

## Environment notes

The build machine got `ezwinports.make`, `BrechtSanders.WinLibs.POSIX.UCRT` (gcc,
for `-race`) and Docker Desktop running during the session. See
[`TOOLING.md`](TOOLING.md).
