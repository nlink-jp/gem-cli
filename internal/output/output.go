// Package output handles formatting and writing LLM responses.
package output

import (
	"fmt"
	"io"
)

// Write outputs the result in the specified format.
func Write(w io.Writer, result string, format string) error {
	switch format {
	case "json", "jsonl":
		// For json/jsonl, the result is already formatted by the model
		_, err := fmt.Fprintln(w, result)
		return err
	default:
		_, err := fmt.Fprintln(w, result)
		return err
	}
}
