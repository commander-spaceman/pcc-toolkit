# JSON Contracts

This document describes the JSON contracts produced by `pcc-core` and consumed by the
Python CLI, tests, golden files, and automation scripts.

## Contract Rules

- Success payloads are written to **stdout** as JSON.
- Error payloads are written to **stderr** as `{"error":"<message>"}`.
- Exit code `0` on success, non-zero on failure.
- Subcommands that accept `--pretty` indent JSON output.
- All field names use `snake_case`.
- Optional fields are omitted from output when unavailable.
- Integer fields use Go's `int` type (platform-dependent size); consumers must handle
  64-bit integers.

---

## Version

```
pcc-core version
```

```json
{
  "version": "0.3.0",
  "target": "me2_ot",
  "capabilities": [
    "pcc_parse_v1",
    "conversation_ast_v1",
    "edit_conversation_v1",
    "..."
  ]
}
```

---

## Parse PCC

```
pcc-core parse-pcc --file <path> [--exports-only] [--export-index N]
                    [--property-tags] [--semantic-props] [--pretty]
```

```json
{
  "file": "BioD_CitHub_LOC_INT.pcc",
  "game_profile": "me2_ot",
  "compressed": true,
  "header": {
    "signature": 3917500530,
    "version": 512,
    "licensee_version": 130,
    "name_count": 1234,
    "name_offset": 64,
    "export_count": 500,
    "export_offset": 123456,
    "import_count": 200,
    "import_offset": 654321
  },
  "names": ["ByteProperty", "IntProperty", "None", "..."],
  "imports": [
    {
      "package_file": -1,
      "class_package": 0,
      "class_name": 5,
      "link": 0,
      "object_name": 10
    }
  ],
  "exports": [
    {
      "class_index": 0,
      "super_index": 0,
      "package_index": 0,
      "object_name": 15,
      "archetype_index": 0,
      "serial_offset": 50000,
      "serial_size": 2048,
      "export_flags": 1,
      "class_name": "BioConversation",
      "object_name_str": "BioD_CitHub_400Conv",
      "serial_data_b64": "<base64 when requested>"
    }
  ]
}
```

---

## Parse TLK

```
pcc-core parse-tlk --file <path> [--search <query>] [--strref <id>]
                   [--dump-all] [--pretty]
```

```json
{
  "file": "BIOGame_INT.tlk",
  "entry_count": 50000,
  "male_count": 25000,
  "female_count": 25000,
  "entries": [
    {
      "strref_id": 0,
      "text": "Hello, Shepard."
    }
  ]
}
```

---

## Resolve TLK

```
pcc-core resolve-tlk --base <path> [--dlc-dir <path>] [--language INT]
                     --strref <id> [--strref <id> ...] [--pretty]
```

```json
{
  "resolved": [
    {
      "strref_id": 12345,
      "text": "I should go.",
      "source_tlk": "DLC_EXP_Part01_INT.tlk",
      "found": true
    }
  ]
}
```

---

## Parse Conversations

```
pcc-core parse-conversations --file <path> [--conv-index N]
                             [--resolve-tlk <path>] [--dlc-dir <path>]
                             [--language INT] [--mode resilient|strict]
                             [--pretty]
```

```json
{
  "file": "BioD_CitHub_LOC_INT.pcc",
  "game_profile": "me2_ot",
  "conversations": [
    {
      "id": "BioD_CitHub_400Conv",
      "export_index": 15,
      "game_profile": "me2_ot",
      "parse_mode": "struct_property_semantic",
      "entries": [
        {
          "id": 0,
          "speaker_id": 0,
          "speaker_tag": "citHub_Ambassador",
          "line_strref": 12345,
          "line_text": "Welcome to the Citadel.",
          "reply_links": [0, 1],
          "reply_choices": [
            {
              "from_entry_id": 0,
              "to_reply_id": 0,
              "order": 0,
              "paraphrase": "Investigate",
              "paraphrase_strref": 67890,
              "category": "Investigate"
            }
          ],
          "skippable": true,
          "non_text_line": false,
          "ambient": false,
          "camera_intimacy": 0,
          "gui_style": 0
        }
      ],
      "replies": [
        {
          "id": 0,
          "line_strref": 54321,
          "line_text": "Tell me more.",
          "target_entry_ids": [1],
          "category": "Neutral",
          "unskippable": false
        }
      ],
      "speakers": [
        { "id": -2, "tag": "player", "friendly_name": "Shepard" },
        { "id": -1, "tag": "owner" },
        { "id": 0, "tag": "citHub_Ambassador", "strref_id": 11111 }
      ],
      "starts": [{ "id": 0, "target_entry_ids": [0], "label": "Start" }],
      "script_list": [{ "id": 0, "tag": "Script_OnStart", "name": "OnStart" }]
    }
  ],
  "errors": []
}
```

Parse modes reported in `parse_mode`:

- `struct_property_semantic` — Schema-guided semantic parsing.
- `row_payload` — Row-mode fallback for partial layouts.
- `row_payload_struct_matrix` — Matrix-based row extraction.
- `row_payload_struct_head` — Head-struct row extraction.
- `count_or_value_fallback` — Minimal fallback for unusual exports.

---

## Layout Graph

```
pcc-core layout-graph --file <path> --conv-index N
                      [--algorithm sugiyama]
                      [--node-width <px>] [--node-height <px>]
                      [--x-spacing <px>] [--y-spacing <px>]
                      [--pretty]
```

