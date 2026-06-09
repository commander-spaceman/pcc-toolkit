# Project Map

**Purpose:** High-level overview of the architecture, main modules, and code navigation points.

## Notes for AI Agents

- **Entry points:** `core/cmd/pcc-core/main.go` for the Go engine (subcommand handlers live in sibling `.go` files in the same package), `cli/src/cli_main.py` for the Typer CLI, and `pyproject.toml` for Python packaging and CLI registration.
- **Main patterns:** Go core owns all domain logic; Python CLI is a thin adapter; communication happens through subprocess calls and JSON stdout/stderr contracts; behavior is guarded by golden-file regression tests.
- **General rule:** Read this file before proposing structural changes or modifying multiple modules.

---

## 1. Go Core Engine

The Go core is the authoritative domain layer for ME2 OT package parsing, dialogue extraction, TLK resolution, graph layout, validation, evidence scanning, line dumping, owner scanning, JSON serialization, PCC encoding, PCC writing, PCC patching, dialogue encoding, and conversation editing.

```text
core/
├── go.mod
├── go.sum
├── cmd/
│   └── pcc-core/
│       ├── main.go          (subcommand dispatcher)
│       ├── parse.go         (parse-pcc)
│       ├── tlk.go           (parse-tlk, resolve-tlk)
│       ├── conversation.go  (parse-conversations)
│       ├── graph.go         (layout-graph)
│       ├── evidence.go      (scan-evidence)
│       ├── validate.go      (validate, batch-validate)
│       ├── serialize.go     (serialize)
│       ├── dump.go          (dump-lines)
│       ├── owners.go        (scan-owners)
│       └── edit.go          (edit-conversation, batch-edit)
└── internal/
    ├── dialenc/
    ├── dialogue/
    ├── dumper/
    ├── editor/
    ├── evidence/
    ├── graph/
    ├── owners/
    ├── pccenc/
    ├── pccpat/
    ├── pccwrt/
    ├── scan/
    ├── serialize/
    └── tlkwrt/
```

**Main responsibilities:**

- Parse ME2 OT `.pcc` packages, including LZO-compressed packages and Unreal property data.
- Build and validate `BioConversation` ASTs, resolve TLK text, compute graph layouts, scan evidence and owners, dump dialogue lines, and emit structured JSON.
- Encode and write PCC packages, patch existing packages with binary-level edits, serialize dialogue back to binary, and edit conversation exports while preserving surrounding data.

**Key files:**

- `core/cmd/pcc-core/main.go`: Main binary entry point and subcommand dispatcher. Subcommand implementations live in sibling files (`parse.go`, `tlk.go`, `conversation.go`, `graph.go`, `evidence.go`, `validate.go`, `serialize.go`, `dump.go`, `owners.go`, `edit.go`).
- PCC reading is delegated to the external `github.com/commander-spaceman/me2pcc` library.
- TLK reading is delegated to the external `github.com/commander-spaceman/me2tlk` library.
- LZO decompression is delegated to the external `github.com/commander-spaceman/me2lzo` library.
- `core/internal/dialogue/parser.go`: Coordinates extraction of conversations from PCC exports and raw serialized data.
- `core/internal/dialogue/parser_semantic.go`: Builds conversation nodes from schema-guided semantic property parsing.
- `core/internal/dialogue/parser_row.go`: Builds entry nodes from row-mode conversation data (matrix-based extraction fallback).
- `core/internal/dialogue/structdb.go`: Contains ME2 dialogue struct metadata used by semantic parsing.
- `core/internal/dialogue/schema.go`: Defines schema helpers for dialogue struct parsing.
- `core/internal/dialogue/types.go`: Dialogue AST node types (EntryNode, ReplyChoice, Conversation, etc.).
- `core/internal/dialogue/validate.go`: Produces validation reports for parsed conversations.
- `core/internal/dialenc/encode.go`: Encodes dialogue AST nodes and reply links back into binary conversation form.
- `core/internal/editor/editor.go`: Edits conversation exports in-place, preserving unchanged data via binary preservation helpers.
- `core/internal/editor/preserve.go`: Property span scanning and byte-level splice for round-trip fidelity.
- `core/internal/editor/conv_ser.go`: Serializes a modified `dialogue.Conversation` AST back into binary property form for PCC writing.
- `core/internal/pccenc/encode.go`: Encodes Unreal properties for PCC writing.
- `core/internal/pccenc/writer.go`: Buffered writer for PCC encoding output.
- `core/internal/pccpat/patch.go`: Applies binary patches to PCC files at specific export offsets.
- `core/internal/pccpat/buildminimal.go`: Builds minimal valid PCC structures for patch injection.
- `core/internal/pccwrt/write.go`: Writes complete PCC packages including header, tables, and export data.
- `core/internal/pccwrt/compress.go`: Handles LZO compression when writing compressed PCC packages.
- `core/internal/graph/layout.go`: Computes deterministic dialogue graph layouts.
- `core/internal/graph/types.go`: Graph domain types (nodes, edges, layout algorithms).
- `core/internal/evidence/builder.go`: Builds evidence reports and enriches hits with conversation AST data.
- `core/internal/evidence/types.go`: Evidence hit types and scan result structures.
- `core/internal/evidence/profile.go`: Evidence scan profile definitions and configuration.
- `core/internal/dumper/lines.go`: Builds normalized dialogue line dump output.
- `core/internal/owners/scanner.go`: Scans Kismet conversation-start exports for conversation owner tags.
- `core/internal/scan/scanner.go`: Runs parallel PCC scanning used by evidence and batch workflows.
- `core/internal/scan/files.go`: Discovers PCC and TLK files from BioGame and DLC directories.
- `core/internal/scan/cache.go`: File modification time cache to skip unchanged packages.
- `core/internal/scan/index.go`: Builds in-memory file index for batch operations.
- `core/internal/scan/types.go`: Scan domain types (scan configuration, result summaries).
- `core/internal/tlkwrt/writer.go`: Writes TLK files from in-memory entries with correct headers and Huffman encoding.
- `core/internal/serialize/writer.go`: Builds the stable JSON output contract consumed by CLI and tests.

