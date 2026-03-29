// Package client provides a Gemini API client for gem-cli.
package client

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"github.com/nlink-jp/gem-cli/internal/config"
	"google.golang.org/genai"
)

// Client wraps the Gemini genai client.
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

// Generate calls the Gemini API and returns the response text.
func (c *Client) Generate(ctx context.Context, opts GenerateOpts) (string, error) {
	gcConfig := &genai.GenerateContentConfig{}

	if opts.SystemPrompt != "" {
		gcConfig.SystemInstruction = genai.NewContentFromText(opts.SystemPrompt, "user")
	}

	if opts.Grounding {
		gcConfig.Tools = []*genai.Tool{
			{GoogleSearch: &genai.GoogleSearch{}},
		}
	}

	if opts.Format == "json" || opts.JSONSchema != "" {
		gcConfig.ResponseMIMEType = "application/json"
	}
	if opts.JSONSchema != "" {
		schema, err := loadJSONSchema(opts.JSONSchema)
		if err != nil {
			return "", err
		}
		gcConfig.ResponseSchema = schema
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

func (c *Client) generateSync(ctx context.Context, cfg *genai.GenerateContentConfig, contents []*genai.Content) (string, error) {
	resp, err := c.inner.Models.GenerateContent(ctx, c.model, contents, cfg)
	if err != nil {
		return "", err
	}
	return extractText(resp), nil
}

func (c *Client) generateStream(ctx context.Context, cfg *genai.GenerateContentConfig, contents []*genai.Content) (string, error) {
	var result strings.Builder
	for resp, err := range c.inner.Models.GenerateContentStream(ctx, c.model, contents, cfg) {
		if err != nil {
			if result.Len() > 0 {
				fmt.Fprintln(os.Stdout)
			}
			return result.String(), err
		}
		text := extractText(resp)
		if text != "" {
			result.WriteString(text)
			fmt.Fprint(os.Stdout, text)
		}
	}
	fmt.Fprintln(os.Stdout)
	return result.String(), nil
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

// RunBatch reads stdin line by line and makes one API call per line.
func RunBatch(ctx context.Context, c *Client, cfg *config.Config, sysPrompt string, stream bool, format string, debug bool) error {
	// TODO: implement batch mode in Phase 3
	return fmt.Errorf("batch mode is not yet implemented")
}
