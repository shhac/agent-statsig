# agent-statsig

Statsig feature flag CLI for AI agents. Wraps the Statsig Console API (v1, base URL `https://statsigapi.net`). Manages gates, dynamic configs, experiments, segments, and tags.

## Project Structure

```
cmd/agent-statsig/main.go     Entry point (version stamped via ldflags)
internal/
  api/                         Statsig Console API HTTP client
    client.go                  Base client (BaseURL field for testability), doAndDecode[T] generic
    types.go                   Shared types: Gate, DynamicConfig, Experiment, Segment, Tag, Rule, Condition
    tags.go                    Tag endpoints
    gates.go                   Gate endpoints
    configs.go                 Dynamic config endpoints
    experiments.go             Experiment endpoints (incl. lifecycle: start/reset/abandon/ship)
    segments.go                Segment endpoints (incl. ID list management)
  cli/                         Cobra command tree
    root.go                    Global flags (--project, --format, --timeout), command registration
    usage.go                   Top-level LLM reference card (progressive disclosure → per-entity usage)
    shared/
      shared.go                WithClient (DI-ready via ClientFactory), project resolution, generics
      testhelper.go            SetupMockServer for httptest-based CLI testing
    project/project.go         Project CRUD: add, update, remove, list, set-default, test
    gate/
      gate.go                  Gate commands: list, get, create, delete, enable/disable, archive, launch, update, check
      rollout.go               Rollout command + FindPublicRule pure helper
      criteria.go              Criteria listing (25 condition types + operators)
      rule.go                  Rule subcommands: list, add, update, remove + FindRuleByID, BuildRuleUpdate, MergeConditionValues
      usage.go                 Per-entity reference card
    config/
      config.go                Dynamic config commands
      rule.go                  Config rule subcommands + ValidateAgainstSchema (santhosh-tekuri/jsonschema/v6)
      usage.go                 Per-entity reference card
    experiment/
      experiment.go            Experiment commands incl. start, reset, abandon, ship
      usage.go                 Per-entity reference card
    segment/
      segment.go               Segment commands incl. ids get/add/remove
      usage.go                 Per-entity reference card
    tag/
      tag.go                   Tag commands: list, get, create, update, delete
      usage.go                 Per-entity reference card
  config/config.go             App config file I/O (~/.config/agent-statsig/config.json)
  credential/
    credential.go              Credential storage (index file + keychain integration)
    keychain.go                macOS Keychain operations (keychainStore/Get/Delete)
  errors/errors.go             APIError type with fixable_by classification
  output/output.go             JSON/YAML/NDJSON formatters, WriteError, PrintJSON
skills/
  agent-statsig/SKILL.md       Claude Code skill definition
```

## Key Design Decisions

- **Single binary, zero runtime deps**: pure Go, CGO_ENABLED=0
- **Structured JSON output to stdout**: errors to stderr as `{error, hint, fixable_by}` JSON
- **Repeatable flags**: `--value`, `--env`, `--id` use StringArrayVar (not comma-separated) to handle values with special characters
- **Default operator**: `--operator` defaults to `any` (case-insensitive match), matching Statsig UI behavior
- **JSON Schema validation**: `santhosh-tekuri/jsonschema/v6` for full draft 2020-12 compliance on dynamic config return values
- **DI for testing**: `shared.ClientFactory` override enables httptest-based CLI command tests
- **Condition types are universal**: the 25 types are a platform-level constant (not per-project). Per-project customization uses `custom_field` and `unit_id`
- **macOS Keychain**: credentials stored in system Keychain (service: `app.paulie.agent-statsig`); falls back to file on Linux/Windows
- **Progressive documentation**: `usage` → per-entity `usage` → `gate criteria` for condition discovery
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

## Testing

Tests use `shared.SetupMockServer(t, handler)` to inject an httptest server via `ClientFactory`. This enables full CLI command testing without real credentials.

```go
func TestGateGet(t *testing.T) {
    out, _ := runGateCmd(t, func(w http.ResponseWriter, r *http.Request) {
        w.Write(entityJSON(api.Gate{Name: "my_gate"}))
    }, "gate", "get", "my_gate")
    // assert on out...
}
```

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
- Tags endpoint: `/console/v1/tags` — CRUD for organizational tags applied to entities
- Tags on entities are validated before create/update to prevent broken state
- 25 condition types with type-specific operators (see `api/types.go` for full mapping)
- Condition types are universal across all Statsig projects (not configurable per-project)
