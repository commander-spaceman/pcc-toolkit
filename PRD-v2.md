# PCC Toolkit v2 PRD

## 1. Purpose

PCC Toolkit v2 is a CLI and GUI toolkit for inspecting, extracting, validating, and analyzing Mass Effect 2 Original Trilogy dialogue data from `.pcc` and `.tlk` files.

The product has two primary users:

- Modders who need a fast read-only explorer for ME2 OT packages, TLK text, dialogue graphs, and evidence trails.
- AI agents and automation scripts that need stable JSON contracts for repeatable analysis and regression testing.

The current implementation is a Go domain engine with thin Python CLI and GUI layers. This PRD defines the next target state, using LegendaryExplorer as the behavioral reference where it applies to ME2 OT dialogue semantics.

## 2. Scope

### In Scope

- Mass Effect 2 Original Trilogy only.
- ME2 OT `.pcc` package parsing, including LZO-compressed packages.
- ME2 OT `.tlk` parsing, text search, and StringRef resolution.
- DLC-aware TLK resolution using `Mount.dlc` priority and language-specific TLK filenames.
- `BioConversation` extraction from package exports.
- Dialogue AST enrichment with speaker, listener, reply, condition, transition, export ID, skippable, ambient, non-text, camera, GUI style, reply type, and reply category metadata where available.
- Deterministic dialogue graph layout and JSON graph metadata for GUI rendering.
- Evidence scanning and tiered evidence reports.
- Normalized dialogue line dumping.
- Kismet conversation owner scanning.
- Golden and real-file regression tests using local game files copied into `dropzone/`.

### Out of Scope

- LE1, LE2, LE3, ME1, and ME3 support.
- Editing or writing PCC dialogue data.
- FaceFX, Wwise, Matinee, or InterpData editing.
- Direct test execution against `C:\Program Files\EA Games\Mass Effect 2`.
- Committing game files, generated outputs, local caches, or built binaries.

## 3. Reference Implementation Findings

LegendaryExplorer reference areas scanned from `ME3Tweaks/LegendaryExplorer` on branch `Beta`:

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

### LEX Semantics To Preserve

- `BioConversation` is parsed from export properties, not from byte pattern scanning when structured properties are available.
- `m_StartingList` maps start-node order to entry indexes.
- `m_SpeakerList` contains conversation speaker tags for ME1/ME2-style conversations.
- LEX synthesizes special speakers before parsed speakers: `player` has speaker ID `-2`, StringRef `125303`, friendly name `"Shepard"`; `owner` has speaker ID `-1`, StringRef `0`, friendly name `No data`.
- Entry nodes use `nSpeakerIndex`; reply nodes are player choices and conceptually use speaker ID `-2`.
- Entry-to-reply links come from `ReplyListNew` entries with `nIndex`, `srParaphrase`, `sParaphrase`, and `Category`.
- Reply-to-entry links come from `EntryList` integer arrays.
- Node conditions use `nConditionalFunc`, `nConditionalParam`, and `bFireConditional`.
- State transitions use `nStateTransition` and `nStateTransitionParam`.
- Dialogue-node identity for sequence correlation uses `nExportID`.
- LEX graph node identity convention is `entry ID < 1000`, `reply UID = reply ID + 1000`, `start UID = start ID + 2000`.
- Dialogue graph edges are typed by source and target role: start to entry, entry to reply, reply to entry.
- Entry reply-choice edge colors are derived from `EReplyCategory`.
- LEX uses curved bezier edges and separate back, edge, and node layers; PCC Toolkit may render differently, but graph metadata should support equivalent presentation.
- ME2/LE2 TLK loading uses base `BIOGame_<language>.tlk`, then DLC TLKs ordered by `Mount.dlc` mount priority.
- ME2 DLC module TLK filenames are resolved through DLC module metadata, including `BIOEngine.ini` `[Engine.DLCModules]` entries and `DLC_<module>_<language>.tlk` names.

## 4. Architecture

```text
pcc-toolkit/
├── core/        Go domain engine and JSON producer
├── cli/         Python Typer command wrapper
├── gui/         Python Dear ImGui renderer
├── tests/       Golden and real-file regression tests
├── dropzone/    Local copied ME2 OT test assets, gitignored
└── output/      Local generated outputs, gitignored
```

### Ownership Rules

- Go core owns all domain logic: package parsing, property decoding, AST building, TLK resolution, graph layout, validation, evidence, line dumping, owner scanning, and JSON serialization.
- Python CLI owns only argument parsing, subprocess dispatch, terminal formatting, and file output.
- Python GUI owns only UI state, widgets, viewport interaction, async subprocess handling, and rendering of JSON returned by the core.
- JSON stdout/stderr is the integration boundary.

