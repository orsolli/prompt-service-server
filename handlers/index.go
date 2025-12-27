package handlers

import (
	"embed"
	"fmt"
	"net/http"
	"strings"
)

type CSPConfig struct {
	AllowedScriptSources  []string
	AllowedStyleSources   []string
	AllowedConnectSources []string
	ScriptHashes          []string
}

func buildCSP(config CSPConfig) string {
	directives := []string{
		"default-src 'self'",
		fmt.Sprintf("script-src 'self' %s %s",
			strings.Join(config.AllowedScriptSources, " "),
			strings.Join(config.ScriptHashes, " ")),
		fmt.Sprintf("connect-src 'self' %s", strings.Join(config.AllowedConnectSources, " ")),
		fmt.Sprintf("style-src 'self' %s", strings.Join(config.AllowedStyleSources, " ")),
		"img-src 'self'",
		"object-src 'none'",
		"base-uri 'self'",
		"form-action 'self'",
	}
	return strings.Join(directives, "; ")
}

func getCSPConfig() CSPConfig {
	return CSPConfig{
		// External script sources (CDNs, etc.)
		AllowedScriptSources: []string{
			"https://esm.sh/",
		},

		// External style sources (CDNs, etc.)
		AllowedStyleSources: []string{"https://cdn.jsdelivr.net/"},

		// External connection sources (for fetch, WebSocket, etc.)
		AllowedConnectSources: []string{
			"https://esm.sh/",
		},

		// SHA256 hashes of allowed inline scripts
		ScriptHashes: []string{
			"'sha256-SCUHckId5Z5Nvwwoo92FGEYK/9vveGoCwh062sCwXY8='", // Hash of the json object `importMap` in importmap.js
		},
	}
}

type IndexHandler struct {
	staticFiles embed.FS
}

func NewIndexHandler(staticFiles embed.FS) *IndexHandler {
	return &IndexHandler{
		staticFiles: staticFiles,
	}
}

func (h *IndexHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Set security headers
	w.Header().Set("Content-Security-Policy", buildCSP(getCSPConfig()))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")

	// Serve index.html with embedded JavaScript
	p, err := h.staticFiles.ReadFile("static/index.html")
	if err != nil {
		http.Error(w, "Failed to read index.html", http.StatusInternalServerError)
		return
	}
	if _, writeErr := w.Write(p); writeErr != nil {
		// Optionally log the error or handle it as needed
		http.Error(w, "Failed to write response", http.StatusInternalServerError)
	}
}
