package probe

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql" // MySQL driver
	"github.com/josesojo2828/net-probe-exporter/internal/config"
)

// MySQLProbe performs MySQL connectivity and optional query health checks.
type MySQLProbe struct {
	name       string
	targetName string
	dsn        string
	query      string
	timeout    time.Duration
}

// NewMySQLProbe creates a new MySQL probe from configuration.
func NewMySQLProbe(cfg config.Probe) *MySQLProbe {
	return &MySQLProbe{
		name:       cfg.Name,
		targetName: cfg.Name,
		dsn:        cfg.MySQL.DSN,
		query:      cfg.MySQL.Query,
		timeout:    cfg.Timeout,
	}
}

// Name returns the probe name.
func (p *MySQLProbe) Name() string { return p.name }

// Type returns "mysql".
func (p *MySQLProbe) Type() string { return "mysql" }

// Probe executes a MySQL connectivity check and returns the result.
func (p *MySQLProbe) Probe(ctx context.Context) Result {
	start := time.Now()

	// Expand env vars in DSN
	dsn := os.ExpandEnv(p.dsn)

	// Open connection using MySQL driver
	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return Result{
			TargetName: p.targetName,
			TargetType: "mysql",
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
			TargetType: "mysql",
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
		TargetType: "mysql",
		Up:         true,
		LatencyMs:  time.Since(start).Seconds() * 1000,
		Extra:      extra,
	}
}
