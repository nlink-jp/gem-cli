package output

import (
	"bytes"
	"testing"

	"github.com/nlink-jp/gem-cli/internal/client"
)

func TestWrite_Text(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, client.Result{Text: "hello world"}, "text")
	if err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "hello world\n" {
		t.Errorf("got %q, want %q", got, "hello world\n")
	}
}

func TestWrite_JSON(t *testing.T) {
	var buf bytes.Buffer
	err := Write(&buf, client.Result{Text: `{"key":"value"}`}, "json")
	if err != nil {
		t.Fatal(err)
	}
	if got := buf.String(); got != "{\"key\":\"value\"}\n" {
		t.Errorf("got %q", got)
	}
}
