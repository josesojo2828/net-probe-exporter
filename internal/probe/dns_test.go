package probe

import (
	"context"
	"testing"
	"time"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
)

func TestDNSProbe_NameAndType(t *testing.T) {
	probe := NewDNSProbe(config.Probe{
		Name:    "test-dns",
		Timeout: 5 * time.Second,
		DNS: &config.DNSProbeConfig{
			Target:     "example.com",
			RecordType: "A",
		},
	})

	if probe.Name() != "test-dns" {
		t.Errorf("expected name 'test-dns', got %q", probe.Name())
	}
	if probe.Type() != "dns" {
		t.Errorf("expected type 'dns', got %q", probe.Type())
	}
}

func TestDNSProbe_ContextCancelled(t *testing.T) {
	probe := NewDNSProbe(config.Probe{
		Name:    "test-dns-cancel",
		Timeout: 5 * time.Second,
		DNS: &config.DNSProbeConfig{
			Target:     "example.com",
			RecordType: "A",
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

func TestDNSProbe_NonExistentDomain(t *testing.T) {
	probe := NewDNSProbe(config.Probe{
		Name:    "test-dns-nonexistent",
		Timeout: 2 * time.Second,
		DNS: &config.DNSProbeConfig{
			Target:     "this-domain-does-not-exist-xyz123.test",
			RecordType: "A",
		},
	})

	result := probe.Probe(context.Background())

	if result.Up {
		t.Error("expected up=false for non-existent domain")
	}
	if result.Error == "" {
		t.Error("expected error message for non-existent domain")
	}
}

func TestDNSProbe_RecordType(t *testing.T) {
	tests := []struct {
		name       string
		recordType string
	}{
		{"A record", "A"},
		{"AAAA record", "AAAA"},
		{"MX record", "MX"},
		{"NS record", "NS"},
		{"CNAME record", "CNAME"},
		{"TXT record", "TXT"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			probe := NewDNSProbe(config.Probe{
				Name:    "test-dns-record-type",
				Timeout: 5 * time.Second,
				DNS: &config.DNSProbeConfig{
					Target:     "example.com",
					RecordType: tt.recordType,
				},
			})

			result := probe.Probe(context.Background())

			if result.TargetType != "dns" {
				t.Errorf("expected target type 'dns', got %q", result.TargetType)
			}
			if result.Extra == nil {
				t.Error("expected Extra map to be set")
			} else {
				if result.Extra["record_type"] != tt.recordType {
					t.Errorf("expected record_type %q in Extra, got %q", tt.recordType, result.Extra["record_type"])
				}
			}
		})
	}
}

func TestDNSProbe_ExtraMapPopulated(t *testing.T) {
	probe := NewDNSProbe(config.Probe{
		Name:    "test-dns-extra",
		Timeout: 5 * time.Second,
		DNS: &config.DNSProbeConfig{
			Target:     "example.com",
			RecordType: "A",
		},
	})

	result := probe.Probe(context.Background())

	if result.Extra == nil {
		t.Fatal("expected Extra map to be non-nil")
	}
	if _, ok := result.Extra["resolved"]; !ok {
		t.Error("expected 'resolved' key in Extra map")
	}
	if _, ok := result.Extra["record_type"]; !ok {
		t.Error("expected 'record_type' key in Extra map")
	}
}
