# AGENTS.md — gem-cli

Gemini CLI client — multimodal prompts, streaming, batch, grounding, structured output.
Part of [cli-series](https://github.com/nlink-jp/cli-series).

## Rules

- Project rules: → [CLAUDE.md](CLAUDE.md)
- Organization conventions: → [CONVENTIONS.md](https://github.com/nlink-jp/.github/blob/main/CONVENTIONS.md)

## Build & test

```sh
make build    # dist/gem-cli
make test     # go test ./...
make check    # vet → test → build
```

## Key structure

```
main.go                  ← entry point, injects version
cmd/root.go              ← Cobra CLI flags + runPrompt orchestration
internal/config/         ← TOML config + GOOGLE_CLOUD_PROJECT / GEM_CLI_MODEL env overrides
internal/client/         ← Gemini client (google.golang.org/genai, Vertex AI backend)
internal/input/          ← stdin reading, multimodal Part assembly (images, PDF, audio, video)
internal/isolation/      ← Nonce-tagged XML data wrapping for prompt injection defense
internal/output/         ← text / json / jsonl formatting
```

## Gotchas

- **SDK**: Uses `google.golang.org/genai` (unified, GA). The old `cloud.google.com/go/vertexai/genai` is deprecated.
- **Data isolation**: Enabled by default. Wraps stdin/file input in `<user_data_{nonce}>`. Use `{{DATA_TAG}}` in system prompt to reference. Disable with `--no-safe-input`.
- **Stream vs non-stream**: Stream writes directly to stdout during generation. Non-stream returns full result then writes once. Don't output twice.
- **Config priority**: CLI flags → env vars → config file → defaults.
- **Module path**: `github.com/nlink-jp/gem-cli`.
- **Env vars**: `GOOGLE_CLOUD_PROJECT` (required), `GEM_CLI_MODEL` (optional, default gemini-2.5-flash).
