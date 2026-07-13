FROM golang:1.25-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 go build -o /net-probe-exporter ./cmd/net-probe-exporter

FROM alpine:3.20
RUN apk --no-cache add ca-certificates tzdata wget
COPY --from=builder /net-probe-exporter /usr/local/bin/net-probe-exporter
EXPOSE 9701
CMD ["net-probe-exporter"]