**Relationships:**

- Exposes a single `pcc-core` executable consumed by `cli/`.
- Produces JSON contracts validated by `tests/` and golden files.
- Should remain the only place where parsing, AST, layout, evidence, line dumping, owner scanning, validation, writing, encoding, patching, and editing logic live.

---

## 2. Python CLI

The CLI is a thin Typer-based command layer. It translates user commands into `pcc-core` subprocess calls and formats results for terminal and JSON consumption by AI agents and automation scripts.

```text
cli/
└── src/
    ├── __init__.py
    ├── __main__.py
    ├── cli_main.py
    ├── engine.py
    └── format.py
```

**Main responsibilities:**

- Define user-facing commands for package inspection, TLK operations, dialogue extraction, evidence search, and batch tasks.
- Convert CLI options into Go engine flags and present returned JSON as tables, summaries, or raw JSON.

**Key files:**

- `cli/src/cli_main.py`: Typer app and public command definitions.
- `cli/src/engine.py`: Subprocess adapter that locates `pcc-core`, builds command arguments, and parses JSON output.
- `cli/src/format.py`: Rich formatting helpers for exports, conversations, validation, evidence, and batch summaries.
- `cli/src/__main__.py`: Module execution hook for the CLI package.

**Relationships:**

- Depends on the Go core executable and its JSON contract.
- Is registered through `pyproject.toml` as the `pcc-toolkit` script.

---

## 3. Tests and Regression Assets

Tests focus on the Go core contract and known-good outputs. Golden files encode stable expectations for parser, TLK, graph, and serialization behavior.

```text
tests/
├── conftest.py
├── test_golden.py
├── test_smoke.py
├── fixtures/
├── golden/
│   ├── batch/
│   ├── conversation/
│   ├── edit/
│   ├── evidence/
│   ├── graph/
│   ├── pcc/
│   └── tlk/
└── regression/
    └── run_probes.py
```

**Main responsibilities:**

- Validate `pcc-core` output shape and stable values against committed golden files.
- Provide regression probes for real ME2 OT samples when local game files are available.

**Key files:**

- `tests/test_golden.py`: Pytest regression checks that execute `pcc-core` and compare against golden files.
- `tests/test_smoke.py`: CLI entry-point smoke tests for help, version, and dev command visibility.
- `tests/regression/run_probes.py`: Probe runner for sample files, including golden regeneration support.
- `tests/golden/`: Versioned known-good JSON outputs for conversations, TLK, graphs, and PCC exports.
- `tests/fixtures/`: Synthetic fixture area for tests that do not require real game files.

**Relationships:**

- Depends on a built `pcc-core` binary in `build/` or on PATH.
- Uses `output/` for optional real game inputs; these files are local artifacts and not source code.
- Golden files should be updated only when parser/output behavior intentionally changes.

---

## 4. Project Configuration and Documentation

Root-level files define package metadata, dependency setup, operating rules, and architecture guidance. These files are the first stop for understanding scope and constraints.

```text
pcc-toolkit/
├── AGENTS.md
├── MAP.md
├── docs/
│   ├── README.md
│   ├── PRD-INSPECT.md
│   ├── PRD-EDITING.md
│   ├── PRD-KISMET.md
│   ├── CONTRACTS.md
│   ├── BUILDING.md
│   └── REFERENCE.md
├── README.md
├── pyproject.toml
├── pytest.ini
└── requirements.txt
```

