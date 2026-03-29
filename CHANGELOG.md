# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Project scaffold: CLI structure, config, Gemini client, multimodal input, data isolation
- Text generation with Gemini 2.5 Pro/Flash via Vertex AI
- Streaming output (`--stream`)
- Multimodal input: `--image` and `--file` (images, PDF, audio, video)
- Google Search Grounding (`--grounding`)
- Structured output (`--format json`, `--json-schema`)
- Prompt injection protection (nonce-tagged XML data isolation)
- Configuration via TOML file + env vars (`GOOGLE_CLOUD_PROJECT`, `GEM_CLI_MODEL`)
