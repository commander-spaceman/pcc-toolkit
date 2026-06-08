# PCC Toolkit Product Requirements Document — Inspect & Extract

## 1. Purpose

PCC Toolkit is a Windows-first CLI toolkit for inspecting, extracting,
validating, and analyzing Mass Effect 2 Original Trilogy dialogue data from `.pcc`
package files and `.tlk` talk files.

The toolkit is designed for AI agents and automation scripts that need deterministic
JSON contracts for repeatable analysis, regression testing, and code-assisted modding
workflows. It provides fast read-only access to ME2 OT dialogue data through a
subprocess-callable Go engine with a thin Python CLI wrapper.

Editing and writing capabilities are covered in [PRD-EDITING.md](PRD-EDITING.md).

## 2. Scope

### 2.1 In Scope

- Mass Effect 2 Original Trilogy only.
- Windows `amd64` distribution for the MVP.
- ME2 OT `.pcc` package parsing, including LZO-compressed packages.
- Unreal package header, name, import, and export table parsing.
- Unreal property tag parsing for supported ME2 OT dialogue structures.
- Semantic property decoding for `BioConversation` data.
- Fallback row-payload parsing for partial or unusual conversation layouts.
- ME2 OT `.tlk` parsing and Huffman text decoding.
- TLK text search and StringRef resolution.
- DLC-aware TLK resolution using `Mount.dlc` priority and ME2 module TLK naming.
- `BioConversation` AST extraction.
- Conversation validation.
- Deterministic dialogue graph layout.
- Evidence scanning from TLK candidate StringRefs into PCC package hits.
- Tiered evidence reports.
- Normalized dialogue line dumping.
- Kismet conversation owner scanning.
- Batch validation and extraction workflows.
- Python Typer CLI as a thin wrapper over the Go core.
- Golden-file, smoke-test, and local real-file regression coverage.

### 2.2 Out of Scope

- Legendary Edition support: LE1, LE2, LE3.
- Other Original Trilogy games: ME1, ME3.
- Editing or writing PCC dialogue data.
- Binary reinjection.
- FaceFX, Wwise, Matinee, or InterpData editing.
- Console package support.
- Running tests directly against `C:\Program Files\EA Games\Mass Effect 2`.
- Committing game files, generated outputs, local caches, built binaries, or release archives.
- Non-Windows distribution for the MVP.

### 2.3 MVP Readiness Definition

The MVP is production-ready for controlled/internal use when:

- `pcc-core.exe version` reports target `me2_ot` and the implemented capability list.
- `go test ./...` passes from `core/`.
- `.venv\Scripts\python.exe -m pytest` passes from the repository root.
- The Windows release zip can be extracted and `pcc-core.exe` can run representative
  workflows against copied local assets.
- CLI commands can inspect packages, parse TLKs, extract conversations, validate,
  generate graph layouts, run batch validation, and scan evidence.
- Known limitations are documented rather than implicit.

The MVP is intentionally read-only. It does not edit, write, or reinject PCC data.

## 3. Reference Implementation Policy

LegendaryExplorer (`ME3Tweaks/LegendaryExplorer`, branch `Beta`) is the behavioral
reference for PCC package semantics, TLK resolution, dialogue structures, graph
semantics, and validation expectations where those behaviors apply to ME2 OT.

Reference areas to consult include:

- `LegendaryExplorerCore/Packages/MEPackageHandler.cs`
- `LegendaryExplorerCore/Dialogue/ConversationExtended.cs`
- `LegendaryExplorerCore/Dialogue/DialogueNodeExtended.cs`
- `LegendaryExplorerCore/Dialogue/ReplyChoiceNode.cs`
- `LegendaryExplorerCore/Dialogue/SpeakerExtended.cs`
- `LegendaryExplorer/Tools/Dialogue Editor/DialogueObjects.cs`
- `LegendaryExplorer/Tools/Dialogue Editor/ConvGraphEditor.cs`
- `LegendaryExplorerCore/GameFilesystem/TlkSystem.cs`
- `LegendaryExplorerCore/GameFilesystem/MountFile.cs`
- `LegendaryExplorerCore/Unreal/ME3Enums.cs`

LegendaryExplorer is GPLv3-licensed and its `CONTRIBUTING.md` includes restrictions
on low-value generative-AI contributions. PCC Toolkit uses LegendaryExplorer as a
behavioral and semantic reference only. Do not copy, paste, translate, or port its
code into this repository unless the project intentionally accepts the resulting
license obligations. Implement behavior independently, and document observed public
behavior, relevant file/class names, and verification outcomes.

