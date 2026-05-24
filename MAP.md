# Project Map (MAP.md)

**Purpose:** This document provides a high-level overview of the code architecture and navigation, designed for both human developers and AI agents.

## 🤖 Notes for AI Agents
- **Entry point:** `core/cmd/pcc-core/main.go` for the main engine logic; `cli/src/cli_main.py` for the public CLI; `gui/src/app.py` for the graphical interface.
- **Main patterns:** Go core containing all domain logic, thin Python CLI/GUI layers, subprocess communication through a JSON contract, parser/serializer behavior guarded by golden-file regression tests.
- **General rules:** Prioritize reading this file to understand the context before proposing code modifications.

## 1. Core Engine
The Go core is the authoritative domain layer. It owns PCC/TLK parsing, dialogue AST construction, graph layout, evidence scanning, validation, and JSON serialization.

```text
core/
├── go.mod
├── go.sum
├── cmd/
│   └── pcc-core/
│       └── main.go
└── internal/
    ├── dialogue/
    ├── evidence/
    ├── graph/
    ├── pcc/
    ├── scan/
    ├── serialize/
    └── tlk/
```

| Archivo / Directorio | Responsabilidad |
| --- | --- |
| `core/` | Go engine containing all domain logic and the JSON-producing binary. |
| `core/go.mod` | Go module definition and engine dependency declarations. |
| `core/go.sum` | Go dependency checksums. |
| `core/cmd/pcc-core/main.go` | Main `pcc-core` binary and subcommand dispatcher. |
| `core/internal/pcc/` | ME2 OT PCC package parsing, export/import/name tables, Unreal strings, properties, and LZO decompression. |
| `core/internal/dialogue/` | `BioConversation` parsing, AST types, ME2 OT schema handling, and structural validation. |
| `core/internal/tlk/` | TLK binary parsing, Huffman decoding, text search, StringRef resolution, and DLC override priority. |
| `core/internal/graph/` | Conversation graph layout computation, including Sugiyama-style positioning. |
| `core/internal/evidence/` | Narrative evidence assembly and contextual profile generation from scan results. |
| `core/internal/scan/` | Game-file scanning, candidate indexes, StringRef parsing, and offset search. |
| `core/internal/serialize/` | Stable JSON output contract combining PCC metadata, conversations, TLK text, and validation. |

## 2. Command-Line Interface
The Python CLI is a thin Typer-based wrapper. It parses user arguments, invokes `pcc-core` as a subprocess, and formats results for terminal use.

```text
cli/
└── src/
    ├── __init__.py
    ├── __main__.py
    ├── cli_main.py
    ├── engine.py
    └── format.py
```

| Archivo / Directorio | Responsabilidad |
| --- | --- |
| `cli/` | Python command-line layer with no domain parsing logic. |
| `cli/src/__init__.py` | Package marker and version-related module surface. |
| `cli/src/__main__.py` | Allows running the CLI package as a Python module. |
| `cli/src/cli_main.py` | Public `pcc-toolkit` Typer app; defines package, TLK, dialogue, evidence, batch, and GUI-launch commands. |
| `cli/src/engine.py` | Subprocess adapter that resolves `pcc-core`, builds flags, runs commands, and parses stdout JSON. |
| `cli/src/format.py` | Rich-based formatting helpers for tables, summaries, validation output, and evidence reports. |

## 3. Graphical Interface
The Python GUI is a Dear ImGui renderer. It manages user interaction and view state while delegating all data extraction and analysis to the Go core.

```text
gui/
└── src/
    ├── __init__.py
    ├── app.py
    ├── engine.py
    ├── state.py
    └── views/
        ├── __init__.py
        ├── dialogue.py
        ├── evidence.py
        ├── package.py
        └── tlk.py
```

| Archivo / Directorio | Responsabilidad |
| --- | --- |
| `gui/` | Python GUI layer for interactive inspection and exploration. |
| `gui/src/app.py` | HelloImGui application entry point, window setup, menus, tabs, status bar, and render loop. |
| `gui/src/state.py` | UI state for loaded paths, selections, graph view, filters, loading state, and errors. |
| `gui/src/engine.py` | GUI subprocess adapter for `pcc-core`, including cancellable asynchronous execution support. |
| `gui/src/views/` | ImGui view modules grouped by feature domain. |
| `gui/src/views/package.py` | PCC package/export inspection tab. |
| `gui/src/views/tlk.py` | TLK lookup, search, and display tab. |
| `gui/src/views/dialogue.py` | Dialogue explorer tab with graph rendering and node details. |
| `gui/src/views/evidence.py` | Evidence-search tab with async process handling. |

## 4. Tests and Regression Data
Tests validate the Go core contract and compare outputs against known-good golden files. Real game samples are expected locally and are not part of source control.

```text
tests/
├── __init__.py
├── conftest.py
├── test_golden.py
├── fixtures/
│   └── __init__.py
├── golden/
│   ├── conversation/
│   ├── graph/
│   ├── pcc/
│   └── tlk/
└── regression/
    └── run_probes.py
```

| Archivo / Directorio | Responsabilidad |
| --- | --- |
| `tests/` | Python regression and contract tests for the toolkit. |
| `tests/test_golden.py` | Runs `pcc-core` and compares stable structure/values against committed golden files. |
| `tests/conftest.py` | Shared pytest configuration and fixtures. |
| `tests/fixtures/` | Synthetic fixture area for tests that do not require real game files. |
| `tests/golden/` | Versioned known-good outputs for conversations, graph layouts, PCC parsing, and TLK behavior. |
| `tests/regression/run_probes.py` | Probe runner for real sample files; validates output shape and can regenerate golden files. |

## 5. Project Configuration and Local Runtime Areas
Root-level configuration defines packaging, dependencies, operational guidance, and local runtime locations. Generated artifacts and real game files should stay outside the source contract.

```text
pcc-toolkit/
├── AGENTS.md
├── MAP.md
├── PRD.md
├── README.md
├── pyproject.toml
├── pytest.ini
├── requirements.txt
├── samples/
└── output/
```

| Archivo / Directorio | Responsabilidad |
| --- | --- |
| `AGENTS.md` | Operational rules for AI agents working in this repository. |
| `MAP.md` | Modular navigation map for humans and AI agents. |
| `PRD.md` | Primary architecture, scope, contract, and migration specification. |
| `README.md` | Minimal repository overview. |
| `pyproject.toml` | Python package metadata, optional dependencies, and CLI entry point. |
| `pytest.ini` | Pytest configuration. |
| `requirements.txt` | Local Python dependency list. |
| `samples/` | Local location for real ME2 OT PCC/TLK files, normally gitignored. |
| `output/` | Generated outputs and runtime files, not source code. |
