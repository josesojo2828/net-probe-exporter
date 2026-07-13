package probe

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
)

// HTTPProbe performs HTTP health checks.
type HTTPProbe struct {
	name          string
	targetName    string
	url           string
	method        string
	expectedCode  int
	timeout       time.Duration
	client        *http.Client
}

// NewHTTPProbe creates a new HTTP probe from configuration.
func NewHTTPProbe(cfg config.Probe) *HTTPProbe {
	return &HTTPProbe{
		name:         cfg.Name,
		targetName:   cfg.Name,
		url:          cfg.HTTP.URL,
		method:       cfg.HTTP.Method,
		expectedCode: cfg.HTTP.ExpectedStatus,
		timeout:      cfg.Timeout,
		client: &http.Client{
			Timeout: cfg.Timeout,
			// Do not follow redirects for health checks
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
	}
}

// Name returns the probe name.
func (p *HTTPProbe) Name() string { return p.name }

// Type returns "http".
func (p *HTTPProbe) Type() string { return "http" }

// Probe executes an HTTP health check and returns the result.
func (p *HTTPProbe) Probe(ctx context.Context) Result {
	start := time.Now()

	req, err := http.NewRequestWithContext(ctx, p.method, p.url, nil)
	if err != nil {
		return Result{
			TargetName: p.targetName,
			TargetType: "http",
			Up:         false,
			LatencyMs:  time.Since(start).Seconds() * 1000,
			Error:      fmt.Sprintf("request creation: %v", err),
		}
	}

	resp, err := p.client.Do(req)
	latencyMs := time.Since(start).Seconds() * 1000
	if err != nil {
		return Result{
			TargetName: p.targetName,
			TargetType: "http",
			Up:         false,
			LatencyMs:  latencyMs,
			Error:      fmt.Sprintf("request failed: %v", err),
		}
	}
	defer resp.Body.Close()

	// Drain body to reuse connections
	io.Copy(io.Discard, resp.Body)

	up := resp.StatusCode == p.expectedCode
	errMsg := ""
	if !up {
		errMsg = fmt.Sprintf("expected status %d, got %d", p.expectedCode, resp.StatusCode)
	}

	return Result{
		TargetName: p.targetName,
		TargetType: "http",
		Up:         up,
		LatencyMs:  latencyMs,
		StatusCode: resp.StatusCode,
		Error:      errMsg,
	}
}