Guiding question:

```text
Does this match how LegendaryExplorer handles the same ME2 OT data?
```

## 4. Architecture

PCC Toolkit is split into a Go domain engine and a thin Python CLI adapter.

```text
pcc-toolkit/
├── core/        Go domain engine and JSON producer
├── cli/         Python Typer command wrapper
├── tests/       Golden, smoke, and regression tests
├── scripts/     Windows build and release scripts
├── output/      Local copied ME2 OT test assets and generated outputs, gitignored except .gitkeep
```

### 4.1 Ownership Rules

- Go core owns all domain logic.
- Python CLI owns argument parsing, subprocess dispatch, terminal formatting, and file output.
- JSON stdout/stderr is the integration boundary.
- No domain parsing, TLK resolution, AST building, graph layout, validation, evidence
  assembly, owner scanning, or serialization logic belongs in Python.

### 4.2 Capability Ownership Matrix

| Capability                             | Go core | Python CLI |
| -------------------------------------- | ------- | ---------- |
| PCC header/table parsing               | All     | None       |
| LZO decompression                      | All     | None       |
| Unreal property parsing                | All     | None       |
| BioConversation AST building           | All     | None       |
| TLK parsing and StringRef resolution   | All     | None       |
| DLC priority and module TLK resolution | All     | None       |
| Graph layout                           | All     | None       |
| Evidence scanning and tiering          | All     | None       |
| Dialogue line dumping                  | All     | None       |
| Conversation owner scanning            | All     | None       |
| Validation                             | All     | None       |
| JSON serialization                     | All     | None       |
| Batch aggregation                      | All     | None       |
| CLI argument parsing                   | None    | All        |
| Terminal formatting                    | None    | All        |
| Error display                          | None    | All        |

## 5. Repository Structure

```text
pcc-toolkit/
├── AGENTS.md
├── MAP.md
├── README.md
├── docs/
│   ├── PRD-READ.md
│   ├── PRD-EDITING.md
│   ├── CONTRACTS.md
│   ├── BUILDING.md
│   └── REFERENCE.md
├── pyproject.toml
├── pytest.ini
├── requirements.txt
├── core/
│   ├── go.mod
│   ├── go.sum
│   ├── cmd/pcc-core/main.go
│   └── internal/
│       ├── dialenc/
│       ├── dialogue/
│       ├── dumper/
│       ├── editor/
│       ├── evidence/
│       ├── graph/
│       ├── owners/
│       ├── pccenc/
│       ├── pccpat/
│       ├── pccwrt/
│       ├── scan/
│       ├── serialize/
│       └── tlkwrt/
├── cli/src/
│   ├── __init__.py
│   ├── __main__.py
│   ├── cli_main.py
│   ├── engine.py
│   └── format.py
├── scripts/
│   ├── build.ps1
│   └── release.ps1
└── tests/
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
    └── regression/run_probes.py
```

Generated or local-only areas:

- `build/`: local compiled binaries, gitignored.
- `release/`: release directories and zip files, gitignored.
- `output/`: copied game assets and generated local outputs, gitignored except `.gitkeep`.
- `.venv/`, `.pytest_cache/`, `__pycache__/`: dependency/test caches, gitignored.

## 6. Dependencies

### 6.1 Go Core

| Package                     |  Version | Purpose                                            |
| --------------------------- | -------: | -------------------------------------------------- |
| `github.com/anchore/go-lzo` | `v0.1.0` | LZO1X decompression for ME2 OT compressed packages |
| `encoding/json`             |   stdlib | JSON contracts                                     |
| `flag`                      |   stdlib | Core subcommand flags                              |
| `runtime`                   |   stdlib | Worker default for scanning                        |

The Go core is otherwise self-contained.

### 6.2 Python CLI

| Package | Version  | Purpose                                 |
| ------- | -------- | --------------------------------------- |
| `typer` | `>=0.15` | CLI command groups and options          |
| `rich`  | `>=13`   | Terminal tables and formatted summaries |

### 6.3 Development

| Package  | Version | Purpose            |
| -------- | ------- | ------------------ |
| `pytest` | `>=9.0` | Python test runner |

Python code should target Python 3.14+ style while maintaining project metadata
compatibility as declared in `pyproject.toml`.

