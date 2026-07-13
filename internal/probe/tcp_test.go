package probe

import (
	"context"
	"net"
	"testing"
	"time"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
)

func TestTCPProbe_Up(t *testing.T) {
	// Start a local TCP listener
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer listener.Close()

	// Accept and close connection in background
	go func() {
		conn, err := listener.Accept()
		if err != nil {
			return
		}
		conn.Close()
	}()

	probe := NewTCPProbe(config.Probe{
		Name:    "test-tcp",
		Timeout: 5 * time.Second,
		TCP: &config.TCPProbeConfig{
			Host: listener.Addr().String(),
		},
	})

	result := probe.Probe(context.Background())

	if !result.Up {
		t.Errorf("expected up=true, got up=%v (error: %s)", result.Up, result.Error)
	}
	if result.LatencyMs <= 0 {
		t.Errorf("expected latency > 0, got %f", result.LatencyMs)
	}
}

func TestTCPProbe_Down(t *testing.T) {
	// Connect to a port that's not listening
	probe := NewTCPProbe(config.Probe{
		Name:    "test-tcp-down",
		Timeout: 100 * time.Millisecond,
		TCP: &config.TCPProbeConfig{
			Host: "127.0.0.1:19999",
		},
	})

	result := probe.Probe(context.Background())

	if result.Up {
		t.Errorf("expected up=false for closed port, got up=%v", result.Up)
	}
	if result.Error == "" {
		t.Error("expected error message for closed port")
	}
}

func TestTCPProbe_Unreachable(t *testing.T) {
	// Try connecting to a non-routable IP with short timeout
	probe := NewTCPProbe(config.Probe{
		Name:    "test-tcp-unreachable",
		Timeout: 500 * time.Millisecond,
		TCP: &config.TCPProbeConfig{
			Host: "203.0.113.1:9999",
		},
	})

	result := probe.Probe(context.Background())

	if result.Up {
		t.Errorf("expected up=false for unreachable, got up=%v", result.Up)
	}
}

func TestTCPProbe_TableDriven(t *testing.T) {
	// Start a listener that responds to connection
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("failed to start listener: %v", err)
	}
	defer listener.Close()

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				return
			}
			conn.Close()
		}
	}()

	tests := []struct {
		name     string
		host     string
		timeout  time.Duration
		expectUp bool
	}{
		{
			name:     "open port",
			host:     listener.Addr().String(),
			timeout:  5 * time.Second,
			expectUp: true,
		},
		{
			name:     "closed port",
			host:     "127.0.0.1:19998",
			timeout:  100 * time.Millisecond,
			expectUp: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probe := NewTCPProbe(config.Probe{
				Name:    "test-tcp-table",
				Timeout: tt.timeout,
				TCP: &config.TCPProbeConfig{
					Host: tt.host,
				},
			})

			result := probe.Probe(context.Background())

			if result.Up != tt.expectUp {
				t.Errorf("up = %v, want %v (error: %s)", result.Up, tt.expectUp, result.Error)
			}
		})
	}
}
