package core

import (
	"fmt"
	"net/http"
	"sync"

	"github.com/google/uuid"
)

// PromptStore stores prompts and manages SSE connections
type PromptStore struct {
	prompts     map[string]*Prompt
	connections map[string][]*SSEConnection
	mutex       sync.RWMutex
}

type SSEConnection struct {
	writer  http.ResponseWriter
	flusher http.Flusher
	key     string
}

type Prompt struct {
	Id       string       `json:"id"`
	Key      string       `json:"-"`
	Message  string       `json:"message"`
	Callback func(string) `json:"-"`
}

func NewPromptStore() *PromptStore {
	return &PromptStore{
		prompts:     make(map[string]*Prompt),
		connections: make(map[string][]*SSEConnection),
	}
}

func (s *PromptStore) AddPrompt(key string, message string, callback func(string)) string {
	s.mutex.Lock()

	prompt := &Prompt{
		Id:       uuid.New().String(),
		Key:      key,
		Message:  message,
		Callback: callback,
	}
	s.prompts[prompt.Id] = prompt
	s.mutex.Unlock()
	s.NotifySSEConnections(prompt)
	return prompt.Id
}

func (s *PromptStore) GetPrompts(key string, id string) []*Prompt {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	prompts := []*Prompt{}
	for _, prompt := range s.prompts {
		if prompt.Key == key || prompt.Id == id {
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
	s.SendEventToConnections(prompt.Key, "new_prompt", prompt.Message)
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

func (s *PromptStore) SendEventToConnections(key string, eventType string, data string) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	if connections, exists := s.connections[key]; exists {
		for _, conn := range connections {
			s.SendEvent(conn.writer, conn.flusher, eventType, data)
		}
	}
}

func (s *PromptStore) SendEvent(w http.ResponseWriter, flusher http.Flusher, eventType string, data string) {
	event := fmt.Sprintf("data: {\"type\": \"%s\", \"content\": \"%s\"}\n\n", eventType, data)
	w.Write([]byte(event))
	flusher.Flush()
}
