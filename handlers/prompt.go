package handlers

import (
	"encoding/json"
	"net/http"
	"prompt-service-server/core"
)

type PromptHandler struct {
	store *core.PromptStore
}

func NewPromptHandler(store *core.PromptStore) *PromptHandler {
	return &PromptHandler{
		store: store,
	}
}

func (h *PromptHandler) Post(w http.ResponseWriter, r *http.Request) {
	// Parse request body
	var req struct {
		PublicKey string `json:"public_key"`
		Message   string `json:"message"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.PublicKey == "" || req.Message == "" {
		http.Error(w, "Missing public_key or message", http.StatusBadRequest)
		return
	}

	// Verify signature (this is a simplified version)
	// In reality, you'd need to verify the signature against the public key
	signature := r.Header.Get("Authorization")
	if signature == "" {
		http.Error(w, "Missing signature", http.StatusUnauthorized)
		return
	}

	// Process the prompt (store in memory for now)
	// In a real implementation, this would store the prompt and wait for response

	// Simulate sending messages
	h.store.AddPrompt(
		req.PublicKey,
		req.Message,
		func(response string) {
			w.WriteHeader(http.StatusOK)
			w.Write([]byte(response))
		},
	)
}

func (h *PromptHandler) Get(w http.ResponseWriter, r *http.Request) {
	// Verify signature
	signature := r.Header.Get("Authorization")
	if signature == "" {
		http.Error(w, "Missing signature", http.StatusUnauthorized)
		return
	}

	// Validate signature against public key
	// This would involve JWT verification

	// Return list of prompts
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(h.store.GetPrompts())
}
