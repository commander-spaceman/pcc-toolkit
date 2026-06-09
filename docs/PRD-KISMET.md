# PCC Toolkit PRD — Kismet / Cinematic Sequence Support

## 1. Purpose

This document defines the scope and requirements for adding Kismet sequence
parsing, editing, and round-trip writing to PCC Toolkit. Kismet is the
visual scripting system in Unreal Engine 3 that controls cinematic presentation
of dialogue: camera placement, character animations, facial expressions,
lighting, and timing.

Currently the toolkit reads and writes `BioConversation` data (the dialogue
text and graph). It does not touch the Kismet sequence layer that drives
_how_ each line is presented on screen. This PRD defines what Kismet support
would look like.

## 2. What Kismet Is

Kismet is Unreal Engine 3's node-based scripting system. In Mass Effect 2,
each `BioConversation` export is linked to a Kismet **sequence** via the
`MatineeSequence` object property. That sequence contains an array of
`SequenceObjects` — individual script nodes connected by output links.

The relevant node types for dialogue are:

```
BioSeqEvt_ConvNode      SeqAct_Interp          InterpData
(m_nNodeID matches       (controls timing,      (camera tracks,
 dialogue node's          camera group,          actor animation,
 nExportID)              audio cues)            FaceFX references)
```

### 2.1 The Chain

```
BioConversation.nExportID
        │
        ▼
BioSeqEvt_ConvNode.m_nNodeID   ──output link──▶   SeqAct_Interp
                                                           │
                                              VariableLinks["Data"]
                                                           │
                                                           ▼
                                                      InterpData
                                                   (camera + animation)
```

1. Each entry/reply node in a `BioConversation` has an `nExportID` (integer).
2. The Kismet sequence contains a `BioSeqEvt_ConvNode` whose `m_nNodeID`
   matches that export ID.
3. The ConvNode is connected via an output link to a `SeqAct_Interp`.
4. The `SeqAct_Interp` has a `VariableLinks` array with a `"Data"` link
   pointing to an `InterpData` export.
5. `InterpData` contains camera tracks, actor animation curves, and
   references to FaceFX assets for lip-sync.

### 2.2 Properties of Key Node Types

**BioSeqEvt_ConvNode:**

- `m_nNodeID` (`IntProperty`) — matches `nExportID` in the dialogue node
- Standard Kismet node properties (output links, position, etc.)

**SeqAct_Interp:**

- `VariableLinks` (`ArrayProperty<StructProperty>`) — each has:
  - `LinkDesc` (`StrProperty`) — e.g. `"Data"`
  - `LinkedVariables` (`ArrayProperty<ObjectProperty>`) — references to
    `InterpData` exports

**InterpData:**

- `InterpGroups` — array of camera, actor, and audio groups
- Each group contains `InterpTracks` with keyframe curves
- References to FaceFX assets for lip-sync

## 3. Scope

### 3.1 In Scope (MVP)

- Parse the Kismet sequence linked to a `BioConversation`
- Map `nExportID` → `BioSeqEvt_ConvNode` → `SeqAct_Interp` → `InterpData`
- Expose this mapping in the conversation AST (add `InterpDataIndex` to
  `EntryNode` and `ReplyNode`)
- Round-trip: when editing a conversation, preserve existing ConvNode →
  InterpData links for unchanged nodes
- When adding new dialogue nodes, create minimal ConvNode + InterpData
  stubs so the game doesn't crash (camera may be static but dialogue plays)
- Dry-run and validation for sequence integrity
- `dialogue export` includes Kismet metadata when `--detailed` is passed

### 3.2 Out of Scope (MVP)

- Editing camera tracks or animation curves
- Creating FaceFX assets
- Full Kismet graph editor
- Wwise audio stream manipulation
- Stage direction parsing (ME3 only)

### 3.3 Future Scope

- Clone InterpData from an existing node as a template for new dialogue
  lines (reuse camera angles, timing)
- Edit basic camera properties (position, field of view)
- Validate audio stream references
- Add new ConvNodes to the Kismet sequence object array

## 4. Current State vs Target

### 4.1 What the Toolkit Already Does

- Parses `BioConversation` exports fully (entries, replies, speakers, starts)
- Edits conversation AST and writes back with round-trip fidelity
- Reads the `MatineeSequence` object reference from the export properties
  (stored as `matinee_sequence_export_id` in the JSON output)
- Does NOT parse the sequence itself

### 4.2 What Needs to Be Built

| Component                         | Package              | Effort |
| --------------------------------- | -------------------- | ------ |
| Kismet sequence parser            | `internal/kismet`    | High   |
| ConvNode → InterpData resolver    | `internal/kismet`    | Medium |
| InterpData struct definitions     | `internal/kismet`    | Medium |
| Export-to-AST enrichment          | `internal/dialogue`  | Low    |
| JSON serialization of Kismet data | `internal/serialize` | Low    |
| ConvNode stub generator           | `internal/kismet`    | Medium |
| InterpData stub generator         | `internal/kismet`    | High   |
| Sequence editor (write path)      | `internal/editor`    | High   |
| Validation for sequences          | `internal/kismet`    | Medium |
| CLI commands                      | `cli/`               | Low    |

### 4.3 LegendaryExplorer Reference Files

- `LegendaryExplorerCore/Kismet/SequenceObjectCreator.cs`
- `LegendaryExplorerCore/Kismet/KismetHelper.cs`
- `LegendaryExplorerCore/Unreal/BinaryConverters/InterpData.cs`
- `LegendaryExplorerCore/Dialogue/ConversationExtended.cs` (lines 160–240)

## 5. Proposed Architecture

### 5.1 New Package: `core/internal/kismet`

