# Tooling

What you need installed to work on this repo, and what is handled automatically.

## Handled automatically (no install)

| Tool | How |
| --- | --- |
| **Go 1.27.x toolchain** | `go.mod` pins `go 1.27.1`; the `go` command downloads the matching toolchain on first use (`GOTOOLCHAIN=auto`, the default). |
| **mockgen** | Pinned as a `tool` directive in `go.mod`. Regenerate the port mocks with `go generate ./internal/testutil/...`. |
| **golang-migrate** | `make migrate-up` / `make migrate-down` run it via `go run ...@<pinned version>`; nothing to install. They need `DATABASE_URL` set. |

## Install yourself

### golangci-lint (for `make lint`)

Version **2.x** (this repo's `.golangci.yml` uses the v2 schema).

```sh
# https://golangci-lint.run/welcome/install/
curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh | sh -s -- -b "$(go env GOPATH)/bin"
```

Lint the e2e files too:

```sh
golangci-lint run --build-tags=e2e
```

### A C compiler (for `make test`)

`make test` runs `go test -race ./...`, and the race detector requires cgo, which
requires a C compiler.

- **Linux/macOS**: gcc or clang (usually already present).
- **Windows**: install [w64devkit](https://github.com/skeeto/w64devkit) or
  TDM-GCC and make sure `gcc` is on `PATH`.

Without a C compiler the suite still runs, just without `-race`:

```sh
go test ./...
```

### Docker (for `make docker-*` and `make test-e2e`)

- `make docker-build` / `make docker-up` build and run the API image.
- `make test-e2e` starts a throwaway Postgres via
  [testcontainers](https://golang.testcontainers.org/); the Docker daemon must be
  running. When it is not, the e2e suite skips cleanly instead of failing.
- To run the e2e suite against an existing database instead of a container, set
  `E2E_DATABASE_URL` to a DSN whose schema already has `migrations/0001_init`
  applied.

## Environment variables

| Variable | Used by | Notes |
| --- | --- | --- |
| `DATABASE_URL` | API server, `make migrate-*` | Postgres/Supabase connection string. |
| `API_KEY` | API server | Shared secret checked against the `X-Api-Key` header. |
| `PORT` | API server | Defaults to `8080`. |
| `FOREMAN_API_URL` | `foreman` CLI | Base URL of the API. Defaults to `http://localhost:8080`. |
| `FOREMAN_API_KEY` | `foreman` CLI | Sent as `X-Api-Key`; overridable with `--api-key`. |
| `E2E_DATABASE_URL` | `make test-e2e` | Run e2e against this DB instead of a container. |
| `E2E_TEST_DOWN` | `make test-e2e` | Set to `1` to also exercise `0001_init.down.sql`. |
