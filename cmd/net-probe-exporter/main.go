package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
	"github.com/josesojo2828/net-probe-exporter/internal/exporter"
	"github.com/josesojo2828/net-probe-exporter/internal/probe"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("error: %v", err)
	}
}

func run() error {
	configPath := os.Getenv("NET_PROBE_CONFIG")
	if configPath == "" {
		configPath = "config.yaml"
	}

	cfg, err := config.Load(configPath)
	if err != nil {
		return fmt.Errorf("loading config: %w", err)
	}

	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
		Level: cfg.Level(),
	})))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	runner := probe.NewRunner(cfg.Probes)
	go runner.Start(ctx)

	exp := exporter.New(runner)

	mux := http.NewServeMux()
	mux.Handle("/metrics", exp.Handler())
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/plain")
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	})

	addr := fmt.Sprintf(":%d", cfg.ListenPort)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	slog.Info("starting net-probe-exporter", "addr", addr)

	go func() {
		sigCh := make(chan os.Signal, 1)
		signal.Notify(sigCh, syscall.SIGINT, syscall.SIGTERM)
		<-sigCh
		slog.Info("shutting down...")
		cancel()
		srv.Shutdown(context.Background())
	}()

	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server: %w", err)
	}

	return nil
}
