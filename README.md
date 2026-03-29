# gem-cli

A CLI client for Google Gemini via Vertex AI. Supports multimodal input (images, PDF, audio, video), streaming, Google Search Grounding, structured output, and prompt injection protection.

[日本語版 README はこちら](README.ja.md)

## Features

- **Multimodal input** — Attach images, PDFs, audio, and video files alongside text prompts
- **Streaming** — Token-by-token output via `--stream`
- **Google Search Grounding** — Web-search-augmented generation via `--grounding`
- **Structured output** — `--format json` and `--json-schema` using Gemini's native `response_schema`
- **Data isolation** — Automatic nonce-tagged XML wrapping to prevent prompt injection (enabled by default)
- **Batch mode** — Process input line by line, one API call per line *(coming soon)*
- **Pipe-friendly** — Reads from stdin, writes to stdout; composable with Unix tools

## Installation

```sh
git clone https://github.com/nlink-jp/gem-cli.git
cd gem-cli
make build
# Binary: dist/gem-cli
```

## Configuration

```sh
gcloud auth application-default login
export GOOGLE_CLOUD_PROJECT="your-project-id"
```

Optional config file at `~/.config/gem-cli/config.toml`:

```toml
[gcp]
project  = "your-project-id"
location = "us-central1"

[model]
name = "gemini-2.5-flash"
```

| Environment variable | Description |
|---|---|
| `GOOGLE_CLOUD_PROJECT` | GCP project ID (required) |
| `GOOGLE_CLOUD_LOCATION` | Vertex AI region (default: us-central1) |
| `GEM_CLI_MODEL` | Model name (default: gemini-2.5-flash) |

## Usage

```sh
# Text prompt
gem-cli "What is the capital of Japan?"

# Image analysis
gem-cli "Describe this image" --image photo.jpg

# Multiple images
gem-cli "Compare these two diagrams" --image a.png --image b.png

# PDF analysis
gem-cli "Summarize this document" --file report.pdf

# Google Search Grounding
gem-cli --grounding "Latest CVE for Log4j"

# Streaming
gem-cli --stream "Write a haiku about security"

# Structured JSON output
gem-cli --format json "List three Go best practices"

# JSON Schema
gem-cli --json-schema person.json "Generate a fictional person"

# Pipe with data isolation
echo "Alice, 34, engineer" | gem-cli \
  -s "Extract fields from <{{DATA_TAG}}>. Return JSON." \
  --format json

# Model override
gem-cli --model gemini-2.5-pro "Complex analysis task" --file data.csv
```

## Data Isolation

When input comes from stdin or files, gem-cli wraps it in a randomly-tagged XML element to prevent prompt injection:

```
<user_data_a3f8b2>
{your data here}
</user_data_a3f8b2>
```

Use `{{DATA_TAG}}` in your system prompt to reference the tag. Disable with `--no-safe-input`.

## Building

```sh
make build      # Current platform → dist/gem-cli
make build-all  # All 5 platforms → dist/
make test       # Run tests
make check      # vet + test + build
```
