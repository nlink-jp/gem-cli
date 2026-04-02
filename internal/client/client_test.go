package client

import (
	"os"
	"path/filepath"
	"testing"

	"google.golang.org/genai"
)

func TestExtractResult_Normal(t *testing.T) {
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
	got := extractResult(resp)
	if got.Text != "hello world" {
		t.Errorf("got %q, want %q", got.Text, "hello world")
	}
	if len(got.Sources) != 0 {
		t.Errorf("expected no sources, got %d", len(got.Sources))
	}
}

func TestExtractResult_NilResponse(t *testing.T) {
	got := extractResult(nil)
	if got.Text != "" {
		t.Errorf("got %q, want empty", got.Text)
	}
}

func TestExtractResult_NoCandidates(t *testing.T) {
	resp := &genai.GenerateContentResponse{}
	got := extractResult(resp)
	if got.Text != "" {
		t.Errorf("got %q, want empty", got.Text)
	}
}

func TestExtractResult_NilContent(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{Content: nil},
		},
	}
	got := extractResult(resp)
	if got.Text != "" {
		t.Errorf("got %q, want empty", got.Text)
	}
}

func TestExtractResult_WithGrounding(t *testing.T) {
	resp := &genai.GenerateContentResponse{
		Candidates: []*genai.Candidate{
			{
				Content: &genai.Content{
					Parts: []*genai.Part{
						{Text: "answer"},
					},
				},
				GroundingMetadata: &genai.GroundingMetadata{
					WebSearchQueries: []string{"query1", "query2"},
					GroundingChunks: []*genai.GroundingChunk{
						{Web: &genai.GroundingChunkWeb{Title: "Title1", URI: "https://example.com/1"}},
						{Web: &genai.GroundingChunkWeb{Title: "Title2", URI: "https://example.com/2"}},
					},
				},
			},
		},
	}
	got := extractResult(resp)
	if got.Text != "answer" {
		t.Errorf("text: got %q, want %q", got.Text, "answer")
	}
	if len(got.SearchQueries) != 2 {
		t.Fatalf("search queries: got %d, want 2", len(got.SearchQueries))
	}
	if got.SearchQueries[0] != "query1" {
		t.Errorf("query[0]: got %q, want %q", got.SearchQueries[0], "query1")
	}
	if len(got.Sources) != 2 {
		t.Fatalf("sources: got %d, want 2", len(got.Sources))
	}
	if got.Sources[0].Title != "Title1" || got.Sources[0].URI != "https://example.com/1" {
		t.Errorf("source[0]: got %+v", got.Sources[0])
	}
}

func TestGroundingChunksToSources_NilWeb(t *testing.T) {
	chunks := []*genai.GroundingChunk{
		{Web: nil},
		{Web: &genai.GroundingChunkWeb{Title: "Valid", URI: "https://example.com", Domain: "example.com"}},
	}
	got := groundingChunksToSources(chunks)
	if len(got) != 1 {
		t.Fatalf("got %d sources, want 1", len(got))
	}
	if got[0].Title != "Valid" {
		t.Errorf("got title %q, want %q", got[0].Title, "Valid")
	}
	if got[0].Domain != "example.com" {
		t.Errorf("got domain %q, want %q", got[0].Domain, "example.com")
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
