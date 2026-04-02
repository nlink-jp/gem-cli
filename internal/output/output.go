// Package output handles formatting and writing LLM responses.
package output

import (
	"fmt"
	"io"

	"github.com/nlink-jp/gem-cli/internal/client"
	"github.com/nlink-jp/gem-cli/internal/grounding"
)

// Write outputs the result in the specified format.
func Write(w io.Writer, result client.Result, format string) error {
	switch format {
	case "json", "jsonl":
		if len(result.Sources) > 0 || len(result.SearchQueries) > 0 {
			_, err := fmt.Fprintln(w, grounding.FormatJSON(result))
			return err
		}
		_, err := fmt.Fprintln(w, result.Text)
		return err
	default:
		_, err := fmt.Fprint(w, result.Text)
		if err != nil {
			return err
		}
		footnotes := grounding.FormatFootnotes(result.Sources)
		if footnotes != "" {
			_, err = fmt.Fprint(w, footnotes)
			if err != nil {
				return err
			}
		}
		_, err = fmt.Fprintln(w)
		return err
	}
}