## 7. Go Core Binary Contract

The core binary is `pcc-core.exe` on Windows.

```text
pcc-core <subcommand> [flags]
```

Core process contract:

- Input is command-line flags and file paths.
- Success payloads are JSON written to stdout.
- Error payloads are JSON written to stderr as `{"error":"..."}`.
- Exit code is `0` on success and non-zero on failure.
- Subcommands that support `--pretty` indent JSON output.
- The target game profile is `me2_ot` for this MVP.

### 7.1 `version`

Returns binary version, target, and capability strings.

Current MVP capabilities:

- `pcc_parse_v1`
- `pcc_property_tags_v1`
- `pcc_semantic_props_v1`
- `conversation_ast_v1`
- `graph_layout_v1`
- `tlk_parse_v1`
- `tlk_dlc_resolve_v1`
- `evidence_scan_v1`
- `validate_v1`
- `serialize_v1`
- `batch_validate_v1`
- `dump_lines_v1`
- `scan_owners_v1`

### 7.2 `parse-pcc`

```text
pcc-core parse-pcc --file <path>
                   [--exports-only]
                   [--export-index <n>]
                   [--property-tags]
                   [--semantic-props]
                   [--pretty]
```

Responsibilities:

- Read raw PCC bytes.
- Detect and decompress ME2 OT LZO-compressed packages.
- Parse package header, name table, import table, and export table.
- Infer game profile.
- Resolve export class names and object names.
- Optionally return export raw serial data as base64.
- Optionally parse low-level property tags or semantic property collections.

Stable output fields include `file`, `game_profile`, `compressed`, `header`,
`names`, `imports`, and `exports`.

### 7.3 `parse-tlk`

```text
pcc-core parse-tlk --file <path>
                   [--search <query>]
                   [--strref <id>]
                   [--dump-all]
                   [--pretty]
```

Responsibilities:

- Parse TLK header.
- Decode Huffman tree and bitstream-backed text entries.
- Expose male/female table counts and total entries.
- Return filtered search results or exact StringRef lookup results.
- Return full entry dump when requested.

### 7.4 `resolve-tlk`

```text
pcc-core resolve-tlk --base <path>
                     [--dlc-dir <path>]
                     [--language <code>]
                     --strref <id> [--strref <id> ...]
                     [--pretty]
```

Responsibilities:

- Load base TLK first.
- Scan DLC folders when `--dlc-dir` is provided.
- Read `Mount.dlc` priority for DLC order.
- Resolve ME2 DLC module TLK filenames from DLC metadata where available.
- Load `DLC_<module>_<language>.tlk` files.
- Return final effective text and source TLK for each requested StringRef.

Default language is `INT`.

### 7.5 `parse-conversations`

```text
pcc-core parse-conversations --file <path>
                             [--conv-index <n>]
                             [--resolve-tlk <tlk_path>]
                             [--dlc-dir <path>]
                             [--language <code>]
                             [--mode resilient|strict]
                             [--pretty]
```

Responsibilities:

- Parse all `BioConversation` exports or a single requested export.
- Prefer schema-guided semantic property parsing.
- Fall back to row-payload modes when semantic parsing cannot decode the export.
- Resolve TLK text when requested.
- Preserve parse warnings without aborting resilient output.
- Return parse errors separately from successful conversations.

Known parse modes include `struct_property_semantic`, `row_payload`,
`row_payload_struct_matrix`, `row_payload_struct_head`, and `count_or_value_fallback`.

### 7.6 `layout-graph`

```text
pcc-core layout-graph --file <path>
                      [--conv-index <n>]
                      [--algorithm sugiyama]
                      [--node-width <px>]
                      [--node-height <px>]
                      [--x-spacing <px>]
                      [--y-spacing <px>]
                      [--pretty]
```

Responsibilities:

- Parse the requested conversation.
- Build a directed graph of start, entry, and reply nodes.
- Use deterministic Sugiyama-style layered layout.
- Return node positions, edge metadata, and node metadata needed by consumers.

Current algorithm support is `sugiyama` only.

### 7.7 `scan-evidence`

```text
pcc-core scan-evidence --query <text>
                        --tlk <path>
                        [--dlc-dir <path>]
                        [--language <code>]
                        [--biogame-root <path>]
                        [--cache <path>]
                        [--workers <n>]
                        [--pretty]
```

