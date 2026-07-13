package probe

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/jackc/pgx/v5/stdlib" // pgx driver registers as "pgx"
	"github.com/josesojo2828/net-probe-exporter/internal/config"
)

// PostgresProbe performs Postgres connectivity and optional query health checks.
type PostgresProbe struct {
	name       string
	targetName string
	dsn        string
	query      string
	timeout    time.Duration
}

// NewPostgresProbe creates a new Postgres probe from configuration.
func NewPostgresProbe(cfg config.Probe) *PostgresProbe {
	return &PostgresProbe{
		name:       cfg.Name,
		targetName: cfg.Name,
		dsn:        cfg.Postgres.DSN,
		query:      cfg.Postgres.Query,
		timeout:    cfg.Timeout,
	}
}

// Name returns the probe name.
func (p *PostgresProbe) Name() string { return p.name }

// Type returns "postgres".
func (p *PostgresProbe) Type() string { return "postgres" }

// Probe executes a Postgres connectivity check and returns the result.
func (p *PostgresProbe) Probe(ctx context.Context) Result {
	start := time.Now()

	// Expand env vars in DSN
	dsn := os.ExpandEnv(p.dsn)

	// Open connection using pgx driver
	db, err := sql.Open("pgx", dsn)
	if err != nil {
		return Result{
			TargetName: p.targetName,
			TargetType: "postgres",
			Up:         false,
			LatencyMs:  time.Since(start).Seconds() * 1000,
			Error:      fmt.Sprintf("open failed: %v", err),
		}
	}
	defer db.Close()

	// Ping (measures connectivity latency)
	pingStart := time.Now()
	err = db.PingContext(ctx)
	pingLatency := time.Since(pingStart).Seconds() * 1000

	if err != nil {
		return Result{
			TargetName: p.targetName,
			TargetType: "postgres",
			Up:         false,
			LatencyMs:  time.Since(start).Seconds() * 1000,
			Error:      fmt.Sprintf("ping failed: %v", err),
		}
	}

	extra := map[string]string{
		"ping_latency_ms": fmt.Sprintf("%.2f", pingLatency),
	}

	// Optional custom query
	if p.query != "" {
		qStart := time.Now()
		rows, err := db.QueryContext(ctx, p.query)
		qDuration := time.Since(qStart).Seconds() * 1000
		if err != nil {
			extra["query_error"] = err.Error()
		} else {
			defer rows.Close()
			count := 0
			for rows.Next() {
				count++
			}
			if rows.Err() != nil {
				extra["query_error"] = rows.Err().Error()
			} else {
				extra["query_duration_ms"] = fmt.Sprintf("%.2f", qDuration)
				extra["rows_count"] = fmt.Sprintf("%d", count)
			}
		}
	}

	return Result{
		TargetName: p.targetName,
		TargetType: "postgres",
		Up:         true,
		LatencyMs:  time.Since(start).Seconds() * 1000,
		Extra:      extra,
	}
}
