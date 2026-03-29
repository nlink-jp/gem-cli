// Package isolation provides prompt injection protection via nonce-tagged XML wrapping.
package isolation

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strings"
)

// Wrap wraps data in a nonce-tagged XML element to isolate it from instructions.
// Returns the wrapped data and the nonce tag name.
func Wrap(data string) (string, string) {
	tag := generateTag()
	wrapped := fmt.Sprintf("<%s>\n%s\n</%s>", tag, data, tag)
	return wrapped, tag
}

// ExpandTag replaces {{DATA_TAG}} in a system prompt with the actual tag name.
func ExpandTag(systemPrompt, tag string) string {
	return strings.ReplaceAll(systemPrompt, "{{DATA_TAG}}", tag)
}

func generateTag() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return "user_data_" + hex.EncodeToString(b)
}
