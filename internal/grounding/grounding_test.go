package grounding

import (
	"encoding/json"
	"testing"

	"github.com/nlink-jp/gem-cli/internal/client"
)

func TestFormatFootnotes_Empty(t *testing.T) {
	got := FormatFootnotes(nil)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestFormatFootnotes_Single(t *testing.T) {
	sources := []client.Source{
		{Title: "Example", URI: "https://example.com"},
	}
	got := FormatFootnotes(sources)
	want := "\n---\nSources:\n[1] Example - https://example.com"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatFootnotes_Multiple(t *testing.T) {
	sources := []client.Source{
		{Title: "First", URI: "https://first.com"},
		{Title: "Second", URI: "https://second.com"},
	}
	got := FormatFootnotes(sources)
	if got == "" {
		t.Fatal("expected non-empty output")
	}
	if !contains(got, "[1] First") || !contains(got, "[2] Second") {
		t.Errorf("missing numbered sources in %q", got)
	}
}

func TestFormatFootnotes_MissingTitle(t *testing.T) {
	sources := []client.Source{
		{Title: "", URI: "https://example.com"},
	}
	got := FormatFootnotes(sources)
	want := "\n---\nSources:\n[1] https://example.com"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatQueries_Empty(t *testing.T) {
	got := FormatQueries(nil)
	if got != "" {
		t.Errorf("got %q, want empty", got)
	}
}

func TestFormatQueries_Multiple(t *testing.T) {
	got := FormatQueries([]string{"query1", "query2"})
	want := "Search queries: query1, query2"
	if got != want {
		t.Errorf("got %q, want %q", got, want)
	}
}

func TestFormatJSON_TextOnly(t *testing.T) {
	result := client.Result{Text: "hello"}
	got := FormatJSON(result)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	if m["text"] != "hello" {
		t.Errorf("text: got %v, want hello", m["text"])
	}
	if _, ok := m["grounding"]; ok {
		t.Error("unexpected grounding field for text-only result")
	}
}

func TestFormatJSON_WithGrounding(t *testing.T) {
	result := client.Result{
		Text:          "answer",
		SearchQueries: []string{"q1"},
		Sources:       []client.Source{{Title: "T", URI: "https://example.com"}},
	}
	got := FormatJSON(result)
	var m map[string]any
	if err := json.Unmarshal([]byte(got), &m); err != nil {
		t.Fatalf("invalid JSON: %v", err)
	}
	g, ok := m["grounding"].(map[string]any)
	if !ok {
		t.Fatal("missing grounding field")
	}
	queries, ok := g["queries"].([]any)
	if !ok || len(queries) != 1 {
		t.Errorf("queries: got %v", g["queries"])
	}
	sources, ok := g["sources"].([]any)
	if !ok || len(sources) != 1 {
		t.Errorf("sources: got %v", g["sources"])
	}
}

func TestIsRedirectURI(t *testing.T) {
	tests := []struct {
		uri  string
		want bool
	}{
		{"https://vertexaisearch.cloud.google.com/grounding-api-redirect/ABC123", true},
		{"https://example.com/page", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := isRedirectURI(tt.uri); got != tt.want {
			t.Errorf("isRedirectURI(%q) = %v, want %v", tt.uri, got, tt.want)
		}
	}
}

func TestResolveRedirects_NoRedirects(t *testing.T) {
	sources := []client.Source{
		{Title: "A", URI: "https://example.com"},
	}
	got := ResolveRedirects(sources)
	if got[0].URI != "https://example.com" {
		t.Errorf("URI changed: got %q", got[0].URI)
	}
}

func TestResolveRedirects_Empty(t *testing.T) {
	got := ResolveRedirects(nil)
	if got != nil {
		t.Errorf("expected nil, got %v", got)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
