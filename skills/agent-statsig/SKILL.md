---
name: agent-statsig
description: Manage Statsig feature gates, dynamic configs, experiments, and segments
triggers:
  - statsig
  - feature gate
  - feature flag
  - dynamic config
  - experiment
  - a/b test
  - segment
  - rollout
  - targeting rule
tools:
  allowed:
    - Bash
    - Read
    - Grep
    - Glob
---

# agent-statsig — Statsig Feature Flag CLI

Manage Statsig feature gates, dynamic configs, experiments, and segments via the Console API.

## When to Use

- User asks about a feature gate, feature flag, or rollout
- User wants to check, modify, or create targeting rules
- User asks about experiment status or wants to start/ship/abandon experiments
- User asks about dynamic config values or wants to modify them
- User wants to manage segment membership

## Process

### Always read before writing

1. **Inspect first**: Run `agent-statsig gate get <name>` (or config/experiment/segment get) to understand the current state before making changes
2. **Check rules**: Run `gate rule list <name>` to see rule IDs before updating/removing rules
3. **Validate criteria**: Run `gate criteria` if unsure which condition types or operators to use

### Making changes

1. Read the current state
2. Identify what needs to change (add rule, modify values, change rollout %)
3. Make the change with the appropriate command
4. Verify by reading again

### Error handling

All errors are JSON to stderr with a classification:
- `fixable_by: agent` — you made a typo or used wrong syntax. Read the hint and retry.
- `fixable_by: human` — credentials or permissions issue. Tell the user.
- `fixable_by: retry` — transient error. Wait and retry once.

## Quick Reference

```bash
# Explore (safe, read-only)
agent-statsig gate list [--search <text>] [--tag <tag>]
agent-statsig gate get <name>
agent-statsig config get <name>
agent-statsig experiment get <name>
agent-statsig segment get <name>

# Modify gates
agent-statsig gate enable <name>
agent-statsig gate disable <name>
agent-statsig gate rollout <name> --percent 50
agent-statsig gate rule add <name> --name "Rule" --criteria email --value user@co.com
agent-statsig gate rule update <name> --rule <id> --add-value new@co.com
agent-statsig gate rule remove <name> --rule <id>

# Modify configs (return values validated against schema)
agent-statsig config rule add <name> --name "Rule" --criteria email --value user@co.com --return-value '{"key":"val"}'

# Experiment lifecycle
agent-statsig experiment start <name>
agent-statsig experiment ship <name> --group <id> --reason "text"
agent-statsig experiment abandon <name> --reason "text"

# Segment IDs
agent-statsig segment ids add <name> --id user1 --id user2
agent-statsig segment ids remove <name> --id user1
```

## Detailed Reference

For full command details with examples, run per-entity usage:
```bash
agent-statsig gate llm-help
agent-statsig config llm-help
agent-statsig experiment llm-help
agent-statsig segment llm-help
agent-statsig usage                  # top-level overview
```

## Key Concepts

### Rules
- Rules are evaluated **top-to-bottom**; first matching rule wins
- Conditions within a rule are **AND-ed** together
- Multiple rules act as **OR** (first match wins)
- Default (no rule matches) = **fail** for gates, **defaultValue** for configs

### Condition Types
25 built-in types. Most common:
- `email` — match by email (operators: any, none, str_contains_any, str_contains_none)
- `user_id` — match by user ID
- `country` — match by country code
- `custom_field` — match any user attribute (use `--field` to specify which)
- `public` — matches everyone (used for rollout percentages)
- `passes_segment` / `passes_gate` — compose with other entities

Default operator is `any` (case-insensitive match). Run `gate criteria` for the full list.

### Dynamic Config Schemas
When a config has a JSON Schema, `--return-value` is validated before the API call.
This catches type errors, missing required fields, and unknown fields locally.

### Environments
Rules can be scoped to environments (staging, production, etc.) using `--env`.
A rule with no environments applies to all environments.

## Project Setup

If the project isn't configured yet:
```bash
agent-statsig project add <alias> --console-key <key> [--client-key <key>]
agent-statsig project test
```
Tell the user to get their Console API key from Statsig Console → Settings → Keys & Environments.