## 5. Current Core Capabilities

The current `pcc-core` binary reports version `0.2.0`, target `me2_ot`, and these capabilities:

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

## 6. Domain Model Target

### Conversation

Required fields:

- `id`: export object name.
- `export_index`: package export index.
- `game_profile`: always `me2_ot` in v2.
- `parse_mode`: parser path used, such as `struct_property_semantic`, `row_payload`, or fallback mode.
- `entries`: entry nodes.
- `replies`: reply nodes.
- `speakers`: synthetic and parsed speakers.
- `starts`: start nodes.
- `warnings`: parser warnings that do not invalidate output.

Future optional fields:

- `script_list`: names parsed from `m_ScriptList`.
- `facefx_sets`: object refs for speaker FaceFX sets.
- `sequence_ref`: object ref or import/export name for `MatineeSequence`.
- `wwise_bank_ref`: detected Wwise bank reference.

### Entry Node

Target fields:

- `id`
- `speaker_id`
- `speaker_tag`
- `listener_index`
- `listener_tag`
- `line_strref`
- `line_text`
- `reply_links`
- `conditional_func`
- `conditional_param`
- `fires_conditional`
- `state_transition`
- `state_transition_param`
- `script_index`
- `export_id`
- `skippable`
- `non_text_line`
- `ambient`
- `camera_intimacy`
- `gui_style`

### Reply Node

Target fields:

- `id`
- `line_strref`
- `line_text`
- `target_entry_ids`
- `condition_refs`
- `category`
- `reply_type`
- `conditional_func`
- `conditional_param`
- `fires_conditional`
- `state_transition`
- `state_transition_param`
- `script_index`
- `export_id`
- `unskippable`
- `non_text_line`
- `ambient`
- `camera_intimacy`
- `gui_style`

### Reply Choice

The AST should preserve the distinction between a reply node and an entry's reply choice metadata. `ReplyListNew` contains the choice edge metadata, while `m_ReplyList` contains the target reply node.

Target edge metadata:

- `from_entry_id`
- `to_reply_id`
- `order`
- `paraphrase`
- `paraphrase_strref`
- `paraphrase_text`
- `category`

This is a current gap: `reply_links` preserves target IDs, but paraphrase and category are not yet modeled as first-class edge metadata in the public AST.

### Speaker

Required behavior:

- Always include synthetic `player` and `owner` speakers when semantic parsing has enough context.
- Preserve negative IDs: `player = -2`, `owner = -1`.
- Preserve parsed `m_SpeakerList` entries by array index.

Target fields:

- `id`
- `tag`
- `display_name`
- `strref_id`
- `friendly_name`
- `facefx_male_ref`
- `facefx_female_ref`

FaceFX refs are deferred metadata, not required for v2 read-only dialogue extraction.

## 7. CLI Contract

Core subcommands remain the source of truth:

- `version`
- `parse-pcc`
- `parse-tlk`
- `resolve-tlk`
- `parse-conversations`
- `layout-graph`
- `scan-evidence`
- `validate`
- `serialize`
- `batch-validate`
- `batch-extract`
- `dump-lines`
- `scan-owners`

CLI commands should stay thin wrappers over these subcommands. CLI must not reimplement parsing, TLK resolution, graph generation, or validation.

## 8. Graph And GUI Requirements

### Graph Output Requirements

The Go graph output must include enough metadata for the GUI to render without reparsing AST internals:

- `conversation_id`
- `node_count`
- `positions`
- `edges`
- `nodes`

Target edge metadata should be expanded to include:

- `source_type`
- `source_id`
- `target_type`
- `target_id`
- `category`
- `paraphrase_text`
- `input_index`

### GUI Requirements

- Render start, entry, and reply nodes with distinct styling.
- Render category-colored reply-choice edges compatible with LEX category semantics.
- Keep hit-testing, pan, zoom, selection, and tab state in Python only.
- Do not duplicate AST parsing or graph layout in Python.
- Read-only is the v2 constraint. Editing support requires a future PRD.

## 9. TLK Resolution Requirements

### Base ME2 OT Resolution

- Base TLK path: `BIOGame/CookedPC/BIOGame_<language>.tlk`.
- Default language: `INT`.
- Missing StringRefs should be represented explicitly, not silently labeled as player or owner text.

### DLC Resolution

