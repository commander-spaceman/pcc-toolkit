"""CLI entry point — Typer dispatch layer."""

import json
import subprocess
import sys
from pathlib import Path

import typer

from engine import (
    EngineError,
    batch_extract as engine_batch_extract,
    batch_validate as engine_batch_validate,
    dump_lines as engine_dump_lines,
    edit_conversation as engine_edit_conversation,
    layout_graph as engine_layout_graph,
    parse_conversations as engine_parse_conversations,
    parse_pcc as engine_parse_pcc,
    parse_tlk as engine_parse_tlk,
    resolve_tlk as engine_resolve_tlk,
    scan_evidence as engine_scan_evidence,
    scan_owners as engine_scan_owners,
    serialize as engine_serialize,
    validate as engine_validate,
    version as engine_version,
)
from format import (
    batch_summary,
    console,
    conversation_table,
    dump_lines_table,
    evidence_report,
    graph_summary,
    narrative_profiles_summary,
    owner_table,
    pcc_export_table,
    tlk_info_table,
    tlk_resolve_table,
    tlk_search_table,
    validation_summary,
)

REPO_ROOT = Path(__file__).resolve().parents[2]
BUILD_DIR = REPO_ROOT / "build"
CORE_MAIN = REPO_ROOT / "core" / "cmd" / "pcc-core"
CORE_MODULE = REPO_ROOT / "core"
_EXE_SUFFIX = ".exe" if sys.platform == "win32" else ""

app = typer.Typer(
    name="pcc-toolkit",
    help="ME2 OT Dialogue Extraction Toolkit",
    no_args_is_help=True,
)


package_app = typer.Typer(help="Inspect PCC packages")
tlk_app = typer.Typer(help="Work with TLK talk files")
dialogue_app = typer.Typer(help="Extract dialogue from BioConversation")
evidence_app = typer.Typer(help="Search for evidence across game files")
batch_app = typer.Typer(help="Batch operations across multiple files")
app.add_typer(package_app, name="package")
app.add_typer(tlk_app, name="tlk")
app.add_typer(dialogue_app, name="dialogue")
app.add_typer(evidence_app, name="evidence")
app.add_typer(batch_app, name="batch")

dev_app = typer.Typer(help="Build, test, and maintenance commands")
app.add_typer(dev_app, name="dev")


@dev_app.command("build-core")
def dev_build_core() -> None:
    """Build the Go core binary into build/."""
    BUILD_DIR.mkdir(parents=True, exist_ok=True)
    target = BUILD_DIR / f"pcc-core{_EXE_SUFFIX}"
    cmd = ["go", "build", "-o", str(target)]
    typer.echo(f"Building pcc-core -> {target}")
    try:
        subprocess.run(cmd, cwd=str(CORE_MAIN), check=True)
    except subprocess.CalledProcessError as e:
        typer.echo(f"Build failed: {e}", err=True)
        raise typer.Exit(code=1)
    typer.echo(f"Built: {target} ({target.stat().st_size} bytes)")


@dev_app.command("test-core")
def dev_test_core() -> None:
    """Run Go core tests."""
    typer.echo("Running go test ./...")
    try:
        subprocess.run(["go", "test", "./..."], cwd=str(CORE_MODULE), check=True)
    except subprocess.CalledProcessError as e:
        typer.echo(f"Tests failed: {e}", err=True)
        raise typer.Exit(code=1)


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
    property_tags: bool = typer.Option(
        False, "--property-tags", help="Include raw property tag data"
    ),
    semantic_props: bool = typer.Option(
        False, "--semantic-props", help="Include semantic property collections"
    ),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    result = engine_parse_pcc(
        file,
        exports_only=True,
        property_tags=property_tags,
        semantic_props=semantic_props,
    )
    exports = result.get("exports", [])
    profile = result.get("game_profile", "unknown")
    compressed = result.get("compressed", False)

    if class_:
        exports = [e for e in exports if e.get("class_name") == class_]

    if json_output:
        typer.echo(json.dumps(result, indent=2))
        return

    console.print(f"File: [bold]{file}[/bold]")
    console.print(f"Profile: {profile}  Compressed: {compressed}")
    console.print(pcc_export_table(exports))


