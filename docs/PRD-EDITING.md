# PCC Toolkit Product Requirements Document — Editing & Writing

## 1. Purpose

This document defines the requirements for editing and writing capabilities in
PCC Toolkit. It covers PCC package writing, PCC binary patching, TLK file writing,
Unreal property encoding, dialogue binary serialization, and conversation editing
with round-trip fidelity.

These capabilities extend the read & inspect foundation defined in
[PRD-INSPECT.md](PRD-INSPECT.md) and share the same scope constraints: ME2 OT only,
Windows amd64, Go engine + Python CLI wrapper.

## 2. Scope

### 2.1 In Scope

- PCC package writing with optional LZO compression.
- Unreal property encoding for PCC write workflows.
- Binary-level patching of PCC packages at specific export offsets.
- Dialogue AST-to-binary encoding (entries, replies, links, speakers).
- Conversation editing with property span scanning and byte-level splice for
  round-trip fidelity.
- TLK file writing with Huffman code table building and string encoding.
- TLK text integration in edit workflows (resolve text → StringRef, write updated TLK).
- Post-edit validation to verify edited conversations remain structurally sound.
- Dry-run mode for previewing edits without writing.
- Automatic backup creation (`--backup`).
- Batch edit across directories with glob matching.
- JSON patch contract for describing conversation modifications.
- Golden-file regression coverage for edit outputs.
- Additive name table entries for new properties not present in the original file.

### 2.2 Out of Scope

- Legendary Edition support: LE1, LE2, LE3.
- Other Original Trilogy games: ME1, ME3.
- Editing non-dialogue PCC data (textures, meshes, materials, levels).
- FaceFX, Wwise, Matinee, or InterpData editing.
- Creating new PCC files from scratch.
- Console package support.
- Non-Windows distribution.

## 3. Architecture

Editing and writing capabilities are implemented across these Go packages:

```text
core/internal/
├── pccenc/      Unreal property encoding
├── pccwrt/      PCC package writing with optional LZO compression
├── pccpat/      Binary-level export patching
├── dialenc/     Dialogue AST → binary property encoding
├── editor/      High-level conversation edit orchestrator
└── tlkwrt/      TLK file writing with Huffman encoding
```

External libraries used:

- [`me2pcc`](https://github.com/commander-spaceman/me2pcc) — PCC reading (used to
  parse the original file before editing).
