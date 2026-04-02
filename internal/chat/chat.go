// Package chat implements the interactive multi-turn REPL for gem-cli.
package chat

import (
	"context"
	"fmt"
	"io"
	"strings"

	"github.com/chzyer/readline"
	"github.com/nlink-jp/gem-cli/internal/client"
	"github.com/nlink-jp/gem-cli/internal/grounding"
	"github.com/nlink-jp/gem-cli/internal/output"
	"github.com/nlink-jp/gem-cli/internal/session"
)

// Sender abstracts the chat send operation for testability.
type Sender interface {
	Send(ctx context.Context, text string) (client.Result, error)
	SendStream(ctx context.Context, text string) (client.Result, error)
}

// REPLOpts configures the REPL loop.
type REPLOpts struct {
	Output      io.Writer        // stdout
	Stderr      io.Writer        // stderr (for prompts and diagnostics)
	Stream      bool
	Format      string
	Quiet       bool
	Debug       bool
	HistFile    string           // readline input history file path
	Session     *session.Session // session for persistence (may be nil)
	SessionFile string           // path to save session (empty = no persistence)
}

// RunREPL runs the interactive chat loop.
// It reads user input via readline and sends messages to the Sender.
// Returns nil on clean exit (EOF, "exit", "quit").
func RunREPL(ctx context.Context, sender Sender, opts REPLOpts) error {
	rlConfig := &readline.Config{
		Prompt:          "> ",
		InterruptPrompt: "^C",
		EOFPrompt:       "exit",
		Stdout:          opts.Output,
		Stderr:          opts.Stderr,
	}
	if opts.HistFile != "" {
		rlConfig.HistoryFile = opts.HistFile
	}

	rl, err := readline.NewEx(rlConfig)
	if err != nil {
		return fmt.Errorf("init readline: %w", err)
	}
	defer rl.Close()

	for {
		line, err := rl.Readline()
		if err != nil {
			// EOF (Ctrl+D) or interrupt
			if err == readline.ErrInterrupt {
				continue // Ctrl+C just cancels current input
			}
			// io.EOF = Ctrl+D → clean exit
			fmt.Fprintln(opts.Output)
			return nil
		}

		text := strings.TrimSpace(line)
		if text == "" {
			continue
		}
		if text == "exit" || text == "quit" {
			return nil
		}

		result, sendErr := sendMessage(ctx, sender, text, opts.Stream)
		if sendErr != nil {
			fmt.Fprintf(opts.Stderr, "Error: %v\n", sendErr)
			continue
		}

		// Save to session if configured
		if opts.Session != nil {
			opts.Session.Append("user", text)
			opts.Session.Append("model", result.Text)
			if opts.SessionFile != "" {
				if saveErr := opts.Session.Save(opts.SessionFile); saveErr != nil {
					fmt.Fprintf(opts.Stderr, "Warning: save session: %v\n", saveErr)
				}
			}
		}

		// Resolve redirect URIs to actual URLs
		if len(result.Sources) > 0 {
			result.Sources = grounding.ResolveRedirects(result.Sources)
		}

		// Print search queries to stderr
		if !opts.Quiet && len(result.SearchQueries) > 0 {
			if q := grounding.FormatQueries(result.SearchQueries); q != "" {
				fmt.Fprintln(opts.Stderr, q)
			}
		}

		// For stream mode, text was already printed; just add footnotes
		if opts.Stream {
			footnotes := grounding.FormatFootnotes(result.Sources)
			if footnotes != "" {
				fmt.Fprintln(opts.Output, footnotes)
			}
		} else {
			if err := output.Write(opts.Output, result, opts.Format); err != nil {
				fmt.Fprintf(opts.Stderr, "Output error: %v\n", err)
			}
		}
		fmt.Fprintln(opts.Output) // blank line between turns
	}
}

func sendMessage(ctx context.Context, sender Sender, text string, stream bool) (client.Result, error) {
	if stream {
		return sender.SendStream(ctx, text)
	}
	return sender.Send(ctx, text)
}
