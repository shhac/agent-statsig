package config

const configUsage = `DYNAMIC CONFIGS — Reference

A dynamic config returns a JSON object whose value varies by targeting rules.
Like gates, rules are evaluated top-to-bottom and use the same condition types.
Unlike gates, each rule has a returnValue (the JSON payload for matching users).

Configs may have a JSON Schema (draft 2020-12). When present, return values
and defaultValue are validated client-side before the API call — type errors,
missing fields, and unknown fields are caught with helpful hints.

READ
  config list [--tag <tag>] [--search <text>] [--limit N] [--page N]
  config get <name>                  Full config with rules, values, and schema
  config rule list <config>          List rules with their IDs
  config schema get <config>         The JSON Schema in clean object form

MODIFY
  config create <name> [--description <text>] [--tag <tag>...]
  config delete <name>
  config enable <name>
  config disable <name>
  config archive <name>
  config update <name> <json>        Raw JSON partial update (escape hatch)
  config schema set <config> <json>  Set/replace the JSON Schema
  config schema clear <config>       Remove the JSON Schema

RULE MANAGEMENT
  config rule add <config>
    --name <rule-name>
    --criteria <type>
    [--operator <op>]                Default: any
    [--value <v>]                    Repeatable
    --return-value <json>            The JSON value this rule returns
    [--pass-percent N]               Default: 100
    [--env <environment>]            Repeatable
    [--field <name>]                 For custom_field criteria

  config rule update <config>
    --rule <rule-id>
    [--pass-percent N --set-percent]
    [--return-value <json>]

  config rule remove <config> --rule <rule-id>

EXAMPLES
  # Add a rule returning custom values for internal users
  config rule add my_config \
    --name "Internal" \
    --criteria email \
    --operator str_contains_any \
    --value "@company.com" \
    --return-value '{"enabledGlobally": true, "allowOrganizers": []}' \
    --env staging

SETTING A SCHEMA
  Pass the schema as a plain JSON object — the CLI compiles it (draft 2020-12)
  and handles the API's string encoding for you:
    config schema set my_config '{"type":"object","properties":{"theme":{"type":"string"}},"required":["theme"]}'

  Before setting, existing values are checked against the new schema:
    ✗ defaultValue or a rule returnValue does not conform → blocked, each
      violation listed. Fix the values first, or pass --force to set anyway.
    • Absent/null values are skipped (nothing to validate).
    • A present empty object {} is a real value — it fails if the schema
      has required fields.

  In raw 'config update' JSON, a schema must be a JSON-encoded STRING
  (API requirement) — prefer 'config schema set', which does this for you.

SCHEMA VALIDATION OF VALUES
  If the config has a schema, --return-value (rule add/update) and
  defaultValue/rules in raw updates are validated before sending:
    ✗ Missing required field → "Schema requires: field1, field2"
    ✗ Unknown field          → "Known fields: field1, field2"
    ✗ Wrong type             → "expected boolean, got string"
  All errors are fixable_by: agent with hints. --force skips the client-side
  check (the server still has the final say).
`