Responsibilities:

- Search TLK text for query matches.
- Convert TLK hits into candidate StringRefs.
- If `--biogame-root` is provided, collect PCC files from a ME2-style tree.
- Scan candidate files for StringRef occurrences.
- Classify hits into evidence tiers.
- Enrich BioConversation matches with AST context where available.
- Return a complete evidence report.

Path rule:

```text
--biogame-root must point to a BioGame-style root containing CookedPC/ and
optionally DLC/. A flat output/ directory is not a scan root.
```

### 7.8 `validate`

```text
pcc-core validate --file <path> [--strict] [--pretty]
```

Responsibilities:

- Parse conversations in resilient mode.
- Produce per-conversation validation results.
- Report valid, warning, and invalid counts.
- Promote warnings consistently under strict mode.
- Exit non-zero when validation fails by selected policy.

### 7.9 `serialize`

```text
pcc-core serialize --file <path>
                   [--game <profile>]
                   [--resolve-tlk <path>]
                   [--dlc-dir <path>]
                   [--language <code>]
                   [--pretty]
```

Responsibilities:

- Run parse, optional TLK resolution, validation, and stable output serialization.
- Serve as the backend for `pcc-toolkit package extract`.
- Keep output stable for regression and automation consumers.

### 7.10 `batch-validate`

```text
pcc-core batch-validate --dir <path>
                        [--glob <pattern>]
                        [--strict]
                        [--output <path>]
                        [--pretty]
```

Aggregates validation across matching PCC files.

### 7.11 `batch-extract`

```text
pcc-core batch-extract --dir <path>
                       [--glob <pattern>]
                       [--output-dir <path>]
                       [--resolve-tlk <path>]
                       [--dlc-dir <path>]
                       [--language <code>]
                       [--pretty]
```

Serializes matching PCC files and optionally writes one JSON file per PCC.

### 7.12 `dump-lines`

```text
pcc-core dump-lines --file <path>
                    [--resolve-tlk <path>]
                    [--dlc-dir <path>]
                    [--language <code>]
                    [--format json|csv]
                    [--pretty]
```

Returns one row per dialogue line with conversation ID, export index, node type,
node ID, speaker tag, StringRef, resolved text, and source file.

### 7.13 `scan-owners`

```text
pcc-core scan-owners --file <path> [--pretty]
```

Scans Kismet sequence exports and reports owner tags associated with conversation
contexts where detected.

## 8. Domain Model And JSON Contracts

### 8.1 Parse Result

```json
{
  "file": "BioD_CitHub_LOC_INT.pcc",
  "game_profile": "me2_ot",
  "conversations": [],
  "errors": []
}
```

### 8.2 Conversation

Required fields:

- `id`: export object name.
- `export_index`: package export index.
- `game_profile`: always `me2_ot` in the MVP.
- `parse_mode`: parser path used.
- `entries`: entry nodes.
- `replies`: reply nodes.
- `speakers`: synthetic and parsed speakers.
- `starts`: start nodes.

Optional fields:

- `script_list`: parsed script entries from script-list properties.
- `matinee_sequence_export_id`: object ref for MatineeSequence where available.
- `warnings`: parser warnings.

### 8.3 Entry Node

Entry node fields:

- `id`
- `speaker_id`
- `speaker_tag`
- `listener_index`
- `listener_tag`
- `line_strref`
- `line_text`
- `reply_links`
- `reply_choices`
- `conditional_func`
- `conditional_param`
- `state_transition`
- `state_transition_param`
- `script_index`
- `script_name`
- `fires_conditional`
- `export_id`
- `skippable`
- `non_text_line`
- `ambient`
- `camera_intimacy`
- `gui_style`

Rules:

- `id` is stable within the conversation.
- `speaker_id` resolves into `speakers` when possible.
- `line_text` appears only when TLK resolution is requested and succeeds.
- `reply_links` preserves linked reply IDs.
- `reply_choices` preserves edge metadata when available.
- Optional booleans and integers are omitted when unavailable.

### 8.4 Reply Node

Reply node fields:

- `id`
- `line_strref`
- `line_text`
- `target_entry_ids`
- `condition_refs`
- `category`
- `reply_type`
- `conditional_func`
- `conditional_param`
- `state_transition`
- `state_transition_param`
- `script_index`
- `script_name`
- `fires_conditional`
- `export_id`
- `unskippable`
- `non_text_line`
- `ambient`
- `camera_intimacy`
- `gui_style`

