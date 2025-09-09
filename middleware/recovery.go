package middleware

import (
	"log"
	"net/http"

	"github.com/aminofox/zentrox"
)

// Recovery is a thin panic-to-500 JSON middleware kept for compatibility.
// Prefer ErrorHandler(...) for standardized error formatting.
func Recovery() zentrox.Handler {
	return func(c *zentrox.Context) {
		defer func() {
			if r := recover(); r != nil {
				log.Printf("panic: %v", r)
				c.SendJSON(http.StatusInternalServerError, zentrox.HTTPError{
					Code:    http.StatusInternalServerError,
					Message: "internal server error",
				})
				c.Abort()
			}
		}()
		c.Forward()
	}
}
