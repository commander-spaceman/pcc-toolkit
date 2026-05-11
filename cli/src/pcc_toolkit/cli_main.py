"""CLI entry point — Typer dispatch layer."""

import json
from pathlib import Path

import typer

from pcc_toolkit.engine import EngineError, parse_pcc as engine_parse_pcc, version as engine_version

app = typer.Typer(
    name="pcc-toolkit",
    help="ME2 OT Dialogue Extraction Toolkit",
    no_args_is_help=True,
)

package_app = typer.Typer(help="Inspect PCC packages")
app.add_typer(package_app, name="package")


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


def main() -> None:
    try:
        app()
    except EngineError as e:
        typer.echo(f"Error: {e}", err=True)
        raise typer.Exit(code=1)


if __name__ == "__main__":
    main()
