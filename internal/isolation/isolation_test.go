package isolation

import (
	"strings"
	"testing"
)

func TestWrap(t *testing.T) {
	data := "some user input"
	wrapped, tag := Wrap(data)

	if tag == "" {
		t.Fatal("tag should not be empty")
	}
	if !strings.HasPrefix(tag, "user_data_") {
		t.Errorf("tag %q should start with user_data_", tag)
	}
	if !strings.Contains(wrapped, "<"+tag+">") {
		t.Errorf("wrapped should contain opening tag <%s>", tag)
	}
	if !strings.Contains(wrapped, "</"+tag+">") {
		t.Errorf("wrapped should contain closing tag </%s>", tag)
	}
	if !strings.Contains(wrapped, data) {
		t.Error("wrapped should contain the original data")
	}
}

func TestWrap_UniqueNonce(t *testing.T) {
	_, tag1 := Wrap("a")
	_, tag2 := Wrap("b")
	if tag1 == tag2 {
		t.Error("two calls to Wrap should produce different nonces")
	}
}

func TestExpandTag(t *testing.T) {
	result := ExpandTag("Extract from <{{DATA_TAG}}>.", "user_data_abc123")
	want := "Extract from <user_data_abc123>."
	if result != want {
		t.Errorf("got %q, want %q", result, want)
	}
}

func TestExpandTag_NoPlaceholder(t *testing.T) {
	input := "No placeholder here."
	result := ExpandTag(input, "user_data_abc123")
	if result != input {
		t.Errorf("should return input unchanged, got %q", result)
	}
}
