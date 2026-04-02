# gem-cli

A CLI client for Google Gemini via Vertex AI. Supports multimodal input (images, PDF, audio, video), streaming, batch processing, Google Search Grounding with citations, interactive chat mode, session persistence, context caching, structured output, and prompt injection protection.

Designed as a Gemini-native counterpart to [lite-llm](https://github.com/nlink-jp/lite-llm) (OpenAI-compatible), with full access to Gemini-specific features.

[日本語版 README はこちら](README.ja.md)

## Features

- **Multimodal input** — Attach images, PDFs, audio, and video files alongside text prompts
- **Streaming** — Token-by-token output via `--stream`
- **Google Search Grounding** — Web-search-augmented generation with source citations via `--grounding`
- **Chat mode** — Interactive multi-turn conversations via `--chat` with readline support
- **Session persistence** — Save/restore chat history via `--session path.json`
- **Context caching** — Cache system prompts to reduce per-turn cost via `--cache`
- **Structured output** — `--format json` and `--json-schema` using Gemini's native `response_schema`
- **Data isolation** — Automatic nonce-tagged XML wrapping to prevent prompt injection (enabled by default)
- **Batch mode** — Process an input file line by line, one API call per line
- **Pipe-friendly** — Reads from stdin, writes to stdout; composable with Unix tools
- **Quiet mode** — `--quiet` / `-q` suppresses warnings for clean pipeline output

## Installation

```sh
git clone https://github.com/nlink-jp/gem-cli.git
cd gem-cli
make build
# Binary: dist/gem-cli
```

Or download a pre-built binary from the [releases page](https://github.com/nlink-jp/gem-cli/releases).

## Quick Start

```sh
# Authenticate with Google Cloud
gcloud auth application-default login
export GOOGLE_CLOUD_PROJECT="your-project-id"

# Ask a question
gem-cli "What is the capital of Japan?"

# Pipe data in (automatically isolated from instructions)
echo "2026-03-29: Revenue $45,200" | gem-cli "Extract the date and amount as JSON" --format json

# Analyze an image
gem-cli "What's in this picture?" --image photo.jpg

# Web search grounding (with source citations)
gem-cli --grounding "Latest news about Log4j vulnerabilities"

# Interactive chat
gem-cli --chat -s "You are a helpful assistant"

# Chat with session persistence
gem-cli --chat --session conv.json

# Batch processing
cat questions.txt | gem-cli --batch --format jsonl \
  --system-prompt "Answer in one sentence."

# Streaming
gem-cli --stream "Write a haiku about cybersecurity"
```

## Configuration

Copy the example config and set your values:

```sh
mkdir -p ~/.config/gem-cli
cp config.example.toml ~/.config/gem-cli/config.toml
```

```toml
# ~/.config/gem-cli/config.toml
[gcp]
project  = "your-project-id"
location = "us-central1"

[model]
name = "gemini-2.5-flash"

[cache]
enabled = false
ttl = "3600s"
```

**Priority order (highest first):** CLI flags → environment variables → config file → compiled-in defaults

| Environment variable | Description |
|---|---|
| `GOOGLE_CLOUD_PROJECT` | GCP project ID (required) |
| `GOOGLE_CLOUD_LOCATION` | Vertex AI region (default: us-central1) |
| `GEM_CLI_MODEL` | Default model name (default: gemini-2.5-flash) |

## Usage

```
gem-cli [flags] [prompt]

Input flags:
  -p, --prompt string              User prompt text
  -f, --file string                Input file path (read as text, use - for stdin)
  -s, --system-prompt string       System prompt text
  -S, --system-prompt-file string  System prompt file path

Multimodal:
      --image strings              Image file path (can be repeated)
      --attach strings             Attach file: PDF, audio, video (can be repeated)

Model:
  -m, --model string               Model name (overrides config)

Execution mode:
      --stream                     Enable streaming output
      --batch                      Batch mode: one request per input line
      --grounding                  Enable Google Search Grounding
      --chat                       Interactive multi-turn chat mode
      --session string             Session file for chat history persistence
      --cache                      Cache system prompt (requires >= 1024 tokens)

Output format:
      --format string              Output format: text (default), json, jsonl
      --json-schema string         JSON Schema file (implies --format json)

Security:
      --no-safe-input              Disable automatic data isolation
  -q, --quiet                      Suppress warnings on stderr
      --debug                      Log API request details to stderr

Config:
  -c, --config string              Config file path
```

## Multimodal Input

### Images

```sh
# Single image
gem-cli "Describe this image in detail" --image photo.jpg

# Multiple images
gem-cli "What are the differences between these two screenshots?" \
  --image before.png --image after.png
```

Supported formats: JPEG, PNG, GIF, WebP

### Documents, Audio, Video

```sh
# PDF analysis
gem-cli "Summarize the key findings" --attach report.pdf

# Audio transcription
gem-cli "Transcribe this audio" --attach recording.mp3

# Video analysis
gem-cli "Describe what happens in this video" --attach clip.mp4

# Combine text + multimodal
gem-cli "Is this diagram consistent with the specification?" \
  --image architecture.png --attach spec.pdf
```

Supported formats: PDF, MP3, WAV, MP4, MOV, TXT, CSV, Markdown

## Data Isolation

When input comes from stdin or a file (`-f`), gem-cli wraps it in a randomly-tagged XML element to prevent prompt injection:

```
<user_data_a3f8b2>
{your data here}
</user_data_a3f8b2>
```

Use `{{DATA_TAG}}` in your **system prompt** to reference the tag by name:

```sh
echo "Alice, 34, engineer" | gem-cli \
  --system-prompt "Extract fields from <{{DATA_TAG}}>. Return JSON with keys: name, age, role." \
  --format json
```

> `{{DATA_TAG}}` is expanded **only in the system prompt**, never in user input.

Disable with `--no-safe-input` when the input is trusted.

## Structured Output

```sh
# JSON object
gem-cli --format json "List three cybersecurity best practices"

# JSON Schema — strict structure enforcement
gem-cli --json-schema person.json "Generate a fictional security analyst"

# Batch + JSONL
cat items.txt | gem-cli --batch --format jsonl \
  --system-prompt "Classify as: food, vehicle, animal, or other."
```

### JSON Schema example

```json
{
  "type": "OBJECT",
  "properties": {
    "name": {"type": "STRING"},
    "age": {"type": "INTEGER"},
    "occupation": {"type": "STRING"}
  },
  "required": ["name", "age", "occupation"]
}
```

> Gemini uses its own schema format (`"type": "STRING"` not `"type": "string"`).

## Google Search Grounding

```sh
# Augment responses with live web search results
gem-cli --grounding "What is the latest version of Go?"

# Combine with JSON output
gem-cli --grounding --format json "Current Bitcoin price in USD and JPY"

# Suppress search queries on stderr for clean pipeline output
gem-cli --grounding --format json -q "Latest CVE for Log4j" | jq .
```

When `--grounding` is enabled, gem-cli displays:
- **Source citations** (stdout): Numbered footnotes appended after the response text
- **Search queries** (stderr): The search queries Gemini used (suppressed with `-q`)
- **URL resolution**: Vertex AI returns redirect tracking URLs; gem-cli automatically resolves them to actual destination URLs via parallel HTTP HEAD requests

Example output (text mode):
```
Go 1.26 was released in March 2026...

---
Sources:
[1] Go Release Notes - https://go.dev/doc/go1.26
[2] Go Blog - https://go.dev/blog/go1.26
```

Example output (JSON mode with `--format json`):
```json
{
  "text": "...",
  "grounding": {
    "queries": ["go 1.26 release"],
    "sources": [
      {"title": "go.dev", "uri": "https://go.dev/doc/go1.26", "domain": "go.dev"}
    ]
  }
}
```

> **Note:** Gemini does not support controlled generation (`ResponseMIMEType` / `ResponseSchema`) together with Google Search. When `--grounding` is used with `--format json`, gem-cli wraps the model's plain-text response and grounding metadata into a JSON structure on the client side. `--json-schema` is ignored when `--grounding` is active.

Grounding uses Google Search via Vertex AI and requires the Vertex AI API to be enabled in your project.

## Chat Mode

Interactive multi-turn conversations with Gemini:

```sh
# Start a chat session
gem-cli --chat

# Chat with system prompt
gem-cli --chat -s "You are a security analyst. Respond concisely."

# Chat with grounding (web search per turn)
gem-cli --chat --grounding

# Chat with streaming
gem-cli --chat --stream

# Start with an initial prompt, then continue interactively
gem-cli --chat -p "Hello, what can you help me with?"
```

In chat mode:
- Type your message and press Enter to send
- Arrow keys, history navigation, and line editing are supported
- Type `exit` or `quit` to end the session
- Press `Ctrl+D` (EOF) to exit
- Press `Ctrl+C` to cancel current input

## Session Persistence

Save and restore chat history across sessions:

```sh
# Start a new session (creates conv.json)
gem-cli --chat --session conv.json

# Resume a previous session (loads conv.json)
gem-cli --chat --session conv.json
```

The session file is a human-readable JSON file containing the conversation history. It is auto-saved after each turn.

## Context Caching

Cache your system prompt to reduce per-turn cost in long conversations:

```sh
gem-cli --chat --cache -s "You are a security analyst with deep expertise in threat hunting."
```

Context caching stores the system prompt on Google's servers (default TTL: 60 minutes) and subsequent turns reference the cache instead of re-sending the full prompt. This reduces both cost and latency for long conversations with large system prompts.

The system prompt must be at least **1024 tokens** for caching to work. If cache creation fails (e.g., prompt too short), gem-cli falls back to the standard system instruction automatically with a warning.

## Building

```sh
make build      # Current platform → dist/gem-cli
make build-all  # All 5 platforms → dist/
make test       # Run tests (31 tests across 6 packages)
make check      # vet + test + build
```

> **Note (sandbox / restricted environments):** If the default Go cache paths
> are not writable, override them:
>
> ```sh
> GOCACHE=/tmp/go-cache GOMODCACHE=/tmp/gopath/pkg/mod make build
> ```