- [`me2lzo`](https://github.com/commander-spaceman/me2lzo) — LZO compression for
  PCC write (`pccwrt/compress.go`).
- [`me2tlk`](https://github.com/commander-spaceman/me2tlk) — TLK reading (used to
  resolve text → StringRef in edit workflows).

### 3.1 Ownership Rules

| Capability                              | Go package           |
| --------------------------------------- | -------------------- |
| Unreal property encoding                | `pccenc`             |
| PCC file writing (header, tables, data) | `pccwrt`             |
| LZO compression for PCC write           | `pccwrt/compress.go` |
| Binary export patching                  | `pccpat`             |
| Dialogue AST → binary encoding          | `dialenc`            |
| Property span scanning (round-trip)     | `editor/preserve.go` |
| Conversation serialization for write    | `editor/conv_ser.go` |
| High-level edit orchestration           | `editor/editor.go`   |
| TLK file writing                        | `tlkwrt`             |

## 4. PCC Writing (`pccwrt`)

### 4.1 Requirements

- Write complete PCC files including header, name table, import table, export table,
  and export serial data.
- Support optional LZO compression for ME2 OT packages.
- Reject non-ME2 profiles with clear errors.
- Write all name table entries as Unreal strings.
- Write import and export tables with correct serial offsets and sizes.
- Preserve header fields from the source file where possible.

### 4.2 Compression

When compression is requested:

- Compress export serial data blocks using ME2 OT LZO chunking.
- Update header compression flags and block table accordingly.
- Preserve the same chunk size as the source package where possible.

## 5. PCC Property Encoding (`pccenc`)

### 5.1 Requirements

- Encode Unreal property values into binary form for PCC writing.
- Support all property types needed for `BioConversation` data:
  `IntProperty`, `FloatProperty`, `BoolProperty`, `StrProperty`,
  `NameProperty`, `ObjectProperty`, `ArrayProperty`, `StructProperty`,
  `ByteProperty`, `EnumProperty`, and `NoneProperty`.
- Handle nested struct and array properties recursively.
- Maintain a name table and auto-add new names when encoding properties
  that reference names not in the source file.
- Produce deterministic output for identical input.

### 5.2 Property Value Model

```go
type PropertyValue struct {
    Name             string
    PropType         string
    Value            interface{}
    ArrayIndex       int
    StructTypeName   string
    ByteSubTypeName  string
    ArrayElementType string
    Properties       []PropertyValue
    Items            []PropertyValue
}
```

## 6. PCC Binary Patching (`pccpat`)

### 6.1 Requirements

- Apply binary patches to PCC files at specific export serial offsets.
- Update the export table serial size and offset entries after patching.
- Support building minimal valid PCC structures for patch injection workflows.
- Reject patches that would corrupt the file structure.

### 6.2 Use Cases

- Direct byte-level export replacement in the conversation editing pipeline.
- Export table metadata updates after serial data changes.

## 7. Dialogue Encoding (`dialenc`)

### 7.1 Requirements

- Encode `dialogue.Conversation` AST nodes back into binary property form:
  - `EntryNode` → entry list struct properties.
  - `ReplyNode` → reply list struct properties.
  - Reply links → entry-to-reply integer arrays.
  - Speaker list → speaker struct properties.
  - Start list → start node data.
  - Script list → script entry data.
- Omit default/unused values to minimize binary size changes.
- Report new name table additions so the caller can track them.
- Preserve all fields present in the original parsed form.

### 7.2 Additive Name Table

When encoding properties that reference names not present in the source file's
name table, `dialenc` reports those additions. The editor layer then appends them
to the name table before writing, ensuring the output PCC remains valid.

## 8. TLK Writing (`tlkwrt`)

### 8.1 Requirements

- Build Huffman code tables from TLK node trees for bitstream encoding.
- Encode strings into compressed bitstreams using the code table.
- Write complete TLK files with correct headers, male/female entry tables,
  Huffman tree nodes, and encoded bitstreams.
- Support incremental additions: add new entries to an existing TLK without
  corrupting existing entries.
- Preserve existing male/female entry counts and structures.

### 8.2 Integration with Edit Workflow

When `--tlk` is provided to the edit subcommand:

1. The TLK is loaded via `me2tlk`.
2. Text in the JSON patch (e.g., new line text, paraphrases) is resolved against
   existing TLK entries or added as new entries.
3. If `--tlk-output` is specified, the modified TLK is written via `tlkwrt`.

## 9. Conversation Editing (`editor`)

### 9.1 High-Level Workflow

```
1. Read PCC           → me2pcc.ReadFileRaw
2. Parse conversation  → dialogue.ParseConversations (resilient mode)
3. Apply JSON patch    → modify conversation AST (entries, replies, speakers)
4. Post-edit validate  → dialogue.ValidateConversation
5. Scan property spans → editor/preserve.go (locates original prop boundaries)
6. Serialize modified  → editor/conv_ser.go (AST → binary properties)
7. Splice unchanged    → editor/preserve.go (insert modified, keep rest)
8. Update name table   → append new names from dialenc
9. Write output PCC    → pccwrt.WritePCC / WritePCCCompressed
```

### 9.2 Round-Trip Fidelity

The editor preserves unchanged surrounding data through two mechanisms:

**Property Span Scanning** (`editor/preserve.go`):

- Parses property tags at the export serial offset to locate the boundaries
  of each top-level property (`m_EntryList`, `m_ReplyList`, `m_SpeakerList`,
  `m_StartingList`, `m_ScriptList`).
- Records the exact byte range of each property header and value.
- Tries multiple offset deltas to handle UE3 export prefix variance.

**Byte-Level Splice**:

- Only the properties the user modifies are re-serialized.
- Unchanged properties are copied verbatim from the original byte stream.
- After splicing, the serial size is updated in the export table entry via `pccpat`.

### 9.3 JSON Patch Contract

The edit subcommand accepts a JSON patch file with this structure:

```json
{
  "entries": [
    {
      "id": 1,
      "line_text": "New dialogue line text",
      "speaker_tag": "NewSpeaker",
      "reply_links": [0, 1],
      "skippable": true
    }
  ],
  "replies": [
    {
      "id": 0,
      "line_text": "New reply text",
      "target_entry_ids": [2],
      "category": "Paragon",
      "unskippable": true
    }
  ],
  "speakers": [
    {
      "id": 5,
      "tag": "NewSpeaker",
      "display_name": "New Speaker Name"
    }
  ],
  "starts": [
    {
      "id": 0,
      "target_entry_ids": [1]
    }
  ]
}
```

Rules:

- Only specified fields are changed; omitted fields are left unchanged.
- `line_text` is resolved to a StringRef via a loaded TLK when `--tlk` is provided.
- `speaker_tag` on entries resolves by matching parsed speakers; new tags add
  speaker entries.
- Unknown IDs produce errors.

### 9.4 Dry Run and Backup

**Dry Run** (`--dry-run`):

- Performs parse, patch application, post-edit validation, and serialization.
- Reports the validation result without writing any files.
- Useful for previewing whether a patch would produce valid output.

**Backup** (`--backup`):

- When writing the output PCC, creates a `.bak` copy of the original file first.
- The backup path is `<original>.bak`.

## 10. Core Subcommands

### 10.1 `edit-conversation`

```text
pcc-core edit-conversation --file <path>
                           --conv-index <n>
                           --patch <path>
                           --output <path>
                           [--dry-run]
                           [--tlk <path>]
                           [--tlk-output <path>]
                           [--backup]
                           [--pretty]
```

### 10.2 `batch-edit`

```text
pcc-core batch-edit --dir <path>
                    --patch <path>
                    [--glob <pattern>]
                    [--output-dir <path>]
                    [--tlk <path>]
                    [--tlk-output <path>]
                    [--dry-run]
                    [--backup]
                    [--pretty]
```

Applies the same JSON patch to every conversation in matching PCC files.

## 11. CLI Interface

The Python CLI exposes editing through `pcc-toolkit dialogue edit`:

```text
pcc-toolkit dialogue edit <file> <conv-index> --patch <path>
                          [--output PATH]
                          [--dry-run]
                          [--tlk PATH]
                          [--tlk-output PATH]
                          [--backup]
```

Batch editing (`pcc-toolkit batch edit`) is defined in the Go core and PRD
but not yet wired to the Python CLI.

## 12. Validation

Post-edit validation runs the same validation pipeline as read-mode validation
(`dialogue/validate.go`) to ensure edited conversations remain structurally sound.

Validation checks include:

- Missing or malformed key dialogue properties.
- Invalid entry-to-reply and reply-to-entry links.
- Orphaned entries and replies.
- Unresolved speaker IDs.
- Missing line StringRefs where expected.

## 13. Known Limitations

- Adding new entries or replies requires the caller to also manage reply link
  and target entry arrays correctly; invalid links are caught by post-edit validation.
- Very large property additions may require expanding the export serial block,
  which the current patcher handles by updating the size in the export table.
- The JSON patch maps entries and replies by ID, not by array index, so reordering
  nodes requires explicit ID reassignment.
- `batch edit` is implemented in the Go core but not yet wired to the Python CLI.
- TLK writing has a known issue with very large bitstreams (small and medium
  TLK files work correctly).
- Editing is ME2 OT only. Legendary Edition packages have different compression
  and name table structures.

## 14. Testing

- Unit tests in `dialenc`, `pccenc`, `pccwrt`, `editor`, and `tlkwrt` validate
  encode/decode round-trips and edge cases.
- Golden tests under `tests/golden/edit/` verify stable edit output for
  representative conversations.
- The `--dry-run` flag enables validation-only integration tests without
  writing files to disk.
- Real-file edit tests can run against copied ME2 OT assets in `output/`.
