# agent-statsig v1 Design

## Overview

Go CLI tool for managing Statsig feature flags, primarily consumed by LLMs (via Claude Code skill), also usable by humans. Wraps the Statsig Console API.

## Entities

### Feature Gates
- CRUD, enable/disable, archive, launch
- Rule manipulation: add/update/remove rules with criteria validation
- Rollout: convenience command for "Everyone" rule at X%
- Criteria discovery: list all 25 condition types and their valid operators
- Per-entity usage command with examples

### Dynamic Configs
- Same CRUD/lifecycle as gates
- Rule manipulation with return value support
- Full JSON Schema validation (santhosh-tekuri/jsonschema/v6, draft 2020-12) — catches type errors, missing required fields, unknown fields client-side before API call
- Per-entity usage command with schema validation examples

### Experiments
- CRUD, archive
- Lifecycle: start, reset, abandon, ship (make_decision)
- Ship requires group ID and decision reason
- Per-entity usage command with lifecycle documentation

### Segments
- CRUD, archive
- ID list management: get/add/remove IDs
- Types: id_list, rule_based
- Per-entity usage command

## Authentication

Two keys per project:
- **Console API key** (required): management operations (CRUD, rules, lifecycle)
- **Client API key** (optional): reserved for future evaluation features

Keys stored in macOS Keychain (service: `app.paulie.agent-statsig`) with JSON payload. Falls back to plaintext file on non-macOS (mode 0600).

## Output Model

- **Get (single + multi)**: `get <id>...` accepts 1..N ids; default output is NDJSON —
  one line per id in input order. An unresolvable id emits
  `{"@unresolved":{"id","reason","fixable_by","hint"?}}` on stdout (item-level miss,
  exits 0). `--format json|yaml` collapses to a `{"data":[…],"@unresolved":[…]}`
  envelope. Only command-level failures (auth/network) go to stderr with exit 1.
- **Lists**: JSON envelope by default (statsig list deliberately stays JSON, unlike gets);
  `--format jsonl` for NDJSON stream.
- **Errors**: always JSON to stderr with `{error, hint, fixable_by}`
- **Mutations**: `{status: "ok", ...}` confirmation to stdout
- **Debug**: `-d/--debug` logs `[debug] METHOD URL` to stderr on all API commands

## Error Classification

| Category | fixable_by | Examples |
|----------|-----------|----------|
| Bad input | agent | Invalid JSON, unknown criteria, wrong name |
| Auth/permission | human | Invalid API key, insufficient permissions |
| Transient | retry | Network error, rate limit (429), server error (5xx) |
| Not found | agent | Entity not found (suggests using list) |

## Flag Patterns

- **Repeatable flags** for multi-value inputs: `--value a --value b` (not comma-separated)
- **Default operator**: `--operator` defaults to `any` (case-insensitive match), matching Statsig UI
- **Repeatable environments**: `--env staging --env production`
- **Repeatable IDs**: `--id user1 --id user2`

## API Details

- Base URL: `https://statsigapi.net`
- Version header: `STATSIG-API-VERSION: 20240601`
- Pagination: page-number-based (`?limit=N&page=N`)
- Rate limits: ~100 mutations/10s per project

## Condition Types

25 universal types (fixed across all Statsig projects, defined in OpenAPI spec):
- Per-project customization via `custom_field` (arbitrary user attributes) and `unit_id` (custom ID types)
- Full mapping with operators in `internal/api/types.go`
- Discovery via `gate criteria` command

## Documentation Architecture

Progressive disclosure pattern:
1. `agent-statsig usage` — top-level overview, common workflows, global flags
2. `gate usage` / `config usage` / `experiment usage` / `segment usage` — detailed per-entity reference with examples
3. `gate criteria` — condition type discovery
4. `skills/agent-statsig/SKILL.md` — Claude Code skill with process guidance

## Testing

- DI via `shared.ClientFactory` override + a mock-server harness (v1 shipped `shared.SetupMockServer`; superseded by `clitest.Run`)
- Full CLI command tests using httptest mock servers
- JSON Schema validation tested with type checking, required fields, unknown fields
- 68%+ total coverage, 75-89% on CLI entity packages

## Future Considerations

- **Log reads**: streaming NDJSON for Statsig event logs
- **Overrides management**: gate/experiment override endpoints
- **Version history**: entity version listing
- **Custom unit ID types**: `unit_id` management commands
