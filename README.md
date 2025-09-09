# Zentrox
![Zentrox](./zentrox.png)
A tiny, fast, and friendly HTTP micro-framework for Go.  
It wraps `net/http` with a minimalist router, chainable middleware, route scopes, and a pragmatic `Context`.

---

## Table of contents

-   [Install](#install)
-   [Quick start](#quick-start)
-   [Routing](#routing)
    -   [HTTP methods](#http-methods)
    -   [Path params](#path-params)
    -   [Wildcards](#wildcards)
    -   [Route scopes (prefix + middleware)](#route-scopes-prefix--middleware)
-   [Middleware](#middleware)
    -   [Global middleware](#global-middleware)
    -   [Per-route middleware](#per-route-middleware)
    -   [Per-scope middleware](#per-scope-middleware)
    -   [Built-ins: Logger, Recovery, CORS, GZIP, JWT](#builtins-logger-recovery-cors-gzip-jwt)
-   [Context API](#context-api)
-   [Binding & Validation](#binding--validation)
    -   [Bind helpers](#bind-helpers)
    -   [Validation tags](#validation-tags)
    -   [Examples](#examples)
-   [Error handling](#error-handling)
-   [Compatibility](#compatibility)
-   [Performance optimizations](#performance-optimizations)
-   [Observability](#observability)
-   [Testing](#testing)
-   [Benchmarks](#performance-benchmarks)
-   [License](#license)

---

## Install

```bash
go get github.com/aminofox/zentrox
```

---

## Quick start

```go
package main

import (
    "github.com/aminofox/zentrox"
    "github.com/aminofox/zentrox/middleware"
)

func main() {
    app := zentrox.NewApp()

    // Global middleware
    app.Plug(middleware.Recovery(), middleware.Logger())

    // Plain text
    app.OnGet("/", func(ctx *zentrox.Context) {
        ctx.SendText(200, "zentrox up!")
    })

    // Path param: /users/42  -> {"id":"42"}
    app.OnGet("/users/:id", func(ctx *zentrox.Context) {
        ctx.SendJSON(200, map[string]string{"id": ctx.Param("id")})
    })

    // Scope + JSON echo
    api := app.Scope("/api")
    api.OnPost("/echo", func(ctx *zentrox.Context) {
        var p map[string]any
        _ = ctx.BindInto(&p)
        ctx.SendJSON(200, p)
    })

    // Start server
    _ = app.Run(":8000")
}
```

Run:

```bash
go run ./examples/basic
# open http://localhost:8000
```

---

## Routing

### HTTP methods

```go
app.OnGet(path, handler)
app.OnPost(path, handler)
app.OnPut(path, handler)
app.OnPatch(path, handler)
app.OnDelete(path, handler)
```

### Path params

```go
app.OnGet("/users/:id", func(ctx *zentrox.Context) {
    id := ctx.Param("id") // "123"
    ctx.SendJSON(200, map[string]string{"id": id})
})
```

### Wildcards

Wildcard must be the last segment:

```go
app.OnGet("/static/*filepath", func(ctx *zentrox.Context) {
    fp := ctx.Param("filepath") // e.g. "css/app.css"
    ctx.SendText(200, fp)
})
```

### Route scopes (prefix + middleware)

```go
api := app.Scope("/api", authMiddleware)

api.OnGet("/profile", func(ctx *zentrox.Context) { /* ... */ })
api.OnPost("/posts",  func(ctx *zentrox.Context) { /* ... */ })

// Scope-level middleware runs after global, before handler
```

---

## Middleware

### Global middleware

```go
app.Plug(middleware.Recovery(), middleware.Logger())
```

### Per-route middleware
Every handler can accept extra middlewares:
```go
func AuthGuard() zentrox.Handler {
    return func(ctx *zentrox.Context) {
        // before
        ctx.Forward() // call next middleware/handler
        // after
    }
}

app.OnGet("/secure", AuthGuard(), (func(c *zentrox.Context) {
    c.SendText(200, "secure ok")
}))
```

### Per-scope middleware
```go
func AuthGuard() zentrox.Handler {
    return func(ctx *zentrox.Context) {
        // before
        ctx.Forward() // call next middleware/handler
        // after
    }
}

api := app.Scope("api", AuthGuard())
api.OnGet("/users", func(ctx *zentrox.Context) {
    ctx.SendText(200, "list ok")
})
```

### Built-ins: Logger, Recovery, CORS, GZIP, JWT

```go
import "github.com/aminofox/zentrox/middleware"

app.Plug(
    middleware.Logger(),       // logs METHOD PATH duration
    middleware.Recovery(),     // catches panics -> 500 JSON
    middleware.CORS(...),      // CORS config
    middleware.Gzip(),         // gzip compression
    middleware.JWT(secretKey), // JWT auth
)
```

---

## Context API

```go
// Inputs
ctx.Param("id")             // path param
ctx.Query("q")              // query string value
ctx.BindInto(&dst)           // auto bind JSON/form/query + validate
ctx.BindJSONInto(&dst)       // JSON + validate
ctx.BindFormInto(&dst)       // form-urlencoded + validate
ctx.BindQueryInto(&dst)      // query params + validate

// Outputs
ctx.SendJSON(200, v)         // JSON response
ctx.SendText(200, "hi")     // plain text
ctx.SendHTML(200, "<h1>")   // HTML
ctx.SendXML(200, obj)        // XML
ctx.SendFile("foo.pdf")     // file response
ctx.PushStream(fn)           // streaming writer
ctx.PushSSE(event, data)     // SSE event

// Flow control
ctx.Forward()                // go to next middleware
```

---

## Binding & Validation

### Bind helpers

The helpers below **bind** the data from the request to the struct and **automatically validate** it according to the `validate` tag:

```go
// Auto-detect binder (JSON/Form/Query), then validate
func (c *Context) BindInto(dst any) error

func (c *Context) BindJSONInto(dst any) error
func (c *Context) BindFormInto(dst any) error
func (c *Context) BindQueryInto(dst any) error
```

### Validation tags

Lightweight validation engine, supporting common rules:

-   `required` — required, no zero-value
-   `min=`, `max=` —
-   number: min/max value
-   string/slice/array: min/max length
-   `len=` — exact length (string/slice/array)
-   `email` — string must be a valid email (basic regex)
-   `oneof=...` — one of the allowed values
-   example: `oneof=small medium large` (separated by spaces or commas)
-   `regex=...` — regular expression matching (Go regex)

> Example struct:

```go
type UserInput struct {
    Name   string `json:"name" validate:"required,min=3,max=50"`
    Email  string `json:"email" validate:"required,email"`
    Role   string `json:"role" validate:"oneof=admin user guest"`
    Zip    string `json:"zip"  validate:"regex=^[0-9]{5}$"`
    Age    int    `json:"age"   validate:"min=18,max=130"`
}
```

> binding + validate:

```go
app.OnPost("/users", func(ctx *zentrox.Context) {
    var in UserInput
    if err := ctx.BindInto(&in); err != nil {
        ctx.SendJSON(400, map[string]string{"error": err.Error()})
        return
    }
    ctx.SendJSON(200, in)
})
```

**Note:**

-   `email` uses a basic regex that is suitable for most APIs; if you need to follow a complex RFC standard, you can replace it with a custom validator.
-   `oneof` supports strings, numbers, and bools. For numbers, the value in `oneof=` will be parsed to the correct type before comparing.
-   `regex` compiles according to the pattern you pass in (if the pattern is invalid, an error will be reported).

---

## Error handling

-   Add `middleware.Recovery()` globally to convert panics to `500` JSON:
    ```json
    { "error": "internal server error" }
    ```
-   For application errors, return your own payload with `ctx.SendJSON(...)`.

---

## Compatibility

-   `App` implements `http.Handler`, so you can mount it anywhere:
    ```go
    http.ListenAndServe(":8000", app)
    ```
-   Works with Go’s standard `net/http` tooling, `httptest`, reverse proxies, etc.

---

## Performance optimizations

-   Context pooling via `sync.Pool` to minimize allocations
-   Precompiled route tree for fast matching
-   Avoid repeated string splits by pre-tokenizing route patterns

---

## Observability

-   Request ID middleware (`X-Request-ID`)
-   Access log middleware with structured output
-   Lightweight tracing middleware (custom, not OpenTelemetry) to measure latency per route

---

## Testing

Minimal example using `httptest`:

```go
package z_test

import (
    "net/http/httptest"
    "testing"

    "github.com/aminofox/zentrox"
)

func TestBasic(t *testing.T) {
    app := zentrox.NewApp()
    app.OnGet("/hi", func(ctx *zentrox.Context) { ctx.SendText(200, "hi") })

    req := httptest.NewRequest("GET", "/hi", nil)
    w := httptest.NewRecorder()

    app.ServeHTTP(w, req)

    if w.Code != 200 || w.Body.String() != "hi" {
        t.Fatalf("unexpected: %d %q", w.Code, w.Body.String())
    }
}
```

Run tests:

```bash
go test ./...
```
---

## Performance Benchmarks

> Machine: **Apple M1 Pro** · `darwin/arm64` · Go toolchain (Xcode SDK)  
> Note: results depend on machine/OS/Go version; please run benchmarks on your own environment for accurate numbers.

### How to run

```bash
# Run ALL benchmarks (skip unit tests)
go test ./z_test -run=^$ -bench=. -benchmem -benchtime=5s

# Measure more clearly per-core (optional)
GOMAXPROCS=1 go test ./z_test -run=^$ -bench=. -benchmem -benchtime=5s

# Try multiple CPU settings (optional)
go test ./z_test -run=^$ -bench=. -benchmem -cpu=1,4,8 -benchtime=3s
```

---

### 📊 Results (sample on Apple M1 Pro)

#### Microbench (router & gzip internals)

| Benchmark             | ns/op   | B/op   | allocs/op |
|-----------------------|--------:|-------:|----------:|
| **Router_Static**     | 168.5   | 60     | 3         |
| **Router_Param**      | 230.2   | 62     | 3         |
| **Gzip_BigJSON**      | 163,546 | 161,523| 7         |

- `Router_*`: only 3 allocations/op — very lean.  
- `Gzip_BigJSON`: compressing a large payload is expected to cost CPU/memory. (B/op is higher because `httptest.ResponseRecorder` keeps the entire compressed body in RAM; in production, data streams to the socket so this is lower.)

#### RPS benchmarks (end-to-end path, with reported `rps`)

| Benchmark                     | ns/op | rps        | B/op  | allocs/op |
|------------------------------|------:|-----------:|------:|----------:|
| **RPS_Static**               | 950.1 | 1,052,544  | 1,136 | 16        |
| **RPS_Param**                | 1100  | 909,131    | 1,152 | 17        |
| **RPS_SmallJSON**            | 1354  | 738,360    | 1,617 | 23        |
| **RPS_SmallJSON_Parallel**   | 1288  | 776,278    | 1,618 | 23        |

> Quick takeaways:
> - ~**1.05M rps** for static routes; ~**0.91M rps** for parameterized routes; ~**0.74–0.78M rps** for small JSON.  
> - `Parallel` nudges rps up by utilizing multiple CPUs, but the benefit is limited because each request is tiny (scheduler/alloc overhead becomes noticeable).

---

### How to read the metrics

- **ns/op**: average time for *one* benchmark operation (nanoseconds). Lower is better. Roughly the inverse of RPS.  
- **B/op**: bytes allocated on the heap per operation. Lower ⇒ less GC pressure.  
- **allocs/op**: number of heap allocations per operation. Lower is better.  
- **rps**: requests per second measured directly in the RPS benchmarks (more intuitive for overall throughput).

💡 Repro tips:
- Use `-benchtime=5s` (or higher) for stable results.  
- For gzip, add `b.SetBytes(size)` in the benchmark to have Go print MB/s.  
- If you aim to measure *stack CPU* only, use a “discard” writer (already used in the RPS benches) to avoid `ResponseRecorder` inflating B/op.

---

<details>
<summary><b>Raw output</b></summary>

```
goos: darwin
goarch: arm64
pkg: github.com/aminofox/zentrox/z_test
cpu: Apple M1 Pro
BenchmarkGzip_BigJSON-10                    7263            163546 ns/op          161523 B/op          7 allocs/op
BenchmarkRouter_Static-10                6943929               168.5 ns/op            60 B/op          3 allocs/op
BenchmarkRouter_Param-10                 5117144               230.2 ns/op            62 B/op          3 allocs/op
BenchmarkRPS_Static-10                   1247601               950.1 ns/op         1052544 rps      1136 B/op         16 allocs/op
BenchmarkRPS_Param-10                    1092390              1100 ns/op            909131 rps      1152 B/op         17 allocs/op
BenchmarkRPS_SmallJSON-10                 819433              1354 ns/op            738360 rps      1617 B/op         23 allocs/op
BenchmarkRPS_SmallJSON_Parallel-10        956000              1288 ns/op            776278 rps      1618 B/op         23 allocs/op
```
</details>


## License

MIT
