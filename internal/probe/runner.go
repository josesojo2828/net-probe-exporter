package probe

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
)

// probeTask pairs a prober with its execution interval.
type probeTask struct {
	prober   Prober
	interval time.Duration
}

// Runner manages and executes all probes concurrently.
type Runner struct {
	mu      sync.RWMutex
	tasks   []probeTask
	results map[string]Result // keyed by probe name
}

// NewRunner creates a new Runner from probe configurations.
func NewRunner(probesCfg []config.Probe) *Runner {
	r := &Runner{
		results: make(map[string]Result),
	}
	for _, cfg := range probesCfg {
		var p Prober
		switch cfg.Type {
		case "http":
			p = NewHTTPProbe(cfg)
		case "tcp":
			p = NewTCPProbe(cfg)
		case "dns":
			p = NewDNSProbe(cfg)
		case "ssl_cert":
			p = NewSSLCertProbe(cfg)
		case "postgres":
			p = NewPostgresProbe(cfg)
		case "mysql":
			p = NewMySQLProbe(cfg)
		}
		if p != nil {
			r.tasks = append(r.tasks, probeTask{
				prober:   p,
				interval: cfg.Interval,
			})
		}
	}
	return r
}

// Start begins running all probes on their configured intervals.
// Blocks until ctx is cancelled.
func (r *Runner) Start(ctx context.Context) {
	var wg sync.WaitGroup

	for _, t := range r.tasks {
		wg.Add(1)
		go r.runProbe(ctx, &wg, t)
	}

	wg.Wait()
	slog.Info("all probes stopped")
}

// Results returns a snapshot of the latest probe results.
func (r *Runner) Results() map[string]Result {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cp := make(map[string]Result, len(r.results))
	for k, v := range r.results {
		cp[k] = v
	}
	return cp
}

func (r *Runner) runProbe(ctx context.Context, wg *sync.WaitGroup, t probeTask) {
	defer wg.Done()

	// Do an immediate first check
	r.executeAndStore(ctx, t.prober)

	// Then run on interval
	ticker := time.NewTicker(t.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			r.executeAndStore(ctx, t.prober)
		}
	}
}

func (r *Runner) executeAndStore(ctx context.Context, p Prober) {
	slog.Debug("running probe", "name", p.Name(), "type", p.Type())
	result := p.Probe(ctx)

	r.mu.Lock()
	r.results[p.Name()] = result
	r.mu.Unlock()

	level := slog.LevelDebug
	if !result.Up {
		level = slog.LevelWarn
	}
	slog.Log(ctx, level, "probe result",
		"name", p.Name(),
		"type", p.Type(),
		"up", result.Up,
		"latency_ms", result.LatencyMs,
		"error", result.Error,
	)
}