Rules:

- Replies represent player choice nodes or reply-like dialogue choices.
- `target_entry_ids` supports multiple targets where the data exposes them.
- `condition_refs` is a normalized compact representation for display and checks.
- Category should align with LegendaryExplorer reply category semantics where possible.

### 8.5 Reply Choice Edge Metadata

Reply choice fields:

- `from_entry_id`
- `to_reply_id`
- `order`
- `paraphrase`
- `paraphrase_strref`
- `paraphrase_text`
- `category`

Reply choices preserve the distinction between reply nodes and entry-to-reply edge
metadata. This is important for graph rendering and LegendaryExplorer parity.

### 8.6 Speaker

Speaker fields:

- `id`
- `tag`
- `display_name`
- `strref_id`
- `friendly_name`
- `facefx_male_animset`
- `facefx_female_animset`

Speaker rules:

- Preserve parsed speaker-list entries by array index.
- Preserve synthetic speaker IDs when present.
- `player` uses ID `-2` and friendly name `Shepard`.
- `owner` uses ID `-1` and represents the conversation owner context.
- FaceFX references are metadata only in this MVP; there is no editing support.

### 8.7 Start Node

Start node fields:

- `id`
- `target_entry_ids`
- `label`

Rules:

- Start nodes represent `m_StartingList` entries or equivalent start data.
- Multiple target entries may be represented.
- Missing start nodes should not by itself make a known stub or reply-only conversation invalid.

### 8.8 Script Entry

Script entry fields:

- `id`
- `tag`
- `name`

Script entries are extracted from script-list structures where available.

## 9. LegendaryExplorer Semantics To Preserve

- `BioConversation` is parsed from export properties when structured properties are available.
- `m_StartingList` maps start-node order to entry indexes.
- `m_SpeakerList` contains conversation speaker tags for ME2-style conversations.
- LEX synthesizes special player and owner speakers before parsed speakers.
- Entry nodes use `nSpeakerIndex` where present.
- Reply nodes are conceptually player choices.
- Entry-to-reply edge metadata comes from ReplyListNew-like structures with order,
  paraphrase, StringRef, and category fields.
- Reply-to-entry links come from reply target indexes or entry-list integer arrays,
  depending on the concrete serialized form.
- Conditions use `nConditionalFunc`, `nConditionalParam`, and `bFireConditional`.
- State transitions use `nStateTransition` and `nStateTransitionParam`.
- Dialogue-node identity for sequence correlation uses `nExportID` where available.
- Dialogue graph edges are typed by source and target role: start to entry, entry to
  reply, reply to entry.
- Reply-choice edge colors should be derivable from reply category metadata.
- ME2 TLK loading uses base `BIOGame_<language>.tlk`, then DLC TLKs ordered by
  `Mount.dlc` priority and module metadata.

## 10. PCC Package Requirements

The package parser must support ME2 OT PCC files including LZO-compressed packages.

Required behavior:

- Validate minimum file size and header bounds before reading tables.
- Decode package version fields and infer ME2 OT profile.
- Reconstruct decompressed package bytes before table parsing when compressed.
- Read name table entries as Unreal strings.
- Read import and export tables with serial offsets and sizes.
- Resolve export class names and object names.
- Keep raw offsets available for evidence scanning and container mapping.
- Reject unsupported package profiles with clear JSON errors.

Unsupported behavior:

- Writing package files.
- Recompressing packages.
- Editing import/export tables.

## 11. TLK Requirements

### 11.1 Base TLK Parsing

The TLK parser must:

- Read header metadata.
- Read male and female string table counts.
- Decode Huffman tree nodes.
- Decode text bitstreams into UTF-8 strings.
- Preserve StringRef IDs.
- Return deterministic search results.

### 11.2 DLC Resolution

The TLK resolver must:

- Load base TLK first.
- Scan DLC roots under `BioGame/DLC` or the supplied DLC directory.
- Read `Mount.dlc` priorities.
- Resolve ME2 module numbers through DLC metadata such as `BIOEngine.ini` entries
  where available.
- Load language-specific module TLKs using `DLC_<module>_<language>.tlk` naming.
- Apply priority ordering so final resolved text matches effective game behavior.
- Include `source_tlk` in resolution output.

Missing StringRefs should be represented as not found rather than silently replaced
with unrelated text.

