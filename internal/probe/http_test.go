package probe

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
)

func TestHTTPProbe_Up(t *testing.T) {
	// Start a test server that returns 200
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		fmt.Fprintln(w, "ok")
	}))
	defer srv.Close()

	probe := NewHTTPProbe(config.Probe{
		Name:    "test-http",
		Timeout: 5 * time.Second,
		HTTP: &config.HTTPProbeConfig{
			URL:            srv.URL + "/health",
			Method:         "GET",
			ExpectedStatus: 200,
		},
	})

	result := probe.Probe(context.Background())

	if !result.Up {
		t.Errorf("expected up=true, got up=%v (error: %s)", result.Up, result.Error)
	}
	if result.LatencyMs <= 0 {
		t.Errorf("expected latency > 0, got %f", result.LatencyMs)
	}
	if result.StatusCode != 200 {
		t.Errorf("expected status 200, got %d", result.StatusCode)
	}
}

func TestHTTPProbe_Down(t *testing.T) {
	// Start a test server that returns 500
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintln(w, "error")
	}))
	defer srv.Close()

	probe := NewHTTPProbe(config.Probe{
		Name:    "test-http-down",
		Timeout: 5 * time.Second,
		HTTP: &config.HTTPProbeConfig{
			URL:            srv.URL + "/fail",
			Method:         "GET",
			ExpectedStatus: 200,
		},
	})

	result := probe.Probe(context.Background())

	if result.Up {
		t.Errorf("expected up=false, got up=%v", result.Up)
	}
	if result.Error == "" {
		t.Error("expected non-empty error message")
	}
}

func TestHTTPProbe_Timeout(t *testing.T) {
	// Start a server that hangs
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(5 * time.Second)
		w.WriteHeader(http.StatusOK)
	}))
	defer srv.Close()

	probe := NewHTTPProbe(config.Probe{
		Name:    "test-http-timeout",
		Timeout: 100 * time.Millisecond,
		HTTP: &config.HTTPProbeConfig{
			URL:            srv.URL + "/slow",
			Method:         "GET",
			ExpectedStatus: 200,
		},
	})

	result := probe.Probe(context.Background())

	if result.Up {
		t.Errorf("expected up=false for timeout, got up=%v", result.Up)
	}
	if result.Error == "" {
		t.Error("expected error message for timeout")
	}
}

func TestHTTPProbe_Unreachable(t *testing.T) {
	probe := NewHTTPProbe(config.Probe{
		Name:    "test-http-unreachable",
		Timeout: 100 * time.Millisecond,
		HTTP: &config.HTTPProbeConfig{
			URL:            "http://192.0.2.1:9999",
			Method:         "GET",
			ExpectedStatus: 200,
		},
	})

	result := probe.Probe(context.Background())

	if result.Up {
		t.Errorf("expected up=false for unreachable, got up=%v", result.Up)
	}
	if result.LatencyMs <= 0 {
		t.Errorf("expected latency > 0 even on error, got %f", result.LatencyMs)
	}
}

func TestHTTPProbe_TableDriven(t *testing.T) {
	tests := []struct {
		name         string
		serverStatus int
		expectStatus int
		expectUp     bool
	}{
		{
			name:         "200 matches expected 200",
			serverStatus: 200,
			expectStatus: 200,
			expectUp:     true,
		},
		{
			name:         "200 does not match expected 201",
			serverStatus: 200,
			expectStatus: 201,
			expectUp:     false,
		},
		{
			name:         "500 does not match expected 200",
			serverStatus: 500,
			expectStatus: 200,
			expectUp:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(tt.serverStatus)
			}))
			defer srv.Close()

			probe := NewHTTPProbe(config.Probe{
				Name:    "test-http-table",
				Timeout: 5 * time.Second,
				HTTP: &config.HTTPProbeConfig{
					URL:            srv.URL,
					Method:         "GET",
					ExpectedStatus: tt.expectStatus,
				},
			})

			result := probe.Probe(context.Background())

			if result.Up != tt.expectUp {
				t.Errorf("up = %v, want %v", result.Up, tt.expectUp)
			}
			if result.StatusCode != tt.serverStatus {
				t.Errorf("status = %d, want %d", result.StatusCode, tt.serverStatus)
			}
		})
	}
}
