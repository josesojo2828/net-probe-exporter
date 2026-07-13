package probe

import (
	"context"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
)

// DNSProbe performs DNS resolution health checks.
type DNSProbe struct {
	name       string
	targetName string
	target     string
	server     string
	recordType string
	timeout    time.Duration
	resolver   *net.Resolver
}

// NewDNSProbe creates a new DNS probe from configuration.
func NewDNSProbe(cfg config.Probe) *DNSProbe {
	resolver := net.DefaultResolver

	if cfg.DNS.Server != "" {
		server := cfg.DNS.Server
		resolver = &net.Resolver{
			PreferGo: true,
			Dial: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{Timeout: cfg.Timeout}
				return d.DialContext(ctx, "udp", server)
			},
		}
	}

	return &DNSProbe{
		name:       cfg.Name,
		targetName: cfg.Name,
		target:     cfg.DNS.Target,
		server:     cfg.DNS.Server,
		recordType: cfg.DNS.RecordType,
		timeout:    cfg.Timeout,
		resolver:   resolver,
	}
}

// Name returns the probe name.
func (p *DNSProbe) Name() string { return p.name }

// Type returns "dns".
func (p *DNSProbe) Type() string { return "dns" }

// Probe executes a DNS resolution check and returns the result.
func (p *DNSProbe) Probe(ctx context.Context) Result {
	start := time.Now()

	var records []string
	var err error

	switch p.recordType {
	case "A", "AAAA":
		addrs, lookupErr := p.resolver.LookupHost(ctx, p.target)
		if lookupErr != nil {
			err = lookupErr
		} else {
			for _, addr := range addrs {
				ip := net.ParseIP(addr)
				if p.recordType == "A" && ip.To4() != nil {
					records = append(records, addr)
				} else if p.recordType == "AAAA" && ip.To4() == nil {
					records = append(records, addr)
				}
			}
		}
	case "MX":
		mxs, lookupErr := p.resolver.LookupMX(ctx, p.target)
		if lookupErr != nil {
			err = lookupErr
		} else {
			for _, mx := range mxs {
				records = append(records, fmt.Sprintf("%s %d", mx.Host, mx.Pref))
			}
		}
	case "NS":
		nss, lookupErr := p.resolver.LookupNS(ctx, p.target)
		if lookupErr != nil {
			err = lookupErr
		} else {
			for _, ns := range nss {
				records = append(records, ns.Host)
			}
		}
	case "CNAME":
		cname, lookupErr := p.resolver.LookupCNAME(ctx, p.target)
		if lookupErr != nil {
			err = lookupErr
		} else {
			records = append(records, cname)
		}
	case "TXT":
		txts, lookupErr := p.resolver.LookupTXT(ctx, p.target)
		if lookupErr != nil {
			err = lookupErr
		} else {
			records = append(records, txts...)
		}
	}

	latencyMs := time.Since(start).Seconds() * 1000

	up := len(records) > 0 && err == nil
	errMsg := ""
	if err != nil {
		errMsg = fmt.Sprintf("dns lookup failed: %v", err)
	}

	extra := map[string]string{
		"resolved":   strings.Join(records, ","),
		"record_type": p.recordType,
	}

	return Result{
		TargetName: p.targetName,
		TargetType: "dns",
		Up:         up,
		LatencyMs:  latencyMs,
		Error:      errMsg,
		Extra:      extra,
	}
}
