package probe

import (
	"testing"
	"time"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
)

func TestMongoDBProbe_Name(t *testing.T) {
	cfg := config.Probe{
		Name: "test-mongo",
		Type: "mongodb",
		MongoDB: &config.MongoDBProbeConfig{
			URI: "mongodb://localhost:27017",
		},
	}
	p := NewMongoDBProbe(cfg)
	if p.Name() != "test-mongo" {
		t.Errorf("expected name %q, got %q", "test-mongo", p.Name())
	}
}

func TestMongoDBProbe_Type(t *testing.T) {
	cfg := config.Probe{
		Name: "test-mongo",
		Type: "mongodb",
		MongoDB: &config.MongoDBProbeConfig{
			URI: "mongodb://localhost:27017",
		},
	}
	p := NewMongoDBProbe(cfg)
	if p.Type() != "mongodb" {
		t.Errorf("expected type %q, got %q", "mongodb", p.Type())
	}
}

func TestMongoDBProbe_Constructor(t *testing.T) {
	cfg := config.Probe{
		Name:     "my-mongo",
		Type:     "mongodb",
		Timeout:  10 * time.Second,
		MongoDB: &config.MongoDBProbeConfig{
			URI:   "mongodb://user:pass@localhost:27017/mydb",
			Query: `{"ping": 1}`,
		},
	}
	p := NewMongoDBProbe(cfg)

	if p.name != "my-mongo" {
		t.Errorf("expected name %q, got %q", "my-mongo", p.name)
	}
	if p.targetName != "my-mongo" {
		t.Errorf("expected targetName %q, got %q", "my-mongo", p.targetName)
	}
	if p.uri != "mongodb://user:pass@localhost:27017/mydb" {
		t.Errorf("expected uri %q, got %q", "mongodb://user:pass@localhost:27017/mydb", p.uri)
	}
	if p.query != `{"ping": 1}` {
		t.Errorf("expected query %q, got %q", `{"ping": 1}`, p.query)
	}
	if p.timeout != 10*time.Second {
		t.Errorf("expected timeout %v, got %v", 10*time.Second, p.timeout)
	}
}

func TestMongoDBProbe_ImplementsProber(t *testing.T) {
	cfg := config.Probe{
		Name: "test-mongo",
		Type: "mongodb",
		MongoDB: &config.MongoDBProbeConfig{
			URI: "mongodb://localhost:27017",
		},
	}
	p := NewMongoDBProbe(cfg)
	// Verify it satisfies the Prober interface at compile time
	var _ Prober = p
}