## 12. Graph Layout Requirements

The graph layout output must include enough metadata for consumers to render a graph
without reparsing AST internals.

Required output fields:

- `conversation_id`
- `node_count`
- `positions`
- `edges`
- `nodes`

Node keys use string form:

- `start:<id>`
- `entry:<id>`
- `reply:<id>`

Edge metadata includes:

- `from`: node key object with `type` and `id`.
- `to`: node key object with `type` and `id`.
- `category`: optional reply category.
- `paraphrase_text`: optional player-facing choice text.
- `input_index`: optional edge order.

Layout rules:

- Deterministic for identical input.
- Start nodes should appear before reachable entries where possible.
- Long or cyclic structures should still produce usable output.
- The consumer owns viewport transforms and hit testing; core owns graph positions.

## 13. Evidence Requirements

Evidence reports connect text search results to package locations and, when possible,
conversation AST context.

Report fields:

- `query`
- `tlk_path`
- `dlc_dir`
- `biogame_root`
- `candidate_strrefs`
- `files_scanned`
- `files_with_hits`
- `total_hits`
- `evidence`
- `errors`

Evidence tiers:

- `bioconversation`: parsed AST node directly references the StringRef.
- `semantic_container`: package/export context is semantically meaningful but not a
  direct AST node match.
- `container_fallback`: byte-level or export-level hit with limited semantic context.

Evidence enrichment rules:

- TLK search creates candidate StringRefs.
- PCC scanning finds candidate StringRefs in package exports.
- BioConversation AST matches outrank container fallback matches.
- Owner context should be included when scan-owner data is available.
- Byte scanning remains a fallback, not the preferred semantic source.

Scan root rule:

```text
scan-evidence --biogame-root expects a ME2 BioGame root containing CookedPC/ and
optionally DLC/. A flat output/ directory is only a sample folder, not a BioGame root.
```

## 14. Validation Requirements

Validation must produce actionable results without hiding parser uncertainty.

Validation should report:

- Missing or malformed key dialogue properties.
- Empty stubs.
- Low-confidence parse modes.
- Invalid entry-to-reply links.
- Invalid reply-to-entry links.
- Orphaned entries.
- Orphaned replies.
- Dangling links.
- Missing or unresolved speaker IDs.
- Missing line StringRefs where expected.
- Cycles or traversal anomalies when relevant.

Validation result statuses:

- `valid`: no blocking issues.
- `warning`: usable output with caveats.
- `invalid`: malformed or semantically broken output.

Strict mode must consistently promote warning/failure behavior and affect exit status
when configured to do so. Warnings should include a `cause` where the parser can
provide one.

## 15. CLI Requirements

The Python CLI command is `pcc-toolkit` and is registered by `pyproject.toml`.

CLI command tree:

```text
pcc-toolkit
├── package
│   ├── list <file> [--class CLASS] [--json]
│   ├── inspect <file> <index> [--json]
│   ├── validate <file> [--strict] [--json]
│   └── extract <file> [--output PATH] [--tlk PATH] [--dlc-dir PATH] [--pretty]
├── tlk
│   ├── info <file> [--json]
│   ├── search <query> --file PATH [--json]
│   ├── resolve <strref> --file PATH [--dlc-dir PATH] [--language CODE] [--json]
│   └── dump <file> [--output PATH]
├── dialogue
│   ├── list <file> [--json]
│   ├── export <file> [--output PATH] [--tlk PATH] [--dlc-dir PATH]
│   │              [--language CODE] [--conv-index N] [--pretty]
│   └── graph <file> --conv-index N [--algorithm sugiyama] [--json]
├── evidence
│   └── scan <query> --tlk PATH [--dlc-dir PATH] [--language CODE]
│                    [--biogame-root PATH] [--output PATH] [--json]
├── batch
│   ├── validate <dir> [--glob PATTERN] [--output PATH] [--json]
│   └── extract <dir> [--glob PATTERN] [--output-dir PATH]
│                    [--tlk PATH] [--dlc-dir PATH] [--language CODE] [--json]
├── dev
│   ├── build-core
│   └── test-core
└── --version
```

CLI rules:

- Every data operation calls the Go core through `engine.py`.
- CLI may format JSON into tables and summaries.
- CLI may write JSON output files.
- CLI must not parse PCC/TLK bytes or build domain objects itself.
- CLI should surface core JSON errors clearly.

