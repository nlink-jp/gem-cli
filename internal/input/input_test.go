package input

import (
	"os"
	"path/filepath"
	"testing"
)

func TestReadSystemPrompt_Text(t *testing.T) {
	result, err := ReadSystemPrompt("hello", "")
	if err != nil {
		t.Fatal(err)
	}
	if result != "hello" {
		t.Errorf("got %q, want hello", result)
	}
}

func TestReadSystemPrompt_File(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "prompt.txt")
	if err := os.WriteFile(path, []byte("from file"), 0o600); err != nil {
		t.Fatal(err)
	}

	result, err := ReadSystemPrompt("", path)
	if err != nil {
		t.Fatal(err)
	}
	if result != "from file" {
		t.Errorf("got %q, want 'from file'", result)
	}
}

func TestReadSystemPrompt_Empty(t *testing.T) {
	result, err := ReadSystemPrompt("", "")
	if err != nil {
		t.Fatal(err)
	}
	if result != "" {
		t.Errorf("got %q, want empty", result)
	}
}

func TestDetectMIME(t *testing.T) {
	tests := []struct {
		path string
		want string
	}{
		{"photo.jpg", "image/jpeg"},
		{"photo.jpeg", "image/jpeg"},
		{"image.png", "image/png"},
		{"doc.pdf", "application/pdf"},
		{"audio.mp3", "audio/mpeg"},
		{"video.mp4", "video/mp4"},
		{"data.csv", "text/csv"},
		{"notes.md", "text/markdown"},
		{"unknown.xyz", "application/octet-stream"},
	}
	for _, tt := range tests {
		got := detectMIME(tt.path)
		if got != tt.want {
			t.Errorf("detectMIME(%q) = %q, want %q", tt.path, got, tt.want)
		}
	}
}

func TestBuildContent_TextOnly(t *testing.T) {
	parts, err := BuildContent("hello", "", false, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 1 {
		t.Fatalf("got %d parts, want 1", len(parts))
	}
}

func TestBuildContent_WithStdin(t *testing.T) {
	parts, err := BuildContent("analyze", "data here", true, nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 1 {
		t.Fatalf("got %d parts, want 1 (text combined)", len(parts))
	}
}

func TestBuildContent_WithImage(t *testing.T) {
	dir := t.TempDir()
	imgPath := filepath.Join(dir, "test.png")
	if err := os.WriteFile(imgPath, []byte("fake png"), 0o600); err != nil {
		t.Fatal(err)
	}

	parts, err := BuildContent("describe", "", false, []string{imgPath}, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(parts) != 2 {
		t.Fatalf("got %d parts, want 2 (text + image)", len(parts))
	}
}
