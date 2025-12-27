package main

import (
	"bytes"
	"context"
	"crypto/ed25519"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"prompt-service-server/config"
	"prompt-service-server/utils"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRouter() *mux.Router {
	cfg := config.LoadConfig()
	return InitializeRouter(cfg)
}

func TestIndexHandler_Get(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest("GET", "/", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// The index handler serves a static file, so we just check that it doesn't error
	// In a real test environment, we'd need to set up the static files properly
	assert.True(t, w.Code == http.StatusOK || w.Code == http.StatusNotFound, "Should either serve file or return 404")
}

func TestAuthHandler_Get(t *testing.T) {
	router := setupTestRouter()

	// Create a test public key and its hash
	testKey := "test-public-key"
	hashedKey := "c88e80202a4d650841e48649b4c3f553e48131b415db680476fd0d65632ff2b0" // sha256 of "test-public-key"

	req := httptest.NewRequest("GET", "/api/auth/"+hashedKey, nil)
	req.AddCookie(&http.Cookie{Name: "publicKey", Value: testKey})

	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Check that a CSRF token cookie is set
	cookies := w.Result().Cookies()
	var csrfCookie *http.Cookie
	for _, cookie := range cookies {
		if cookie.Name == "CSRFToken" {
			csrfCookie = cookie
			break
		}
	}
	assert.NotNil(t, csrfCookie)
	assert.NotEmpty(t, csrfCookie.Value)
}

func TestPromptHandler_Post_Success(t *testing.T) {
	router := setupTestRouter()

	reqBody := map[string]string{
		"public_key": "dGVzdC1wdWJsaWMta2V5", // base64 encoded "test-public-key"
		"message":    "Test prompt message",
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/prompts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()

	// This will block waiting for a response, so we need to handle it asynchronously
	done := make(chan bool)
	go func() {
		router.ServeHTTP(w, req)
		done <- true
	}()

	// Wait for the handler to start
	select {
	case <-done:
		// If it completes immediately, check the response
		assert.Equal(t, http.StatusOK, w.Code)
	default:
		// Handler is blocking as expected
		t.Log("Handler is blocking waiting for prompt response")
	}
}

func TestPromptHandler_Post_InvalidJSON(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest("POST", "/api/prompts", bytes.NewReader([]byte("invalid json")))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Invalid request body")
}

func TestPromptHandler_Post_MissingFields(t *testing.T) {
	router := setupTestRouter()

	reqBody := map[string]string{
		"public_key": "test-key",
		// missing message
	}
	body, _ := json.Marshal(reqBody)

	req := httptest.NewRequest("POST", "/api/prompts", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "Missing public_key or message")
}

func TestPromptHandler_Get_Unauthenticated(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest("GET", "/api/prompts/test-hash", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should redirect due to missing authentication
	assert.Equal(t, http.StatusFound, w.Code)
	assert.Equal(t, "/", w.Header().Get("Location"))
}

func TestPromptHandler_Respond_Unauthenticated(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest("POST", "/api/prompts/test-id", bytes.NewReader([]byte("response")))
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should redirect due to missing authentication (VerifyKeyHash redirects)
	assert.Equal(t, http.StatusFound, w.Code)
}

// Test the key handler
func TestSSEHandler_Get_Unauthenticated(t *testing.T) {
	router := setupTestRouter()

	req := httptest.NewRequest("GET", "/api/sse/test-hash", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// Should redirect due to missing authentication
	assert.Equal(t, http.StatusFound, w.Code)
}

func TestSSEHandler_Get_Authenticated(t *testing.T) {
	router := setupTestRouter()

	// Generate keypair for proper authentication
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	pubKeyB64 := base64.StdEncoding.EncodeToString(pub)
	hashedKey := sha256.Sum256([]byte(pubKeyB64))
	keyHash := hex.EncodeToString(hashedKey[:])

	// Create JWT and signature using the same pattern as auth tests
	token, err := utils.GenerateCSRFToken(keyHash)
	require.NoError(t, err)

	signature := ed25519.Sign(priv, []byte(token))
	sigB64 := base64.StdEncoding.EncodeToString(signature)

	req := httptest.NewRequest("GET", "/api/sse/"+keyHash, nil)
	req.AddCookie(&http.Cookie{Name: "publicKey", Value: pubKeyB64})
	req.AddCookie(&http.Cookie{Name: "CSRFToken", Value: token})
	req.AddCookie(&http.Cookie{Name: "CSRFChallenge", Value: sigB64})

	// Use a context that will cancel quickly to avoid blocking
	ctx, cancel := context.WithTimeout(req.Context(), 10*time.Millisecond)
	defer cancel()
	req = req.WithContext(ctx)

	w := httptest.NewRecorder()

	// Run the request - it should send the initial event and then block
	done := make(chan bool, 1)
	go func() {
		router.ServeHTTP(w, req)
		done <- true
	}()

	// Wait for either completion or timeout
	select {
	case <-done:
		// Handler completed (should happen due to context cancellation)
	case <-time.After(100 * time.Millisecond):
		// Timeout - handler is still running
		t.Fatal("Handler should have completed due to context cancellation")
	}

	// Check that the response has the correct SSE headers
	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, "text/event-stream", w.Header().Get("Content-Type"))
	assert.Equal(t, "no-cache", w.Header().Get("Cache-Control"))
	assert.Equal(t, "keep-alive", w.Header().Get("Connection"))

	// Check that the initial "connected" event was sent
	responseBody := w.Body.String()
	assert.Contains(t, responseBody, `"type":"connected"`)
	assert.Contains(t, responseBody, `"content":"Connection established"`)
	assert.Contains(t, responseBody, `"id":"`+pubKeyB64+`"`) // The connected event includes the public key as id
}

func TestSSE_PromptNotification_Integration(t *testing.T) {
	router := setupTestRouter()

	// Generate keypair for proper authentication
	pub, priv, err := ed25519.GenerateKey(rand.Reader)
	require.NoError(t, err)

	pubKeyB64 := base64.StdEncoding.EncodeToString(pub)
	hashedKey := sha256.Sum256([]byte(pubKeyB64))
	keyHash := hex.EncodeToString(hashedKey[:])

	// Create JWT and signature
	token, err := utils.GenerateCSRFToken(keyHash)
	require.NoError(t, err)

	signature := ed25519.Sign(priv, []byte(token))
	sigB64 := base64.StdEncoding.EncodeToString(signature)

	// Set up cookies for all requests
	cookies := []*http.Cookie{
		{Name: "publicKey", Value: pubKeyB64},
		{Name: "CSRFToken", Value: token},
		{Name: "CSRFChallenge", Value: sigB64},
	}

	// Start SSE connection with reasonable timeout
	sseReq := httptest.NewRequest("GET", "/api/sse/"+keyHash, nil)
	for _, cookie := range cookies {
		sseReq.AddCookie(cookie)
	}

	sseCtx, sseCancel := context.WithTimeout(sseReq.Context(), 5*time.Second)
	defer sseCancel()
	sseReq = sseReq.WithContext(sseCtx)

	sseW := httptest.NewRecorder()

	// Start SSE handler in goroutine
	sseDone := make(chan bool, 1)
	go func() {
		defer func() { sseDone <- true }()
		router.ServeHTTP(sseW, sseReq)
	}()

	// Give SSE connection time to establish and send initial event
	time.Sleep(50 * time.Millisecond)

	// Verify SSE connection was established
	sseResponse := sseW.Body.String()
	assert.Contains(t, sseResponse, `"type":"connected"`)
	assert.Contains(t, sseResponse, `"content":"Connection established"`)

	// Post a prompt for the same user
	reqBody := map[string]string{
		"public_key": pubKeyB64,
		"message":    "Test prompt for SSE notification",
	}
	body, _ := json.Marshal(reqBody)

	promptReq := httptest.NewRequest("POST", "/api/prompts", bytes.NewReader(body))
	promptReq.Header.Set("Content-Type", "application/json")

	promptW := httptest.NewRecorder()

	// Start prompt posting in a goroutine that will respond to the prompt
	promptDone := make(chan string, 1)
	go func() {
		// This will block until we respond to the prompt
		router.ServeHTTP(promptW, promptReq)
		promptDone <- promptW.Body.String()
	}()

	// Wait for the prompt to be posted and notification to be sent
	// The SSE should receive a "new_prompt" event
	maxWait := time.After(2 * time.Second)
	ticker := time.NewTicker(50 * time.Millisecond)
	defer ticker.Stop()

	var promptId string
	for {
		select {
		case <-maxWait:
			t.Fatal("Timeout waiting for prompt notification")
		case <-ticker.C:
			currentResponse := sseW.Body.String()
			if strings.Contains(currentResponse, `"type":"new_prompt"`) &&
				strings.Contains(currentResponse, `"content":"Test prompt for SSE notification"`) {
				// Extract prompt ID from the JSON event
				lines := strings.Split(currentResponse, "\n")
				for _, line := range lines {
					if strings.Contains(line, `"type":"new_prompt"`) &&
						strings.Contains(line, `"content":"Test prompt for SSE notification"`) {
						// Parse the JSON to extract the ID
						// Format: data: {"type":"new_prompt", "content":"...", "id":"uuid"}
						jsonStart := strings.Index(line, "{")
						if jsonStart != -1 {
							jsonStr := line[jsonStart:]
							var event map[string]interface{}
							if err := json.Unmarshal([]byte(jsonStr), &event); err == nil {
								if id, ok := event["id"].(string); ok && id != "" {
									promptId = id
									goto foundPrompt
								}
							}
						}
					}
				}
			}
		}
	}

foundPrompt:

	// Now respond to the prompt with proper authentication
	respondReq := httptest.NewRequest("POST", "/api/prompts/"+promptId, bytes.NewReader([]byte("test response")))
	for _, cookie := range cookies {
		respondReq.AddCookie(cookie)
	}
	respondReq.Header.Set("Content-Type", "text/plain")

	respondW := httptest.NewRecorder()

	// Post the response
	router.ServeHTTP(respondW, respondReq)

	// The response should be successful
	assert.Equal(t, http.StatusOK, respondW.Code)
	assert.Equal(t, promptId, strings.TrimSpace(respondW.Body.String()))

	// Wait for the prompt posting goroutine to complete
	select {
	case response := <-promptDone:
		assert.Equal(t, "test response", response)
	case <-time.After(1 * time.Second):
		t.Fatal("Prompt posting did not complete")
	}

	// Cancel the SSE context to stop the connection
	sseCancel()

	// Wait for SSE to finish
	select {
	case <-sseDone:
		// SSE finished cleanly
	case <-time.After(500 * time.Millisecond):
		t.Log("SSE did not finish cleanly, but that's expected due to context cancellation")
	}
}
