package handler

import (
	"net/http"
	"testing"
)

func TestHealthCheck(t *testing.T) {
	w := doRequest(http.MethodGet, "/api/health", nil)
	assertStatus(t, w, http.StatusOK)

	var body map[string]string
	decodeJSON(t, w, &body)

	if body["status"] != "ok" {
		t.Errorf("expected status = \"ok\", got %q", body["status"])
	}
	if body["database"] != "ok" {
		t.Errorf("expected database = \"ok\", got %q", body["database"])
	}
}
