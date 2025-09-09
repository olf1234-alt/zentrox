package middleware

import (
	"log"
	"time"

	"github.com/aminofox/zentrox"
)

func Logger() zentrox.Handler {
	return func(c *zentrox.Context) {
		start := time.Now()
		c.Forward()
		log.Printf("%s %s %v", c.Request.Method, c.Request.URL.Path, time.Since(start))
	}
}
