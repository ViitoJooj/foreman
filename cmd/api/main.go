// Command api runs the harness HTTP API server.
package main

import (
	"context"
	"errors"
	"log"
	nethttp "net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ViitoJooj/foreman/internal/adapters/driven/bus"
	"github.com/ViitoJooj/foreman/internal/adapters/driven/supabase"
	"github.com/ViitoJooj/foreman/internal/adapters/driving/http"
	"github.com/ViitoJooj/foreman/internal/config"
	"github.com/ViitoJooj/foreman/internal/core/service"
)

const shutdownTimeout = 10 * time.Second

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	pool, err := supabase.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer pool.Close()

	companies := supabase.NewCompanyRepository(pool)
	agents := supabase.NewAgentRepository(pool)
	channels := supabase.NewChannelRepository(pool)
	tasks := supabase.NewTaskRepository(pool)
	messages := supabase.NewMessageRepository(pool)
	messageBus := bus.NewInProcess()

	router := http.NewRouter(cfg.APIKey, http.Handlers{
		Companies: http.NewCompanyHandler(service.NewCompany(companies)),
		Agents:    http.NewAgentHandler(service.NewAgent(agents)),
		Channels:  http.NewChannelHandler(service.NewChannel(channels)),
		Tasks:     http.NewTaskHandler(service.NewCreateTask(tasks), service.NewListTasks(tasks)),
		Messages:  http.NewMessageHandler(service.NewPostMessage(messages, messageBus)),
	})

	srv := &nethttp.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 10 * time.Second,
	}

	errCh := make(chan error, 1)
	go func() {
		log.Printf("api listening on %s", srv.Addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, nethttp.ErrServerClosed) {
			errCh <- err
		}
	}()

	select {
	case err := <-errCh:
		return err
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), shutdownTimeout)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	}
}
