# agent-statsig

Statsig feature flag CLI for AI agents. Manage gates, dynamic configs, experiments, and segments from the command line with structured JSON output optimized for LLM consumption.

## Features

- **Four entity types**: feature gates, dynamic configs, experiments, segments
- **Full CRUD + lifecycle**: create, read, update, delete, enable/disable, archive, and entity-specific operations (rollout, start/ship experiments, manage segment IDs)
- **Rule manipulation**: add, update, and remove targeting rules with criteria validation
- **JSON Schema validation**: dynamic config return values validated client-side against the config's schema
- **Structured output**: JSON/YAML/NDJSON output with classified errors (`fixable_by: agent|human|retry`)
- **Secure credential storage**: macOS Keychain integration, multi-project support via aliases
- **Progressive documentation**: top-level overview → per-entity reference → criteria discovery
- **Zero runtime dependencies**: single compiled binary

## Installation

### Homebrew

```bash
brew install shhac/tap/agent-statsig
```

### Go Install

```bash
go install github.com/shhac/agent-statsig/cmd/agent-statsig@latest
```

### Build from Source

```bash
git clone https://github.com/shhac/agent-statsig.git
cd agent-statsig
make build
```

## Quick Start

### 1. Add a project

Get your Console API key from **Statsig Console → Settings → Keys & Environments**.

```bash
agent-statsig project add myproject --console-key "console-xxx" --client-key "client-xxx"
```

### 2. Test connectivity

```bash
agent-statsig project test
```

### 3. Explore

```bash
agent-statsig gate list
agent-statsig gate get my_feature_gate
agent-statsig config list
agent-statsig experiment list --search "checkout"
```

### 4. Modify

```bash
# Enable a gate
agent-statsig gate enable my_gate

# Roll out to 50%
agent-statsig gate rollout my_gate --percent 50

# Add a targeting rule
agent-statsig gate rule add my_gate \
  --name "Internal team" \
  --criteria email \
  --operator str_contains_any \
  --value "@mycompany.com"

# Multiple values
agent-statsig gate rule add my_gate \
  --name "Beta users" \
  --criteria email \
  --value "alice@example.com" \
  --value "bob@example.com"

# Start an experiment
agent-statsig experiment start my_experiment
```

## Documentation

The CLI has layered documentation for progressive discovery:

```bash
agent-statsig usage              # Top-level overview + common workflows
agent-statsig gate usage            # Feature gates detailed reference
agent-statsig config usage          # Dynamic configs + schema validation
agent-statsig experiment usage      # Experiments + lifecycle
agent-statsig segment usage         # Segments + ID list management
agent-statsig gate criteria      # List all 25 condition types + operators
```

## Output Formats

```bash
agent-statsig gate get my_gate                    # JSON (default, pretty)
agent-statsig gate list --format jsonl             # NDJSON (one per line)
agent-statsig gate get my_gate --format yaml       # YAML
```

## Error Output

All errors are written to stderr as structured JSON:

```json
{
  "error": "Gate 'my_gat' not found",
  "hint": "Check the entity name — use 'list' to see available items",
  "fixable_by": "agent"
}
```

| `fixable_by` | Meaning |
|-------------|---------|
| `agent` | Typo, wrong name, bad syntax — retry with a fix |
| `human` | Missing credentials, permission denied — needs human action |
| `retry` | Network error, rate limit, server error — wait and retry |

## Multi-Project Support

```bash
agent-statsig project add production --console-key "console-xxx"
agent-statsig project add staging --console-key "console-yyy"
agent-statsig project set-default staging
agent-statsig -p production gate list
```

## Claude Code Skill

Install as a Claude Code skill for automatic discovery:

```bash
npx skills add shhac/agent-statsig
```

## License

MIT
