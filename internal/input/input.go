// Package input handles reading prompts, stdin, and multimodal files.
package input

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/genai"
)

// ReadSystemPrompt returns the system prompt from text or file.
func ReadSystemPrompt(text, filePath string) (string, error) {
	if text != "" {
		return text, nil
	}
	if filePath != "" {
		data, err := os.ReadFile(filePath)
		if err != nil {
			return "", fmt.Errorf("read system prompt file: %w", err)
		}
		return string(data), nil
	}
	return "", nil
}

// ReadStdin reads all of stdin if it's not a terminal.
func ReadStdin() (string, bool, error) {
	stat, _ := os.Stdin.Stat()
	if stat.Mode()&os.ModeCharDevice != 0 {
		return "", false, nil
	}
	data, err := io.ReadAll(os.Stdin)
	if err != nil {
		return "", false, fmt.Errorf("read stdin: %w", err)
	}
	s := string(data)
	if strings.TrimSpace(s) == "" {
		return "", false, nil
	}
	return s, true, nil
}

// BuildContent assembles a []*genai.Part slice from prompt text, stdin data,
// and multimodal file paths.
func BuildContent(prompt, stdinData string, hasStdin bool, images, files []string) ([]*genai.Part, error) {
	var parts []*genai.Part

	// Text
	var textParts []string
	if prompt != "" {
		textParts = append(textParts, prompt)
	}
	if hasStdin {
		textParts = append(textParts, stdinData)
	}
	if len(textParts) > 0 {
		t := genai.NewPartFromText(strings.Join(textParts, "\n\n"))
		parts = append(parts, t)
	}

	// Images
	for _, path := range images {
		p, err := fileToInlineData(path)
		if err != nil {
			return nil, fmt.Errorf("image %s: %w", path, err)
		}
		parts = append(parts, p)
	}

	// Other files
	for _, path := range files {
		p, err := fileToInlineData(path)
		if err != nil {
			return nil, fmt.Errorf("file %s: %w", path, err)
		}
		parts = append(parts, p)
	}

	return parts, nil
}

func fileToInlineData(path string) (*genai.Part, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	mime := detectMIME(path)
	return genai.NewPartFromBytes(data, mime), nil
}

func detectMIME(path string) string {
	ext := strings.ToLower(filepath.Ext(path))
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	case ".pdf":
		return "application/pdf"
	case ".mp3":
		return "audio/mpeg"
	case ".wav":
		return "audio/wav"
	case ".mp4":
		return "video/mp4"
	case ".mov":
		return "video/quicktime"
	case ".txt":
		return "text/plain"
	case ".md":
		return "text/markdown"
	case ".csv":
		return "text/csv"
	default:
		return "application/octet-stream"
	}
}
