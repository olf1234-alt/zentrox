package main

import (
	"log"
	"time"

	"github.com/aminofox/zentrox"
	"github.com/aminofox/zentrox/middleware"
)

func main() {
	app := zentrox.NewApp()

	// Set Timeout early in the chain to cut off slow handlers early.
	app.Plug(
		middleware.Timeout(middleware.TimeoutOptions{
			Duration: 500 * time.Millisecond, // 0.5s
		}),
	)

	// Handler fast (OK in timeout)
	app.OnGet("/fast", func(c *zentrox.Context) {
		time.Sleep(100 * time.Millisecond)
		c.SendJSON(200, map[string]string{"status": "fast ok"})
	})

	// Handler slow (over timeout -> 504 Gateway Timeout)
	app.OnGet("/slow", func(c *zentrox.Context) {
		time.Sleep(1200 * time.Millisecond)
		c.SendJSON(200, map[string]string{"status": "slow done (you should not see this body)"})
	})

	// Handler writes header early then sleeps — check writer wrapped timeoutWriter
	app.OnGet("/header-then-sleep", func(c *zentrox.Context) {
		c.SetHeader("X-Debug", "set-before-sleep")
		time.Sleep(2 * time.Second)
		c.SendText(200, "done") // will not come if timeout
	})

	log.Println("listening on :8000")
	_ = app.Run(":8000")
}
