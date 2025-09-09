package z_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/aminofox/zentrox"
	"github.com/aminofox/zentrox/middleware"
)

func newApp() *zentrox.App {
	app := zentrox.NewApp()
	app.Plug(middleware.ErrorHandler(middleware.DefaultErrorHandler()))
	return app
}

func TestRouter_Static(t *testing.T) {
	app := newApp()
	app.OnGet("/hi", func(c *zentrox.Context) {
		c.SendText(http.StatusOK, "hello")
	})

	req := httptest.NewRequest(http.MethodGet, "/hi", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if got := w.Body.String(); got != "hello" {
		t.Fatalf("want body %q, got %q", "hello", got)
	}
}

func TestRouter_ParamsAndWildcard(t *testing.T) {
	app := newApp()
	app.OnGet("/users/:id/files/*path", func(c *zentrox.Context) {
		id := c.Param("id")
		path := c.Param("path")
		c.SendJSON(http.StatusOK, map[string]string{"id": id, "path": path})
	})

	req := httptest.NewRequest(http.MethodGet, "/users/42/files/a/b/c.txt", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("want 200, got %d", w.Code)
	}

	want := `{"id":"42","path":"a/b/c.txt"}` + "\n" // SendJSON usually uses json.NewEncoder(...).Encode(v) so it ends with a newline \n. So the returned body will have \n at the end.
	got := w.Body.String()
	if got != want {
		t.Fatalf("want %s, got %s", want, got)
	}
}

func TestRouter_404And405(t *testing.T) {
	app := newApp()
	app.OnGet("/onlyget", func(c *zentrox.Context) { c.SendText(200, "ok") })

	// 404
	req := httptest.NewRequest(http.MethodGet, "/notfound", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusNotFound {
		t.Fatalf("want 404, got %d", w.Code)
	}

	// 405 (path exists but method not allowed)
	req = httptest.NewRequest(http.MethodPost, "/onlyget", nil)
	w = httptest.NewRecorder()
	app.ServeHTTP(w, req)
	if w.Code != http.StatusMethodNotAllowed {
		t.Fatalf("want 405, got %d", w.Code)
	}
}

func TestRouter_AutoHEAD(t *testing.T) {
	app := newApp()
	app.OnGet("/page", func(c *zentrox.Context) {
		c.SendText(http.StatusOK, "body")
	})

	req := httptest.NewRequest(http.MethodHead, "/page", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != 200 {
		t.Fatalf("want 200, got %d", w.Code)
	}
	if l := w.Body.Len(); l != 0 {
		t.Fatalf("HEAD should have empty body, got %d bytes", l)
	}
}
