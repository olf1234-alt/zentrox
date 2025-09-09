package middleware

import (
	"time"

	"github.com/aminofox/zentrox"
	"github.com/aminofox/zentrox/telemetry"
)

// MetricsConfig wires a telemetry.Registry into middleware.
type MetricsConfig struct {
	Registry *telemetry.Registry
}

// Metrics records request count and latency histogram.
func Metrics(cfg MetricsConfig) zentrox.Handler {
	if cfg.Registry == nil {
		cfg.Registry = telemetry.NewRegistry()
	}
	return func(c *zentrox.Context) {
		start := time.Now()
		c.Forward()
		elapsed := time.Since(start)
		cfg.Registry.Requests.Add(1)
		cfg.Registry.Latency.Observe(float64(elapsed) / float64(time.Millisecond))
	}
}
