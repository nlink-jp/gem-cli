# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [0.3.0] - 2026-04-02

### Added
- **Chat mode** — `--chat` enables interactive multi-turn conversations with Gemini
  - readline support (arrow keys, input history, line editing) via `chzyer/readline`
  - Works with all existing flags (`--grounding`, `--stream`, `--format`, `-s`)
- **Grounding citations** — `--grounding` now displays search queries (stderr) and source footnotes (stdout)
  - Text mode: numbered `[1] Title - URL` footnotes after response
  - JSON mode: structured `grounding` field with queries and sources
- **Session persistence** — `--session path.json` saves/restores chat history across sessions
- **Context caching** — `--cache` caches system prompt to reduce per-turn cost in chat mode
  - Graceful fallback if cache creation fails (token minimum not met, etc.)
- **URL resolution** — grounding source URLs automatically resolved from Vertex AI redirect tracking URLs to actual destination URLs via parallel HTTP HEAD requests
- New `internal/grounding`, `internal/chat`, `internal/session` packages with full test coverage

### Changed
- `--grounding` with `--format json` now wraps response on client side (Gemini API does not support controlled generation with Search tool)
- `--json-schema` is ignored when `--grounding` is active
- `Generate()` now returns `Result` struct (with `Text`, `SearchQueries`, `Sources`) instead of bare `string`
- Output formatting supports grounding metadata in both text and JSON modes

## [0.2.0] - 2026-03-29

### Added
- **`-p, --prompt`** flag for explicit prompt text (lite-llm compatible)
- **`-f, --file`** flag for text input file (read as text, use `-` for stdin; lite-llm compatible)
- **Batch mode** — `--batch` processes stdin line by line with per-line data isolation
- **`--attach`** flag for multimodal file attachment (renamed from `--file` to avoid conflict with `-f`)
- Expanded README with full flag reference, Quick Start, dedicated sections for Multimodal, Data Isolation, Structured Output, Grounding

### Fixed
- Stream mode no longer double-prints output
- Batch mode correctly defers stdin reading to RunBatch
- SystemInstruction role set to empty (not "user")

## [0.1.0] - 2026-03-29

### Added
- Text generation via Gemini 2.5 Pro/Flash on Vertex AI (`google.golang.org/genai` unified SDK)
- Streaming output (`--stream`)
- Multimodal input: `--image` for images
- Google Search Grounding (`--grounding`)
- Structured output (`--format json`, `--json-schema`)
- Prompt injection protection (nonce-tagged XML data isolation)
- Configuration via TOML file + env vars
- Unit tests for all 6 packages (31 tests)
