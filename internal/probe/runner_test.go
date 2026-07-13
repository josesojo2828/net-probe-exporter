package probe

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
)

func TestRunner_StartsAndStops(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	}))
	defer srv.Close()

	cfg := []config.Probe{
		{
			Name:     "test-http",
			Type:     "http",
			Interval: 50 * time.Millisecond,
			Timeout:  5 * time.Second,
			HTTP: &config.HTTPProbeConfig{
				URL:            srv.URL + "/health",
				Method:         "GET",
				ExpectedStatus: 200,
			},
		},
	}

	runner := NewRunner(cfg)
	ctx, cancel := context.WithCancel(context.Background())

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		runner.Start(ctx)
	}()

	// Give it time to run at least one probe cycle
	time.Sleep(150 * time.Millisecond)

	results := runner.Results()
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}

	result, ok := results["test-http"]
	if !ok {
		t.Fatal("expected result for test-http")
	}
	if !result.Up {
		t.Errorf("expected up=true, got up=%v (error: %s)", result.Up, result.Error)
	}
	if result.LatencyMs <= 0 {
		t.Errorf("expected latency > 0, got %f", result.LatencyMs)
	}

	cancel()
	wg.Wait()
}

func TestRunner_MultipleProbes(t *testing.T) {
	srv1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer srv1.Close()

	srv2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv2.Close()

	cfg := []config.Probe{
		{
			Name:     "up-probe",
			Type:     "http",
			Interval: 100 * time.Millisecond,
			Timeout:  5 * time.Second,
			HTTP: &config.HTTPProbeConfig{
				URL:            srv1.URL,
				Method:         "GET",
				ExpectedStatus: 200,
			},
		},
		{
			Name:     "down-probe",
			Type:     "http",
			Interval: 100 * time.Millisecond,
			Timeout:  5 * time.Second,
			HTTP: &config.HTTPProbeConfig{
				URL:            srv2.URL,
				Method:         "GET",
				ExpectedStatus: 200,
			},
		},
	}

	runner := NewRunner(cfg)
	ctx, cancel := context.WithCancel(context.Background())

	go runner.Start(ctx)

	// Let probes run a couple cycles
	time.Sleep(250 * time.Millisecond)
	cancel()

	results := runner.Results()

	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}

	// Check up-probe
	if r, ok := results["up-probe"]; !ok {
		t.Error("missing up-probe result")
	} else if !r.Up {
		t.Errorf("up-probe expected up=true, got %v", r.Up)
	}

	// Check down-probe
	if r, ok := results["down-probe"]; !ok {
		t.Error("missing down-probe result")
	} else if r.Up {
		t.Errorf("down-probe expected up=false, got %v", r.Up)
	}
}
