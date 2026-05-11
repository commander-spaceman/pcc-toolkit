"""CLI entry point — Typer dispatch layer."""

import json
from pathlib import Path

import typer

from pcc_toolkit.engine import (
    EngineError,
    layout_graph as engine_layout_graph,
    parse_conversations as engine_parse_conversations,
    parse_pcc as engine_parse_pcc,
    parse_tlk as engine_parse_tlk,
    resolve_tlk as engine_resolve_tlk,
    scan_evidence as engine_scan_evidence,
    serialize as engine_serialize,
    validate as engine_validate,
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
evidence_app = typer.Typer(help="Search for evidence across game files")
app.add_typer(package_app, name="package")
app.add_typer(tlk_app, name="tlk")
app.add_typer(dialogue_app, name="dialogue")
app.add_typer(evidence_app, name="evidence")


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


@package_app.command("validate")
def package_validate(
    file: Path = typer.Argument(..., help="Path to PCC file"),
    strict: bool = typer.Option(False, "--strict", help="Fail on warnings"),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    result = engine_validate(file, strict=strict)
    if json_output:
        typer.echo(json.dumps(result, indent=2))
        return

    summary = result.get("report_summary", {})
    typer.echo(f"File: {file}")
    typer.echo(f"Total: {summary.get('total', 0)}  Valid: {summary.get('valid', 0)}  "
               f"Warning: {summary.get('warning', 0)}  Invalid: {summary.get('invalid', 0)}")
    typer.echo()

    for r in result.get("results", []):
        status_icon = {"valid": "[OK]", "warning": "[!!]", "invalid": "[XX]"}.get(r.get("status", ""), "[??]")
        typer.echo(f"  {status_icon} {r['conversation_id']} (#{r['export_index']}) — {r['status']}")
        for issue in r.get("issues", []):
            loc = f" [{issue.get('node_type', '')}#{issue.get('node_id', '')}]" if issue.get("node_id") is not None else ""
            typer.echo(f"      {issue['severity']}{loc}: {issue['message']}")
        typer.echo()


@package_app.command("extract")
def package_extract(
    file: Path = typer.Argument(..., help="Path to PCC file"),
    output: Path = typer.Option(None, "--output", help="Output JSON file"),
    tlk: Path = typer.Option(None, "--tlk", help="TLK file for text resolution"),
    dlc_dir: Path = typer.Option(None, "--dlc-dir", help="DLC directory for TLK overrides"),
    pretty: bool = typer.Option(False, "--pretty", help="Pretty-print JSON"),
) -> None:
    kwargs = {}
    if tlk:
        kwargs["resolve_tlk"] = str(tlk)
    if dlc_dir:
        kwargs["dlc_dir"] = str(dlc_dir)

    result = engine_serialize(file, **kwargs)
    data = json.dumps(result, indent=2, ensure_ascii=False) if pretty else json.dumps(result, ensure_ascii=False)

    if output:
        output.write_text(data, encoding="utf-8")
        typer.echo(f"Saved to {output}")
        return

    typer.echo(data)


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


@dialogue_app.command("graph")
def dialogue_graph(
    file: Path = typer.Argument(..., help="Path to PCC file"),
    conv_index: int = typer.Option(..., "--conv-index", help="Conversation export index"),
    algorithm: str = typer.Option("sugiyama", "--algorithm", help="Layout algorithm"),
    node_width: float = typer.Option(240, "--node-width", help="Node width in pixels"),
    node_height: float = typer.Option(64, "--node-height", help="Node height in pixels"),
    x_spacing: float = typer.Option(80, "--x-spacing", help="Horizontal spacing"),
    y_spacing: float = typer.Option(120, "--y-spacing", help="Vertical spacing"),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    result = engine_layout_graph(
        file,
        conv_index=conv_index,
        algorithm=algorithm,
        node_width=int(node_width),
        node_height=int(node_height),
        x_spacing=int(x_spacing),
        y_spacing=int(y_spacing),
    )

    if json_output:
        typer.echo(json.dumps(result, indent=2))
        return

    typer.echo(f"Conversation: {result.get('conversation_id', '?')}")
    typer.echo(f"Nodes: {result.get('node_count', 0)}")
    typer.echo(f"Edges: {len(result.get('edges', []))}")
    typer.echo()
    for key, pos in result.get("positions", {}).items():
        typer.echo(f"  {key}: ({pos[0]:.1f}, {pos[1]:.1f})")


if __name__ == "__main__":
    main()


# ── evidence ─────────────────────────────────────────────────────────────────

@evidence_app.command("scan")
def evidence_scan(
    query: str = typer.Argument(..., help="Text to search for in dialogue"),
    tlk: Path = typer.Option(..., "--tlk", help="Path to base TLK file"),
    dlc_dir: Path = typer.Option(None, "--dlc-dir", help="DLC directory for TLK overrides"),
    biogame_root: Path = typer.Option(None, "--biogame-root", help="BioGame root for PCC scanning"),
    output: Path = typer.Option(None, "--output", help="Output JSON file"),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    result = engine_scan_evidence(
        query=query,
        tlk=str(tlk),
        dlc_dir=str(dlc_dir) if dlc_dir else None,
        biogame_root=str(biogame_root) if biogame_root else None,
    )

    data = json.dumps(result, indent=2, ensure_ascii=False)

    if output:
        output.write_text(data, encoding="utf-8")
        typer.echo(f"Saved to {output}")
        return

    if json_output:
        typer.echo(data)
        return

    typer.echo(f"Query: {query}")
    typer.echo(f"TLK: {tlk}")
    typer.echo(f"Files scanned: {result.get('files_scanned', 0)}")
    typer.echo(f"Files with hits: {result.get('files_with_hits', 0)}")
    typer.echo(f"Total hits: {result.get('total_hits', 0)}")
    typer.echo()

    for ev in result.get("evidence", []):
        typer.echo(f"--- StrRef #{ev['strref']} ---")
        if ev.get("text"):
            typer.echo(f"  Text: {ev['text']}")

        tier1 = ev.get("bioconversation", [])
        if tier1:
            typer.echo(f"  Tier 1 — BioConversation ({len(tier1)}):")
            for hit in tier1:
                typer.echo(f"    {hit.get('conversation_id', '?')} in {hit.get('file_path', '?')}")

        tier2 = ev.get("semantic_container", [])
        if tier2:
            typer.echo(f"  Tier 2 — Semantic container ({len(tier2)}):")
            for hit in tier2:
                typer.echo(f"    {hit.get('export_name', '?')} [{hit.get('class_name', '?')}] in {hit.get('file_path', '?')}")

        tier3 = ev.get("container_fallback", [])
        if tier3:
            typer.echo(f"  Tier 3 — Container fallback ({len(tier3)}):")
            for hit in tier3[:5]:
                typer.echo(f"    {hit.get('export_name', '?')} [{hit.get('class_name', '?')}] in {hit.get('file_path', '?')}")
            if len(tier3) > 5:
                typer.echo(f"    ... and {len(tier3) - 5} more")
        typer.echo()
