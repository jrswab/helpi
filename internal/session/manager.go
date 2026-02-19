package session

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	"github.com/jrswab/helpi/internal/llm"
)

type Manager interface {
	Get(userID int64) ([]llm.Message, error)
	Save(userID int64, messages []llm.Message) error
	Delete(userID int64) error
}

type manager struct {
	path        string
	maxMessages int
	mu          sync.RWMutex
}

func NewManager(path string, maxMessages int) (Manager, error) {
	if path == "" {
		path = "./data/sessions"
	}
	if maxMessages == 0 {
		maxMessages = 50
	}

	if err := os.MkdirAll(path, 0755); err != nil {
		return nil, fmt.Errorf("failed to create session directory: %w", err)
	}

	return &manager{
		path:        path,
		maxMessages: maxMessages,
	}, nil
}

func (m *manager) Get(userID int64) ([]llm.Message, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	path := m.sessionPath(userID)
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return []llm.Message{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("failed to read session: %w", err)
	}

	var messages []llm.Message
	if err := json.Unmarshal(data, &messages); err != nil {
		return nil, fmt.Errorf("failed to parse session: %w", err)
	}

	return messages, nil
}

func (m *manager) Save(userID int64, messages []llm.Message) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.maxMessages > 0 && len(messages) > m.maxMessages {
		messages = messages[len(messages)-m.maxMessages:]
	}

	data, err := json.Marshal(messages)
	if err != nil {
		return fmt.Errorf("failed to marshal session: %w", err)
	}

	path := m.sessionPath(userID)
	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write session: %w", err)
	}

	return nil
}

func (m *manager) Delete(userID int64) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	path := m.sessionPath(userID)
	if err := os.Remove(path); os.IsNotExist(err) {
		return nil
	} else if err != nil {
		return fmt.Errorf("failed to delete session: %w", err)
	}

	return nil
}

func (m *manager) sessionPath(userID int64) string {
	return filepath.Join(m.path, fmt.Sprintf("%d.json", userID))
}
