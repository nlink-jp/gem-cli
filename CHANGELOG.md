# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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
