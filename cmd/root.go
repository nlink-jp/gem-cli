// Package cmd implements the gem-cli CLI commands.
package cmd

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/nlink-jp/gem-cli/internal/chat"
	"github.com/nlink-jp/gem-cli/internal/client"
	"github.com/nlink-jp/gem-cli/internal/config"
	"github.com/nlink-jp/gem-cli/internal/grounding"
	"github.com/nlink-jp/gem-cli/internal/input"
	"github.com/nlink-jp/gem-cli/internal/isolation"
	"github.com/nlink-jp/gem-cli/internal/output"
	"github.com/nlink-jp/gem-cli/internal/session"
	"github.com/spf13/cobra"
	"google.golang.org/genai"
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
		flagPrompt          string
		flagInputFile       string
		flagChat            bool
		flagSession         string
		flagCache           bool
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
				prompt:           flagPrompt,
				inputFile:        flagInputFile,
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
				chat:             flagChat,
				session:          flagSession,
				cache:            flagCache,
			})
		},
	}

	// Input flags
	root.Flags().StringVarP(&flagPrompt, "prompt", "p", "", "User prompt text")
	root.Flags().StringVarP(&flagInputFile, "file", "f", "", "Input file path (read as text, use - for stdin)")
	root.Flags().StringVarP(&flagSystemPrompt, "system-prompt", "s", "", "System prompt text")
	root.Flags().StringVarP(&flagSystemPromptFile, "system-prompt-file", "S", "", "System prompt file path")
	root.Flags().StringSliceVar(&flagImages, "image", nil, "Image file path (can be repeated)")
	root.Flags().StringSliceVar(&flagFiles, "attach", nil, "Attach file: PDF, audio, video (can be repeated)")

	// Model / endpoint
	root.Flags().StringVarP(&flagModel, "model", "m", "", "Model name (overrides config)")

	// Execution mode
	root.Flags().BoolVar(&flagStream, "stream", false, "Enable streaming output")
	root.Flags().BoolVar(&flagBatch, "batch", false, "Batch mode: one request per input line")
	root.Flags().BoolVar(&flagGrounding, "grounding", false, "Enable Google Search Grounding")
	root.Flags().BoolVar(&flagChat, "chat", false, "Interactive multi-turn chat mode")
	root.Flags().StringVar(&flagSession, "session", "", "Session file path for chat history persistence")
	root.Flags().BoolVar(&flagCache, "cache", false, "Cache system prompt in chat mode (requires >= 1024 tokens)")

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
	prompt           string
	inputFile        string
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
	chat             bool
	session          string
	cache            bool
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

	// Build prompt: -p flag > positional arg
	prompt := opts.prompt
	if prompt == "" && len(args) > 0 {
		prompt = args[0]
	}

	// Read input file (-f) as text data
	var fileData string
	if opts.inputFile != "" {
		data, err := input.ReadInputFile(opts.inputFile)
		if err != nil {
			return err
		}
		fileData = data
	}

	// Read system prompt
	sysPrompt, err := input.ReadSystemPrompt(opts.systemPrompt, opts.systemPromptFile)
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

	// Chat and batch are mutually exclusive
	if opts.chat && opts.batch {
		return fmt.Errorf("--chat and --batch are mutually exclusive")
	}

	// Chat mode: interactive multi-turn conversation
	if opts.chat {
		return runChat(ctx, gemClient, sysPrompt, opts)
	}

	// Batch mode: stdin is consumed line by line by RunBatch, not here
	if opts.batch {
		return client.RunBatch(ctx, gemClient, cfg, sysPrompt, opts.stream, opts.format, opts.debug)
	}

	// Read stdin if available (non-batch mode only)
	stdinData, hasStdin, err := input.ReadStdin()
	if err != nil {
		return err
	}

	if prompt == "" && !hasStdin && fileData == "" && len(opts.images) == 0 && len(opts.files) == 0 {
		return fmt.Errorf("no input provided: pass a prompt argument, pipe data via stdin, or use --image/--file")
	}

	// Apply data isolation to stdin and file input
	var dataTag string
	if fileData != "" && !opts.noSafeInput {
		fileData, dataTag = isolation.Wrap(fileData)
	}
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
	userContent, err := input.BuildContent(prompt, stdinData, hasStdin, fileData, opts.images, opts.files)
	if err != nil {
		return err
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

	// Resolve redirect URIs to actual URLs
	if len(result.Sources) > 0 {
		result.Sources = grounding.ResolveRedirects(result.Sources)
	}

	// Print search queries to stderr if grounding was used
	if !opts.quiet && len(result.SearchQueries) > 0 {
		if q := grounding.FormatQueries(result.SearchQueries); q != "" {
			fmt.Fprintln(os.Stderr, q)
		}
	}

	// Stream mode already writes text to stdout during generation;
	// still need to print grounding footnotes
	if opts.stream {
		footnotes := grounding.FormatFootnotes(result.Sources)
		if footnotes != "" {
			fmt.Fprintln(os.Stdout, footnotes)
		}
		return nil
	}
	return output.Write(os.Stdout, result, opts.format)
}

