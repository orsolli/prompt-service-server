package handlers

import (
	"net/http"
	"net/http/httptest"
	"prompt-service-server/config"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func TestCORSMiddleware_PostPromptsUnrestricted(t *testing.T) {
	cfg := &config.Config{
		AllowedOrigins: "", // Not configured
	}

	corsMiddleware := NewCORSMiddleware(cfg)

	// Create a test handler
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Create router
	r := mux.NewRouter()
	r.Use(corsMiddleware.Handler)
	r.HandleFunc("/api/prompts", handler).Methods("POST")

	// Test with an origin header
	req := httptest.NewRequest("POST", "/api/prompts", strings.NewReader("{}"))
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should allow the origin
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "false", w.Header().Get("Access-Control-Allow-Credentials"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
}

func TestCORSMiddleware_PostPromptsPreflightOptions(t *testing.T) {
	cfg := &config.Config{
		AllowedOrigins: "",
	}

	corsMiddleware := NewCORSMiddleware(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := mux.NewRouter()
	r.Use(corsMiddleware.Handler)
	r.HandleFunc("/api/prompts", handler).Methods("POST", "OPTIONS")

	// Test preflight OPTIONS request
	req := httptest.NewRequest("OPTIONS", "/api/prompts", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should return NoContent and set CORS headers
	assert.Equal(t, http.StatusNoContent, w.Code)
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Contains(t, w.Header().Get("Access-Control-Allow-Methods"), "POST")
}

func TestCORSMiddleware_OtherEndpointsNoOriginConfig(t *testing.T) {
	cfg := &config.Config{
		AllowedOrigins: "", // Not configured
	}

	corsMiddleware := NewCORSMiddleware(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := mux.NewRouter()
	r.Use(corsMiddleware.Handler)
	r.HandleFunc("/api/prompts/{id}", handler).Methods("GET")

	// Test with an origin header but no AllowedOrigins configured
	req := httptest.NewRequest("GET", "/api/prompts/123", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should not set CORS headers
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_OtherEndpointsWithAllowedOrigin(t *testing.T) {
	cfg := &config.Config{
		AllowedOrigins: "https://example.com,https://app.example.com",
	}

	corsMiddleware := NewCORSMiddleware(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := mux.NewRouter()
	r.Use(corsMiddleware.Handler)
	r.HandleFunc("/api/prompts/{id}", handler).Methods("GET")

	// Test with allowed origin
	req := httptest.NewRequest("GET", "/api/prompts/123", nil)
	req.Header.Set("Origin", "https://example.com")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should set CORS headers
	assert.Equal(t, "https://example.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddleware_OtherEndpointsWithDisallowedOrigin(t *testing.T) {
	cfg := &config.Config{
		AllowedOrigins: "https://example.com",
	}

	corsMiddleware := NewCORSMiddleware(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := mux.NewRouter()
	r.Use(corsMiddleware.Handler)
	r.HandleFunc("/api/prompts/{id}", handler).Methods("GET")

	// Test with disallowed origin
	req := httptest.NewRequest("GET", "/api/prompts/123", nil)
	req.Header.Set("Origin", "https://malicious.com")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should not set CORS headers
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}

func TestCORSMiddleware_WildcardAllowedOrigins(t *testing.T) {
	cfg := &config.Config{
		AllowedOrigins: "*",
	}

	corsMiddleware := NewCORSMiddleware(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := mux.NewRouter()
	r.Use(corsMiddleware.Handler)
	r.HandleFunc("/api/sse/{id}", handler).Methods("GET")

	// Test with any origin
	req := httptest.NewRequest("GET", "/api/sse/123", nil)
	req.Header.Set("Origin", "https://any-origin.com")
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should allow any origin
	assert.Equal(t, "https://any-origin.com", w.Header().Get("Access-Control-Allow-Origin"))
	assert.Equal(t, "true", w.Header().Get("Access-Control-Allow-Credentials"))
}

func TestCORSMiddleware_NoOriginHeader(t *testing.T) {
	cfg := &config.Config{
		AllowedOrigins: "https://example.com",
	}

	corsMiddleware := NewCORSMiddleware(cfg)

	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	r := mux.NewRouter()
	r.Use(corsMiddleware.Handler)
	r.HandleFunc("/api/prompts", handler).Methods("POST")

	// Test without Origin header
	req := httptest.NewRequest("POST", "/api/prompts", strings.NewReader("{}"))
	w := httptest.NewRecorder()

	r.ServeHTTP(w, req)

	// Should not set CORS headers when no origin is present
	assert.Empty(t, w.Header().Get("Access-Control-Allow-Origin"))
}
