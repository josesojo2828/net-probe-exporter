package probe

import (
	"testing"
	"time"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
)

func TestMySQLProbe_Name(t *testing.T) {
	cfg := config.Probe{
		Name: "test-mysql",
		Type: "mysql",
		MySQL: &config.MySQLProbeConfig{
			DSN: "root:password@tcp(localhost:3306)/testdb",
		},
		Timeout: 5 * time.Second,
	}
	p := NewMySQLProbe(cfg)
	if p.Name() != "test-mysql" {
		t.Errorf("expected name 'test-mysql', got %q", p.Name())
	}
}

func TestMySQLProbe_Type(t *testing.T) {
	cfg := config.Probe{
		Name: "test-mysql",
		Type: "mysql",
		MySQL: &config.MySQLProbeConfig{
			DSN: "root:password@tcp(localhost:3306)/testdb",
		},
		Timeout: 5 * time.Second,
	}
	p := NewMySQLProbe(cfg)
	if p.Type() != "mysql" {
		t.Errorf("expected type 'mysql', got %q", p.Type())
	}
}

func TestMySQLProbe_Constructor(t *testing.T) {
	cfg := config.Probe{
		Name: "my-mysql-probe",
		Type: "mysql",
		MySQL: &config.MySQLProbeConfig{
			DSN:   "user:pass@tcp(localhost:3306)/mydb",
			Query: "SELECT 1",
		},
		Timeout: 10 * time.Second,
	}
	p := NewMySQLProbe(cfg)
	if p.dsn != cfg.MySQL.DSN {
		t.Errorf("expected dsn %q, got %q", cfg.MySQL.DSN, p.dsn)
	}
	if p.query != cfg.MySQL.Query {
		t.Errorf("expected query %q, got %q", cfg.MySQL.Query, p.query)
	}
	if p.timeout != cfg.Timeout {
		t.Errorf("expected timeout %v, got %v", cfg.Timeout, p.timeout)
	}
}
