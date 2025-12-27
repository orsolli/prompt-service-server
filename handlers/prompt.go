package handlers

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"prompt-service-server/config"
	"prompt-service-server/core"
	"prompt-service-server/utils"

	"github.com/gorilla/mux"
)

var cfg = config.LoadConfig()

type PromptHandler struct {
	store *core.PromptStore
}

func NewPromptHandler(store *core.PromptStore) *PromptHandler {
	return &PromptHandler{
		store: store,
	}
}

func (h *PromptHandler) Post(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > cfg.MaxRequestBodySize {
		http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}

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

	// Validate PublicKey is valid base64
	if _, err := base64.StdEncoding.DecodeString(req.PublicKey); err != nil {
		http.Error(w, "Invalid public_key format", http.StatusBadRequest)
		return
	}

	signal := utils.NewSignal()

	defer h.store.RemovePrompt(h.store.AddPrompt(
		req.PublicKey,
		req.Message,
		func(response string) {
			signal.Signal(response)
		},
	))
	response := signal.Wait()
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(response))
}

func (h *PromptHandler) Get(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	keyHash := vars["id"]
	// Authenticate and verify CSRF for this request
	key, err := AuthenticateAndVerifyCSRF(w, r, keyHash)
	if err != nil {
		// Error response already written by helper
		return
	}

	// Validate signature against public key
	// This would involve JWT verification

	// Return list of prompts
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(h.store.GetPrompts(key, ""))
}

func (h *PromptHandler) Respond(w http.ResponseWriter, r *http.Request) {
	if r.ContentLength > cfg.MaxRequestBodySize {
		http.Error(w, "Request body too large", http.StatusRequestEntityTooLarge)
		return
	}

	vars := mux.Vars(r)
	id := vars["id"]
	var keyHash string
	var prompt *core.Prompt
	for _, p := range h.store.GetPrompts("", id) {
		if p.Id == id {
			prompt = p
			hashedKey := sha256.Sum256([]byte(prompt.Key))
			keyHash = hex.EncodeToString(hashedKey[:])
			break
		}
	}

	// Authenticate and verify CSRF for this request
	if _, err := AuthenticateAndVerifyCSRF(w, r, keyHash); err != nil {
		// Error response already written by helper
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	response := make([]byte, r.ContentLength)
	r.Body.Read(response)
	prompt.Callback(string(response))
	h.store.SendEventToConnections(prompt.Key, "prompt_responded", string(response), prompt.Id)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(prompt.Id))
}
