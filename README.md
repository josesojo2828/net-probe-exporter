# net-probe-exporter

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?logo=prometheus)](https://prometheus.io)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Prometheus exporter customizado para healthchecks HTTP, TCP, SSL, DNS y bases de datos. Monitoreá uptime, latencia, certificados SSL y más con dashboards en Grafana.

## Features

- **Healthchecks HTTP** — verifica status codes y tiempos de respuesta
- **Healthchecks TCP** — verifica conectividad a nivel de transporte
- **SSL Cert** — días hasta expiración, issuer, subject, fechas de validez
- **DNS** — resolución de registros A, AAAA, MX, NS, CNAME, TXT
- **Bases de datos** — Postgres, MySQL, MongoDB (query custom)
- **Métricas Prometheus** — `net_probe_up`, `net_probe_latency_ms`, `net_probe_http_status_code`, `net_probe_scrapes_total`, `net_probe_ssl_days_until_expiry`, `net_probe_db_query_duration_ms`
- **Concurrencia nativa** — cada probe corre en su propia goroutine
- **Dashboard Grafana** — pre-provisionado con paneles de status, latencia, uptime y SSL
- **Configuración YAML** — declarativa y fácil de versionar
- **Retención automática** — Prometheus configurado con 30d / 10GB

## Quick Start

```bash
# Clonar
git clone https://github.com/josesojo2828/net-probe-exporter.git
cd net-probe-exporter

# Copiar y editar config
cp config.yaml.example config.yaml

# Copiar credenciales de Grafana (opcional, default admin/admin)
cp .env.example .env

# Levantar todo el stack
docker compose up -d
```

### URLs

| Servicio | URL |
|----------|-----|
| Métricas | `http://localhost:15001/metrics` |
| Prometheus | `http://localhost:15002` |
| Grafana | `http://localhost:15003` |

## Configuración

Creá un `config.yaml`:

```yaml
listen_port: 9701
log_level: "info"

probes:
  - name: "api-production"
    type: http
    interval: 30s
    timeout: 5s
    http:
      url: "https://api.tudominio.com/health"
      method: GET
      expected_status: 200

  - name: "api-ssl"
    type: ssl_cert
    interval: 6h
    timeout: 10s
    ssl_cert:
      target: "api.tudominio.com"
      port: 443
```

### Tipos de probe

| Tipo | Config | Métricas |
|------|--------|----------|
| `http` | `url`, `method`, `expected_status` | `net_probe_up`, `net_probe_latency_ms`, `net_probe_http_status_code` |
| `tcp` | `host` | `net_probe_up`, `net_probe_latency_ms` |
| `ssl_cert` | `target`, `port`, `sni` | `net_probe_up`, `net_probe_latency_ms`, `net_probe_ssl_days_until_expiry`, `net_probe_ssl_info` |
| `dns` | `target`, `server`, `record_type` | `net_probe_up`, `net_probe_latency_ms` + DNS records en extra |
| `postgres` | `dsn`, `query` | `net_probe_up`, `net_probe_db_query_duration_ms` |
| `mysql` | `dsn`, `query` | `net_probe_up`, `net_probe_db_query_duration_ms` |
| `mongodb` | `uri`, `query` | `net_probe_up`, `net_probe_db_query_duration_ms` |

## Dashboard

El dashboard incluye:

- **Service Status** — tabla con UP/DOWN por site
- **Average Latency** — time series de latencia por target
- **HTTP Status Code** — último código HTTP
- **Total Scrapes** — contador de checks
- **Scrapes by Result** — donut chart UP vs DOWN
- **SSL Days Until Expiry** — días restantes del certificado (verde >30, amarillo >15, naranja >7, rojo <7)
- **Uptime (7d)** — porcentaje de disponibilidad semanal (verde ≥99.9%)
- **Current Latency** — latencia actual por target

Filtrable por site con el dropdown superior.

## Retención de datos

Prometheus está configurado con:

- **30 días** de retención temporal
- **10 GB** máximo de almacenamiento

Los datos se borran automáticamente.

## Development

```bash
make test        # Tests rápidos
make test-all    # Tests completos
make coverage    # Reporte de cobertura
make lint        # Verificar formato
make build       # Compilar binario
```

## Docker Compose (Monitoring Stack)

```bash
# 1. Copiar configuración
cp config.yaml.example config.yaml
cp .env.example .env

# 2. Editar config.yaml con tus targets

# 3. Levantar
docker compose up -d

# 4. Ver logs
docker compose logs -f
```

### Servicios

| Servicio | Puerto host | Descripción |
|----------|-------------|-------------|
| `net-probe-exporter` | 15001 | Métricas Prometheus del healthcheck |
| `prometheus` | 15002 | Almacenamiento de métricas con retención 30d |
| `grafana` | 15003 | Dashboards provisionados automáticamente |

## Docker

```bash
docker build -t net-probe-exporter .
docker run -p 15001:9701 -v $(pwd)/config.yaml:/config.yaml net-probe-exporter
```

## Licencia

MIT
