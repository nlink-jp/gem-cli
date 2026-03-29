# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.1.0] - 2026-03-29

### Added
- **Text generation** via Gemini 2.5 Pro/Flash on Vertex AI (`google.golang.org/genai` unified SDK)
- **Streaming output** — `--stream` for token-by-token display
- **Multimodal input** — `--image` and `--file` for images, PDFs, audio, and video
- **Google Search Grounding** — `--grounding` for web-search-augmented generation
- **Structured output** — `--format json` and `--json-schema` using Gemini's native `response_schema`
- **Batch mode** — `--batch` processes stdin line by line, one API call per line, with per-line data isolation
- **Prompt injection protection** — Automatic nonce-tagged XML wrapping of stdin/file input; `{{DATA_TAG}}` expansion in system prompts
- **Configuration** — TOML config file + env var overrides (`GOOGLE_CLOUD_PROJECT`, `GEM_CLI_MODEL`)
- Unit tests for all 6 packages (31 tests)
