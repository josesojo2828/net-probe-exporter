package config

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config represents the top-level configuration.
type Config struct {
	ListenPort int     `yaml:"listen_port"`
	LogLevel   string  `yaml:"log_level"`
	Probes     []Probe `yaml:"probes"`
}

// Probe is a generic probe configuration with type-specific settings.
// It uses custom unmarshaling to handle the union-like structure.
type Probe struct {
	Name     string        `yaml:"name"`
	Type     string        `yaml:"type"`
	Interval time.Duration `yaml:"interval"`
	Timeout  time.Duration `yaml:"timeout"`

	// HTTP specific settings
	HTTP *HTTPProbeConfig `yaml:"http,omitempty"`
	// TCP specific settings
	TCP *TCPProbeConfig `yaml:"tcp,omitempty"`
	// DNS specific settings
	DNS *DNSProbeConfig `yaml:"dns,omitempty"`
	// SSL certificate specific settings
	SSL *SSLCertProbeConfig `yaml:"ssl_cert,omitempty"`
	// Postgres specific settings
	Postgres *PostgresProbeConfig `yaml:"postgres,omitempty"`
	// MySQL specific settings
	MySQL *MySQLProbeConfig `yaml:"mysql,omitempty"`
}

// HTTPProbeConfig holds HTTP-specific probe settings.
type HTTPProbeConfig struct {
	URL            string `yaml:"url"`
	Method         string `yaml:"method"`
	ExpectedStatus int    `yaml:"expected_status"`
}

// TCPProbeConfig holds TCP-specific probe settings.
type TCPProbeConfig struct {
	Host string `yaml:"host"`
}

// DNSProbeConfig holds DNS-specific probe settings.
type DNSProbeConfig struct {
	Target     string `yaml:"target"`
	Server     string `yaml:"server,omitempty"`
	RecordType string `yaml:"record_type,omitempty"`
}

// SSLCertProbeConfig holds SSL certificate probe settings.
type SSLCertProbeConfig struct {
	Target string `yaml:"target"`
	Port   int    `yaml:"port,omitempty"`
	SNI    string `yaml:"sni,omitempty"`
}

// PostgresProbeConfig holds Postgres-specific probe settings.
type PostgresProbeConfig struct {
	DSN   string `yaml:"dsn"`
	Query string `yaml:"query,omitempty"`
}

// MySQLProbeConfig holds MySQL-specific probe settings.
type MySQLProbeConfig struct {
	DSN   string `yaml:"dsn"`
	Query string `yaml:"query,omitempty"`
}

// Defaults
const (
	DefaultListenPort = 9701
	DefaultInterval   = 30 * time.Second
	DefaultTimeout    = 5 * time.Second
	DefaultLogLevel   = "info"
)

// Level returns the slog.Level for the configured log level.
func (c *Config) Level() slog.Level {
	switch c.LogLevel {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Load reads and parses a YAML config file.
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("reading config file: %w", err)
	}

	cfg := &Config{
		ListenPort: DefaultListenPort,
		LogLevel:   DefaultLogLevel,
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("parsing config: %w", err)
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("invalid config: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.ListenPort < 1 || c.ListenPort > 65535 {
		return fmt.Errorf("listen_port %d out of range [1-65535]", c.ListenPort)
	}

	for i, p := range c.Probes {
		if err := p.validate(); err != nil {
			return fmt.Errorf("probe %d (%q): %w", i, p.Name, err)
		}
	}

	return nil
}

func (p *Probe) validate() error {
	if p.Name == "" {
		return fmt.Errorf("name is required")
	}
	if p.Interval <= 0 {
		p.Interval = DefaultInterval
	}
	if p.Timeout <= 0 {
		p.Timeout = DefaultTimeout
	}
	if p.Timeout >= p.Interval {
		return fmt.Errorf("timeout (%s) must be less than interval (%s)", p.Timeout, p.Interval)
	}

	switch p.Type {
	case "http":
		if p.HTTP == nil {
			return fmt.Errorf("http config is required for http probe type")
		}
		if p.HTTP.URL == "" {
			return fmt.Errorf("http.url is required")
		}
		if p.HTTP.Method == "" {
			p.HTTP.Method = "GET"
		}
		if p.HTTP.ExpectedStatus == 0 {
			p.HTTP.ExpectedStatus = 200
		}
	case "tcp":
		if p.TCP == nil {
			return fmt.Errorf("tcp config is required for tcp probe type")
		}
		if p.TCP.Host == "" {
			return fmt.Errorf("tcp.host is required")
		}
	case "dns":
		if p.DNS == nil {
			return fmt.Errorf("dns config is required for dns probe type")
		}
		if p.DNS.Target == "" {
			return fmt.Errorf("dns.target is required")
		}
		if p.DNS.RecordType == "" {
			p.DNS.RecordType = "A"
		}
		validRecordTypes := map[string]bool{"A": true, "AAAA": true, "MX": true, "NS": true, "CNAME": true, "TXT": true}
		if !validRecordTypes[p.DNS.RecordType] {
			return fmt.Errorf("dns.record_type must be one of: A, AAAA, MX, NS, CNAME, TXT (got %q)", p.DNS.RecordType)
		}
	case "ssl_cert":
		if p.SSL == nil {
			return fmt.Errorf("ssl_cert config is required for ssl_cert probe type")
		}
		if p.SSL.Target == "" {
			return fmt.Errorf("ssl_cert.target is required")
		}
		if p.SSL.Port == 0 {
			p.SSL.Port = 443
		}
	case "postgres":
		if p.Postgres == nil {
			return fmt.Errorf("postgres config is required for postgres probe type")
		}
		if p.Postgres.DSN == "" {
			return fmt.Errorf("postgres.dsn is required")
		}
	case "mysql":
		if p.MySQL == nil {
			return fmt.Errorf("mysql config is required for mysql probe type")
		}
		if p.MySQL.DSN == "" {
			return fmt.Errorf("mysql.dsn is required")
		}
	default:
		return fmt.Errorf("unsupported probe type %q (supported: http, tcp, dns, ssl_cert, postgres, mysql)", p.Type)
	}

	return nil
}
