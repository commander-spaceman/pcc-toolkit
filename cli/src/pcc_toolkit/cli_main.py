"""CLI entry point — Typer dispatch layer."""

import json
from pathlib import Path

import typer

from pcc_toolkit.engine import (
    EngineError,
    parse_conversations as engine_parse_conversations,
    parse_pcc as engine_parse_pcc,
    parse_tlk as engine_parse_tlk,
    resolve_tlk as engine_resolve_tlk,
    version as engine_version,
)

app = typer.Typer(
    name="pcc-toolkit",
    help="ME2 OT Dialogue Extraction Toolkit",
    no_args_is_help=True,
)

package_app = typer.Typer(help="Inspect PCC packages")
tlk_app = typer.Typer(help="Work with TLK talk files")
dialogue_app = typer.Typer(help="Extract dialogue from BioConversation")
app.add_typer(package_app, name="package")
app.add_typer(tlk_app, name="tlk")
app.add_typer(dialogue_app, name="dialogue")


@app.callback(invoke_without_command=True)
def callback(
    _version: bool = typer.Option(
        False,
        "--version",
        help="Show version and capabilities",
        is_eager=True,
    ),
) -> None:
    if _version:
        result = engine_version()
        typer.echo(json.dumps(result, indent=2))


# ── package ────────────────────────────────────────────────────────────────

