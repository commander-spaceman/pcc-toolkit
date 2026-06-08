# Building and Release Guide

## Prerequisites

- **Go 1.25+** — for the core engine (`core/`).
- **Python 3.14+** — for the CLI (`cli/`).
- **Windows amd64** — the only supported platform for builds and releases.
- **PowerShell 7+** — for build and release scripts.

## Project Dependencies

The Go core depends on three external libraries:

| Library  | Repository                                    | Purpose                         |
| -------- | --------------------------------------------- | ------------------------------- |
| `me2lzo` | `github.com/commander-spaceman/me2lzo` v1.0.0 | LZO compression / decompression |
| `me2pcc` | `github.com/commander-spaceman/me2pcc`        | PCC package reading             |
| `me2tlk` | `github.com/commander-spaceman/me2tlk`        | TLK file reading                |

Python CLI dependencies (`typer`, `rich`, `pytest`) are declared in `pyproject.toml`.

## Environment Setup

```powershell
# Create virtual environment
python -m venv .venv
.venv\Scripts\Activate.ps1

# Install CLI in editable mode
.venv\Scripts\python.exe -m pip install -e ".[cli]"

# Install test dependencies
.venv\Scripts\python.exe -m pip install -e ".[dev]"
```

## Build

### Build Core (standalone)

```powershell
.\scripts\build.ps1 -Version "0.3.0"
```

Output: `build/pcc-core.exe`

The script:

- Reads the default version from `pyproject.toml` if `-Version` is omitted.
- Injects the version via Go ldflags: `-X main.version=x.y.z`.
- Produces a statically-linked Windows amd64 binary.

### Build Core (via CLI)

```powershell
.venv\Scripts\pcc-toolkit.exe dev build-core
```

### Build Options

| Flag       | Description                            |
| ---------- | -------------------------------------- |
| `-Version` | Version string for ldflags injection   |
| `-Output`  | Custom output path (default: `build/`) |

## Test

### Go Core Tests

```powershell
cd core
go test ./...
go vet ./...
go fmt ./...
```

### Python Tests

```powershell
.venv\Scripts\python.exe -m pytest
```

### Real-File Regression Probes

Copy ME2 OT `.pcc` and `.tlk` files from your local install into `output/`:

```powershell
.venv\Scripts\python.exe tests\regression\run_probes.py --samples-dir output
```

## Release

### Create Local Release

```powershell
.\scripts\release.ps1 -Version "0.3.0" -Archive
```

Produces:

- `release/pcc-toolkit-v0.3.0-windows-amd64/` with `pcc-core.exe` and `INSTALL.txt`.
- `release/pcc-toolkit-v0.3.0-windows-amd64.zip` (when `-Archive` is specified).

### Published Artifacts

| Artifact       | Platform        | Built from           |
| -------------- | --------------- | -------------------- |
| `pcc-core.exe` | Windows (amd64) | `core/cmd/pcc-core/` |

Build and release outputs are gitignored and never committed.

### Release Smoke Tests

Before publishing a release, verify:

1. Extract the release zip into a clean temporary directory.
2. Run `pcc-core.exe version` — confirms target `me2_ot` and capabilities.
3. Run `parse-pcc`, `parse-tlk`, `resolve-tlk` against copied local assets.
4. Run `parse-conversations`, `layout-graph`, `validate`, `dump-lines`.
5. Run `scan-evidence` against a temporary `BioGame/CookedPC` tree.
6. Run `edit-conversation` with `--dry-run` to verify editing pipeline.

## Development Workflow

1. Make changes in `core/` for domain logic or `cli/` for CLI commands.
2. Run `go test ./...` from `core/` after Go changes.
3. Run `.venv\Scripts\python.exe -m pytest` after CLI changes.
4. Run `go fmt ./...` from `core/` before committing Go code.
5. Rebuild `pcc-core.exe` if the binary is out of date vs source.

## Reproducibility

- Go dependencies are locked in `core/go.sum`.
- Version is injected via ldflags at build time.
- Python dependencies are declared in `pyproject.toml` with minimum versions.
- The build is deterministic for a given Go toolchain version and platform target.
