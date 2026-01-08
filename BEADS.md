# BEADS Workflow & Issue Tracking

> This project uses **bd** (beads) for issue tracking. Run `bd onboard` to get started.
>
> Reference: [steveyegge/beads](https://github.com/steveyegge/beads)

## Quick Reference

```bash
bd ready              # Find unblocked work (use --json for agents)
bd show <id>          # View issue details
bd create "Title" -t bug|feature|task -p 0-4  # Create issue
bd update <id> --status in_progress  # Claim work
bd close <id> --reason "Done"        # Complete work
bd dep add <id1> <id2> --type blocks # Link dependencies
bd export             # Export to .beads/issues.jsonl
bd sync               # Sync with git
```

## Core Concept

Beads solves the "50 First Dates" problem - agents wake up with no memory. Unlike markdown TODOs:
- Issues are **structured data**, not text to parse
- Stored as **JSONL in `.beads/`**, versioned like code
- **Hash-based IDs** (`bd-a1b2`) prevent merge collisions
- **Dependency tracking** ensures work happens in order

## Git Integration

**`.beads/` is tracked in git** - issues merge like code:

```bash
# Before committing code changes:
bd export -o .beads/issues.jsonl
git add .beads/
git commit -m "Your changes"
```

Only use `bd init --stealth` for personal/local tracking without committing.

## Session Start Protocol

At the beginning of each session:

```bash
bd stats              # Check project health
bd ready              # Show available work (unblocked)
```

## Session Completion Protocol

When ending a work session, complete ALL steps. **Work is NOT complete until `git push` succeeds.**

1. **File issues for remaining work**
   - Create issues for anything needing follow-up
   - Link discoveries: `bd dep add <new-id> <parent-id> --type discovered-from`

2. **Run quality gates** (if code changed)
   - Tests, linters, builds

3. **Update issue status**
   - Close finished work: `bd close <id> --reason "Done"`
   - Update in-progress items

4. **Export and push** (MANDATORY)
   ```bash
   bd export
   git add .beads/
   git pull --rebase
   git commit -m "Update beads" --allow-empty
   git push
   git status  # MUST show "up to date with origin"
   ```

5. **Verify**
   - All changes committed AND pushed
   - `git status` shows clean state

6. **Hand off**
   - Provide context for next session

## Critical Rules

- Work is NOT complete until `git push` succeeds
- NEVER stop before pushing - that leaves work stranded locally
- NEVER say "ready to push when you are" - YOU must push
- If push fails, resolve and retry until it succeeds
- Always use `--json` flag for agent integrations

## Init Modes

| Mode | Command | Use Case |
|------|---------|----------|
| Standard | `bd init` | Shared project, .beads/ committed |
| Stealth | `bd init --stealth` | Personal tracking, not committed |
| Contributor | `bd init --contributor` | OSS fork workflow |
| Team | `bd init --team` | Branch-based collaboration |
| Protected | `bd init --branch beads-sync` | Protected main branch |

## Dependency Types

| Type | Effect | Use Case |
|------|--------|----------|
| `blocks` | Affects ready queue | Hard dependency |
| `related` | Informational | Soft link |
| `parent-child` | Hierarchy | Epics/subtasks |
| `discovered-from` | Audit trail | Bugs found during work |

Only `blocks` dependencies affect `bd ready` output.

## Best Practices

1. **Claim before working**: `bd update <id> --status in_progress`
2. **Create discovery issues**: File bugs found during other work
3. **Link everything**: Use `discovered-from` for audit trail
4. **Export before commit**: `bd export` ensures JSONL is current
5. **Use JSON output**: `bd ready --json` for agent consumption
6. **Delete old closed**: Periodically clean up, they stay in git history
