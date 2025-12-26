package handlers

import (
	"net/http"
	"prompt-service-server/core"
	"time"

	"github.com/gorilla/mux"
)

type SSEHandler struct {
	store *core.PromptStore
}

func NewSSEHandler(store *core.PromptStore) *SSEHandler {
	return &SSEHandler{store: store}
}

func (h *SSEHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyHash := vars["id"]

	// Authenticate and verify CSRF for this request
	if _, err := AuthenticateAndVerifyCSRF(w, r, keyHash); err != nil {
		// Error response already written by helper
		return
	}
	// cookieKey is the base64-encoded public key string (as in the cookie)
	cookie, _ := r.Cookie("publicKey")
	cookieKey := ""
	if cookie != nil {
		cookieKey = cookie.Value
	}

	// Validate signature against public key

	// Set SSE headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("Access-Control-Allow-Origin", "*")

	// Create connection object
	flusher := w.(http.Flusher)

	// Add connection to store (this would track connections by keyHash)
	connection := h.store.AddSSEConnection(cookieKey, w, flusher)

	// Send initial connection confirmation
	h.store.SendEvent(w, flusher, "connected", "Connection established", cookieKey)

	// Keep connection alive
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send heartbeat
			h.store.SendEvent(w, flusher, "heartbeat", "alive", cookieKey)
		case <-r.Context().Done():
			// Remove connection
			h.store.RemoveSSEConnection(cookieKey, connection)
			return
		}
	}
}
