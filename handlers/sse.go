package handlers

import (
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
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

	// Check if cookie contains matching public key
	cookie, err := r.Cookie("publicKey")
	if err != nil {
		// Redirect to root
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Verify the cookie's public key matches the hash
	cookieKey := cookie.Value
	hashedKey := sha256.Sum256([]byte(cookieKey))
	if hex.EncodeToString(hashedKey[:]) != keyHash {
		// Redirect to root
		http.Redirect(w, r, "/", http.StatusFound)
		return
	}

	// Verify signature (similar to prompt handlers)
	signature, err := r.Cookie("CSRFChallenge")
	if err != nil {
		http.Error(w, "Missing signature", http.StatusUnauthorized)
		return
	}

	// Verify token (similar to prompt handlers)
	token, err := r.Cookie("CSRFToken")
	if err != nil {
		http.Error(w, "Missing token", http.StatusUnauthorized)
		return
	}

	publicKey, err := base64.StdEncoding.DecodeString(cookieKey)
	if err != nil {
		http.Error(w, "Failed to decode", http.StatusUnauthorized)
		return
	}

	signatureBytes, err := base64.StdEncoding.DecodeString(signature.Value)
	if err != nil {
		http.Error(w, "Failed to decode", http.StatusUnauthorized)
		return
	}

	valid := ed25519.Verify(publicKey, []byte(token.Value), signatureBytes)
	if valid == false {
		http.Error(w, "Invalid signature", http.StatusUnauthorized)
		return
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
	h.sendEvent(w, flusher, "connected", "Connection established")

	// Keep connection alive
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Send heartbeat
			h.sendEvent(w, flusher, "heartbeat", "alive")
		case <-r.Context().Done():
			// Remove connection
			h.store.RemoveSSEConnection(cookieKey, connection)
			return
		}
	}
}

func (h *SSEHandler) sendEvent(w http.ResponseWriter, flusher http.Flusher, eventType, data string) {
	event := fmt.Sprintf("event: %s\ndata: %s\n\n", eventType, data)
	w.Write([]byte(event))
	flusher.Flush()
}
