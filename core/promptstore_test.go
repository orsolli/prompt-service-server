package core

import (
	"net/http"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewPromptStore(t *testing.T) {
	store := NewPromptStore()
	assert.NotNil(t, store)
	assert.NotNil(t, store.prompts)
	assert.NotNil(t, store.connections)
}

func TestAddPrompt(t *testing.T) {
	store := NewPromptStore()
	key := "test-key"
	message := "test message"
	called := false
	callback := func(response string) {
		called = true
		assert.Equal(t, "test response", response)
	}

	id := store.AddPrompt(key, message, callback)
	assert.NotEmpty(t, id)

	// Verify prompt was added
	prompts := store.GetPrompts(key, "")
	require.Len(t, prompts, 1)
	assert.Equal(t, id, prompts[0].Id)
	assert.Equal(t, key, prompts[0].Key)
	assert.Equal(t, message, prompts[0].Message)
	assert.NotNil(t, prompts[0].Callback)

	// Test callback
	prompts[0].Callback("test response")
	assert.True(t, called)
}

func TestGetPrompts(t *testing.T) {
	store := NewPromptStore()

	// Add prompts
	id1 := store.AddPrompt("key1", "message1", func(string) {})
	id2 := store.AddPrompt("key1", "message2", func(string) {})
	id3 := store.AddPrompt("key2", "message3", func(string) {})

	// Get prompts for key1
	prompts := store.GetPrompts("key1", "")
	assert.Len(t, prompts, 2)
	promptIds := []string{prompts[0].Id, prompts[1].Id}
	assert.Contains(t, promptIds, id1)
	assert.Contains(t, promptIds, id2)

	// Get prompts for key2
	prompts = store.GetPrompts("key2", "")
	assert.Len(t, prompts, 1)
	assert.Equal(t, id3, prompts[0].Id)

	// Get specific prompt by ID
	prompts = store.GetPrompts("", id1)
	assert.Len(t, prompts, 1)
	assert.Equal(t, id1, prompts[0].Id)
}

func TestRemovePrompt(t *testing.T) {
	store := NewPromptStore()

	id := store.AddPrompt("key", "message", func(string) {})

	// Verify prompt exists
	prompts := store.GetPrompts("key", "")
	assert.Len(t, prompts, 1)

	// Remove prompt
	store.RemovePrompt(id)

	// Verify prompt is gone
	prompts = store.GetPrompts("key", "")
	assert.Len(t, prompts, 0)
}

// MockResponseWriter is a mock http.ResponseWriter for testing
type MockResponseWriter struct {
	data []byte
}

func (m *MockResponseWriter) Header() http.Header {
	return make(http.Header)
}

func (m *MockResponseWriter) Write(data []byte) (int, error) {
	m.data = append(m.data, data...)
	return len(data), nil
}

func (m *MockResponseWriter) WriteHeader(statusCode int) {
	// No-op for testing
}

// MockFlusher is a mock http.Flusher for testing
type MockFlusher struct{}

func (m *MockFlusher) Flush() {
	// No-op for testing
}

func TestAddSSEConnection(t *testing.T) {
	store := NewPromptStore()
	key := "test-key"

	w := &MockResponseWriter{}
	flusher := &MockFlusher{}

	conn := store.AddSSEConnection(key, w, flusher)
	assert.NotNil(t, conn)
	assert.Equal(t, w, conn.writer)
	assert.Equal(t, flusher, conn.flusher)
	assert.Equal(t, key, conn.key)

	// Verify connection was added
	assert.Len(t, store.connections[key], 1)
	assert.Equal(t, conn, store.connections[key][0])
}

func TestRemoveSSEConnection(t *testing.T) {
	store := NewPromptStore()
	key := "test-key"

	w := &MockResponseWriter{}
	flusher := &MockFlusher{}

	conn := store.AddSSEConnection(key, w, flusher)
	assert.Len(t, store.connections[key], 1)

	// Remove connection
	store.RemoveSSEConnection(key, conn)
	assert.Len(t, store.connections[key], 0)
}

func TestSendEventToConnections(t *testing.T) {
	store := NewPromptStore()
	key := "test-key"

	w := &MockResponseWriter{}
	flusher := &MockFlusher{}

	store.AddSSEConnection(key, w, flusher)

	// Send event
	store.SendEventToConnections(key, "test_event", "test data", "test-id")

	// Verify event was sent
	assert.Contains(t, string(w.data), `"type":"test_event"`)
	assert.Contains(t, string(w.data), `"content":"test data"`)
	assert.Contains(t, string(w.data), `"id":"test-id"`)
}

func TestSendEvent(t *testing.T) {
	store := NewPromptStore()

	w := &MockResponseWriter{}
	flusher := &MockFlusher{}

	store.SendEvent(w, flusher, "test_event", "test data", "test-id")

	// Verify event format
	eventData := string(w.data)
	assert.True(t, strings.HasPrefix(eventData, "data: "))
	assert.Contains(t, eventData, `"type":"test_event"`)
	assert.Contains(t, eventData, `"content":"test data"`)
	assert.Contains(t, eventData, `"id":"test-id"`)
	assert.True(t, strings.HasSuffix(eventData, "\n\n"))
}

func TestNotifySSEConnections(t *testing.T) {
	store := NewPromptStore()
	key := "test-key"

	w := &MockResponseWriter{}
	flusher := &MockFlusher{}

	store.AddSSEConnection(key, w, flusher)

	// Add prompt and notify
	prompt := &Prompt{
		Id:      "test-id",
		Key:     key,
		Message: "test message",
	}
	store.NotifySSEConnections(prompt)

	// Verify notification was sent
	assert.Contains(t, string(w.data), `"type":"new_prompt"`)
	assert.Contains(t, string(w.data), `"content":"test message"`)
}
