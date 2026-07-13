package probe

import (
	"testing"
	"time"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
)

func TestPostgresProbe_Name(t *testing.T) {
	cfg := config.Probe{
		Name: "test-postgres",
		Type: "postgres",
		Postgres: &config.PostgresProbeConfig{
			DSN: "postgres://localhost:5432/testdb",
		},
		Timeout: 5 * time.Second,
	}
	p := NewPostgresProbe(cfg)
	if p.Name() != "test-postgres" {
		t.Errorf("expected name 'test-postgres', got %q", p.Name())
	}
}

func TestPostgresProbe_Type(t *testing.T) {
	cfg := config.Probe{
		Name: "test-postgres",
		Type: "postgres",
		Postgres: &config.PostgresProbeConfig{
			DSN: "postgres://localhost:5432/testdb",
		},
		Timeout: 5 * time.Second,
	}
	p := NewPostgresProbe(cfg)
	if p.Type() != "postgres" {
		t.Errorf("expected type 'postgres', got %q", p.Type())
	}
}

func TestPostgresProbe_Constructor(t *testing.T) {
	cfg := config.Probe{
		Name: "my-pg-probe",
		Type: "postgres",
		Postgres: &config.PostgresProbeConfig{
			DSN:   "postgres://user:pass@localhost:5432/mydb",
			Query: "SELECT 1",
		},
		Timeout: 10 * time.Second,
	}
	p := NewPostgresProbe(cfg)
	if p.dsn != cfg.Postgres.DSN {
		t.Errorf("expected dsn %q, got %q", cfg.Postgres.DSN, p.dsn)
	}
	if p.query != cfg.Postgres.Query {
		t.Errorf("expected query %q, got %q", cfg.Postgres.Query, p.query)
	}
	if p.timeout != cfg.Timeout {
		t.Errorf("expected timeout %v, got %v", cfg.Timeout, p.timeout)
	}
}
