// Package session manages chat session persistence.
package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"

	"google.golang.org/genai"
)

// Message represents a single chat message.
type Message struct {
	Role string `json:"role"` // "user" or "model"
	Text string `json:"text"`
}

// Session holds the conversation history.
type Session struct {
	Model    string    `json:"model"`
	Messages []Message `json:"messages"`
}

// Load reads a session from a JSON file.
// Returns an empty session if the file does not exist.
func Load(path string) (*Session, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return &Session{}, nil
		}
		return nil, fmt.Errorf("read session: %w", err)
	}
	var s Session
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("parse session: %w", err)
	}
	return &s, nil
}

// Save writes the session to a JSON file.
func (s *Session) Save(path string) error {
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return fmt.Errorf("marshal session: %w", err)
	}
	if err := os.WriteFile(path, data, 0o600); err != nil {
		return fmt.Errorf("write session: %w", err)
	}
	return nil
}

// Append adds a message to the session.
func (s *Session) Append(role, text string) {
	s.Messages = append(s.Messages, Message{Role: role, Text: text})
}

// ToHistory converts the session messages to genai Content for chat history.
func (s *Session) ToHistory() []*genai.Content {
	if len(s.Messages) == 0 {
		return nil
	}
	contents := make([]*genai.Content, len(s.Messages))
	for i, m := range s.Messages {
		contents[i] = genai.NewContentFromText(m.Text, genai.Role(m.Role))
	}
	return contents
}
