# pcc-toolkit

CLI toolkit for inspecting, extracting, and analyzing Mass Effect 2 Original Trilogy
dialogue from `.pcc` package files and `.tlk` talk files. Built for AI agent workflows.

**WIP** — Read-only. Windows-only. ME2 OT only.

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

| Command                        | Description                                        |
| ------------------------------ | -------------------------------------------------- |
| `pcc-toolkit package list`     | List exports in a PCC package                      |
| `pcc-toolkit package inspect`  | Show details for a single export                   |
| `pcc-toolkit package validate` | Validate conversations in a PCC                    |
| `pcc-toolkit package extract`  | Full serialization with optional TLK resolution    |
| `pcc-toolkit tlk info`         | TLK header and entry counts                        |
| `pcc-toolkit tlk search`       | Full-text search across TLK entries                |
| `pcc-toolkit tlk resolve`      | Resolve a StringRef through base + DLC TLKs        |
| `pcc-toolkit tlk dump`         | Dump all TLK entries                               |
| `pcc-toolkit dialogue list`    | List all BioConversation exports in a PCC          |
| `pcc-toolkit dialogue export`  | Export parsed conversations as JSON                |
| `pcc-toolkit dialogue graph`   | Compute Sugiyama graph layout for a conversation   |
| `pcc-toolkit evidence scan`    | Search TLK text, then scan PCCs for StringRef hits |
| `pcc-toolkit batch validate`   | Validate all PCCs in a directory                   |
| `pcc-toolkit batch extract`    | Extract all PCCs in a directory                    |

| `pcc-toolkit dev build-core` | Build the Go core from the CLI |
| `pcc-toolkit dev test-core` | Run Go core tests from the CLI |
| `pcc-toolkit --version` | Show version and capabilities |

## Capabilities

PCC parsing, LZO decompression, Unreal property tags, BioConversation AST extraction,
TLK parsing with Huffman decoding, DLC-aware StringRef resolution, deterministic
Sugiyama graph layout, evidence scanning with tiered reports, conversation validation,
dialogue line dumping, and Kismet owner scanning.

## Architecture

```
core/     ── Go engine (all domain logic, JSON stdout/stderr)
cli/      ── Python Typer CLI (argument parsing, terminal formatting)
tests/    ── Golden files, smoke tests, and regression probes
```

The Go core is the single source of truth. Python adapters are thin wrappers that call
`pcc-core` as a subprocess and format its JSON output.

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

- Go 1.22+ (core engine)
- Python 3.11+ (CLI)
- Windows amd64

## License

MIT
