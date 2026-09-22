package spine_test

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/azarov-serge/spine"
)

func TestPingAndJSONError(t *testing.T) {
	app := spine.New(spine.Config{ServiceName: "test"})
	app.GET("/ping", func(c *spine.Context) error {
		return c.Text(http.StatusOK, "pong")
	})
	app.GET("/missing", func(c *spine.Context) error {
		return spine.NotFound("item not found")
	})

	t.Run("ping", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/ping", nil)
		app.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want 200", rec.Code)
		}
		if got := rec.Body.String(); got != "pong" {
			t.Fatalf("body = %q, want %q", got, "pong")
		}
	})

	t.Run("not found", func(t *testing.T) {
		rec := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/missing", nil)
		app.Handler().ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("status = %d, want 404", rec.Code)
		}

		var body map[string]string
		if err := json.Unmarshal(rec.Body.Bytes(), &body); err != nil {
			t.Fatalf("json: %v", err)
		}
		if body["code"] != "not_found" {
			t.Fatalf("code = %q, want not_found", body["code"])
		}
	})
}