@package_app.command("inspect")
def package_inspect(
    file: Path = typer.Argument(..., help="Path to PCC file"),
    index: int = typer.Argument(..., help="Export index to inspect"),
    property_tags: bool = typer.Option(
        False, "--property-tags", help="Include raw property tag data"
    ),
    semantic_props: bool = typer.Option(
        False, "--semantic-props", help="Include semantic property collections"
    ),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    result = engine_parse_pcc(
        file,
        export_index=index,
        property_tags=property_tags,
        semantic_props=semantic_props,
    )
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

    console.print(f"File: [bold]{file}[/bold]")
    validation_summary(result)


@package_app.command("extract")
def package_extract(
    file: Path = typer.Argument(..., help="Path to PCC file"),
    output: Path = typer.Option(None, "--output", help="Output JSON file"),
    tlk: Path = typer.Option(None, "--tlk", help="TLK file for text resolution"),
    dlc_dir: Path = typer.Option(
        None, "--dlc-dir", help="DLC directory for TLK overrides"
    ),
    pretty: bool = typer.Option(False, "--pretty", help="Pretty-print JSON"),
) -> None:
    kwargs = {}
    if tlk:
        kwargs["resolve_tlk"] = str(tlk)
    if dlc_dir:
        kwargs["dlc_dir"] = str(dlc_dir)

    result = engine_serialize(file, **kwargs)
    data = (
        json.dumps(result, indent=2, ensure_ascii=False)
        if pretty
        else json.dumps(result, ensure_ascii=False)
    )

    if output:
        output.write_text(data, encoding="utf-8")
        typer.echo(f"Saved to {output}")
        return

    typer.echo(data)


@package_app.command("scan-owners")
def package_scan_owners(
    file: Path = typer.Argument(..., help="Path to PCC file"),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    result = engine_scan_owners(file)
    if json_output:
        typer.echo(json.dumps(result, indent=2))
        return

    owners = result.get("owners", {})
    console.print(f"File: [bold]{file}[/bold]")
    if owners:
        console.print(owner_table(owners))
    else:
        console.print("  No conversation owners found.")


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

    console.print(f"File: [bold]{file}[/bold]")
    console.print(tlk_info_table(result))


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

    console.print(f"Search: [bold]'{query}'[/bold] in [dim]{file}[/dim]")
    console.print(f"Results: {len(results)}")
    if results:
        console.print(tlk_search_table(results))


@tlk_app.command("resolve")
def tlk_resolve(
    strref: int = typer.Argument(..., help="StringRef ID to resolve"),
    file: Path = typer.Option(..., "--file", help="Path to TLK file"),
    dlc_dir: Path = typer.Option(None, "--dlc-dir", help="DLC directory for overrides"),
    language: str = typer.Option(
        "INT", "--language", help="TLK language code (INT, DEU, FRA, etc.)"
    ),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    if dlc_dir:
        result = engine_resolve_tlk(
            base=file, dlc_dir=dlc_dir, strrefs=[strref], language=language
        )
    else:
        result = engine_parse_tlk(file, strref=strref)

    if json_output:
        typer.echo(json.dumps(result, indent=2))
        return

    entries = result.get("entries", [])
    results = result.get("results", [])
    items = entries + results
    if not items:
        console.print(f"StringRef [bold]#{strref}[/bold] not found")
        return
    console.print(tlk_resolve_table(items))


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

    console.print(f"File: [bold]{file}[/bold]")
    console.print(f"Profile: {result.get('game_profile', 'unknown')}")
    console.print(conversation_table(conversations))
    for c in conversations:
        for w in c.get("warnings", []):
            console.print(f"  [yellow]⚠ {w}[/yellow]")


@dialogue_app.command("export")
def dialogue_export(
    file: Path = typer.Argument(..., help="Path to PCC file"),
    output: Path = typer.Option(None, "--output", help="Output JSON file"),
    tlk: Path = typer.Option(None, "--tlk", help="TLK file for text resolution"),
    dlc_dir: Path = typer.Option(
        None, "--dlc-dir", help="DLC directory for TLK overrides"
    ),
    language: str = typer.Option("INT", "--language", help="TLK language code"),
    conv_index: int = typer.Option(
        None, "--conv-index", help="Export a single conversation"
    ),
    pretty: bool = typer.Option(False, "--pretty", help="Pretty-print JSON"),
) -> None:
    kwargs = {}
    if tlk:
        kwargs["resolve_tlk"] = str(tlk)
    if dlc_dir:
        kwargs["dlc_dir"] = str(dlc_dir)
    kwargs["language"] = language
    if conv_index is not None:
        kwargs["conv_index"] = conv_index

    result = engine_parse_conversations(file, **kwargs)

    data = (
        json.dumps(result, indent=2, ensure_ascii=False)
        if pretty
        else json.dumps(result, ensure_ascii=False)
    )

    if output:
        output.write_text(data, encoding="utf-8")
        typer.echo(f"Saved to {output}")
    else:
        typer.echo(data)


@dialogue_app.command("graph")
def dialogue_graph(
    file: Path = typer.Argument(..., help="Path to PCC file"),
    conv_index: int = typer.Option(
        ..., "--conv-index", help="Conversation export index"
    ),
    algorithm: str = typer.Option("sugiyama", "--algorithm", help="Layout algorithm"),
    node_width: float = typer.Option(240, "--node-width", help="Node width in pixels"),
    node_height: float = typer.Option(
        64, "--node-height", help="Node height in pixels"
    ),
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

    console.print(f"Conversation: [bold]{result.get('conversation_id', '?')}[/bold]")
    console.print(
        f"Nodes: {result.get('node_count', 0)}  Edges: {len(result.get('edges', []))}"
    )
    console.print(graph_summary(result))


@dialogue_app.command("dump-lines")
def dialogue_dump_lines(
    file: Path = typer.Argument(..., help="Path to PCC file"),
    tlk: Path = typer.Option(None, "--tlk", help="TLK file for text resolution"),
    dlc_dir: Path = typer.Option(
        None, "--dlc-dir", help="DLC directory for TLK overrides"
    ),
    language: str = typer.Option("INT", "--language", help="TLK language code"),
    output_format: str = typer.Option(
        "json", "--format", help="Output format: json or csv"
    ),
    pretty: bool = typer.Option(False, "--pretty", help="Pretty-print JSON"),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    result = engine_dump_lines(
        file,
        resolve_tlk=str(tlk) if tlk else None,
        dlc_dir=str(dlc_dir) if dlc_dir else None,
        language=language,
        output_format=output_format,
        pretty=pretty,
    )

    lines = result.get("lines", [])
    if output_format == "csv":
        typer.echo(
            "conversation_id,export_index,node_type,node_id,"
            "speaker_tag,strref,line_text,line_status,file"
        )
        for line in lines:
            text = (line.get("line_text") or "").replace('"', '""')
            typer.echo(
                f"{line.get('conversation_id', '')},"
                f"{line.get('export_index', '')},"
                f"{line.get('node_type', '')},"
                f"{line.get('node_id', '')},"
                f"{line.get('speaker_tag', '')},"
                f"{line.get('strref', '')},"
                f'"{text}",'
                f"{line.get('line_status', '')},"
                f"{line.get('file', '')}"
            )
        return

    if json_output:
        typer.echo(json.dumps(result, indent=2))
        return

    console.print(f"File: [bold]{file}[/bold]")
    console.print(f"Lines: {len(lines)}")
    if lines:
        console.print(dump_lines_table(lines))


@dialogue_app.command("edit")
def dialogue_edit(
    file: Path = typer.Argument(..., help="Path to PCC file"),
    conv_index: int = typer.Option(
        ..., "--conv-index", help="Export index of the conversation to edit"
    ),
    patch_file: Path = typer.Option(..., "--patch", help="Path to JSON patch file"),
    output: Path = typer.Option(None, "--output", help="Path for the output PCC file"),
    dry_run: bool = typer.Option(
        False, "--dry-run", help="Validate without writing output"
    ),
) -> None:
    result = engine_edit_conversation(
        file,
        conv_index=conv_index,
        patch_file=patch_file,
        output=output,
        dry_run=dry_run,
    )
    status = result.get("status", "unknown")
    validation = result.get("validation")
    out = result.get("output", str(output) if output else "(dry-run)")
    console.print(f"[green]{status}[/green] -> {out}")
    if validation:
        total = validation.get("total", 0)
        valid = validation.get("valid", 0)
        warning = validation.get("warning", 0)
        invalid = validation.get("invalid", 0)
        console.print(
            f"  validation: [green]{valid} valid[/green]"
            f" / [yellow]{warning} warnings[/yellow]"
            f" / [red]{invalid} invalid[/red]"
            f" ({total} total)"
        )


# ── batch ────────────────────────────────────────────────────────────────────


@batch_app.command("validate")
def batch_validate(
    dir: Path = typer.Argument(..., help="Directory to scan for PCC files"),
    glob_pattern: str = typer.Option("*.pcc", "--glob", help="Glob pattern"),
    strict: bool = typer.Option(False, "--strict", help="Fail on warnings"),
    output: Path = typer.Option(None, "--output", help="Output JSON file"),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    result = engine_batch_validate(
        dir,
        glob_pattern=glob_pattern,
        strict=strict,
        output=str(output) if output else None,
    )
    if json_output:
        typer.echo(json.dumps(result, indent=2))
        return

    console.print(f"Directory: [bold]{dir}[/bold]")
    console.print(f"Pattern: {glob_pattern}")
    batch_summary(result)


@batch_app.command("extract")
def batch_extract(
    dir: Path = typer.Argument(..., help="Directory to scan for PCC files"),
    glob_pattern: str = typer.Option("*.pcc", "--glob", help="Glob pattern"),
    output_dir: Path = typer.Option(None, "--output-dir", help="Output directory"),
    tlk: Path = typer.Option(None, "--tlk", help="TLK file for text resolution"),
    dlc_dir: Path = typer.Option(
        None, "--dlc-dir", help="DLC directory for TLK overrides"
    ),
    language: str = typer.Option("INT", "--language", help="TLK language code"),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    result = engine_batch_extract(
        dir,
        glob_pattern=glob_pattern,
        output_dir=str(output_dir) if output_dir else None,
        resolve_tlk=str(tlk) if tlk else None,
        dlc_dir=str(dlc_dir) if dlc_dir else None,
        language=language,
    )
    if json_output:
        typer.echo(json.dumps(result, indent=2))
        return

    typer.echo(f"Dir: {dir}  Pattern: {glob_pattern}")
    typer.echo(f"Files found: {result.get('files_found', 0)}")
    typer.echo(
        f"Files OK: {result.get('files_ok', 0)}  Errors: {result.get('files_error', 0)}"
    )
    for r in result.get("results", []):
        if r.get("error"):
            typer.echo(f"  ERR: {r['file']} — {r['error']}")
        else:
            typer.echo(f"  {r.get('conversations', 0):>3} conv  {r.get('file', '')}")


# ── evidence ─────────────────────────────────────────────────────────────────


@evidence_app.command("scan")
def evidence_scan(
    query: str = typer.Argument(..., help="Text to search for in dialogue"),
    tlk: Path = typer.Option(..., "--tlk", help="Path to base TLK file"),
    dlc_dir: Path = typer.Option(
        None, "--dlc-dir", help="DLC directory for TLK overrides"
    ),
    language: str = typer.Option("INT", "--language", help="TLK language code"),
    biogame_root: Path = typer.Option(
        None, "--biogame-root", help="BioGame root for PCC scanning"
    ),
    cache: Path = typer.Option(
        None, "--cache", help="File cache JSON for scan persistence"
    ),
    workers: int = typer.Option(
        0, "--workers", help="Number of concurrent workers (0=CPU count)"
    ),
    output: Path = typer.Option(None, "--output", help="Output JSON file"),
    json_output: bool = typer.Option(False, "--json", help="Output as JSON"),
) -> None:
    result = engine_scan_evidence(
        query=query,
        tlk=str(tlk),
        dlc_dir=str(dlc_dir) if dlc_dir else None,
        language=language,
        biogame_root=str(biogame_root) if biogame_root else None,
        cache=str(cache) if cache else None,
        workers=workers,
    )

    data = json.dumps(result, indent=2, ensure_ascii=False)

    if output:
        output.write_text(data, encoding="utf-8")
        typer.echo(f"Saved to {output}")
        return

    if json_output:
        typer.echo(data)
        return

    evidence_report(result)
    narrative_profiles_summary(result.get("narrative_profiles", []))


if __name__ == "__main__":
    main()
