# Issue Tracker: Beads

Issues for this repo live in Beads and are stored under `.beads/` in git.

Use the `br` CLI for issue operations unless the repo instructions for a specific task say otherwise.

## Conventions

* Beads is the single source of truth for task status, priority, type, and dependencies.
* Use Beads issue IDs as stable identifiers in commits, notes, and cross-agent threads.
* Use labels for triage states; see `triage-labels.md`.
* After Beads edits, run `br sync --flush-only`.
* If JSONL looks stale, trust `br show <id>` first, then sync.

## Common commands

```bash
br ready
br list --status open
br show <id>
br create --title "..." --type task --priority 2
br update <id> --status in_progress
br close <id> --reason "Completed"
br sync --flush-only
```

## Dependencies

For child issues, do not use `br list --parent`.

Use:

```bash
br show <parent> --json | jq -r '.[0].dependents[] | select(.dependency_type=="parent-child") | .id'
```

To create a normal random child issue ID, create the child first, then link it:

```bash
br create --title "..." --type task --priority 2
br dep add <child-id> <parent-id> --type parent-child
```

Only use `br create --parent <parent>` if dotted child IDs are desired.

## Long descriptions and notes

For multiline descriptions, acceptance criteria, or notes, avoid shell-sensitive inline strings. Use quoted heredocs.

Use `* ` bullets inside heredocs because leading `- ` lines can be parsed as CLI flags by some commands.

```bash
notes=$(cat <<'EOF'
Summary text.

* First item
* Second item with `backticks`, $vars, and $(commands) kept literal
EOF
)

br update <id> --add-label ready-for-agent --notes "$notes"
```

## When a skill says "publish to the issue tracker"

Create or update Beads issues with `br`. If the work comes from a PRD or plan, create independently actionable issues, set their type and priority, and add dependencies where sequencing matters.

## When a skill says "fetch the relevant ticket"

Use `br show <id>` for human-readable context, or `br show <id> --json` when structured data is needed.
