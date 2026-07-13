package probe

import (
	"context"
	"testing"
	"time"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
)

func TestSSLCertProbe_NameAndType(t *testing.T) {
	probe := NewSSLCertProbe(config.Probe{
		Name:    "test-ssl",
		Timeout: 5 * time.Second,
		SSL: &config.SSLCertProbeConfig{
			Target: "example.com",
			Port:   443,
		},
	})

	if probe.Name() != "test-ssl" {
		t.Errorf("expected name 'test-ssl', got %q", probe.Name())
	}
	if probe.Type() != "ssl_cert" {
		t.Errorf("expected type 'ssl_cert', got %q", probe.Type())
	}
}

func TestSSLCertProbe_ContextCancelled(t *testing.T) {
	probe := NewSSLCertProbe(config.Probe{
		Name:    "test-ssl-cancel",
		Timeout: 5 * time.Second,
		SSL: &config.SSLCertProbeConfig{
			Target: "example.com",
			Port:   443,
		},
	})

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	result := probe.Probe(ctx)

	if result.Up {
		t.Error("expected up=false for cancelled context")
	}
	if result.Error == "" {
		t.Error("expected error message for cancelled context")
	}
	if result.LatencyMs <= 0 {
		t.Errorf("expected latency > 0 even on error, got %f", result.LatencyMs)
	}
}

func TestSSLCertProbe_ConnectionError(t *testing.T) {
	probe := NewSSLCertProbe(config.Probe{
		Name:    "test-ssl-error",
		Timeout: 1 * time.Second,
		SSL: &config.SSLCertProbeConfig{
			Target: "127.0.0.1",
			Port:   19999,
		},
	})

	result := probe.Probe(context.Background())

	if result.Up {
		t.Error("expected up=false for connection error")
	}
	if result.Error == "" {
		t.Error("expected error message for connection error")
	}
}

func TestSSLCertProbe_DefaultPort(t *testing.T) {
	// Default port is set by config.validate(), not the constructor.
	// This test verifies the probe uses whatever port is configured.
	probe := NewSSLCertProbe(config.Probe{
		Name:    "test-ssl-default-port",
		Timeout: 5 * time.Second,
		SSL: &config.SSLCertProbeConfig{
			Target: "example.com",
			Port:   443,
		},
	})

	if probe.port != 443 {
		t.Errorf("expected port 443, got %d", probe.port)
	}
}

func TestSSLCertProbe_SNIDefault(t *testing.T) {
	probe := NewSSLCertProbe(config.Probe{
		Name:    "test-ssl-sni-default",
		Timeout: 5 * time.Second,
		SSL: &config.SSLCertProbeConfig{
			Target: "example.com",
			Port:   443,
			SNI:    "", // should default to target
		},
	})

	if probe.sni != "example.com" {
		t.Errorf("expected SNI to default to target 'example.com', got %q", probe.sni)
	}
}

func TestSSLCertProbe_SNICustom(t *testing.T) {
	probe := NewSSLCertProbe(config.Probe{
		Name:    "test-ssl-sni-custom",
		Timeout: 5 * time.Second,
		SSL: &config.SSLCertProbeConfig{
			Target: "example.com",
			Port:   443,
			SNI:    "custom.example.com",
		},
	})

	if probe.sni != "custom.example.com" {
		t.Errorf("expected SNI 'custom.example.com', got %q", probe.sni)
	}
}

func TestSSLCertProbe_ExtraMapPopulated(t *testing.T) {
	// This test requires network access to example.com
	// Skip in CI if needed
	probe := NewSSLCertProbe(config.Probe{
		Name:    "test-ssl-extra",
		Timeout: 5 * time.Second,
		SSL: &config.SSLCertProbeConfig{
			Target: "example.com",
			Port:   443,
		},
	})

	result := probe.Probe(context.Background())

	if !result.Up {
		t.Skipf("skipping: example.com not reachable: %s", result.Error)
	}

	if result.Extra == nil {
		t.Fatal("expected Extra map to be non-nil")
	}

	requiredKeys := []string{"days_until_expiry", "issuer", "subject", "valid_from", "valid_to"}
	for _, key := range requiredKeys {
		if _, ok := result.Extra[key]; !ok {
			t.Errorf("expected key %q in Extra map", key)
		}
	}
}
