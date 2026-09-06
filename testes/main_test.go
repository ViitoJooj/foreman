//go:build e2e

// Package e2e holds end-to-end tests that run against a real Postgres and, where
// applicable, real Docker. They never use mocks or fakes. Run them with:
//
//	go test -tags=e2e ./testes/...
//
// By default a throwaway Postgres container is started (Docker required) with the
// 0001_init schema applied. Set E2E_DATABASE_URL to run against an existing
// database (e.g. a Supabase project that already has the migrations applied).
package e2e

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"

	"github.com/ViitoJooj/foreman/internal/adapters/driven/supabase"
)

// testPool is the shared connection pool for every e2e test in this package.
var testPool *pgxpool.Pool

func TestMain(m *testing.M) {
	ctx := context.Background()

	dsn := os.Getenv("E2E_DATABASE_URL")
	var cleanup func()

	if dsn == "" {
		container, err := postgres.Run(ctx,
			"postgres:17-alpine",
			postgres.WithDatabase("foreman_e2e"),
			postgres.WithUsername("foreman"),
			postgres.WithPassword("foreman"),
			postgres.WithInitScripts(filepath.Join("..", "migrations", "0001_init.up.sql")),
			testcontainers.WithWaitStrategy(
				wait.ForLog("database system is ready to accept connections").
					WithOccurrence(2).
					WithStartupTimeout(90*time.Second),
			),
		)
		if err != nil {
			fmt.Fprintln(os.Stderr, "e2e: skipping - could not start postgres container:", err)
			os.Exit(0)
		}
		cleanup = func() { _ = testcontainers.TerminateContainer(container) }

		dsn, err = container.ConnectionString(ctx, "sslmode=disable")
		if err != nil {
			fmt.Fprintln(os.Stderr, "e2e: connection string:", err)
			cleanup()
			os.Exit(1)
		}
	}

	pool, err := supabase.NewPool(ctx, dsn)
	if err != nil {
		fmt.Fprintln(os.Stderr, "e2e: open pool:", err)
		if cleanup != nil {
			cleanup()
		}
		os.Exit(1)
	}
	testPool = pool

	code := m.Run()

	pool.Close()
	if cleanup != nil {
		cleanup()
	}
	os.Exit(code)
}
