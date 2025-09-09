package main

import (
	"fmt"
	"time"

	"github.com/aminofox/zentrox"
	"github.com/aminofox/zentrox/middleware"
	"github.com/aminofox/zentrox/telemetry"
)

func main() {
	app := zentrox.NewApp()

	// Attach Gzip early in the chain (before the logger), as it changes writer/headers.
	reg := telemetry.NewRegistry()
	app.Plug(
		middleware.Metrics(middleware.MetricsConfig{
			Registry: reg,
		}),
		middleware.Gzip(),
	)

	app.OnGet("/json", func(c *zentrox.Context) {
		c.SetHeader("Content-Type", "application/json; charset=utf-8")
		c.SendJSON(200, map[string]any{
			"message": "hello, gzip!",
			"time":    time.Now().Format(time.RFC3339),
			"tips":    "use curl --compressed to auto-decompress on client",
		})
	})

	// Return text/plain (also compressed if client accepts)
	app.OnGet("/text", func(c *zentrox.Context) {
		c.SetHeader("Content-Type", "text/plain; charset=utf-8")
		c.SendText(200, longText(8)) //long text to see compression benefits
	})

	// Returning “image/*” will be ignored by the middleware
	app.OnGet("/image", func(c *zentrox.Context) {
		c.SetHeader("Content-Type", "image/png")
		c.SendBytes(200, []byte{0x89, 0x50, 0x4E, 0x47})
	})

	fmt.Println("Gzip example running on :8000")
	_ = app.Run(":8000")
}

func longText(lines int) string {
	s := ""
	for i := 0; i < lines; i++ {
		s += "Lorem ipsum dolor sit amet, consectetur adipiscing elit. Sed do eiusmod tempor incididunt ut labore et dolore magna aliqua.\n"
	}
	return s
}
