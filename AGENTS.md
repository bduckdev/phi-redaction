# AGENTS.md

## Build & Test Commands
- Build: `make build` or `go build -o main cmd/api/main.go`
- Run: `make run` or `go run cmd/api/main.go`
- Test all: `make test` or `go test ./... -v`
- Single test: `go test -v -run TestName ./path/to/package`
- Live reload: `make watch` (requires `air`)

## Code Style
- **Imports**: stdlib first, blank line, external deps, blank line, internal (`phi-redactor/...`)
- **Formatting**: `gofmt` standard. No special config.
- **Types**: Use custom types for enums (`EntityType`, `Confidence` as `string` types with constants)
- **Naming**: PascalCase exports, camelCase internal. Receiver names: single letter (`l` for Lexer, `s` for Server)
- **Errors**: Return errors, don't panic. Fatal only in main/startup paths.
- **Tests**: Table-driven with `[]struct{name, input, want}`. Use `github.com/google/go-cmp/cmp` for diffs.
- **Packages**: `internal/` for app code, `pkg/` for reusable utilities, `cmd/` for entrypoints

## Project Structure
- HTTP server uses chi router (`go-chi/chi/v5`)
- Detectors return `[]finding.Finding` with start/end byte positions
- Lexer tokenizes text preserving position info for redaction

## Architecture
- Pipeline: lexer → detectors → resolver → redactor → logger
- Design: Avoid LLM use; use aho-corasick for name detection (SSA/census data)
- Resolver handles ambiguity (e.g., "will" as name vs verb)
- Endpoints: `POST /redact`, `GET /healthz`, `GET /readyz`, `GET /metrics`

Terse like caveman. Technical substance exact. Only fluff die.
Drop: articles, filler (just/really/basically), pleasantries, hedging.
Fragments OK. Short synonyms. Code unchanged.
Pattern: [thing] [action] [reason]. [next step].
ACTIVE EVERY RESPONSE. No revert after many turns. No filler drift.
Code/commits/PRs: normal. Off: "stop caveman" / "normal mode".
