package main

import (
	"log"
	"net/http"
	"time"

	"github.com/aminofox/zentrox"
	"github.com/aminofox/zentrox/middleware"
	"github.com/aminofox/zentrox/telemetry"
)

func main() {
	app := zentrox.NewApp()

	// Shared metrics registry
	reg := telemetry.NewRegistry()

	// Recommended order
	app.Plug(
		middleware.RequestID(middleware.DefaultRequestID()),
		middleware.SimpleTrace(middleware.DefaultSimpleTrace()),
		middleware.AccessLog(middleware.AccessLogConfig{}),
		middleware.ErrorHandler(middleware.DefaultErrorHandler()),
		// Metrics is cheap; you can place it before or after AccessLog.
		middleware.Metrics(middleware.MetricsConfig{Registry: reg}),
	)

	// Metrics endpoint (plain text)
	app.OnGet("/metrics", func(c *zentrox.Context) {
		telemetry.MetricsHandler(reg).ServeHTTP(c.Writer, c.Request)
	})

	app.OnGet("/", func(c *zentrox.Context) {
		c.SendText(http.StatusOK, "hello, simple observability")
	})

	app.OnGet("/work", func(c *zentrox.Context) {
		time.Sleep(150 * time.Millisecond)
		c.SendJSON(200, map[string]any{
			"request_id": c.RequestID(),
			"trace_id":   getString(c, zentrox.TraceID),
			"span_id":    getString(c, zentrox.SpanID),
			"status":     "ok",
		})
	})

	log.Println("listening on :8000")
	_ = app.Run(":8000")
}

func getString(c *zentrox.Context, key string) string {
	if v, ok := c.Get(key); ok {
		if s, _ := v.(string); s != "" {
			return s
		}
	}
	return ""
}