@package_app.command("list")
def package_list(
    file: Path = typer.Argument(..., help="Path to PCC file"),
    class_: str = typer.Option(None, "--class", help="Filter by class name"),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    result = engine_parse_pcc(file, exports_only=True)
    exports = result.get("exports", [])
    profile = result.get("game_profile", "unknown")
    compressed = result.get("compressed", False)

    if class_:
        exports = [e for e in exports if e.get("class_name") == class_]

    if json_output:
        typer.echo(json.dumps(result, indent=2))
        return

    typer.echo(f"File: {file}")
    typer.echo(f"Profile: {profile}  Compressed: {compressed}")
    typer.echo(f"Exports: {len(exports)}")
    typer.echo()
    typer.echo(f"{'Index':>6}  {'Class':<30}  {'Object':<40}  {'Size':>8}  {'Offset':>10}")
    typer.echo("-" * 100)
    for e in exports:
        typer.echo(
            f"{e['index']:>6}  {e.get('class_name', ''):<30}  "
            f"{e.get('object_name', ''):<40}  {e['serial_size']:>8}  {e['serial_offset']:>10}"
        )


@package_app.command("inspect")
def package_inspect(
    file: Path = typer.Argument(..., help="Path to PCC file"),
    index: int = typer.Argument(..., help="Export index to inspect"),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    result = engine_parse_pcc(file, export_index=index)
    if json_output:
        typer.echo(json.dumps(result, indent=2))
        return

    exp = result
    typer.echo(f"Export #{exp['index']}")
    typer.echo(f"  Class:  {exp.get('class_name', 'N/A')}")
    typer.echo(f"  Object: {exp.get('object_name', 'N/A')}")
    typer.echo(f"  Serial: offset={exp['serial_offset']}, size={exp['serial_size']}")
    if exp.get("serial_data"):
        typer.echo(f"  Data:   {len(exp['serial_data'])} chars (base64)")


# ── tlk ────────────────────────────────────────────────────────────────────

@tlk_app.command("info")
def tlk_info(
    file: Path = typer.Argument(..., help="Path to TLK file"),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    result = engine_parse_tlk(file)
    if json_output:
        typer.echo(json.dumps(result, indent=2))
        return

    h = result.get("header", {})
    typer.echo(f"File: {file}")
    typer.echo(f"  Magic:       0x{h.get('magic', 0):08X}")
    typer.echo(f"  Version:     {h.get('version', '?')} (min {h.get('min_version', '?')})")
    typer.echo(f"  Male entries:   {h.get('male_entry_count', 0)}")
    typer.echo(f"  Female entries: {h.get('female_entry_count', 0)}")
    typer.echo(f"  Huffman nodes:  {h.get('tree_node_count', 0)}")
    typer.echo(f"  Bitstream len:  {h.get('data_len', 0)} bytes")
    typer.echo(f"  Total entries:  {result.get('total_entries', 0)}")


@tlk_app.command("search")
def tlk_search(
    query: str = typer.Argument(..., help="Text to search for"),
    file: Path = typer.Option(..., "--file", help="Path to TLK file"),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    result = engine_parse_tlk(file, search=query)
    results = result.get("results", [])

    if json_output:
        typer.echo(json.dumps(result, indent=2))
        return

    typer.echo(f"Searching for '{query}' in {file}")
    typer.echo(f"Results: {len(results)}")
    typer.echo()
    for r in results:
        typer.echo(f"  #{r['string_id']}: {r['text']}")


@tlk_app.command("resolve")
def tlk_resolve(
    strref: int = typer.Argument(..., help="StringRef ID to resolve"),
    file: Path = typer.Option(..., "--file", help="Path to TLK file"),
    dlc_dir: Path = typer.Option(None, "--dlc-dir", help="DLC directory for overrides"),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    if dlc_dir:
        result = engine_resolve_tlk(base=file, dlc_dir=dlc_dir, strrefs=[strref])
    else:
        result = engine_parse_tlk(file, strref=strref)

    if json_output:
        typer.echo(json.dumps(result, indent=2))
        return

    entries = result.get("entries", [])
    results = result.get("results", [])
    items = entries + results
    if not items:
        typer.echo(f"StringRef #{strref} not found")
        return
    for item in items:
        source = item.get("source_tlk", "")
        label = f" (source: {source})" if source else ""
        typer.echo(f"#{item['string_id']}: {item['text']}{label}")


@tlk_app.command("dump")
def tlk_dump(
    file: Path = typer.Argument(..., help="Path to TLK file"),
    output: Path = typer.Option(None, "--output", help="Output JSON file"),
) -> None:
    result = engine_parse_tlk(file, dump_all=True)
    data = json.dumps(result, indent=2, ensure_ascii=False)

    if output:
        output.write_text(data, encoding="utf-8")
        typer.echo(f"Written to {output}")
    else:
        typer.echo(data)


def main() -> None:
    try:
        app()
    except EngineError as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


# ── dialogue ────────────────────────────────────────────────────────────────

@dialogue_app.command("list")
def dialogue_list(
    file: Path = typer.Argument(..., help="Path to PCC file"),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    result = engine_parse_conversations(file)
    conversations = result.get("conversations", [])
    errors = result.get("errors", [])

    if json_output:
        typer.echo(json.dumps(result, indent=2))
        return

    typer.echo(f"File: {file}")
    typer.echo(f"Profile: {result.get('game_profile', 'unknown')}")
    typer.echo(f"Conversations: {len(conversations)}")
    if errors:
        typer.echo(f"Errors: {len(errors)}")
    typer.echo()
    typer.echo(f"{'Index':>6}  {'ID':<50}  {'Mode':<30}  {'Entries':>7}  {'Replies':>7}")
    typer.echo("-" * 110)
    for c in conversations:
        typer.echo(
            f"{c['export_index']:>6}  {c['id']:<50}  {c['parse_mode']:<30}  "
            f"{len(c['entries']):>7}  {len(c['replies']):>7}"
        )
        for w in c.get("warnings", []):
            typer.echo(f"        ⚠ {w}")


@dialogue_app.command("export")
def dialogue_export(
    file: Path = typer.Argument(..., help="Path to PCC file"),
    output: Path = typer.Option(None, "--output", help="Output JSON file"),
    tlk: Path = typer.Option(None, "--tlk", help="TLK file for text resolution"),
    dlc_dir: Path = typer.Option(None, "--dlc-dir", help="DLC directory for TLK overrides"),
    conv_index: int = typer.Option(None, "--conv-index", help="Export a single conversation"),
    pretty: bool = typer.Option(False, "--pretty", help="Pretty-print JSON"),
) -> None:
    kwargs = {}
    if tlk:
        kwargs["resolve_tlk"] = str(tlk)
    if dlc_dir:
        kwargs["dlc_dir"] = str(dlc_dir)
    if conv_index is not None:
        kwargs["conv_index"] = conv_index

    result = engine_parse_conversations(file, **kwargs)

    data = json.dumps(result, indent=2, ensure_ascii=False) if pretty else json.dumps(result, ensure_ascii=False)

    if output:
        output.write_text(data, encoding="utf-8")
        typer.echo(f"Saved to {output}")
    else:
        typer.echo(data)


if __name__ == "__main__":
    main()
