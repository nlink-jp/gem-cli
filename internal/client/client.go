// Package client provides a Gemini API client for gem-cli.
package client

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/nlink-jp/gem-cli/internal/config"
	"github.com/nlink-jp/gem-cli/internal/isolation"
	"google.golang.org/genai"
)

// Source represents a grounding source with title and URI.
type Source struct {
	Title  string
	URI    string
	Domain string
}

// Result holds the response text and optional grounding metadata.
type Result struct {
	Text          string
	SearchQueries []string
	Sources       []Source
}

// Generator is the interface for generating content. Extracted for testability.
type Generator interface {
	Generate(ctx context.Context, opts GenerateOpts) (Result, error)
}

// Client wraps the Gemini genai client and implements Generator.
type Client struct {
	inner *genai.Client
	model string
}

// New creates a Client configured for Vertex AI.
func New(ctx context.Context, cfg *config.Config) (*Client, error) {
	client, err := genai.NewClient(ctx, &genai.ClientConfig{
		Backend:  genai.BackendVertexAI,
		Project:  cfg.GCP.Project,
		Location: cfg.GCP.Location,
	})
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}
	return &Client{inner: client, model: cfg.Model.Name}, nil
}

// GenerateOpts holds parameters for a single generation call.
type GenerateOpts struct {
	SystemPrompt string
	Content      []*genai.Part
	Stream       bool
	Grounding    bool
	Format       string
	JSONSchema   string
	Debug        bool
}

// Generate calls the Gemini API and returns the result with grounding metadata.
func (c *Client) Generate(ctx context.Context, opts GenerateOpts) (Result, error) {
	gcConfig := &genai.GenerateContentConfig{}

	if opts.SystemPrompt != "" {
		gcConfig.SystemInstruction = genai.NewContentFromText(opts.SystemPrompt, "")
	}

	if opts.Grounding {
		gcConfig.Tools = []*genai.Tool{
			{GoogleSearch: &genai.GoogleSearch{}},
		}
	}

	// Controlled generation (ResponseMIMEType/ResponseSchema) is incompatible
	// with Google Search tool. When grounding is enabled, skip these settings;
	// JSON output with grounding metadata is handled by the output layer instead.
	if !opts.Grounding {
		if opts.Format == "json" || opts.JSONSchema != "" {
			gcConfig.ResponseMIMEType = "application/json"
		}
		if opts.JSONSchema != "" {
			schema, err := loadJSONSchema(opts.JSONSchema)
			if err != nil {
				return Result{}, err
			}
			gcConfig.ResponseSchema = schema
		}
	}

	if opts.Debug {
		fmt.Fprintf(os.Stderr, "[debug] model=%s grounding=%v format=%s parts=%d\n",
			c.model, opts.Grounding, opts.Format, len(opts.Content))
	}

	contents := []*genai.Content{
		genai.NewContentFromParts(opts.Content, "user"),
	}

	if opts.Stream {
		return c.generateStream(ctx, gcConfig, contents)
	}
	return c.generateSync(ctx, gcConfig, contents)
}

func (c *Client) generateSync(ctx context.Context, cfg *genai.GenerateContentConfig, contents []*genai.Content) (Result, error) {
	resp, err := c.inner.Models.GenerateContent(ctx, c.model, contents, cfg)
	if err != nil {
		return Result{}, err
	}
	return extractResult(resp), nil
}

func (c *Client) generateStream(ctx context.Context, cfg *genai.GenerateContentConfig, contents []*genai.Content) (Result, error) {
	var textBuf strings.Builder
	var lastGrounding *genai.GroundingMetadata
	for resp, err := range c.inner.Models.GenerateContentStream(ctx, c.model, contents, cfg) {
		if err != nil {
			if textBuf.Len() > 0 {
				fmt.Fprintln(os.Stdout)
			}
			return Result{Text: textBuf.String()}, err
		}
		text := extractText(resp)
		if text != "" {
			textBuf.WriteString(text)
			fmt.Fprint(os.Stdout, text)
		}
		if gm := extractGroundingMetadata(resp); gm != nil {
			lastGrounding = gm
		}
	}
	fmt.Fprintln(os.Stdout)
	r := Result{Text: textBuf.String()}
	if lastGrounding != nil {
		r.SearchQueries = lastGrounding.WebSearchQueries
		r.Sources = groundingChunksToSources(lastGrounding.GroundingChunks)
	}
	return r, nil
}

func extractResult(resp *genai.GenerateContentResponse) Result {
	r := Result{Text: extractText(resp)}
	if resp == nil || len(resp.Candidates) == 0 {
		return r
	}
	if gm := resp.Candidates[0].GroundingMetadata; gm != nil {
		r.SearchQueries = gm.WebSearchQueries
		r.Sources = groundingChunksToSources(gm.GroundingChunks)
	}
	return r
}

func extractText(resp *genai.GenerateContentResponse) string {
	if resp == nil || len(resp.Candidates) == 0 || resp.Candidates[0].Content == nil {
		return ""
	}
	var parts []string
	for _, part := range resp.Candidates[0].Content.Parts {
		if part.Text != "" {
			parts = append(parts, part.Text)
		}
	}
	return strings.Join(parts, "")
}

