# Documentation

## Index

| File               | Purpose                                                                                                          |
| ------------------ | ---------------------------------------------------------------------------------------------------------------- |
| [PRD-INSPECT.md][] | Product requirements: PCC parsing, TLK reading, dialogue extraction, graph layout, evidence scanning, validation |
| [PRD-EDITING.md][] | Product requirements: PCC writing, TLK writing, conversation editing, round-trip fidelity, batch edit            |
| [CONTRACTS.md][]   | JSON contract reference for all `pcc-core` subcommands                                                           |
| [BUILDING.md][]    | Build, test, and release guide                                                                                   |
| [REFERENCE.md][]   | LegendaryExplorer reference notes, semantic parity items, and known divergences                                  |

## Reading Order

**New contributors** should read in this order:

1. [PRD-INSPECT.md][] — the inspect/extract foundation.
2. [PRD-EDITING.md][] — the editing and writing layer.
3. [CONTRACTS.md][] — what the tool produces.
4. [BUILDING.md][] — how to build and test.
5. [REFERENCE.md][] — how the project relates to LegendaryExplorer.

## Where to Start

- Adding a **new CLI command** → [CONTRACTS.md][] for the JSON shape, then [PRD-INSPECT.md][] or [PRD-EDITING.md][] for scope.
- Changing **parser behavior** → [PRD-INSPECT.md][] and [REFERENCE.md][] for LEX parity.
- Changing **editing or writing** → [PRD-EDITING.md][] and [REFERENCE.md][].
- **Build/release tooling** → [BUILDING.md][].
- **Understanding LEX semantics** → [REFERENCE.md][].

[PRD-INSPECT.md]: PRD-INSPECT.md
[PRD-EDITING.md]: PRD-EDITING.md
[CONTRACTS.md]: CONTRACTS.md
[BUILDING.md]: BUILDING.md
[REFERENCE.md]: REFERENCE.md
