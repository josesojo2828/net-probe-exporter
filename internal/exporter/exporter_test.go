package exporter

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
	"github.com/josesojo2828/net-probe-exporter/internal/probe"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/prometheus/client_golang/prometheus/testutil"
)

func TestExporter_MetricsFormat(t *testing.T) {
	// Setup: an HTTP server that's up
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	}))
	defer srv.Close()

	probeCfg := []config.Probe{
		{
			Name:     "test-service",
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

	runner := probe.NewRunner(probeCfg)
	ctx, cancel := context.WithCancel(context.Background())
	go runner.Start(ctx)
	time.Sleep(100 * time.Millisecond)
	cancel()

	exp := New(runner)
	registry := prometheus.NewRegistry()
	registry.MustRegister(exp)

	// Check metrics via HTTP
	handler := promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
	req := httptest.NewRequest("GET", "/metrics", nil)
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	body := rec.Body.String()

	// Check that expected metrics are present
	checks := []string{
		"net_probe_up{target=\"test-service\",type=\"http\"} 1",
		"net_probe_latency_ms{target=\"test-service\",type=\"http\"}",
		"net_probe_http_status_code{target=\"test-service\"} 200",
		"net_probe_scrapes_total{result=\"up\",target=\"test-service\"}",
	}

	for _, check := range checks {
		if !strings.Contains(body, check) {
			t.Errorf("expected metric not found:\n  %q\n  in body:\n%s", check, body)
		}
	}
}

func TestExporter_DownMetric(t *testing.T) {
	// No server -> probe will fail
	probeCfg := []config.Probe{
		{
			Name:     "down-service",
			Type:     "tcp",
			Interval: 50 * time.Millisecond,
			Timeout:  100 * time.Millisecond,
			TCP: &config.TCPProbeConfig{
				Host: "127.0.0.1:19999",
			},
		},
	}

	runner := probe.NewRunner(probeCfg)
	ctx, cancel := context.WithCancel(context.Background())
	go runner.Start(ctx)
	time.Sleep(200 * time.Millisecond)
	cancel()

	exp := New(runner)

	// Verify via testutil
	expected := `
		# HELP net_probe_up 1 if the target is up, 0 otherwise
		# TYPE net_probe_up gauge
		net_probe_up{target="down-service",type="tcp"} 0
	`

	if err := testutil.CollectAndCompare(exp, strings.NewReader(expected),
		"net_probe_up"); err != nil {
		t.Errorf("unexpected metrics: %v", err)
	}
}
