package z_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/aminofox/zentrox"
	"github.com/aminofox/zentrox/middleware"
)

func TestTimeout_Triggers504(t *testing.T) {
	app := zentrox.NewApp()
	app.Plug(middleware.Timeout(middleware.TimeoutOptions{Duration: 50 * time.Millisecond}))

	app.OnGet("/slow", func(c *zentrox.Context) {
		time.Sleep(150 * time.Millisecond)
		c.SendText(http.StatusOK, "done")
	})

	req := httptest.NewRequest(http.MethodGet, "/slow", nil)
	w := httptest.NewRecorder()
	app.ServeHTTP(w, req)

	if w.Code != http.StatusGatewayTimeout && w.Code != 504 {
		t.Fatalf("want 504, got %d", w.Code)
	}
}
