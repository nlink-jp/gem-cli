# CLAUDE.md — gem-cli

**Organization rules (mandatory): https://github.com/nlink-jp/.github/blob/main/CONVENTIONS.md**

## This project

Gemini CLI client for Vertex AI. Supports multimodal input (images, PDF, audio, video),
streaming, Google Search Grounding, structured output, and prompt injection protection.

## Key structure

```
main.go                  ← entry point
cmd/root.go              ← Cobra CLI, flag wiring, runPrompt orchestration
internal/config/         ← TOML config, env var overrides
internal/client/         ← Gemini API client (google.golang.org/genai, BackendVertexAI)
internal/input/          ← stdin/file reading, multimodal Part assembly
internal/isolation/      ← Prompt injection protection (nonce-tagged XML)
internal/output/         ← Response formatting (text/json/jsonl)
```

## Build & test

```sh
make build    # dist/gem-cli
make test     # go test ./...
make check    # vet → test → build
```

## Environment

```sh
export GOOGLE_CLOUD_PROJECT="your-project-id"
gcloud auth application-default login
```

## Design notes

- Uses `google.golang.org/genai` (unified SDK, GA v1.x) — NOT the deprecated `cloud.google.com/go/vertexai/genai`
- Data isolation: stdin/file input wrapped in `<user_data_{nonce}>` tags, `{{DATA_TAG}}` expanded in system prompt only
- Stream mode writes to stdout during generation; non-stream uses output.Write after completion
- Batch mode stub exists but is not yet implemented (Phase 3)