**Main responsibilities:**

- Document architecture, repository conventions, setup metadata, and test configuration.
- Keep agents and developers aligned on ME2 OT scope, JSON contracts, and the Go-core/Python-wrapper separation.

**Key files:**

- `AGENTS.md`: Operational rules for AI agents, including scope, verification, and repository conventions.
- `docs/PRD-INSPECT.md`: Canonical architecture, domain model, contracts, and verification for inspect & extract.
- `docs/PRD-EDITING.md`: Editing, writing, round-trip fidelity, and batch edit requirements.
- `docs/PRD-KISMET.md`: Kismet/cinematic sequence support requirements and phase plan.
- `docs/CONTRACTS.md`: JSON contracts for all pcc-core subcommands.
- `docs/BUILDING.md`: Build, test, and release guide.
- `docs/REFERENCE.md`: LegendaryExplorer reference notes and known divergences.
- `docs/README.md`: Documentation index and reading-order guide.
- `MAP.md`: Concise navigation map for the repository.
- `pyproject.toml`: Python project metadata, dependencies, and script entry point.
- `pytest.ini`: Pytest configuration.
- `requirements.txt`: Local Python dependency list.
- `README.md`: Minimal project overview.

**Relationships:**

- Guides changes across `core/`, `cli/`, and `tests/`.
- `pyproject.toml` connects the Python CLI package to user-facing command execution.
- `AGENTS.md` and the PRDs in `docs/` should be consulted before architecture or behavior changes.

---

## 5. Local Runtime Areas

These folders hold local inputs and generated outputs. They are useful during development but are not part of the source architecture.

```text
pcc-toolkit/
├── output/
├── samples/
```

**Main responsibilities:**

- Store local ME2 OT sample files used by regression probes and manual testing.
- Store generated output and other local artifacts.

**Key files:**

- `output/`: Expected location for real `.pcc` and `.tlk` files copied from the local ME2 OT install for tests and manual runs.

**Relationships:**

- `tests/` and regression probes read from `output/` when available.
- CLI and batch workflows may write generated data to `output/`.
- Neither folder should be treated as a source-of-truth module.

---

## 6. Build and Release Tooling

Scripts used during the build and release process.

```text
pcc-toolkit/
├── scripts/
│   ├── README.md
│   ├── build.ps1
│   ├── release.ps1
│   ├── apply.ps1
│   ├── backup-pcc.ps1
│   ├── restore.ps1
│   ├── run-game.ps1
│   ├── trace-game.ps1
│   ├── trace_game.py
│   └── conv_to_screenplay.py
```

**Main responsibilities:**

- Build the Go core binary with version injection.
- Package portable releases for distribution.
- Provide game modding utilities (backup, apply, restore).
- Provide game debugging and tracing tools.

**Key files:**

- `scripts/build.ps1`: Standalone PowerShell script to build the Windows `pcc-core` binary with ldflags version injection. Defaults the version from `pyproject.toml`.
- `scripts/release.ps1`: Release packaging script that builds the core binary and creates a portable release directory or zip archive.
- `scripts/apply.ps1`: Copies modified files from `output/` into the game's `CookedPC/` directory.
- `scripts/backup-pcc.ps1`: Backs up original ME2 OT game files before modding.
- `scripts/restore.ps1`: Restores original game files from backup.
- `scripts/run-game.ps1`: Launches ME2 OT with real-time log and file-access monitoring.
- `scripts/trace_game.py`: ETW kernel-file tracer that captures every `.pcc`/`.tlk`/`.upk` file the game opens. Requires Administrator.
- `scripts/trace-game.ps1`: PowerShell wrapper for `trace_game.py` that auto-elevates and displays results.
- `scripts/conv_to_screenplay.py`: Converts a conversation JSON export to a readable screenplay format.
- `scripts/README.md`: Scripts directory reference and usage guide.

**Relationships:**

- Build scripts operate on `core/cmd/pcc-core/` to produce the Go binary.
- Modding scripts (`apply.ps1`, `backup-pcc.ps1`, `restore.ps1`) interact with the local ME2 OT install and `output/`.
- Generated binaries and release packages are local artifacts and are gitignored.
- The Python CLI provides `pcc-toolkit dev build-core` as an alternative when the CLI is installed, but `scripts/build.ps1` works without Python.

---

## 7. Legacy Code

The `pcc-dialog-toolkit-legacy/` directory at the workspace root contains the predecessor
Python/Go hybrid implementation. It is preserved for reference only, is outside the
current MVP scope, and should not be modified or depended upon.
