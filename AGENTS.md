# PCC Toolkit v2 - ME2 OT Dialogue Extraction Toolkit

## Instruction Entry Point

Use this file as the operational entry point for AI agents working on this project.

- Architecture and design live in `docs/PRD-INSPECT.md` (read/inspect) and `docs/PRD-EDITING.md` (edit/write).
- Fast repository navigation lives in `MAP.md`.
- Persistent agent context should live in the `memory` MCP, not in session lifecycle files.
- `.opencode/` is reserved for MCP runtime files, not project instructions.

## Operating Model

Agents should work task-by-task, using the smallest correct change.

1. Understand the request and inspect the relevant code before changing files.
2. Use `memory` MCP for durable context worth keeping across sessions.
3. Follow the conventions and verification rules in this file.
4. Keep changes isolated, incremental, and reversible where possible.
5. Do not push, commit, or update external trackers unless explicitly asked.

## Read First

1. `docs/PRD-INSPECT.md` - canonical architecture, contracts, domain model, dependencies, and verification for read & inspect.
2. `docs/PRD-EDITING.md` - editing, writing, round-trip fidelity, and batch edit requirements.
3. `MAP.md` - current repository structure and module navigation.
4. `AGENTS.md` - operating model, conventions, verification, and repository map.

## Repository Map

| Path                        | What it contains                                                                    | When to read                                                          |
| --------------------------- | ----------------------------------------------------------------------------------- | --------------------------------------------------------------------- |
| `docs/PRD-INSPECT.md`       | Canonical architecture, domain model, contracts, dependencies, and verification     | Before architecture or feature work                                   |
| `docs/PRD-EDITING.md`       | Editing, writing, round-trip fidelity, and batch edit requirements                  | Before editing or write-path work                                     |
| `docs/CONTRACTS.md`         | JSON contracts for all pcc-core subcommands                                         | Before adding/modifying CLI commands or golden tests                  |
| `docs/BUILDING.md`          | Build, test, and release guide                                                      | Before build/release tooling changes                                  |
| `docs/REFERENCE.md`         | LegendaryExplorer reference notes and known divergences                             | When LEX semantics need clarification                                 |
| `MAP.md`                    | Current repository structure and navigation guide                                   | Before structural changes or multi-module work                        |
| `core/`                     | Go engine; domain logic belongs here                                                | Parsing, AST, layout, evidence, validation                            |
| `cli/`                      | Python CLI; thin dispatch wrapper                                                   | CLI arg parsing and output formatting                                 |
| `tests/test_golden.py`      | Golden regression checks for `pcc-core` output                                      | Contract and parser regression validation                             |
| `tests/test_smoke.py`       | CLI entry-point smoke tests                                                         | CLI packaging and command visibility checks                           |
| `tests/golden/`             | Known-good regression outputs                                                       | Port validation and parser regression checks                          |
| `tests/regression/`         | Probe/regression runners                                                            | Golden or probe validation workflows                                  |
| `tests/fixtures/synthetic/` | Synthetic test fixture builders/data                                                | Unit tests that do not require game files                             |
| `pyproject.toml`            | Python packaging metadata, version, script entry point                              | Packaging, version, dependency, or CLI registration changes           |
| `pytest.ini`                | Pytest configuration                                                                | Test discovery/configuration changes                                  |
| `requirements.txt`          | Local Python dependency list                                                        | Environment setup or dependency changes                               |
| `output/`                   | Real ME2 OT PCC/TLK files and generated local outputs, gitignored except `.gitkeep` | Local input for golden tests, real-file probes, and runtime artifacts |
| `scripts/`                  | Build and release automation scripts                                                | Before building pcc-core or creating a release                        |

## LegendaryExplorer Reference

The GitHub MCP is configured and should be used as the primary way to consult:

```text
ME3Tweaks/LegendaryExplorer
```

Always consult LegendaryExplorer when implementing or changing behavior for PCC parsing, TLK resolution, dialogue editing, conversation graph layout, package structures, or validation semantics.

Treat LegendaryExplorer as the reference implementation unless `docs/PRD-INSPECT.md` or `docs/REFERENCE.md` explicitly says otherwise. The guiding question is:

> Does this match how LegendaryExplorer handles it?

LegendaryExplorer is GPLv3-licensed and its `CONTRIBUTING.md` includes restrictions on low-value generative-AI contributions. Use it as a behavioral and semantic reference only: do not copy, paste, translate, or port its code into this repository unless the project intentionally accepts the resulting license obligations. Prefer documenting observed behavior, public file/class names consulted, and independently implemented logic.

## External Libraries

The Go core depends on three external libraries maintained in separate repositories:

| Library  | Repository                                    | Purpose                         |
| -------- | --------------------------------------------- | ------------------------------- |
| `me2lzo` | `github.com/commander-spaceman/me2lzo` v1.0.0 | LZO compression / decompression |
| `me2pcc` | `github.com/commander-spaceman/me2pcc`        | PCC package reading             |
| `me2tlk` | `github.com/commander-spaceman/me2tlk`        | TLK file reading                |

These libraries are imported via `core/go.mod` and are maintained independently
from this repository. Local clones for development live at sibling directories
relative to the workspace root:

