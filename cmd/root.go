// Package cmd implements the gem-cli CLI commands.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"

	"github.com/nlink-jp/gem-cli/internal/client"
	"github.com/nlink-jp/gem-cli/internal/config"
	"github.com/nlink-jp/gem-cli/internal/input"
	"github.com/nlink-jp/gem-cli/internal/isolation"
	"github.com/nlink-jp/gem-cli/internal/output"
	"github.com/spf13/cobra"
)

var appVersion string

// Execute runs the CLI with the given version string.
func Execute(version string) {
	appVersion = version
	if err := newRootCmd().Execute(); err != nil {
		os.Exit(1)
	}
}

func newRootCmd() *cobra.Command {
	var (
		flagConfig          string
		flagModel           string
		flagSystemPrompt    string
		flagSystemPromptFile string
		flagFormat          string
		flagJSONSchema      string
		flagStream          bool
		flagBatch           bool
		flagNoSafeInput     bool
		flagQuiet           bool
		flagDebug           bool
		flagGrounding       bool
		flagImages          []string
		flagFiles           []string
	)

	root := &cobra.Command{
		Use:   "gem-cli [flags] [prompt]",
		Short: "Gemini CLI client — multimodal prompts, streaming, batch, grounding",
		Long: `gem-cli is a CLI client for Google Gemini via Vertex AI.

Supports text, images, PDFs, audio, and video as input. Includes prompt
injection protection, batch processing, structured output, and Google
Search Grounding.

Examples:
  gem-cli "What is the capital of Japan?"
  gem-cli "Describe this image" --image photo.jpg
  gem-cli "Summarize this document" --file report.pdf
  echo "log data" | gem-cli "Find anomalies"
  gem-cli --grounding "Latest news about Log4j"
  gem-cli --stream "Write a haiku about security"
  cat items.txt | gem-cli --batch --format jsonl -s "Classify each item"`,
		SilenceUsage: true,
		RunE: func(cmd *cobra.Command, args []string) error {
			return runPrompt(cmd, args, runOpts{
				configPath:       flagConfig,
				model:            flagModel,
				systemPrompt:     flagSystemPrompt,
				systemPromptFile: flagSystemPromptFile,
				format:           flagFormat,
				jsonSchema:       flagJSONSchema,
				stream:           flagStream,
				batch:            flagBatch,
				noSafeInput:      flagNoSafeInput,
				quiet:            flagQuiet,
				debug:            flagDebug,
				grounding:        flagGrounding,
				images:           flagImages,
				files:            flagFiles,
			})
		},
	}

	// Input flags
	root.Flags().StringVarP(&flagSystemPrompt, "system-prompt", "s", "", "System prompt text")
	root.Flags().StringVarP(&flagSystemPromptFile, "system-prompt-file", "S", "", "System prompt file path")
	root.Flags().StringSliceVar(&flagImages, "image", nil, "Image file path (can be repeated)")
	root.Flags().StringSliceVar(&flagFiles, "file", nil, "File path: PDF, audio, video (can be repeated)")

	// Model / endpoint
	root.Flags().StringVarP(&flagModel, "model", "m", "", "Model name (overrides config)")

	// Execution mode
	root.Flags().BoolVar(&flagStream, "stream", false, "Enable streaming output")
	root.Flags().BoolVar(&flagBatch, "batch", false, "Batch mode: one request per input line")
	root.Flags().BoolVar(&flagGrounding, "grounding", false, "Enable Google Search Grounding")

	// Output format
	root.Flags().StringVar(&flagFormat, "format", "text", "Output format: text, json, jsonl")
	root.Flags().StringVar(&flagJSONSchema, "json-schema", "", "JSON Schema file (implies --format json)")

	// Security
	root.Flags().BoolVar(&flagNoSafeInput, "no-safe-input", false, "Disable automatic data isolation")
	root.Flags().BoolVarP(&flagQuiet, "quiet", "q", false, "Suppress warnings on stderr")
	root.Flags().BoolVar(&flagDebug, "debug", false, "Log request details to stderr")

	// Config
	root.Flags().StringVarP(&flagConfig, "config", "c", "", "Config file path")

	root.Version = appVersion

	return root
}

type runOpts struct {
	configPath       string
	model            string
	systemPrompt     string
	systemPromptFile string
	format           string
	jsonSchema       string
	stream           bool
	batch            bool
	noSafeInput      bool
	quiet            bool
	debug            bool
	grounding        bool
	images           []string
	files            []string
}

func runPrompt(cmd *cobra.Command, args []string, opts runOpts) error {
	// Load config
	cfg, err := config.Load(opts.configPath)
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	// CLI overrides
	if opts.model != "" {
		cfg.Model.Name = opts.model
	}

	// Build prompt from args
	prompt := ""
	if len(args) > 0 {
		prompt = args[0]
	}

	// Read system prompt
	sysPrompt, err := input.ReadSystemPrompt(opts.systemPrompt, opts.systemPromptFile)
	if err != nil {
		return err
	}

	// Read stdin if available
	stdinData, hasStdin, err := input.ReadStdin()
	if err != nil {
		return err
	}

	if prompt == "" && !hasStdin && len(opts.images) == 0 && len(opts.files) == 0 {
		return fmt.Errorf("no input provided: pass a prompt argument, pipe data via stdin, or use --image/--file")
	}

	// Apply data isolation
	var dataTag string
	if hasStdin && !opts.noSafeInput {
		stdinData, dataTag = isolation.Wrap(stdinData)
		if !opts.quiet {
			fmt.Fprintf(os.Stderr, "Data isolation: enabled (tag=%s)\n", dataTag)
		}
	}
	if dataTag != "" {
		sysPrompt = isolation.ExpandTag(sysPrompt, dataTag)
	}

	// Build user content: prompt + stdin + multimodal files
	userContent, err := input.BuildContent(prompt, stdinData, hasStdin, opts.images, opts.files)
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	// Create Gemini client
	gemClient, err := client.New(ctx, cfg)
	if err != nil {
		return fmt.Errorf("create client: %w", err)
	}

	// Execute
	if opts.batch {
		return client.RunBatch(ctx, gemClient, cfg, sysPrompt, opts.stream, opts.format, opts.debug)
	}

	result, err := gemClient.Generate(ctx, client.GenerateOpts{
		SystemPrompt: sysPrompt,
		Content:      userContent,
		Stream:       opts.stream,
		Grounding:    opts.grounding,
		Format:       opts.format,
		JSONSchema:   opts.jsonSchema,
		Debug:        opts.debug,
	})
	if err != nil {
		return fmt.Errorf("generate: %w", err)
	}

	// Stream mode already writes to stdout during generation
	if opts.stream {
		return nil
	}
	return output.Write(os.Stdout, result, opts.format)
}