## 16. Build, Release, And Installation

### 16.1 Build Core

Standalone build:

```powershell
.\scripts\build.ps1 -Version "0.2.0"
```

Default build target is Windows. Output:

```text
build/pcc-core.exe
```

CLI-assisted build:

```powershell
.venv\Scripts\pcc-toolkit.exe dev build-core
```

### 16.2 Install CLI

```powershell
.venv\Scripts\python.exe -m pip install -e .[cli]
```

### 16.3 Create Local Release

```powershell
.\scripts\release.ps1 -Version "0.2.0" -Archive
```

Expected local artifacts:

```text
release/pcc-toolkit-v0.2.0-windows-amd64/
├── pcc-core.exe
└── INSTALL.txt

release/pcc-toolkit-v0.2.0-windows-amd64.zip
```

Published MVP binary:

| Binary         | Platform      | Built from           |
| -------------- | ------------- | -------------------- |
| `pcc-core.exe` | Windows amd64 | `core/cmd/pcc-core/` |

Build and release outputs are never committed.

## 17. Testing And Verification

### 17.1 Automated Tests

Go tests:

```powershell
cd core
go test ./...
```

Python tests:

```powershell
.venv\Scripts\python.exe -m pytest
```

### 17.2 Golden Files

Golden files live under `tests/golden/` and define stable expected outputs for:

- Conversation parsing.
- Graph layout.
- PCC export listing.
- TLK info, search, and resolution.
- Serialization and validation outputs.

Golden rules:

- Do not edit golden files manually unless explicitly regenerating them.
- Regeneration must be justified by an intentional contract or parser behavior change.
- Real input files must be copied to `output/`; they are not committed.

### 17.3 Real-File Probes

Real-file probes use copied local game assets:

```powershell
.venv\Scripts\python.exe tests\regression\run_probes.py --samples-dir output
```

Do not run tests directly against the game installation tree.

### 17.4 Release Smoke Tests

Before declaring an MVP release, validate:

- Extract release zip into a clean temporary directory.
- Run extracted `pcc-core.exe version`.
- Run `parse-pcc` against a copied PCC.
- Run `parse-tlk` and `resolve-tlk` against copied TLKs.
- Run `parse-conversations`, `layout-graph`, `validate`, and `dump-lines`.
- Run `scan-evidence` against a temporary `BioGame/CookedPC` tree.
- Run representative CLI commands against the repository build.

## 18. Known Limitations

- Windows-only distribution for MVP.
- Read-only data access; no PCC writing or editing.
- ME2 OT only.
- `pcc-core.exe --help` is minimal because core is primarily consumed through CLI.
- `scan-evidence --biogame-root` requires a BioGame-style directory and will not scan
  a flat `output/` directory as package root.
- Semantic parsing can still fall back for unusual or malformed conversation exports.
- Evidence tiers may include fallback byte-level hits when AST context is unavailable.

## 19. Roadmap After MVP

### 19.1 Contract Hardening

- Add formal JSON schema or schema-like validation for core outputs.
- Add more golden files for edge-case conversations and DLC precedence.
- Improve core help output for direct `pcc-core.exe` users.

### 19.2 Semantic Expansion

- Expand semantic parsing for additional ME2 dialogue fields.
- Improve script, sequence, and FaceFX metadata coverage where useful for read-only display.
- Add more owner/context extraction where Kismet structures expose it.

### 19.3 Evidence Improvements

- Improve ranking of semantic container hits.
- Add saved evidence report browsing.

### 19.4 Future Scope Decisions

- Editing support requires a separate PRD.
- Legendary Edition support requires a separate PRD and different package/compression assumptions.
- Non-Windows distribution can be revisited after the Windows MVP is stable.

## 20. Reconstruction Notes

If this project is rebuilt in another language, preserve these invariants:

- Keep one domain implementation as the source of truth.
- Keep UI and CLI adapters thin.
- Preserve JSON contracts at process boundaries.
- Preserve ME2 OT-only scope unless intentionally expanded.
- Preserve LegendaryExplorer semantic parity without copying GPLv3 code.
- Preserve golden-file regression strategy.
- Preserve local asset rules and never commit game files.
- Preserve deterministic graph layout output.
- Preserve TLK/DLC priority behavior.
- Preserve parser warnings and validation causes rather than hiding uncertainty.

The implementation language can change; the contracts, scope, and verification model
should not change without an explicit product decision.
