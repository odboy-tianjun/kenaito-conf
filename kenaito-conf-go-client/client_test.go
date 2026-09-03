package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestFetchChangedSendsToken(t *testing.T) {
	var received map[string]interface{}

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/fetchChangedConfigs" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&received); err != nil {
			t.Errorf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{}}`))
	}))
	defer srv.Close()

	c := New(&Config{
		BaseURL: srv.URL,
		AppName: "demo-app",
		Env:     "daily",
		Token:   "test-token",
	})
	c.lastTime = 1699999999999

	if err := c.fetchChanged(); err != nil {
		t.Fatalf("fetchChanged() error = %v", err)
	}

	if got := received["token"]; got != "test-token" {
		t.Fatalf("token = %v, want %q", got, "test-token")
	}
	if got := received["lastTime"]; got != float64(1699999999999) {
		t.Fatalf("lastTime = %v, want %v", got, 1699999999999)
	}
}