```
core/internal/kismet/
├── parser.go          Parse Kismet sequence from a PCC export
├── types.go           ConvNode, InterpData, InterpGroup, InterpTrack structs
├── resolver.go        Map nExportID → InterpData via ConvNode chain
├── stub.go            Generate minimal ConvNode + InterpData for new nodes
├── validate.go        Sequence integrity checks
└── types_test.go
```

### 5.2 Data Model

```go
// ConvNode represents a BioSeqEvt_ConvNode within a Kismet sequence.
type ConvNode struct {
    ExportIndex int    // PCC export index of the ConvNode
    NodeID      int    // m_nNodeID — matches dialogue node's nExportID
    OutputLinks []int  // export indices of linked nodes
}

// InterpDataRef holds the resolved InterpData for a dialogue node.
type InterpDataRef struct {
    InterpDataIndex int  // PCC export index of the InterpData
    ConvNodeIndex   int  // PCC export index of the bridging ConvNode
    ActInterpIndex  int  // PCC export index of the SeqAct_Interp
}

// SequenceMapping maps dialogue nExportIDs to their cinematic data.
type SequenceMapping struct {
    Nodes map[int]InterpDataRef // key = nExportID of dialogue node
}
```

### 5.3 Integration with Conversation AST

Add to `dialogue.EntryNode` and `dialogue.ReplyNode`:

```go
// InterpDataExportIndex is the PCC export index of the InterpData
// that drives this node's cinematic presentation.
// -1 means no cinematic data (or not yet resolved).
InterpDataExportIndex int `json:"interp_data_export_index,omitempty"`
```

### 5.4 JSON Contract

```json
{
  "id": 21,
  "speaker_id": 2,
  "speaker_tag": "hench_tali",
  "line_strref": 260225,
  "line_text": "Good luck, Lia'Vael.",
  "n_export_id": 260225,
  "interp_data_export_index": 1234
}
```

## 6. Phase Plan

### Phase A — Parse-Only (read path)

1. Read the `MatineeSequence` object property from a `BioConversation` export
2. Parse the sequence's `SequenceObjects` array
3. Identify all `BioSeqEvt_ConvNode` entries
4. For each ConvNode, follow output links to `SeqAct_Interp`
5. For each SeqAct_Interp, resolve the `"Data"` variable link to `InterpData`
6. Build the `nExportID → InterpDataIndex` mapping
7. Enrich parsed conversations with `interp_data_export_index`
8. Expose in `dialogue export --detailed`

### Phase B — Round-Trip Preservation

1. When editing a conversation, preserve all ConvNode → InterpData links
   for unchanged nodes
2. The property span scanner in `editor/preserve.go` must skip over Kismet
   data (it already does — it only touches `m_EntryList`, `m_ReplyList`, etc.)
3. If `MatineeSequence` is unchanged, the sequence is preserved as-is

### Phase C — Stub Generation for New Nodes

1. When a new entry/reply is added via `edit-conversation`, generate a
   minimal `BioSeqEvt_ConvNode` with the correct `m_nNodeID`
2. Generate a minimal `SeqAct_Interp` linked to the ConvNode
3. Generate a minimal `InterpData` with empty group/track arrays
4. Add all three to the PCC export table
5. Update the sequence's `SequenceObjects` array to include the new ConvNode
6. Update the SeqAct_Interp's VariableLinks to point to the new InterpData

### Phase D — Template-Based InterpData

1. Allow the user to specify a template node: "clone the cinematic from
   entry 16 for this new entry"
2. Copy InterpData from the template, adjust references
3. This gives the new line a proper camera angle and animation

## 7. Dependencies

| Dependency    | Status | Notes                                                                  |
| ------------- | ------ | ---------------------------------------------------------------------- |
| `me2pcc`      | ✅     | Already reads/writes PCC exports, needed for InterpData                |
| `me2tlk`      | ✅     | Not directly used; cinematic data is not in TLK                        |
| LEX reference | ✅     | `KismetHelper`, `SequenceObjectCreator`, `InterpData` binary converter |

## 8. Known Limitations

- **No camera/animation editing.** The MVP only preserves or stubs cinematic
  data. Editing camera tracks requires a curve editor UI.
- **No audio stream management.** WwiseStream references are preserved but
  not modified.
- **New ConvNodes require expanding the SequenceObjects array.** This changes
  the serial size of the sequence export, which the patcher must handle.
- **InterpData binary format** is complex (keyframe curves, group tracks).
  Parsing it fully is significant work; the MVP needs only enough to clone
  and stub.
- **ME2 OT only.** Kismet node structure differs in ME1, ME3, and LE.
- **Stage directions** (`m_aStageDirections`) are ME3-only and out of scope.

## 9. Verification

- Golden tests: parse a real BioConversation, verify the sequence mapping
  matches LegendaryExplorer output
- Round-trip test: edit a conversation (no Kismet changes), verify the
  sequence is byte-identical after writing
- Stub test: add a new dialogue node, verify the output PCC has a valid
  ConvNode → ActInterp → InterpData chain
- In-game smoke test: load a modified conversation in ME2, verify no crash

## 10. Relationship to Lia'Vael Beyond Citadel

For the Lia'Vael mod project, Kismet support at Phase C level means:

- New dialogue lines added to `cithub_rpplot_giver2_d_dlg` would play
  with static/default camera (no crash, but no fancy cinematic)
- Phase D (template-based cloning) would let new lines reuse existing
  camera work — e.g., clone entry 21 (Tali saying "Good luck, Lia'Vael")
  as the cinematic base for a new line
- Full cinematic editing is not needed for the mod's Phase 1 (Citadel
  Technical MVP); static camera is acceptable for initial testing
