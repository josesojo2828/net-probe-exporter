package probe

import "context"

// Result holds the outcome of a single probe execution.
type Result struct {
	TargetName string
	TargetType string
	Up         bool
	LatencyMs  float64
	StatusCode int // HTTP status code (0 for TCP probes)
	Error      string
}

// Prober is the interface implemented by HTTP and TCP probes.
type Prober interface {
	Name() string
	Type() string
	Probe(ctx context.Context) Result
}
