package client

import (
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/genai"
)

func TestExtractText_Normal(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Text: "hello "},
						{Text: "world"},
					},
				},
			},
		},
	}
	got := extractText(resp)
	if got != "hello world" {
		t.Errorf("got %q, want %q", got, "hello world")
	}
}

func TestExtractText_NilResponse(t *testing.T) {
	got := extractText(nil)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestExtractText_NoCandidates(t *testing.T) {
	resp := &genai.GenerateContentResponse{}
	got := extractText(resp)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestExtractText_NilContent(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{Content: nil},
		},
	}
	got := extractText(resp)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestLoadJSONSchema_Valid(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "schema.json")
	content := `{"type":"OBJECT","properties":{"name":{"type":"STRING"}}}`
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	schema, err := loadJSONSchema(path)
	if err != nil {
		t.Fatalf("loadJSONSchema: %v", err)
	}
	if schema == nil {
		t.Fatal("schema should not be nil")
	}
}

func TestLoadJSONSchema_InvalidJSON(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "bad.json")
	if err := os.WriteFile(path, []byte("not json"), 0o600); err != nil {
		t.Fatal(err)
	}

	_, err := loadJSONSchema(path)
	if err == nil {
		t.Error("expected error for invalid JSON")
	}
}

func TestLoadJSONSchema_MissingFile(t *testing.T) {
	_, err := loadJSONSchema("/nonexistent/schema.json")
	if err == nil {
		t.Error("expected error for missing file")
	}
}
