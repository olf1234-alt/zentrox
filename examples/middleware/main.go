package main

import (
	"log"
	"net/http"
	"time"

	"github.com/aminofox/zentrox"
	"github.com/aminofox/zentrox/middleware"
)

func main() {
	app := zentrox.NewApp()

	// CORS + GZIP
	app.Plug(
		middleware.CORS(middleware.DefaultCORS()),
		middleware.StrictCORS(middleware.StrictCORSConfig{
			RequireKnownOrigin:  true,
			AllowedOriginsExact: []string{"https://yourapp.com", "https://admin.yourapp.com"},
		}))

	// JWT protected group
	secret := []byte("supersecret")
	auth := app.Scope("/auth",
		middleware.JWT(middleware.JWTConfig{Secret: secret}),
		middleware.JWTChecks(middleware.JWTChecksConfig{
			AllowedAlgs: []string{"RS256", "ES256"},
			Issuer:      "https://issuer.example.com",
			Audience:    "api://zentrox",
			ClockSkew:   60 * time.Second,
			RequireExp:  true,
			RequireNbf:  false,
		}))

	app.OnGet("/", func(c *zentrox.Context) {
		c.SendText(200, "public ok")
	})

	app.OnGet("/token", func(c *zentrox.Context) {
		now := time.Now().Unix()
		tok, _ := middleware.SignHS256(map[string]any{"sub": "123", "name": "demo", "exp": now + 3600}, secret)
		c.SendJSON(200, map[string]any{"token": tok})
	})

	auth.OnGet("/me", func(c *zentrox.Context) {
		u, _ := c.Get("user")
		c.SendJSON(http.StatusOK, map[string]any{"user": u})
	})

	log.Println("listening on :8000")
	_ = app.Run(":8000")
}
