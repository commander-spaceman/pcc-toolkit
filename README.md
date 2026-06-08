# pcc-toolkit

CLI toolkit for inspecting, extracting, editing, and analyzing Mass Effect 2 Original Trilogy
dialogue from `.pcc` package files and `.tlk` talk files. Built for AI agent workflows.

**MVP** — Windows-only. ME2 OT only. Production-ready for controlled/internal use.

## Quick Start

```powershell
# Create and activate a virtual environment
python -m venv .venv
.venv\Scripts\Activate.ps1

# Build the Go core engine
.\scripts\build.ps1

# Install the CLI
.venv\Scripts\python.exe -m pip install -e ".[cli]"
```

## What It Does

### Package

| Command                           | Description                                     |
| --------------------------------- | ----------------------------------------------- |
| `pcc-toolkit package list`        | List exports in a PCC package                   |
| `pcc-toolkit package inspect`     | Show details for a single export                |
| `pcc-toolkit package validate`    | Validate conversations in a PCC                 |
| `pcc-toolkit package extract`     | Full serialization with optional TLK resolution |
| `pcc-toolkit package scan-owners` | Scan Kismet sequences for conversation owners   |

### TLK

| Command                   | Description                                 |
| ------------------------- | ------------------------------------------- |
| `pcc-toolkit tlk info`    | TLK header and entry counts                 |
| `pcc-toolkit tlk search`  | Full-text search across TLK entries         |
| `pcc-toolkit tlk resolve` | Resolve a StringRef through base + DLC TLKs |
| `pcc-toolkit tlk dump`    | Dump all TLK entries                        |

### Dialogue

| Command                           | Description                                      |
| --------------------------------- | ------------------------------------------------ |
| `pcc-toolkit dialogue list`       | List all BioConversation exports in a PCC        |
| `pcc-toolkit dialogue export`     | Export parsed conversations as JSON              |
| `pcc-toolkit dialogue graph`      | Compute Sugiyama graph layout for a conversation |
| `pcc-toolkit dialogue dump-lines` | Dump flat dialogue lines (JSON/CSV)              |
| `pcc-toolkit dialogue edit`       | Edit conversation entries/replies via JSON patch |

### Evidence

| Command                     | Description                                        |
| --------------------------- | -------------------------------------------------- |
| `pcc-toolkit evidence scan` | Search TLK text, then scan PCCs for StringRef hits |

### Batch

| Command                      | Description                      |
| ---------------------------- | -------------------------------- |
| `pcc-toolkit batch validate` | Validate all PCCs in a directory |
| `pcc-toolkit batch extract`  | Extract all PCCs in a directory  |

### Dev

| Command                      | Description                    |
| ---------------------------- | ------------------------------ |
| `pcc-toolkit dev build-core` | Build the Go core from the CLI |
| `pcc-toolkit dev test-core`  | Run Go core tests from the CLI |
| `pcc-toolkit --version`      | Show version and capabilities  |

## Capabilities

### PCC Packages

Read, write, and patch ME2 OT `.pcc` package files.

- **Reading** — Package header, name/import/export tables, and Unreal property tag
  parsing via [`me2pcc`](https://github.com/commander-spaceman/me2pcc).
- **LZO** — Compression and decompression of ME2 OT LZO-compressed packages via
  [`me2lzo`](https://github.com/commander-spaceman/me2lzo).
- **Writing** — Full PCC package serialization with optional LZO compression (`pccwrt`).
- **Encoding** — Unreal property encoding for PCC write workflows (`pccenc`).
- **Patching** — Binary-level patches applied at specific export offsets (`pccpat`).

### TLK Talk Files

Parse, search, resolve, and write ME2 OT `.tlk` files.

- **Reading** — Header parsing, Huffman tree decoding, and bitstream text extraction
  via [`me2tlk`](https://github.com/commander-spaceman/me2tlk).
- **Writing** — Huffman code table building, string encoding, and TLK file
  serialization (`tlkwrt`).
- **Resolution** — DLC-aware StringRef resolution using `Mount.dlc` priority and ME2
  module TLK naming.

### Dialogue

Extract and inspect `BioConversation` data from PCC exports.

- **AST Extraction** — Schema-guided semantic property parsing with row-payload
  fallback modes for partial or unusual conversation layouts.
- **Serialization** — Binary encoding of dialogue AST nodes and reply links back
  into conversation form (`dialenc`).
- **Line Dumping** — Flat JSON/CSV output with one row per dialogue line including
  speaker, text, node type, and source file.

### Editing

Modify conversation data while preserving surrounding package bytes.

- **Conversation Editing** — Edit entry/reply nodes via JSON patch input with
  property span scanning and byte-level splice for round-trip fidelity (`editor`).
- **Dry Run & Backup** — Preview changes without writing, automatic backup creation.

### Graph Layout

Deterministic Sugiyama-style layered layout for dialogue graphs.

- Node positions, typed edges (start→entry, entry→reply, reply→entry), and
  reply-choice metadata for rendering without reparsing AST internals.

### Evidence & Discovery

Connect TLK text search results to PCC package locations.

- **Evidence Scanning** — Tiered reports classifying hits as `bioconversation`,
  `semantic_container`, or `container_fallback` with AST enrichment.
- **Owner Scanning** — Kismet sequence scanning for conversation owner tags.

### Validation

Actionable per-conversation validation with resilient and strict modes.

- Detects missing properties, empty stubs, invalid links, orphaned nodes,
  unresolved speakers, missing StringRefs, and traversal anomalies.

### Batch Operations

Apply operations across directories of PCC files.

- `validate`, `extract`, and `edit` with glob pattern matching and optional
  per-file JSON output.

## Architecture

```
core/     ── Go engine (all domain logic, JSON stdout/stderr)
cli/      ── Python Typer CLI (argument parsing, terminal formatting)
tests/    ── Golden files, smoke tests, and regression probes
```

The Go core is the single source of truth. Python adapters are thin wrappers that call
`pcc-core` as a subprocess and format its JSON output.

See [MAP.md](MAP.md) for the full module navigation guide,
[docs/PRD-INSPECT.md](docs/PRD-INSPECT.md) for the inspect/extract architecture, and
[docs/PRD-EDITING.md](docs/PRD-EDITING.md) for editing and writing.

## Testing

```powershell
cd core && go test ./...                           # Go core tests
.venv\Scripts\python.exe -m pytest                  # Python CLI + golden tests
```

For real-file regression, copy ME2 OT `.pcc`/`.tlk` files from your local install into
`output/`, then run:

```powershell
.venv\Scripts\python.exe tests\regression\run_probes.py --samples-dir output
```

## Requirements

- Go 1.25+ (core engine)
- Python 3.14+ (CLI)
- Windows amd64

## License

MIT