func extractGroundingMetadata(resp *genai.GenerateContentResponse) *genai.GroundingMetadata {
	if resp == nil || len(resp.Candidates) == 0 {
		return nil
	}
	return resp.Candidates[0].GroundingMetadata
}

func groundingChunksToSources(chunks []*genai.GroundingChunk) []Source {
	var sources []Source
	for _, chunk := range chunks {
		if chunk.Web != nil {
			sources = append(sources, Source{
				Title:  chunk.Web.Title,
				URI:    chunk.Web.URI,
				Domain: chunk.Web.Domain,
			})
		}
	}
	return sources
}

func loadJSONSchema(path string) (*genai.Schema, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read JSON schema %s: %w", path, err)
	}
	var schema genai.Schema
	if err := json.Unmarshal(data, &schema); err != nil {
		return nil, fmt.Errorf("parse JSON schema: %w", err)
	}
	return &schema, nil
}

// CreateCache creates a cached content resource for the given system prompt.
// Returns the cache name to be set on GenerateContentConfig.CachedContent.
func (c *Client) CreateCache(ctx context.Context, systemPrompt string, ttl time.Duration) (string, error) {
	cached, err := c.inner.Caches.Create(ctx, c.model, &genai.CreateCachedContentConfig{
		SystemInstruction: genai.NewContentFromText(systemPrompt, ""),
		TTL:               ttl,
		DisplayName:       "gem-cli-chat-cache",
	})
	if err != nil {
		return "", fmt.Errorf("create cache: %w", err)
	}
	return cached.Name, nil
}

// RunBatch reads stdin line by line and makes one API call per line.
// Each line is treated as user input with data isolation applied.
func RunBatch(ctx context.Context, c *Client, cfg *config.Config, sysPrompt string, stream bool, format string, debug bool) error {
	scanner := bufio.NewScanner(os.Stdin)
	scanner.Buffer(make([]byte, 0, 1024*1024), 1024*1024) // 1MB line buffer

	lineNum := 0
	for scanner.Scan() {
		line := scanner.Text()
		if strings.TrimSpace(line) == "" {
			continue
		}
		lineNum++

		// Apply data isolation per line
		wrappedLine, tag := isolation.Wrap(line)
		lineSysPrompt := sysPrompt
		if tag != "" {
			lineSysPrompt = isolation.ExpandTag(sysPrompt, tag)
		}

		parts := []*genai.Part{genai.NewPartFromText(wrappedLine)}

		if debug {
			fmt.Fprintf(os.Stderr, "[batch] line %d: %d chars\n", lineNum, len(line))
		}

		result, err := c.Generate(ctx, GenerateOpts{
			SystemPrompt: lineSysPrompt,
			Content:      parts,
			Stream:       false,
			Format:       format,
			Debug:        debug,
		})
		if err != nil {
			fmt.Fprintf(os.Stderr, "[batch] line %d error: %v\n", lineNum, err)
			continue
		}

		fmt.Println(strings.TrimSpace(result.Text))
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("read stdin: %w", err)
	}
	return nil
}

// ChatSender wraps a genai.Chat and provides Send/SendStream for the chat REPL.
type ChatSender struct {
	chat  *genai.Chat
	model string
}

// NewChat creates a ChatSender with a new genai chat session.
func (c *Client) NewChat(ctx context.Context, gcConfig *genai.GenerateContentConfig, history []*genai.Content) (*ChatSender, error) {
	chat, err := c.inner.Chats.Create(ctx, c.model, gcConfig, history)
	if err != nil {
		return nil, fmt.Errorf("create chat: %w", err)
	}
	return &ChatSender{chat: chat, model: c.model}, nil
}

// Send sends a text message and returns the full result.
func (cs *ChatSender) Send(ctx context.Context, text string) (Result, error) {
	resp, err := cs.chat.Send(ctx, genai.NewPartFromText(text))
	if err != nil {
		return Result{}, err
	}
	return extractResult(resp), nil
}

// SendStream sends a text message with streaming and returns the result.
func (cs *ChatSender) SendStream(ctx context.Context, text string) (Result, error) {
	var textBuf strings.Builder
	var lastGrounding *genai.GroundingMetadata
	part := genai.Part{Text: text}
	for resp, err := range cs.chat.SendMessageStream(ctx, part) {
		if err != nil {
			if textBuf.Len() > 0 {
				fmt.Fprintln(os.Stdout)
			}
			return Result{Text: textBuf.String()}, err
		}
		t := extractText(resp)
		if t != "" {
			textBuf.WriteString(t)
			fmt.Fprint(os.Stdout, t)
		}
		if gm := extractGroundingMetadata(resp); gm != nil {
			lastGrounding = gm
		}
	}
	fmt.Fprintln(os.Stdout)
	r := Result{Text: textBuf.String()}
	if lastGrounding != nil {
		r.SearchQueries = lastGrounding.WebSearchQueries
		r.Sources = groundingChunksToSources(lastGrounding.GroundingChunks)
	}
	return r, nil
}
