# agent-statsig

Statsig feature flag CLI for AI agents. Wraps the Statsig Console API (v1, base URL `https://statsigapi.net`).

## Project Structure

```
cmd/agent-statsig/main.go     Entry point (version stamped via ldflags)
internal/
  api/                         Statsig Console API HTTP client
    client.go                  Base HTTP client, auth, error classification
    types.go                   Shared types: Gate, DynamicConfig, Experiment, Segment, Rule, Condition
    gates.go                   Gate endpoints
    configs.go                 Dynamic config endpoints
    experiments.go             Experiment endpoints (incl. lifecycle: start/reset/abandon/ship)
    segments.go                Segment endpoints (incl. ID list management)
  cli/                         Cobra command tree
    root.go                    Global flags (--project, --format, --timeout), command registration
    usage.go                   LLM reference card
    shared/shared.go           WithClient helper, project resolution, paginated list output
    project/project.go         Project CRUD: add, update, remove, list, set-default, test
    gate/gate.go               Gate commands: list, get, create, delete, enable/disable, archive, launch, update, rollout, check, criteria
    gate/rule.go               Gate rule subcommands: list, add, update, remove + criteria validation
    config/config.go           Dynamic config commands (same shape as gate)
    config/rule.go             Config rule subcommands + JSON schema validation for return values
    experiment/experiment.go   Experiment commands incl. start, reset, abandon, ship
    segment/segment.go         Segment commands incl. ids get/add/remove
  config/config.go             App config file I/O (~/.config/agent-statsig/config.json)
  credential/credential.go     Keychain-backed credential storage (console key + client key per project)
  errors/errors.go             APIError type with fixable_by classification
  output/output.go             JSON/YAML/NDJSON formatters, WriteError, PrintJSON
```

## Key Design Decisions

- **Single binary, zero runtime deps**: pure Go, CGO_ENABLED=0
- **Structured JSON output to stdout**: errors to stderr as `{error, hint, fixable_by}` JSON
- **macOS Keychain**: credentials stored in system Keychain (service: `app.paulie.agent-statsig`); falls back to file on Linux/Windows
- **Project aliases**: one credential per project (console key + optional client key), switchable via `--project` flag or `AGENT_STATSIG_PROJECT` env var
- **Cobra CLI framework**: same pattern as other agent-* tools
- **API version pinned**: `STATSIG-API-VERSION: 20240601` header sent on all requests

## Dev Workflow

```bash
make build          # Build binary
make test           # Run all tests
make test-short     # Skip integration tests
make lint           # golangci-lint
make fmt            # gofmt + goimports
make dev ARGS="..." # Run in dev mode
make vet            # go vet
```

## Adding a New Entity Type

1. Add API methods in `internal/api/<entity>.go`
2. Add types to `internal/api/types.go`
3. Create CLI package `internal/cli/<entity>/`
4. Register in `internal/cli/root.go`
5. Update reference card in `internal/cli/usage.go`
6. Add tests

## Releasing

Uses goreleaser. Platforms: darwin/linux/windows × amd64/arm64.
Distributed via GitHub releases and Homebrew (`shhac/tap`).

## Statsig API Notes

- Base URL: `https://statsigapi.net`
- Auth: `STATSIG-API-KEY` header with Console API key
- Entity IDs: all endpoints accept entity `name` or internal Statsig ID
- Pagination: `?limit=N&page=N` (1-indexed), response includes `pagination.nextPage`
- Rate limits: ~100 mutations/10s, ~900/15min per project
- PATCH = partial update, POST to `/{id}` = full replacement
- Rules: conditions within a rule are AND-ed; rules evaluated top-to-bottom (first match wins)
- 25 condition types with type-specific operators (see `api/types.go` for full mapping)
