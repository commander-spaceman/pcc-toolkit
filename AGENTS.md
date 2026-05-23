# PCC Toolkit v2 - ME2 OT Dialogue Extraction Toolkit

## Instruction Entry Point

Use this file as the operational entry point for AI agents working on this project.

- Architecture and design live in `PRD.md`.
- Persistent agent context should live in the `memory` MCP, not in session lifecycle files.
- `.opencode/` is reserved for MCP runtime files, not project instructions.

## Operating Model

Agents should work task-by-task, using the smallest correct change.

1. Understand the request and inspect the relevant code before changing files.
2. Use `memory` MCP for durable context worth keeping across sessions.
3. Follow the conventions and verification rules in this file.
4. Keep changes isolated, incremental, and reversible where possible.
5. Do not push, commit, or update external trackers unless explicitly asked.

There is no required session lifecycle file. Do not maintain `.opencode/current.md`, `.opencode/history.md`, or `.opencode/checkpoints.md`.

## Read First

1. `PRD.md` - architecture, AST spec, migration plan, dependencies.
2. `AGENTS.md` - operating model, conventions, verification, and repository map.

## Repository Map

| Path | What it contains | When to read |
|------|------------------|--------------|
| `PRD.md` | Full architecture, AST spec, dependencies, migration plan | Before architecture or feature work |
| `.opencode/start-github-mcp.ps1` | GitHub MCP wrapper that loads `.env` | MCP runtime only |
| `.opencode/memory.jsonl` | Local memory MCP storage | MCP runtime only |
| `build/` | Local compiled binaries, gitignored | Runtime artifacts only |
| `core/` | Go engine; domain logic belongs here | Parsing, AST, layout, evidence, validation |
| `cli/` | Python CLI; thin dispatch wrapper | CLI arg parsing and output formatting |
| `gui/` | Python GUI; thin renderer | ImGui views and interaction |
| `tests/golden/` | Known-good regression outputs | Port validation and parser regression checks |
| `tests/regression/` | Probe/regression runners | Golden or probe validation workflows |
| `tests/fixtures/synthetic/` | Synthetic test fixture builders/data | Unit tests that do not require game files |
| `samples/` | Real ME2 OT PCC/TLK files, gitignored | Local input for golden tests |
| `output/` | Generated local outputs, gitignored except `.gitkeep` | Runtime artifacts only |

## LegendaryExplorer Reference

The GitHub MCP is configured and should be used as the primary way to consult:

```text
ME3Tweaks/LegendaryExplorer
```

Always consult LegendaryExplorer when implementing or changing behavior for PCC parsing, TLK resolution, dialogue editing, conversation graph layout, package structures, or validation semantics.

Treat LegendaryExplorer as the reference implementation unless `PRD.md` explicitly says otherwise. The guiding question is:

> Does this match how LegendaryExplorer handles it?

## Operational Rules

- Scope is Mass Effect 2 Original Trilogy only. Do not add LE1, LE2, LE3, ME1, or ME3 behavior unless the task explicitly changes project scope.
- ME2 OT compressed package support is LZO-only per `PRD.md`.
- Go core contains all domain logic. Python CLI and GUI are thin layers only.
- Go core writes success payloads as JSON to stdout and error payloads as JSON to stderr.
- One feature at a time. Validate equivalence against old toolkit or golden files before moving on.
- Golden files are the structural contract. Do not edit them manually unless the task is explicitly to regenerate and justify them.
- Prefer additive, low-intrusion changes over broad rewrites.
- Avoid touching unrelated systems.

## Language Policy

- Use English for repository-facing content.
- Documentation, code comments, identifiers, user-facing strings, error messages, and test descriptions should be in English.
- Spanish is fine for local planning notes or conversation with the user.

## Code Conventions

- Match existing local patterns before introducing new structure.
- Add comments only to explain non-obvious why, subtle invariants, or documented workarounds.
- Remove debug `print()` and `fmt.Println()` calls before finishing.
- Avoid orphaned TODOs. Every TODO must reference a concrete follow-up.
- Go: run `go fmt ./...`; return `error` as the last value for fallible functions; do not panic in library code.
- Go: use `CamelCase` for exported names and `camelCase` for unexported names.
- Python: target Python 3.11+ style, PEP 8, max 100 columns, type hints with `| None`, built-in generics, and double-quoted strings.
- Tests: prefer concrete expected outputs over tests that only assert no crash.

## Build, Test, and Lint

- Go core tests: from `core/`, run `go test ./...`.
- Go formatting: from `core/`, run `go fmt ./...`.
- Python CLI/GUI tests: `pytest` once implemented.

No single top-level build command is guaranteed. If one is added, document it here.

## Verification Checklist

- Choose the smallest verification set that proves the change.
- From `core/`, run `go test ./...` for core parser, domain logic, or integration-sensitive Go changes.
- Run `pytest` when Python code is affected and tests exist.
- Compare against golden files when parser output changes.
- Consult LegendaryExplorer through the GitHub MCP when LEX semantics matter.
- Final summaries should mention what changed, what verification ran, and any skipped or failing checks.
