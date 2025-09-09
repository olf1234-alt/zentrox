package main

import (
	"log"

	"github.com/aminofox/zentrox"
	"github.com/aminofox/zentrox/openapi"
)

type User struct {
	ID    string `json:"id"`
	Name  string `json:"name"`
	Email string `json:"email"`
}

type CreateUser struct {
	Name  string `json:"name"  validate:"required,min=2"`
	Email string `json:"email" validate:"required,email"`
}

type AppError struct {
	Detail string `json:"detail"`
	Status int    `json:"status"`
	Title  string `json:"Title"`
	Type   string `json:"Type"`
}

func main() {
	app := zentrox.NewApp().SetEnableOpenAPI(true)

	b := openapi.New(
		"Zentrox Example API",
		"1.0.0",
		openapi.WithServer("http://localhost:8000", "local"),
		openapi.WithDescription("Example demonstrating OpenAPI without helper wrappers"),
	)

	app.OnGet("/users/:id", func(c *zentrox.Context) {
		u := User{ID: c.Param("id"), Name: "Alice", Email: "alice@example.com"}
		c.SendJSON(200, u)
	})

	app.OnPost("/users", func(c *zentrox.Context) {
		var in CreateUser
		if err := c.BindJSONInto(&in); err != nil {
			c.Problemf(400, "invalid payload", err.Error())
			return
		}
		u := User{ID: "123", Name: in.Name, Email: in.Email}
		c.SendJSON(201, u)
	})

	app.OnGet("/healthz", func(c *zentrox.Context) {
		c.SendJSON(200, map[string]any{"status": "ok"})
	})

	// Convention: change ":id" -> "{id}" in the path when declaring in spec.
	openapi.Register(b, "GET", "/users/{id}",
		openapi.Op().
			SetSummary("Get user by ID").
			SetTag("users").
			// path param corresponding to :id
			PathParam("id", "string", true, "User ID").
			ResponseJSON(200, User{}, "OK").
			ResponseProblem(404, "User not found", AppError{}).
			ResponseProblem(500, "Internal server error", AppError{}),
	)

	openapi.Register(b, "POST", "/users",
		openapi.Op().
			SetSummary("Create new user").
			SetTag("users").
			RequestJSON(CreateUser{}, true, "payload").
			ResponseJSON(201, User{}, "Created").
			ResponseProblem(400, "Bad request", AppError{}),
	)

	openapi.Register(b, "GET", "/healthz",
		openapi.Op().
			SetSummary("Health check").
			SetTag("system").
			ResponseJSON(200, map[string]string{"status": "ok"}, "OK"),
	)

	log.Println("listening on :8000")
	_ = app.Run(":8000")
}
