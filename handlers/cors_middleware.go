package handlers

import (
	"net/http"
	"prompt-service-server/config"
	"strings"
)

// CORSMiddleware handles CORS for different endpoints
type CORSMiddleware struct {
	cfg *config.Config
}

func NewCORSMiddleware(cfg *config.Config) *CORSMiddleware {
	return &CORSMiddleware{cfg: cfg}
}

// Handler wraps an http.Handler with CORS logic
func (m *CORSMiddleware) Handler(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")

		// Check if this is the POST /api/prompts endpoint (unrestricted CORS)
		if r.URL.Path == "/api/prompts" && (r.Method == "POST" || r.Method == "OPTIONS") {
			// Allow any origin for POST /api/prompts
			if origin != "" {
				w.Header().Set("Access-Control-Allow-Origin", origin)
				w.Header().Set("Access-Control-Allow-Credentials", "false")
			}
			w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

			// Handle preflight OPTIONS request
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusNoContent)
				return
			}

			next.ServeHTTP(w, r)
			return
		}

		// For all other endpoints, restrict to allowed origins if configured
		if origin != "" && m.cfg.AllowedOrigins != "" {
			allowedOrigins := strings.Split(m.cfg.AllowedOrigins, ",")
			allowed := false
			for _, allowedOrigin := range allowedOrigins {
				trimmedOrigin := strings.TrimSpace(allowedOrigin)
				if trimmedOrigin == origin || trimmedOrigin == "*" {
					allowed = true
					w.Header().Set("Access-Control-Allow-Origin", origin)
					break
				}
			}

			if allowed {
				w.Header().Set("Access-Control-Allow-Credentials", "true")
				w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
				w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Cookie")

				// Handle preflight OPTIONS request
				if r.Method == "OPTIONS" {
					w.WriteHeader(http.StatusNoContent)
					return
				}
			}
		}

		next.ServeHTTP(w, r)
	})
}
