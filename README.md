# net-probe-exporter

[![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)](https://go.dev)
[![Prometheus](https://img.shields.io/badge/Prometheus-E6522C?logo=prometheus)](https://prometheus.io)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](LICENSE)

Prometheus exporter customizado para healthchecks HTTP/TCP. Monitoreá el uptime y latencia de tus servicios críticos y exponé métricas limpias para dashboards en Grafana.

## Features

- **Healthchecks HTTP** — verifica status codes y tiempos de respuesta
- **Healthchecks TCP** — verifica conectividad a nivel de transporte
- **Métricas Prometheus** — `net_probe_up`, `net_probe_latency_ms`, `net_probe_http_status_code`, `net_probe_scrapes_total`
- **Concurrencia nativa** — cada probe corre en su propia goroutine con intervalos configurables
- **Configuración YAML** — declarativa y fácil de versionar

## Quick Start

```bash
# Clonar
git clone https://github.com/josesojo2828/net-probe-exporter.git
cd net-probe-exporter

# Copiar y editar config
cp config.yaml.example config.yaml

# Build y ejecutar
make build
./net-probe-exporter
```

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

  - name: "postgres-local"
    type: tcp
    interval: 15s
    timeout: 3s
    tcp:
      host: "localhost:5432"
```

También podés usar la variable de entorno `NET_PROBE_CONFIG` para apuntar a otra ruta.

## Métricas

| Métrica | Tipo | Labels | Descripción |
|---------|------|--------|-------------|
| `net_probe_up` | Gauge | `target`, `type` | 1 si el target responde, 0 si no |
| `net_probe_latency_ms` | Gauge | `target`, `type` | Latencia en milisegundos |
| `net_probe_http_status_code` | Gauge | `target` | Último HTTP status code |
| `net_probe_scrapes_total` | Counter | `target`, `result` | Conteo de checks por resultado |

## Endpoints

- `GET /metrics` — métricas Prometheus
- `GET /health` — health check del exporter

## Dashboard Sugerido (Grafana)

Panel básico:

```promql
# Uptime de servicios
net_probe_up{type="http"}

# Latencia promedio
avg(net_probe_latency_ms) by (target)
```

## Development

```bash
make test        # Tests rápidos
make test-all    # Tests completos
make coverage    # Reporte de cobertura
make lint        # Verificar formato
```

## Docker

```bash
docker build -t net-probe-exporter .
docker run -p 9701:9701 -v $(pwd)/config.yaml:/config.yaml net-probe-exporter
```

## Licencia

MIT