- DLC root: `BioGame/DLC`.
- Scan `DLC_*` folders.
- Read `Mount.dlc` from each DLC cooked directory.
- Sort DLCs by mount priority before loading TLKs.
- For ME2 OT, resolve module numbers from `BIOEngine.ini` `[Engine.DLCModules]` where available.
- Load module TLKs using `DLC_<module>_<language>.tlk`.
- Resolution should prefer the final effective text after applying the same priority order as LegendaryExplorer.

## 10. Evidence Requirements

Evidence reports must distinguish:

- Candidate StringRefs from TLK search.
- Raw scan hits in package exports.
- BioConversation AST matches.
- Semantic container matches.
- Container fallback matches.
- Kismet owner context where available.

Evidence enrichment should prefer AST context over byte-offset context. Byte scanning remains a fallback, not the primary semantic source.

## 11. Validation Requirements

Validation must report at least:

- Missing or malformed `m_EntryList`.
- Missing or malformed `m_ReplyList`.
- Missing or malformed `m_StartingList`.
- Invalid entry-to-reply links.
- Invalid reply-to-entry links.
- Missing or unresolved speaker IDs.
- Missing line StringRefs where expected.
- Cycles in graph traversal when relevant.
- Missing key properties that force fallback parsing.

Strict mode must change result severity or process exit behavior consistently. It must not be accepted and ignored.

## 12. Testing And Verification

### Local Asset Rules

- Real ME2 OT files must be copied from `C:\Program Files\EA Games\Mass Effect 2` into `dropzone/`.
- Do not run golden tests directly against the install tree.
- Do not commit copied game files.

### Commands

- Go tests: from `core/`, run `go test ./...`.
- Python tests: from repo root, run `.venv\Scripts\python.exe -m pytest`.
- Real-file probes: run `.venv\Scripts\python.exe tests\regression\run_probes.py --samples-dir dropzone`.

### Current Real-File Probe Set

The regression runner currently validates package headers, conversation parsing, graph layout, TLK info/search/resolve, validation, serialization, and known conversation edge cases against copied ME2 OT assets.

## 13. Roadmap

### Milestone A: Contract Cleanup

- Add explicit reply-choice edge metadata to conversation AST.
- Add graph edge category and paraphrase metadata.
- Ensure all current CLI flags are documented in `PRD.md`, `PRD-v2.md`, and `MAP.md` when changed.
- Add JSON schema or schema-like validation for core outputs.

### Milestone B: LEX Semantic Parity

- Expand semantic parsing for fields present in `DialogueNodeExtended` but missing from the public AST.
- Add script list extraction from `m_ScriptList`.
- Add optional FaceFX object reference extraction for speakers.
- Add optional sequence reference extraction from `MatineeSequence`.
- Add optional Wwise bank reference extraction where feasible for ME2 OT.

### Milestone C: TLK/DLC Hardening

- Verify DLC TLK resolution against multiple official DLCs and installed language variants.
- Validate `BIOEngine.ini` module-number parsing across official DLC folder layouts.
- Add golden probes for DLC override precedence.

### Milestone D: GUI Usability

- Render category-colored edges from graph metadata.
- Show speaker, line, condition, transition, and export ID metadata in node details.
- Add evidence-to-conversation navigation.
- Add graceful loading and error states for missing core binary or missing game files.

### Milestone E: Release Readiness

- Add repeatable build instructions for `pcc-core` without committing binaries.
- Add smoke tests for installed `.venv\Scripts\pcc-toolkit.exe` entry points.
- Add CI-friendly tests that do not require game assets.
- Keep real-file tests as local/manual regression gates due to game-file licensing.

## 14. Acceptance Criteria

- `go test ./...` passes from `core/`.
- `.venv\Scripts\python.exe -m pytest` passes when required assets are present in `dropzone/`.
- `.venv\Scripts\python.exe tests\regression\run_probes.py --samples-dir dropzone` passes on the maintained local ME2 OT sample set.
- `pcc-core version` reports `me2_ot` and the implemented capability list.
- Python CLI and GUI do not contain domain parsing logic.
- New parser behavior is validated against golden files or explicitly regenerated golden outputs.
- Any semantic behavior touching PCC parsing, TLK resolution, dialogue graph semantics, or validation cites LegendaryExplorer behavior during implementation review.

## 15. Open Questions

- Should reply-choice edge metadata become part of `parse-conversations`, `layout-graph`, or both?
- Should FaceFX/Wwise/Sequence refs be exposed as stable public JSON now, or remain internal evidence enrichments until GUI needs them?
- Should `PRD.md` eventually be replaced by this document, or should `PRD-v2.md` remain the forward-looking product contract while `PRD.md` keeps historical migration notes?
- Should `batch-extract` and `dump-lines` get first-class CLI commands beyond the existing engine wrappers?
