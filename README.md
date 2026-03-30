# agent-statsig

Statsig feature flag CLI for AI agents. Manage gates, dynamic configs, experiments, and segments from the command line with structured JSON output optimized for LLM consumption.

## Features

- **Four entity types**: feature gates, dynamic configs, experiments, segments
- **Full CRUD + lifecycle**: create, read, update, delete, enable/disable, archive, and entity-specific operations (rollout, start/ship experiments, manage segment IDs)
- **Rule manipulation**: add, update, and remove targeting rules with criteria validation
- **Structured output**: JSON/YAML/NDJSON output with classified errors (`fixable_by: agent|human|retry`)
- **Secure credential storage**: macOS Keychain integration, multi-project support via aliases
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

Get your Console API key from **Statsig Console → Settings → API Keys**.

```bash
agent-statsig project add myproject --console-key "console-xxx"
```

Optionally add a client key for evaluation features:

```bash
agent-statsig project update myproject --client-key "client-xxx"
```

### 2. Test connectivity

```bash
agent-statsig project test
```

### 3. List gates

```bash
agent-statsig gate list
agent-statsig gate list --tag core --search "onboarding"
```

### 4. Inspect a gate

```bash
agent-statsig gate get my_feature_gate
```

### 5. Modify a gate

```bash
# Enable/disable
agent-statsig gate enable my_feature_gate

# Set rollout percentage
agent-statsig gate rollout my_feature_gate --percent 50

# Add a targeting rule
agent-statsig gate rule add my_feature_gate \
  --name "Internal team" \
  --criteria email \
  --operator str_contains_any \
  --values "@mycompany.com" \
  --pass-percent 100
```

## Usage Reference

Run `agent-statsig usage` for the full LLM-optimized reference card.

### Global Flags

| Flag | Description |
|------|-------------|
| `-p, --project <alias>` | Project alias (or set `AGENT_STATSIG_PROJECT`) |
| `--format json\|yaml\|jsonl` | Output format (default: json) |
| `--timeout <ms>` | Request timeout in milliseconds |

### Commands

| Command | Description |
|---------|-------------|
| `project` | Manage Statsig project connections |
| `gate` | Manage feature gates |
| `config` | Manage dynamic configs |
| `experiment` | Manage experiments |
| `segment` | Manage segments |
| `usage` | Show LLM-optimized reference card |

### Error Output

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
# Add multiple projects
agent-statsig project add production --console-key "console-xxx"
agent-statsig project add staging --console-key "console-yyy"

# Set default
agent-statsig project set-default staging

# Override per-command
agent-statsig -p production gate list

# Or via environment variable
AGENT_STATSIG_PROJECT=production agent-statsig gate list
```

## License

MIT
