package exporter

import (
	"net/http"
	"strconv"

	"github.com/josesojo2828/net-probe-exporter/internal/probe"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

// Metrics descriptors
var (
	upDesc = prometheus.NewDesc(
		"net_probe_up",
		"1 if the target is up, 0 otherwise",
		[]string{"target", "type"},
		nil,
	)

	latencyDesc = prometheus.NewDesc(
		"net_probe_latency_ms",
		"Latency of the last probe in milliseconds",
		[]string{"target", "type"},
		nil,
	)

	httpStatusCodeDesc = prometheus.NewDesc(
		"net_probe_http_status_code",
		"Last HTTP status code returned by the target",
		[]string{"target"},
		nil,
	)

	scrapesTotalDesc = prometheus.NewDesc(
		"net_probe_scrapes_total",
		"Total number of probe scrapes by result",
		[]string{"target", "result"},
		nil,
	)

	sslDaysUntilExpiryDesc = prometheus.NewDesc(
		"net_probe_ssl_days_until_expiry",
		"Number of days until the SSL certificate expires",
		[]string{"target", "type"},
		nil,
	)

	sslInfoDesc = prometheus.NewDesc(
		"net_probe_ssl_info",
		"SSL certificate metadata",
		[]string{"target", "issuer", "subject", "valid_from", "valid_to"},
		nil,
	)

	dbQueryDurationDesc = prometheus.NewDesc(
		"net_probe_db_query_duration_ms",
		"Duration of the last custom database query in milliseconds",
		[]string{"target"},
		nil,
	)
)

// Exporter implements prometheus.Collector and serves probe metrics.
type Exporter struct {
	runner       *probe.Runner
	scrapeCounts map[string]map[string]float64 // target -> result -> count
}

// New creates a new Exporter that reads results from the given runner.
func New(runner *probe.Runner) *Exporter {
	return &Exporter{
		runner:       runner,
		scrapeCounts: make(map[string]map[string]float64),
	}
}

// Describe implements prometheus.Collector.
func (e *Exporter) Describe(ch chan<- *prometheus.Desc) {
	ch <- upDesc
	ch <- latencyDesc
	ch <- httpStatusCodeDesc
	ch <- scrapesTotalDesc
	ch <- sslDaysUntilExpiryDesc
	ch <- sslInfoDesc
	ch <- dbQueryDurationDesc
}

// Collect implements prometheus.Collector.
func (e *Exporter) Collect(ch chan<- prometheus.Metric) {
	results := e.runner.Results()

	for _, r := range results {
		up := 0.0
		if r.Up {
			up = 1.0
		}

		ch <- prometheus.MustNewConstMetric(
			upDesc, prometheus.GaugeValue, up, r.TargetName, r.TargetType,
		)

		ch <- prometheus.MustNewConstMetric(
			latencyDesc, prometheus.GaugeValue, r.LatencyMs, r.TargetName, r.TargetType,
		)

		if r.TargetType == "http" && r.StatusCode > 0 {
			ch <- prometheus.MustNewConstMetric(
				httpStatusCodeDesc, prometheus.GaugeValue, float64(r.StatusCode), r.TargetName,
			)
		}

		if r.TargetType == "ssl_cert" && r.Extra != nil {
			if daysStr, ok := r.Extra["days_until_expiry"]; ok {
				if days, err := strconv.ParseFloat(daysStr, 64); err == nil {
					ch <- prometheus.MustNewConstMetric(
						sslDaysUntilExpiryDesc, prometheus.GaugeValue, days, r.TargetName, r.TargetType,
					)
				}
			}
			if issuer, ok := r.Extra["issuer"]; ok {
				subject := r.Extra["subject"]
				validFrom := r.Extra["valid_from"]
				validTo := r.Extra["valid_to"]
				ch <- prometheus.MustNewConstMetric(
					sslInfoDesc, prometheus.GaugeValue, 1, r.TargetName, issuer, subject, validFrom, validTo,
				)
			}
		}

		if (r.TargetType == "postgres" || r.TargetType == "mysql" || r.TargetType == "mongodb") && r.Extra != nil {
			if qDur, ok := r.Extra["query_duration_ms"]; ok {
				if dur, err := strconv.ParseFloat(qDur, 64); err == nil {
					ch <- prometheus.MustNewConstMetric(
						dbQueryDurationDesc, prometheus.GaugeValue, dur, r.TargetName,
					)
				}
			}
		}

		// Track scrape counts per target
		resultLabel := "up"
		if !r.Up {
			resultLabel = "down"
		}
		if e.scrapeCounts[r.TargetName] == nil {
			e.scrapeCounts[r.TargetName] = make(map[string]float64)
		}
		e.scrapeCounts[r.TargetName][resultLabel]++
		count := e.scrapeCounts[r.TargetName][resultLabel]

		ch <- prometheus.MustNewConstMetric(
			scrapesTotalDesc, prometheus.CounterValue, count, r.TargetName, resultLabel,
		)
	}
}

// Handler returns an http.Handler that serves Prometheus metrics via the
// default registry (which includes this exporter as a collector).
func (e *Exporter) Handler() http.Handler {
	registry := prometheus.NewRegistry()
	registry.MustRegister(e)
	return promhttp.HandlerFor(registry, promhttp.HandlerOpts{})
}