```json
{
  "conversation_id": "BioD_CitHub_400Conv",
  "node_count": 25,
  "positions": {
    "start:0": { "x": 100.0, "y": 0.0 },
    "entry:0": { "x": 100.0, "y": 100.0 },
    "reply:0": { "x": 300.0, "y": 100.0 }
  },
  "edges": [
    {
      "from": { "type": "start", "id": "0" },
      "to": { "type": "entry", "id": "0" }
    },
    {
      "from": { "type": "entry", "id": "0" },
      "to": { "type": "reply", "id": "0" },
      "category": "Neutral",
      "paraphrase_text": "Tell me more.",
      "input_index": 0
    }
  ],
  "nodes": [
    {
      "key": "entry:0",
      "type": "entry",
      "id": "0",
      "speaker": "citHub_Ambassador",
      "line_strref": 12345,
      "line_text": "Welcome to the Citadel."
    }
  ]
}
```

---

## Scan Evidence

```
pcc-core scan-evidence --query <text> --tlk <path>
                        [--dlc-dir <path>] [--language INT]
                        [--biogame-root <path>] [--cache <path>]
                        [--workers N] [--pretty]
```

```json
{
  "query": "Shepard",
  "tlk_path": "BIOGame_INT.tlk",
  "dlc_dir": "",
  "biogame_root": "C:\\BioGame",
  "candidate_strrefs": [1, 2, 3],
  "files_scanned": 50,
  "files_with_hits": 12,
  "total_hits": 45,
  "evidence": [
    {
      "tier": "bioconversation",
      "strref_id": 12345,
      "file": "BioD_CitHub_LOC_INT.pcc",
      "export_index": 15,
      "export_name": "BioD_CitHub_400Conv",
      "node_type": "entry",
      "node_id": 0,
      "context": "speaker: citHub_Ambassador"
    }
  ],
  "errors": []
}
```

Evidence tiers:

- `bioconversation` — Direct AST node references the StringRef.
- `semantic_container` — Export context is meaningful but not a direct AST match.
- `container_fallback` — Byte-level hit with limited context.

---

## Validate

```
pcc-core validate --file <path> [--strict] [--pretty]
```

```json
{
  "file": "BioD_CitHub_LOC_INT.pcc",
  "game_profile": "me2_ot",
  "total": 5,
  "valid": 4,
  "warning": 1,
  "invalid": 0,
  "conversations": [
    {
      "id": "BioD_CitHub_400Conv",
      "export_index": 15,
      "status": "valid",
      "warnings": [],
      "issues": []
    }
  ]
}
```

---

## Serialize

```
pcc-core serialize --file <path> [--resolve-tlk <path>]
                   [--dlc-dir <path>] [--language INT] [--pretty]
```

Returns a combined payload with parse results, validation report, and
optional TLK-resolved text. This is the backend for `pcc-toolkit package extract`.

---

## Dump Lines

```
pcc-core dump-lines --file <path> [--resolve-tlk <path>]
                     [--dlc-dir <path>] [--language INT]
                     [--format json|csv] [--pretty]
```

```json
[
  {
    "conversation_id": "BioD_CitHub_400Conv",
    "export_index": 15,
    "node_type": "entry",
    "node_id": 0,
    "speaker_tag": "citHub_Ambassador",
    "line_strref": 12345,
    "line_text": "Welcome to the Citadel.",
    "file": "BioD_CitHub_LOC_INT.pcc"
  }
]
```

---

## Scan Owners

```
pcc-core scan-owners --file <path> [--pretty]
```

```json
{
  "file": "BioD_CitHub_LOC_INT.pcc",
  "owners": [
    {
      "conversation_export_index": 15,
      "conversation_name": "BioD_CitHub_400Conv",
      "owner_tag": "citHub_Ambassador",
      "kismet_export_index": 0,
      "kismet_export_name": "SeqAct_StartConversation_0"
    }
  ]
}
```

---

## Edit Conversation

```
pcc-core edit-conversation --file <path> --conv-index N --patch <path>
                           --output <path> [--dry-run] [--tlk <path>]
                           [--tlk-output <path>] [--backup] [--pretty]
```

```json
{
  "status": "written",
  "output": "output.pcc",
  "validation": {
    "id": "BioD_CitHub_400Conv",
    "export_index": 15,
    "status": "valid",
    "warnings": [],
    "issues": []
  }
}
```

When `--dry-run` is used, `status` is `"dry_run"` and `output` is omitted.

---

## Batch Operations

### Batch Validate

```
pcc-core batch-validate --dir <path> [--glob <pattern>]
                        [--strict] [--output <path>] [--pretty]
```

```json
{
  "files_scanned": 50,
  "files_with_conversations": 30,
  "total_conversations": 150,
  "valid": 140,
  "warning": 8,
  "invalid": 2,
  "errors": []
}
```

### Batch Extract

```
pcc-core batch-extract --dir <path> [--glob <pattern>]
                       [--output-dir <path>] [--resolve-tlk <path>]
                       [--dlc-dir <path>] [--language INT] [--pretty]
```

Each matching PCC produces one JSON output file in the output directory.

### Batch Edit

```
pcc-core batch-edit --dir <path> --patch <path>
                    [--glob <pattern>] [--output-dir <path>]
                    [--tlk <path>] [--tlk-output <path>]
                    [--dry-run] [--backup] [--pretty]
```

```json
{
  "files_edited": 5,
  "files_skipped": 0,
  "errors": []
}
```
