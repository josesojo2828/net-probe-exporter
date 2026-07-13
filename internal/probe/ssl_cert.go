package probe

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"time"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
)

// SSLCertProbe performs TLS certificate health checks.
type SSLCertProbe struct {
	name       string
	targetName string
	target     string
	port       int
	sni        string
	timeout    time.Duration
}

// NewSSLCertProbe creates a new SSL certificate probe from configuration.
func NewSSLCertProbe(cfg config.Probe) *SSLCertProbe {
	sni := cfg.SSL.SNI
	if sni == "" {
		sni = cfg.SSL.Target
	}

	return &SSLCertProbe{
		name:       cfg.Name,
		targetName: cfg.Name,
		target:     cfg.SSL.Target,
		port:       cfg.SSL.Port,
		sni:        sni,
		timeout:    cfg.Timeout,
	}
}

// Name returns the probe name.
func (p *SSLCertProbe) Name() string { return p.name }

// Type returns "ssl_cert".
func (p *SSLCertProbe) Type() string { return "ssl_cert" }

// Probe executes a TLS certificate check and returns the result.
func (p *SSLCertProbe) Probe(ctx context.Context) Result {
	start := time.Now()

	addr := fmt.Sprintf("%s:%d", p.target, p.port)

	tlsConfig := &tls.Config{
		ServerName: p.sni,
	}

	dialer := &tls.Dialer{
		NetDialer: &net.Dialer{Timeout: p.timeout},
		Config:    tlsConfig,
	}
	netConn, err := dialer.DialContext(ctx, "tcp", addr)
	latencyMs := time.Since(start).Seconds() * 1000

	if err != nil {
		return Result{
			TargetName: p.targetName,
			TargetType: "ssl_cert",
			Up:         false,
			LatencyMs:  latencyMs,
			Error:      fmt.Sprintf("tls connection failed: %v", err),
		}
	}
	defer netConn.Close()

	tlsConn, ok := netConn.(*tls.Conn)
	if !ok {
		return Result{
			TargetName: p.targetName,
			TargetType: "ssl_cert",
			Up:         false,
			LatencyMs:  latencyMs,
			Error:      "connection is not a TLS connection",
		}
	}

	state := tlsConn.ConnectionState()
	if len(state.PeerCertificates) == 0 {
		return Result{
			TargetName: p.targetName,
			TargetType: "ssl_cert",
			Up:         false,
			LatencyMs:  latencyMs,
			Error:      "no peer certificates returned",
		}
	}

	cert := state.PeerCertificates[0]
	daysUntilExpiry := int(time.Until(cert.NotAfter).Hours() / 24)

	up := daysUntilExpiry > 0

	extra := map[string]string{
		"days_until_expiry": fmt.Sprintf("%d", daysUntilExpiry),
		"issuer":            cert.Issuer.CommonName,
		"subject":           cert.Subject.CommonName,
		"valid_from":        cert.NotBefore.Format(time.RFC3339),
		"valid_to":          cert.NotAfter.Format(time.RFC3339),
	}

	return Result{
		TargetName: p.targetName,
		TargetType: "ssl_cert",
		Up:         up,
		LatencyMs:  latencyMs,
		Extra:      extra,
	}
}
