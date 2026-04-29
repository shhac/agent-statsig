package tag

const tagUsage = `TAGS — Reference

Tags are labels that organize gates, configs, experiments, and segments.
Each tag has a name, description, and isCore flag. Tags must be created
before they can be applied to entities.

READ
  tag list [--search <text>] [--limit N] [--page N]
  tag get <id>

MODIFY
  tag create <name> [--description <text>] [--is-core]
  tag update <id> [--name <name>] [--description <text>] [--is-core]
  tag delete <id>

APPLYING TAGS TO ENTITIES
  Tags are applied by name when creating or updating gates, configs, and experiments.
  The --tag flag is repeatable. Tags are validated before the entity is created/updated.

  gate create <name> --tag <tag-name> [--tag <tag-name>...]
  gate update <name> '{}' --tag <tag-name>
  config create <name> --tag <tag-name>
  experiment create <name> --tag <tag-name>

FILTERING BY TAG
  Entity list commands support filtering by tag:
  gate list --tag <tag-name>
  config list --tag <tag-name>
  experiment list --tag <tag-name>
  segment list --tag <tag-name>

EXAMPLES
  # Create a core tag
  tag create "mobile" --description "Mobile platform features" --is-core

  # List all tags
  tag list

  # Create a gate with tags
  gate create my_gate --tag mobile --tag ios

  # Update a gate's tags
  gate update my_gate '{}' --tag mobile --tag android
`
