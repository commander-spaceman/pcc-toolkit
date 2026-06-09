# scripts/

Build, release, and modding utilities for pcc-toolkit.

## Build & Release

| Script        | Purpose                                                                                          |
| ------------- | ------------------------------------------------------------------------------------------------ |
| `build.ps1`   | Compile `pcc-core.exe` from `core/cmd/pcc-core/`. Pass `-Version` to inject ldflags.             |
| `release.ps1` | Package a portable release: builds the binary and creates a zip. Pass `-Version` and `-Archive`. |

## Dialogue Tools

| Script                  | Purpose                                                                                                                |
| ----------------------- | ---------------------------------------------------------------------------------------------------------------------- |
| `conv_to_screenplay.py` | Convert a conversation JSON to a readable screenplay format. Usage: `python conv_to_screenplay.py <conversation.json>` |

## Game Modding (ME2 OT)

| Script           | Purpose                                                                                 |
| ---------------- | --------------------------------------------------------------------------------------- |
| `backup-pcc.ps1` | Back up original game files to `~/Desktop/MassEffect2Backups`. Run once before modding. |
| `apply.ps1`      | Copy modified files from `output/` into the game's `CookedPC/` folder.                  |
| `restore.ps1`    | Restore original files from the backup. Run if the game breaks.                         |
