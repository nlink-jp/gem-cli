package session

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoad_NewFile(t *testing.T) {
	s, err := Load("/nonexistent/path/session.json")
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if len(s.Messages) != 0 {
		t.Errorf("expected empty messages, got %d", len(s.Messages))
	}
}

func TestSaveAndLoad_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "test-session.json")

	s := &Session{
		Model: "gemini-2.5-flash",
		Messages: []Message{
			{Role: "user", Text: "hello"},
			{Role: "model", Text: "hi there"},
			{Role: "user", Text: "how are you?"},
			{Role: "model", Text: "I'm doing well!"},
		},
	}
	if err := s.Save(path); err != nil {
		t.Fatalf("Save: %v", err)
	}

	loaded, err := Load(path)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded.Model != "gemini-2.5-flash" {
		t.Errorf("model: got %q, want %q", loaded.Model, "gemini-2.5-flash")
	}
	if len(loaded.Messages) != 4 {
		t.Fatalf("messages: got %d, want 4", len(loaded.Messages))
	}
	if loaded.Messages[0].Role != "user" || loaded.Messages[0].Text != "hello" {
		t.Errorf("message[0]: got %+v", loaded.Messages[0])
	}
	if loaded.Messages[1].Role != "model" || loaded.Messages[1].Text != "hi there" {
		t.Errorf("message[1]: got %+v", loaded.Messages[1])
	}
}

func TestLoad_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}
	_, err := Load(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestAppend(t *testing.T) {
	s := &Session{}
	s.Append("user", "hello")
	s.Append("model", "hi")
	if len(s.Messages) != 2 {
		t.Fatalf("got %d messages, want 2", len(s.Messages))
	}
	if s.Messages[0].Role != "user" || s.Messages[0].Text != "hello" {
		t.Errorf("message[0]: got %+v", s.Messages[0])
	}
}

func TestToHistory_Empty(t *testing.T) {
	s := &Session{}
	h := s.ToHistory()
	if h != nil {
		t.Errorf("expected nil history, got %v", h)
	}
}

func TestToHistory_WithMessages(t *testing.T) {
	s := &Session{
		Messages: []Message{
			{Role: "user", Text: "hello"},
			{Role: "model", Text: "hi"},
		},
	}
	h := s.ToHistory()
	if len(h) != 2 {
		t.Fatalf("got %d contents, want 2", len(h))
	}
	if h[0].Role != "user" {
		t.Errorf("history[0].Role: got %q, want %q", h[0].Role, "user")
	}
	if h[1].Role != "model" {
		t.Errorf("history[1].Role: got %q, want %q", h[1].Role, "model")
	}
}