func runChat(ctx context.Context, gemClient *client.Client, sysPrompt string, opts runOpts) error {
	if opts.debug {
		fmt.Fprintf(os.Stderr, "[debug] entering chat mode: cache=%v session=%q grounding=%v stream=%v\n",
			opts.cache, opts.session, opts.grounding, opts.stream)
	}
	gcConfig := &genai.GenerateContentConfig{}
	if opts.grounding {
		gcConfig.Tools = []*genai.Tool{
			{GoogleSearch: &genai.GoogleSearch{}},
		}
	}

	// Context caching: cache system prompt to reduce per-turn cost
	if opts.cache && sysPrompt != "" {
		ttl := 60 * time.Minute
		if !opts.quiet {
			fmt.Fprint(os.Stderr, "Creating context cache...")
		}
		// Use a timeout for cache creation to avoid hanging
		cacheCtx, cacheCancel := context.WithTimeout(ctx, 30*time.Second)
		cacheName, cacheErr := gemClient.CreateCache(cacheCtx, sysPrompt, ttl)
		cacheCancel()
		if cacheErr != nil {
			if !opts.quiet {
				fmt.Fprintf(os.Stderr, "\nWarning: cache creation failed (continuing without cache): %v\n", cacheErr)
			}
			// Fall back to non-cached system instruction
			gcConfig.SystemInstruction = genai.NewContentFromText(sysPrompt, "")
		} else {
			gcConfig.CachedContent = cacheName
			if !opts.quiet {
				fmt.Fprintf(os.Stderr, " done (TTL=%s)\n", ttl)
			}
		}
	} else if sysPrompt != "" {
		gcConfig.SystemInstruction = genai.NewContentFromText(sysPrompt, "")
	}

	// Load session history if specified
	var sess *session.Session
	var history []*genai.Content
	if opts.session != "" {
		var err error
		sess, err = session.Load(opts.session)
		if err != nil {
			return fmt.Errorf("load session: %w", err)
		}
		history = sess.ToHistory()
		if len(sess.Messages) > 0 && !opts.quiet {
			fmt.Fprintf(os.Stderr, "Loaded session: %d messages\n", len(sess.Messages))
		}
	} else {
		sess = &session.Session{}
	}

	chatSender, err := gemClient.NewChat(ctx, gcConfig, history)
	if err != nil {
		return fmt.Errorf("create chat: %w", err)
	}

	// Build prompt from args/flags
	prompt := opts.prompt

	// If an initial prompt was provided, send it as the first message
	if prompt != "" {
		var result client.Result
		var sendErr error
		if opts.stream {
			result, sendErr = chatSender.SendStream(ctx, prompt)
		} else {
			result, sendErr = chatSender.Send(ctx, prompt)
		}
		if sendErr != nil {
			return fmt.Errorf("initial message: %w", sendErr)
		}
		sess.Append("user", prompt)
		sess.Append("model", result.Text)
		if opts.session != "" {
			if saveErr := sess.Save(opts.session); saveErr != nil {
				fmt.Fprintf(os.Stderr, "Warning: save session: %v\n", saveErr)
			}
		}
		if !opts.quiet && len(result.SearchQueries) > 0 {
			if q := grounding.FormatQueries(result.SearchQueries); q != "" {
				fmt.Fprintln(os.Stderr, q)
			}
		}
		if !opts.stream {
			if err := output.Write(os.Stdout, result, opts.format); err != nil {
				return err
			}
		} else {
			footnotes := grounding.FormatFootnotes(result.Sources)
			if footnotes != "" {
				fmt.Fprintln(os.Stdout, footnotes)
			}
		}
		fmt.Fprintln(os.Stdout)
	}

	// Check if stdin is a terminal — if not, don't start REPL
	stat, _ := os.Stdin.Stat()
	if stat.Mode()&os.ModeCharDevice == 0 {
		// Piped input — already handled via prompt, just exit
		if prompt != "" {
			return nil
		}
		return fmt.Errorf("--chat requires a terminal for interactive input (or provide a prompt)")
	}

	return chat.RunREPL(ctx, chatSender, chat.REPLOpts{
		Output:      os.Stdout,
		Stderr:      os.Stderr,
		Stream:      opts.stream,
		Format:      opts.format,
		Quiet:       opts.quiet,
		Debug:       opts.debug,
		Session:     sess,
		SessionFile: opts.session,
	})
}
