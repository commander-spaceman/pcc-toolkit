# Project Map

**Purpose:** High-level overview of the architecture, main modules, and code navigation points.

## Notes for AI Agents

- **Entry points:** `core/cmd/pcc-core/main.go` for the Go engine, `cli/src/cli_main.py` for the Typer CLI, and `pyproject.toml` for Python packaging and CLI registration.
- **Main patterns:** Go core owns all domain logic; Python CLI is a thin adapter; communication happens through subprocess calls and JSON stdout/stderr contracts; behavior is guarded by golden-file regression tests.
- **General rule:** Read this file before proposing structural changes or modifying multiple modules.

---

## 1. Go Core Engine

The Go core is the authoritative domain layer for ME2 OT package parsing, dialogue extraction, TLK resolution, graph layout, validation, evidence scanning, line dumping, owner scanning, and JSON serialization.

```text
core/
├── go.mod
├── cmd/
│   └── pcc-core/
│       └── main.go
└── internal/
    ├── pcc/
    ├── dialogue/
    ├── tlk/
    ├── graph/
    ├── evidence/
    ├── dumper/
    ├── owners/
    ├── scan/
    └── serialize/
```

**Main responsibilities:**

- Parse ME2 OT `.pcc` packages, including LZO-compressed packages and Unreal property data.
- Build and validate `BioConversation` ASTs, resolve TLK text, compute graph layouts, scan evidence and owners, dump dialogue lines, and emit structured JSON.

**Key files:**

- `core/cmd/pcc-core/main.go`: Main binary and subcommand dispatcher for `parse-pcc`, `parse-tlk`, `parse-conversations`, `layout-graph`, `scan-evidence`, `validate`, `serialize`, `dump-lines`, `scan-owners`, and batch operations.
- `core/internal/pcc/reader.go`: Reads PCC headers, names, imports, exports, and package metadata.
- `core/internal/pcc/decompress.go`: Handles ME2 OT LZO decompression.
- `core/internal/pcc/properties.go`: Parses Unreal property tags and semantic property collections.
- `core/internal/pcc/unreal_props.go`: Decodes low-level Unreal property payloads.
- `core/internal/dialogue/parser.go`: Coordinates extraction of conversations from PCC exports and raw serialized data.
- `core/internal/dialogue/parser_semantic.go`: Builds conversation nodes from schema-guided semantic property parsing.
- `core/internal/dialogue/structdb.go`: Contains ME2 dialogue struct metadata used by semantic parsing.
- `core/internal/dialogue/schema.go`: Defines schema helpers for dialogue struct parsing.
- `core/internal/dialogue/validate.go`: Produces validation reports for parsed conversations.
- `core/internal/graph/layout.go`: Computes deterministic dialogue graph layouts.
- `core/internal/evidence/builder.go`: Builds evidence reports and enriches hits with conversation AST data.
- `core/internal/dumper/lines.go`: Builds normalized dialogue line dump output.
- `core/internal/owners/scanner.go`: Scans Kismet conversation-start exports for conversation owner tags.
- `core/internal/scan/scanner.go`: Runs parallel PCC scanning used by evidence and batch workflows.
- `core/internal/tlk/reader.go`: Parses TLK files and decodes text entries.
- `core/internal/tlk/resolver.go`: Resolves StringRefs with base TLK and DLC override priority.
- `core/internal/serialize/writer.go`: Builds the stable JSON output contract consumed by CLI and tests.

**Relationships:**

- Exposes a single `pcc-core` executable consumed by `cli/`.
- Produces JSON contracts validated by `tests/` and golden files.
- Should remain the only place where parsing, AST, layout, evidence, line dumping, owner scanning, and validation logic live.

---

## 2. Python CLI

The CLI is a thin Typer-based command layer. It translates user commands into `pcc-core` subprocess calls and formats results for terminal and JSON consumption by AI agents and automation scripts.

```text
cli/
└── src/
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
│   ├── conversation/
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
- Uses `dropzone/` for optional real game inputs; these files are local artifacts and not source code.
- Golden files should be updated only when parser/output behavior intentionally changes.

---

## 4. Project Configuration and Documentation

Root-level files define package metadata, dependency setup, operating rules, and architecture guidance. These files are the first stop for understanding scope and constraints.

```text
pcc-toolkit/
├── AGENTS.md
├── MAP.md
├── PRD-MVP.md
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
- `PRD-MVP.md`: Canonical MVP architecture, domain model, contracts, dependencies, and verification.
- `MAP.md`: Concise navigation map for the repository.
- `pyproject.toml`: Python project metadata, dependencies, and script entry point.
- `pytest.ini`: Pytest configuration.
- `requirements.txt`: Local Python dependency list.
- `README.md`: Minimal project overview.

**Relationships:**

- Guides changes across `core/`, `cli/`, and `tests/`.
- `pyproject.toml` connects the Python CLI package to user-facing command execution.
- `AGENTS.md` and `PRD-MVP.md` should be consulted before architecture or behavior changes.

---

## 5. Local Runtime Areas

These folders hold local inputs and generated outputs. They are useful during development but are not part of the source architecture.

```text
pcc-toolkit/
├── dropzone/
└── output/
```

**Main responsibilities:**

- Store local ME2 OT sample files used by regression probes and manual testing.
- Store generated output and other local artifacts.

**Key files:**

- `dropzone/`: Expected location for real `.pcc` and `.tlk` files copied from the local ME2 OT install for tests and manual runs.
- `output/`: Generated output directory; contains runtime artifacts rather than source code.

**Relationships:**

- `tests/` and regression probes read from `dropzone/` when available.
- CLI and batch workflows may write generated data to `output/`.
- Neither folder should be treated as a source-of-truth module.

---

## 6. Build and Release Tooling

Scripts used during the build and release process.

```text
pcc-toolkit/
├── scripts/
│   ├── build.ps1
│   └── release.ps1
```

**Main responsibilities:**

- Build the Go core binary with version injection.
- Package portable releases for distribution.

**Key files:**

- `scripts/build.ps1`: Standalone PowerShell script to build the Windows `pcc-core` binary with ldflags version injection. Defaults the version from `pyproject.toml`.
- `scripts/release.ps1`: Release packaging script that builds the core binary and creates a portable release directory or zip archive.

**Relationships:**

- Both scripts operate on `core/cmd/pcc-core/` to produce the Go binary.
- Generated binaries and release packages are local artifacts and are gitignored.
- The Python CLI provides `pcc-toolkit dev build-core` as an alternative when the CLI is installed, but `scripts/build.ps1` works without Python.
