package core

import (
	"fmt"
	"net/http"
	"sync"
)

// PromptStore stores prompts and manages SSE connections
type PromptStore struct {
	prompts     map[string][]*Prompt
	connections map[string][]*SSEConnection
	mutex       sync.RWMutex
}

type SSEConnection struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	key     string
}

type Prompt struct {
	key      string
	message  string
	callback func(string)
}

func NewPromptStore() *PromptStore {
	return &PromptStore{
		prompts:     make(map[string][]*Prompt),
		connections: make(map[string][]*SSEConnection),
	}
}

func (s *PromptStore) AddPrompt(key string, message string, callback func(string)) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	prompt := &Prompt{
		key:      key,
		message:  message,
		callback: callback,
	}
	s.prompts[key] = append(s.prompts[key], prompt)
	s.NotifySSEConnections(prompt)
}

func (s *PromptStore) GetPrompts() []*Prompt {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	var prompts []*Prompt
	for _, keyprompt := range s.prompts {
		for _, prompt := range keyprompt {
			prompts = append(prompts, prompt)
		}
	}
	return prompts
}

// Add this to PromptStore
func (s *PromptStore) RemovePrompt(id string) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	delete(s.prompts, id)
}

func (s *PromptStore) NotifySSEConnections(prompt *Prompt) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	event := fmt.Sprintf("event: %s\ndata: %s\n\n", "Prompt", prompt.message)
	s.SendEventToConnections(prompt.key, event)
}

// Add this to PromptStore
func (s *PromptStore) AddSSEConnection(key string, writer http.ResponseWriter, flusher http.Flusher) *SSEConnection {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	connection := &SSEConnection{
		writer:  writer,
		flusher: flusher,
		key:     key,
	}
	s.connections[key] = append(s.connections[key], connection)
	return connection
}

func (s *PromptStore) RemoveSSEConnection(key string, connection *SSEConnection) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if connections, exists := s.connections[key]; exists {
		for i, conn := range connections {
			if conn == connection {
				s.connections[key] = append(connections[:i], connections[i+1:]...)
				break
			}
		}
	}
}

func (s *PromptStore) SendEventToConnections(key string, event string) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if connections, exists := s.connections[key]; exists {
		for _, conn := range connections {
			conn.writer.Write([]byte(event))
			conn.flusher.Flush()
		}
	}
}
