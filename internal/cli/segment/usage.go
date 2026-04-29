package segment

const segmentUsage = `SEGMENTS — Reference

A segment is a reusable group of users. Define targeting once, then reference
it from multiple gates or experiments via passes_segment / fails_segment
conditions. Segment types: id_list (explicit user IDs) or rule_based (rules).

READ
  segment list [--tag <tag>] [--search <text>] [--limit N] [--page N]
  segment get <name>

MODIFY
  segment create <name> [--description <text>] [--type id_list|rule_based]
  segment delete <name>
  segment archive <name>

ID LIST MANAGEMENT (for id_list segments)
  segment ids get <name>             List IDs in the segment
  segment ids add <name> --id <id>   Add IDs (repeatable: --id user1 --id user2)
  segment ids remove <name> --id <id>

EXAMPLES
  # Create a segment for internal team members
  segment create internal_team --type id_list

  # Add users
  segment ids add internal_team --id user123 --id user456

  # Reference from a gate rule
  gate rule add my_gate --name "Internal" --criteria passes_segment --value internal_team
`
