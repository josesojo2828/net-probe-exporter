.PHONY: build run test test-all test-race coverage lint clean docker-build help

BINARY=net-probe-exporter
COMPOSE=docker compose

build:
	$(COMPOSE) build

run:
	$(COMPOSE) up -d

test:
	$(COMPOSE) run --rm tester

test-all:
	$(COMPOSE) run --rm tester go test ./... -v -count=1

test-race:
	$(COMPOSE) run --rm tester go test ./... -race -count=1 -short

coverage:
	$(COMPOSE) run --rm tester go test ./... -coverprofile=coverage.out -short
	$(COMPOSE) run --rm tester go tool cover -html=coverage.out -o coverage.html

lint:
	$(COMPOSE) run --rm tester gofmt -l .

clean:
	$(COMPOSE) down -v
	rm -f coverage.out coverage.html

docker-build:
	$(COMPOSE) build

help:
	@echo "Comandos (todos vía docker compose):"
	@echo "  make build       Build de la imagen"
	@echo "  make run         Levantar el exporter"
	@echo "  make test        Tests (corto)"
	@echo "  make test-all    Tests completos"
	@echo "  make test-race   Tests con -race"
	@echo "  make coverage    Reporte de cobertura"
	@echo "  make lint        Chequear formato"
	@echo "  make clean       Bajar contenedores"