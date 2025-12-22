package handlers

import (
	"encoding/json"
	"net/http"
	"prompt-service-server/core"
	"prompt-service-server/utils"

	"github.com/gorilla/mux"
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

	signal := utils.NewSignal()
	h.store.AddPrompt(
		req.PublicKey,
		req.Message,
		func(response string) {
			signal.Signal(response)
		},
	)
	response := signal.Wait()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func (h *PromptHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyHash := vars["id"]
	// Authenticate and verify CSRF for this request
	if _, err := AuthenticateAndVerifyCSRF(w, r, keyHash); err != nil {
		// Error response already written by helper
		return
	}

	// Validate signature against public key
	// This would involve JWT verification

	// Return list of prompts
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(h.store.GetPrompts())
}
