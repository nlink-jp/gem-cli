// Package grounding formats grounding metadata for CLI output.
package grounding

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/nlink-jp/gem-cli/internal/client"
)

// FormatFootnotes returns numbered source citations.
// Returns empty string if there are no sources.
func FormatFootnotes(sources []client.Source) string {
	if len(sources) == 0 {
		return ""
	}
	var b strings.Builder
	b.WriteString("\n---\nSources:")
	for i, s := range sources {
		uri := displayURI(s)
		if s.Title != "" {
			fmt.Fprintf(&b, "\n[%d] %s - %s", i+1, s.Title, uri)
		} else {
			fmt.Fprintf(&b, "\n[%d] %s", i+1, uri)
		}
	}
	return b.String()
}

// FormatQueries returns a one-line summary of search queries for stderr.
// Returns empty string if there are no queries.
func FormatQueries(queries []string) string {
	if len(queries) == 0 {
		return ""
	}
	return "Search queries: " + strings.Join(queries, ", ")
}

// FormatJSON returns a JSON string with text and grounding metadata.
func FormatJSON(result client.Result) string {
	m := map[string]any{
		"text": result.Text,
	}
	if len(result.SearchQueries) > 0 || len(result.Sources) > 0 {
		g := map[string]any{}
		if len(result.SearchQueries) > 0 {
			g["queries"] = result.SearchQueries
		}
		if len(result.Sources) > 0 {
			sources := make([]map[string]string, len(result.Sources))
			for i, s := range result.Sources {
				sources[i] = map[string]string{
					"title":  s.Title,
					"uri":    displayURI(s),
					"domain": s.Domain,
				}
			}
			g["sources"] = sources
		}
		m["grounding"] = g
	}
	data, _ := json.Marshal(m)
	return string(data)
}

// ResolveRedirects resolves redirect URIs in sources concurrently.
// Vertex AI returns grounding-api-redirect URLs; this resolves them
// to actual destination URLs via HTTP HEAD requests.
func ResolveRedirects(sources []client.Source) []client.Source {
	if len(sources) == 0 {
		return sources
	}

	// Check if any source needs resolving
	needsResolve := false
	for _, s := range sources {
		if isRedirectURI(s.URI) {
			needsResolve = true
			break
		}
	}
	if !needsResolve {
		return sources
	}

	resolved := make([]client.Source, len(sources))
	copy(resolved, sources)

	httpClient := &http.Client{
		Timeout: 10 * time.Second,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			// Stop following redirects — we just want the Location header
			return http.ErrUseLastResponse
		},
	}

	var wg sync.WaitGroup
	for i := range resolved {
		if !isRedirectURI(resolved[i].URI) {
			continue
		}
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			req, err := http.NewRequestWithContext(ctx, http.MethodHead, resolved[idx].URI, nil)
			if err != nil {
				return
			}
			resp, err := httpClient.Do(req)
			if err != nil {
				return
			}
			resp.Body.Close()
			if loc := resp.Header.Get("Location"); loc != "" {
				resolved[idx].URI = loc
			}
		}(i)
	}
	wg.Wait()
	return resolved
}

func isRedirectURI(uri string) bool {
	return strings.Contains(uri, "grounding-api-redirect")
}

// displayURI returns the best available URI for display.
// If the URI is a redirect and domain is available, constructs a domain-based URL.
func displayURI(s client.Source) string {
	if s.URI == "" {
		return s.Domain
	}
	return s.URI
}