| Library  | Local path                   |
| -------- | ---------------------------- |
| `me2lzo` | `..\dev\me2lzo` (or similar) |
| `me2pcc` | `..\dev\me2pcc` (or similar) |
| `me2tlk` | `..\dev\me2tlk` (or similar) |

Ask the user for the exact local path if unsure.

When a bug is traced to one of these libraries:

1. Fix the library in its own repository.
2. Tag or push the fix, then update the dependency in this project:
   ```powershell
   cd core
   go get github.com/commander-spaceman/<library>@<version-or-commit>
   go mod tidy
   ```
3. Run `go test ./...` from `core/` to confirm the fix resolves the issue.
4. Commit the updated `go.mod` and `go.sum` in this repository.

Do not vendor, fork, or inline code from these libraries into this repository.
The library boundaries keep each module independently testable and replaceable.

## Operational Rules

- Scope is Mass Effect 2 Original Trilogy only. Do not add LE1, LE2, LE3, ME1, or ME3 behavior unless the task explicitly changes project scope.
- ME2 OT compressed package support is LZO-only per `docs/PRD-INSPECT.md`.
- Go core contains all domain logic. Python CLI is a thin layer only.
- Go core writes success payloads as JSON to stdout and error payloads as JSON to stderr.
- One feature at a time. Validate behavior against golden files, regression probes, and LegendaryExplorer semantics before moving on.
- Golden files are the structural contract. Do not edit them manually unless the task is explicitly to regenerate and justify them.
- `scan-evidence --biogame-root` expects a ME2-style `BioGame` root containing `CookedPC/` and optionally `DLC/`; a flat `output/` directory is not a substitute for scan workflows.
- Prefer additive, low-intrusion changes over broad rewrites.
- Avoid touching unrelated systems.
- Python commands must use this repository's virtual environment. On Windows, prefer `.venv\Scripts\python.exe -m pytest` and `.venv\Scripts\pcc-toolkit.exe`; do not run bare `pytest` or system Python unless the venv is unavailable and the user approves.
- Real ME2 OT test assets must be copied into `output/` from the local install at `C:\Program Files\EA Games\Mass Effect 2`. Do not run golden tests directly against the install tree, and never commit copied game files.

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
- Python: target Python 3.14+ style, PEP 8, max 100 columns, type hints with `| None`, built-in generics, and double-quoted strings.
- Tests: prefer concrete expected outputs over tests that only assert no crash.

## Build, Test, and Lint

- Go core tests: from `core/`, run `go test ./...`.
- Go formatting: from `core/`, run `go fmt ./...`.
- Python CLI tests: from the repository root, run `.venv\Scripts\python.exe -m pytest`.
- Real-file regression probes: copy only the needed ME2 OT `.pcc`/`.tlk` files from `C:\Program Files\EA Games\Mass Effect 2` into `output/`, then run `.venv\Scripts\python.exe tests\regression\run_probes.py --samples-dir output`.

No single top-level build command is guaranteed. If one is added, document it here.

## Distribution and Release

### Build pcc-core (Go core binary)

**Standalone (no Python required):**

```powershell
.\scripts\build.ps1 -Version "0.2.0"
```

Builds a single static binary to `build/pcc-core.exe` with ldflags version injection.
The default build target is Windows.

**Via Python CLI (requires installed CLI):**

```
pcc-toolkit dev build-core
```

### Install CLI

From the repository root, inside the project virtual environment:

```
.venv\Scripts\python.exe -m pip install -e .[cli]    # CLI
```

After installation, the `pcc-toolkit` command is available.

### Create a local release

```powershell
.\scripts\release.ps1 -Version "0.2.0" -Archive
```

This builds the Go core binary and creates:

- `release/pcc-toolkit-v0.2.0-windows-amd64/` directory with:
  - `pcc-core.exe` — the Go engine binary
  - `INSTALL.txt` — install instructions
- `release/pcc-toolkit-v0.2.0-windows-amd64.zip` (if `-Archive`)

### Published binaries

The following binaries are produced for distribution (never committed to the repository):

| Binary         | Platform        | Built from                          |
| -------------- | --------------- | ----------------------------------- |
| `pcc-core.exe` | Windows (amd64) | `core/cmd/pcc-core/` via `go build` |

These are excluded from version control via `.gitignore` rules (`*.exe`, `build/`).
For public distribution they are attached to GitHub Releases as Windows archives.

### Reproducibility

- Go dependencies are locked in `core/go.sum`.
- Version is injected via `-ldflags "-X main.version=x.y.z"` at build time.
- Python dependencies are declared in `pyproject.toml` with minimum versions.
- The build is deterministic for a given Go toolchain version and platform target.

## Verification Checklist

- Choose the smallest verification set that proves the change.
- From `core/`, run `go test ./...` for core parser, domain logic, or integration-sensitive Go changes.
- Run `.venv\Scripts\python.exe -m pytest` when Python code is affected and tests exist.
- Use `output/` as the local sample directory for golden tests. If real game files are needed, copy them from `C:\Program Files\EA Games\Mass Effect 2` into `output/` first.
- Compare against golden files when parser output changes.
- Consult LegendaryExplorer through the GitHub MCP when LEX semantics matter.
- Final summaries should mention what changed, what verification ran, and any skipped or failing checks.
