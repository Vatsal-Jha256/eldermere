package httpapi

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/config"
	"github.com/Vatsal-Jha256/eldermere/apps/server/internal/storage"
)

func TestHealthz(t *testing.T) {
	router := NewRouter(config.Config{AppEnv: "test"}, slog.Default(), storage.NewMemoryStore())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}

	var body healthResponse
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if body.Status != "ok" {
		t.Fatalf("expected ok status, got %q", body.Status)
	}
	if body.App != "eldermere" {
		t.Fatalf("expected app eldermere, got %q", body.App)
	}
}

func TestCreateSession(t *testing.T) {
	store := storage.NewMemoryStore()
	router := NewRouter(config.Config{AppEnv: "test"}, slog.Default(), store)

	req := httptest.NewRequest(http.MethodPost, "/api/v1/sessions", bytes.NewBufferString(`{"display_name":"Tester"}`))
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusCreated {
		t.Fatalf("expected status 201, got %d", res.Code)
	}

	var body storage.PlayerSession
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if body.PlayerID == "" || body.Token == "" {
		t.Fatalf("expected player id and token, got %#v", body)
	}

	ok, err := store.VerifyPlayerSession(req.Context(), body.PlayerID, body.Token)
	if err != nil {
		t.Fatalf("verify session: %v", err)
	}
	if !ok {
		t.Fatal("expected session token to verify")
	}
}

func TestCORSPreflight(t *testing.T) {
	router := NewRouter(config.Config{AppEnv: "test"}, slog.Default(), storage.NewMemoryStore())

	req := httptest.NewRequest(http.MethodOptions, "/api/v1/sessions", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusNoContent {
		t.Fatalf("expected status 204, got %d", res.Code)
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("allow origin = %q, want *", got)
	}
	if got := res.Header().Get("Access-Control-Allow-Methods"); got != "GET, POST, OPTIONS" {
		t.Fatalf("allow methods = %q", got)
	}
	if got := res.Header().Get("Access-Control-Allow-Headers"); got != "Content-Type, Authorization" {
		t.Fatalf("allow headers = %q", got)
	}
}

func TestCORSHeadersOnAPIResponse(t *testing.T) {
	router := NewRouter(config.Config{AppEnv: "test"}, slog.Default(), storage.NewMemoryStore())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if res.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", res.Code)
	}
	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "*" {
		t.Fatalf("allow origin = %q, want *", got)
	}
}

func TestCORSAllowsConfiguredOrigin(t *testing.T) {
	router := NewRouter(config.Config{
		AppEnv:         "test",
		AllowedOrigins: []string{"https://eldermere.pages.dev"},
	}, slog.Default(), storage.NewMemoryStore())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "https://eldermere.pages.dev")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "https://eldermere.pages.dev" {
		t.Fatalf("allow origin = %q", got)
	}
}

func TestCORSRejectsUnconfiguredOrigin(t *testing.T) {
	router := NewRouter(config.Config{
		AppEnv:         "test",
		AllowedOrigins: []string{"https://eldermere.pages.dev"},
	}, slog.Default(), storage.NewMemoryStore())

	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)
	req.Header.Set("Origin", "https://example.invalid")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	if got := res.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("allow origin = %q, want empty", got)
	}
}
