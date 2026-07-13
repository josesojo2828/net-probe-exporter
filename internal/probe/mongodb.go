package probe

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/josesojo2828/net-probe-exporter/internal/config"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// MongoDBProbe performs MongoDB connectivity and optional command health checks.
type MongoDBProbe struct {
	name       string
	targetName string
	uri        string
	query      string
	timeout    time.Duration
}

// NewMongoDBProbe creates a new MongoDB probe from configuration.
func NewMongoDBProbe(cfg config.Probe) *MongoDBProbe {
	return &MongoDBProbe{
		name:       cfg.Name,
		targetName: cfg.Name,
		uri:        cfg.MongoDB.URI,
		query:      cfg.MongoDB.Query,
		timeout:    cfg.Timeout,
	}
}

// Name returns the probe name.
func (p *MongoDBProbe) Name() string { return p.name }

// Type returns "mongodb".
func (p *MongoDBProbe) Type() string { return "mongodb" }

// Probe executes a MongoDB connectivity check and returns the result.
func (p *MongoDBProbe) Probe(ctx context.Context) Result {
	start := time.Now()

	// Expand env vars in URI
	uri := os.ExpandEnv(p.uri)

	// Connect to MongoDB
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return Result{
			TargetName: p.targetName,
			TargetType: "mongodb",
			Up:         false,
			LatencyMs:  time.Since(start).Seconds() * 1000,
			Error:      fmt.Sprintf("connect failed: %v", err),
		}
	}
	defer client.Disconnect(ctx)

	// Ping (measures connectivity latency)
	pingStart := time.Now()
	err = client.Ping(ctx, nil)
	pingLatency := time.Since(pingStart).Seconds() * 1000

	if err != nil {
		return Result{
			TargetName: p.targetName,
			TargetType: "mongodb",
			Up:         false,
			LatencyMs:  time.Since(start).Seconds() * 1000,
			Error:      fmt.Sprintf("ping failed: %v", err),
		}
	}

	extra := map[string]string{
		"ping_latency_ms": fmt.Sprintf("%.2f", pingLatency),
	}

	// Optional query (command via RunCommand)
	// The user provides a BSON document like {"buildInfo": 1}
	if p.query != "" {
		qStart := time.Now()
		var result bson.M
		err := bson.UnmarshalExtJSON([]byte(p.query), false, &result)
		if err != nil {
			extra["query_error"] = fmt.Sprintf("invalid query json: %v", err)
		} else {
			cmdResult := client.Database("admin").RunCommand(ctx, result)
			qDuration := time.Since(qStart).Seconds() * 1000
			extra["query_duration_ms"] = fmt.Sprintf("%.2f", qDuration)
			if cmdResult.Err() != nil {
				extra["query_error"] = cmdResult.Err().Error()
			}
		}
	}

	return Result{
		TargetName: p.targetName,
		TargetType: "mongodb",
		Up:         true,
		LatencyMs:  time.Since(start).Seconds() * 1000,
		Extra:      extra,
	}
}
