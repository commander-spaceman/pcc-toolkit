# PCC Toolkit v2 — ME2 OT Dialogue Extraction Toolkit

## Instruction Entry Point

Use this file as the **operational entry point** for AI agents working on this project.

- Architecture and design lives in `DESIGN.md`.
- Session state lives in `.opencode/`.

## Session Lifecycle

Every agent session follows this protocol:

### Open

1. Read `.opencode/current.md` — understand where the last session left off.
2. Pick **one** task. If multiple are pending, take the highest-priority one.
3. Update `.opencode/current.md`: set phase, task, branch, plan.
4. If using Notion, move the phase page to `In Progress`.

### Work

- Document in `.opencode/current.md` as you go, not at the end.
- Work incrementally, following the phase order in `DESIGN.md`.
- Prefer additive, low-intrusion changes over replacements.
- Keep modifications isolated and reversible where possible.
- Avoid touching unrelated systems.

### Close

1. Verify deliverables against `.opencode/checkpoints.md`.
2. If complete: mark phase/task as `Done` in Notion.
3. Move `.opencode/current.md` summary to `.opencode/history.md` (append-only).
4. Clear `.opencode/current.md` back to the template.
5. Push branch to remote.
6. No temporary files, no debug prints, no orphaned TODOs.

## Read First

1. `DESIGN.md` — complete architecture, AST spec, migration plan, dependencies.
2. `.opencode/checkpoints.md` — what "done" looks like.

## Repository Map

| Path | What it contains | When to read |
|------|-----------------|--------------|
| `DESIGN.md` | Full architecture, AST spec, dependencies, migration plan | Before any work |
| `.opencode/current.md` | Active session state | Every session start |
| `.opencode/history.md` | Append-only session log | For historical context |
| `.opencode/checkpoints.md` | Objective completion criteria | Before closing any phase |
| `.opencode/conventions.md` | Code style rules (Go + Python) | Before writing code |
| `.opencode/verification.md` | How to prove work is correct | Before marking a task done |
| `core/` | Go engine — all domain logic | For implementing parsing, AST, layout, evidence |
| `cli/` | Python CLI — thin dispatch wrapper | For CLI arg parsing and formatting |
| `gui/` | Python GUI — thin renderer | For ImGui views and interaction |
| `tests/golden/` | Known-good output files for regression | During port validation |
| `samples/` | Real ME2 OT PCC/TLK files (gitignored) | Input for golden tests |

## Operational Rules (High Impact)

- **Always consult the LegendaryExplorer repository** when implementing features. Use GitHub search against `github.com/ME3Tweaks/LegendaryExplorer` to understand how the official tool handles PCC parsing, TLK resolution, dialogue editing, and graph rendering. LEX is the reference implementation — match its behavior unless `DESIGN.md` specifies otherwise.
- Go core contains ALL domain logic. Python CLI and GUI are thin layers only.
- All Go ↔ Python communication is JSON over stdout.
- One feature at a time. Validate equivalence against old toolkit before moving on.
- Golden files are the structural contract. Never edit them manually.

## Language Policy

- Use English for all repository-facing content.
- Documentation, code comments, variable names, function/class identifiers, user-facing strings, error messages, and test descriptions should be written in English.
- **Exception**: Notion Kanban pages and reports may be written in Spanish.

## Notion Tracking Workflow

The Kanban board `PCC Dialog Toolkit - Kanban` is the execution log for toolkit development.

- **When starting a new phase**: create a new page in the Kanban. Set `Phase` to the phase name, `Status` to `TO-DO`.
- **When beginning work**: move `Status` to `In Progress`. Update `.opencode/current.md`.
- **During the phase**: keep the Notion page updated with scope, deliverables, risks, and verification notes as work advances.
- **When complete**: move `Status` to `Done`. Verify against `.opencode/checkpoints.md` first.
- If extra tasks appear outside the original phase plan, add them as new Kanban items.
- Keep Notion updates aligned with repository state (do not mark `Done` without corresponding code/docs progress).

## Build, Test, and Lint

- **Go core** (`core/`): `go test ./...`, `go build ./cmd/pcc-core/`
- **Python CLI** (`cli/`): `pytest` (once implemented)
- **Python GUI** (`gui/`): `pytest` (once implemented)

No top-level build command exists yet. When one is added, document it here.

## Guiding Question

> "Does this match how LegendaryExplorer handles it?"
