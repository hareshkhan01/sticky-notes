# Taste

## Workflow
- Prefers to run commands and perform actions themselves; wants guidance and instructions rather than having the assistant execute things directly. Confidence: 0.85

## Preferences
- Prefers semantic versioning (e.g. 1.0.0) over raw commit hashes for version display. Confidence: 0.85
- Designs tools with maximum user flexibility: composable CLI flags/options with sensible defaults that users can override later (e.g., `stick init fish latest` as shorthand + `stick startup --limit N` for runtime override). Explicitly values giving end-users power and control. Confidence: 0.9
- Expects configuration changes to be achievable via CLI commands rather than requiring manual config file editing. Wants toggle-able settings exposed as subcommands/flags. Confidence: 0.85
- Wants defensive config validation: user-set config values (even via manual YAML edits) should be hard-capped and enforced at load time to prevent unreasonable values. Confidence: 0.8
