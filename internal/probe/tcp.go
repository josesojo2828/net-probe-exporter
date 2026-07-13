package probe

import (
	"context"
	"fmt"
	"net"
	"time"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
)

// TCPProbe performs TCP health checks.
type TCPProbe struct {
	name       string
	targetName string
	host       string
	timeout    time.Duration
}

// NewTCPProbe creates a new TCP probe from configuration.
func NewTCPProbe(cfg config.Probe) *TCPProbe {
	return &TCPProbe{
		name:       cfg.Name,
		targetName: cfg.Name,
		host:       cfg.TCP.Host,
		timeout:    cfg.Timeout,
	}
}

// Name returns the probe name.
func (p *TCPProbe) Name() string { return p.name }

// Type returns "tcp".
func (p *TCPProbe) Type() string { return "tcp" }

// Probe executes a TCP health check and returns the result.
func (p *TCPProbe) Probe(ctx context.Context) Result {
	start := time.Now()

	dialer := net.Dialer{Timeout: p.timeout}
	conn, err := dialer.DialContext(ctx, "tcp", p.host)
	latencyMs := time.Since(start).Seconds() * 1000

	if err != nil {
		return Result{
			TargetName: p.targetName,
			TargetType: "tcp",
			Up:         false,
			LatencyMs:  latencyMs,
			Error:      fmt.Sprintf("connection failed: %v", err),
		}
	}
	conn.Close()

	return Result{
		TargetName: p.targetName,
		TargetType: "tcp",
		Up:         true,
		LatencyMs:  latencyMs,
	}
}
