# LegendaryExplorer Reference Notes

## Purpose

This document records observed LegendaryExplorer behaviors, known divergences
between PCC Toolkit and LEX, and design decisions made during implementation.
It exists to answer the guiding question from the PRD:

> Does this match how LegendaryExplorer handles the same ME2 OT data?

## Reference Policy

LegendaryExplorer (`ME3Tweaks/LegendaryExplorer`, branch `Beta`) is the behavioral
reference for:

- PCC package semantics and table parsing.
- TLK parsing, Huffman decoding, and DLC resolution.
- `BioConversation` dialogue structure parsing.
- Graph node typing, edge semantics, and reply choice categories.
- Validation expectations for dialogue data.

**Rule:** Use LEX as a behavioral and semantic reference only. Do not copy, paste,
translate, or port LEX code into this repository. Implement behavior independently
and document observed public behavior, relevant file/class names, and verification
outcomes.

## Key LEX Source Files Consulted

| LEX File                                                     | What it covers                                 |
| ------------------------------------------------------------ | ---------------------------------------------- |
| `LegendaryExplorerCore/Packages/MEPackageHandler.cs`         | PCC reading, header, name/import/export tables |
| `LegendaryExplorerCore/Dialogue/ConversationExtended.cs`     | Conversation model and property extraction     |
| `LegendaryExplorerCore/Dialogue/DialogueNodeExtended.cs`     | Entry node property model                      |
| `LegendaryExplorerCore/Dialogue/ReplyChoiceNode.cs`          | Reply/reply-choice node model                  |
| `LegendaryExplorerCore/Dialogue/SpeakerExtended.cs`          | Speaker list parsing and synthetic speakers    |
| `LegendaryExplorer/Tools/Dialogue Editor/DialogueObjects.cs` | Dialogue editor domain types                   |
| `LegendaryExplorer/Tools/Dialogue Editor/ConvGraphEditor.cs` | Graph layout and rendering                     |
| `LegendaryExplorerCore/GameFilesystem/TlkSystem.cs`          | TLK loading and Huffman decoding               |
| `LegendaryExplorerCore/GameFilesystem/MountFile.cs`          | DLC mount priority and module TLK resolution   |
| `LegendaryExplorerCore/Unreal/ME3Enums.cs`                   | Unreal property type enums                     |

## Semantic Parity Items

### Speaker IDs

LEX synthesizes special speakers before parsed speakers:

- Player: ID `-2`, friendly name `Shepard`.
- Owner: ID `-1`, represents the conversation owner context.

PCC Toolkit follows this convention. Parsed speaker-list entries are preserved
by array index starting from ID `0`.

### Entry Nodes

- `nSpeakerIndex` maps to the speaker list where present.
- `m_StartingList` maps start-node order to entry indexes.
- Properties like `bSkippable`, `bNonTextLine`, `bAmbient`,
  `nCameraIntimacy`, and `eGuiStyle` are parsed where available.

### Reply Nodes and Reply Choices

LEX distinguishes between reply nodes (player choice containers) and reply choice
edges (entry-to-reply metadata with paraphrase, order, and category).

PCC Toolkit preserves this distinction:

- `ReplyNode` fields represent the reply itself.
- `ReplyChoice` fields represent the edge from an entry to a reply.

### Conditions and State Transitions

- `nConditionalFunc` and `nConditionalParam` define entry/reply conditions.
- `bFireConditional` controls whether the condition fires.
- `nStateTransition` and `nStateTransitionParam` define state transitions.
- `nExportID` provides dialogue-node identity for sequence correlation.

### TLK DLC Resolution

ME2 TLK loading uses:

1. Base `BIOGame_<language>.tlk`.
2. DLC TLKs ordered by `Mount.dlc` priority.
3. ME2 module TLK filenames resolved from DLC metadata (`BIOEngine.ini`
   entries) when available.
4. Language-specific files: `DLC_<module>_<language>.tlk`.

## Known Divergences

### Parse Modes

PCC Toolkit uses multiple parse modes (`struct_property_semantic`,
`row_payload`, `row_payload_struct_matrix`, `row_payload_struct_head`,
`count_or_value_fallback`) to handle conversation data that LEX may parse
through different code paths. The goal is resilient extraction, not
pixel-perfect AST equivalence.

### Graph Layout

PCC Toolkit uses a deterministic Sugiyama-style layered layout. LEX uses a
different layout approach tied to its WPF rendering pipeline. Node positions
will not match LEX pixel-for-pixel, but edge typing and node metadata should
be semantically equivalent.

### Compressed Package Handling

PCC Toolkit uses the `me2lzo` library for LZO, which implements ME2 OT's
specific LZO chunking scheme. LEX uses its own LZO implementation. Both
should produce equivalent decompressed data.

### Property Tag Parsing

PCC Toolkit parses Unreal property tags independently from `me2pcc`. LEX
parses them through `MEPackageHandler` and `PropertyCollection`. While the
set of supported property types and their binary layouts should match,
specific parser implementation details may differ.

### Round-Trip Fidelity

PCC Toolkit's editor uses property span scanning (`editor/preserve.go`) to
preserve unchanged bytes. LEX has its own preservation approach through its
property collection and serialization pipeline. Both aim to minimize
unintended changes, but the mechanisms differ.

## Verification Notes

- Golden files under `tests/golden/` encode expected outputs that have been
  manually compared against LEX for representative conversations.
- When LEX and PCC Toolkit diverge on a parse result, the decision defaults
  to LEX behavior unless there's a documented reason to differ.
- New features should be verified against LEX for semantic correctness before
  being considered complete.

## How We Use LegendaryExplorer

LegendaryExplorer is GPLv3-licensed. PCC Toolkit is MIT-licensed. To ensure we
use LEX as a reference without creating license obligations, we follow a strict
clean-room approach:

1. **Observe, don't copy.** We study LEX behavior by running it against real
   ME2 OT files, inspecting its public source files to understand data
   structures and algorithms, and consulting its public documentation and
   community resources.

2. **Document behavior, not code.** This file records what LEX does — which
   field maps to what, what values speakers get, how DLC TLKs are ordered.
   We never transcribe, translate, or port LEX source code into this repository.

3. **Implement independently.** All PCC Toolkit code is written from scratch
   based on documented behavior, file format specifications, and direct
   observation of ME2 OT binary files. No LEX source is used as a starting
   point for any implementation.

4. **Trace, don't copy.** We reference LEX file and class names (e.g.,
   `MEPackageHandler.cs`, `ConversationExtended.cs`) for traceability — so
   anyone can verify our behavioral claims by consulting the relevant LEX
   source. These references are documentation, not code ancestry.

This approach uses LegendaryExplorer the same way any developer uses
reference documentation: as a source of truth about how ME2 OT files
work, not as a source of implementation code.
