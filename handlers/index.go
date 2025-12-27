package handlers

import (
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

type IndexHandler struct{}

func NewIndexHandler() *IndexHandler {
	return &IndexHandler{}
}

func (h *IndexHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Set security headers
	w.Header().Set("Content-Security-Policy", buildCSP(getCSPConfig()))
	w.Header().Set("X-Content-Type-Options", "nosniff")
	w.Header().Set("X-Frame-Options", "DENY")

	// Check if user has keys in localStorage
	// This is a client-side check, but we can set up the redirect logic here
	// For now, we'll just serve the index.html
	http.ServeFile(w, r, "static/index.html")
}
