"""Go core subprocess interface for the CLI."""

import json
import subprocess
import sys
from pathlib import Path
from typing import Any


CORE_BINARY = "pcc-core"

_EXE_SUFFIX = ".exe" if sys.platform == "win32" else ""


class EngineError(Exception):
    """Raised when the Go core returns a non-zero exit code."""


def _resolve_binary() -> Path:
    binary = Path(CORE_BINARY + _EXE_SUFFIX)
    if binary.is_absolute():
        return binary
    resolved = _find_binary()
    if resolved:
        return resolved
    return binary


def _find_binary() -> Path | None:
    core_dir = Path(__file__).resolve().parents[2] / "build"
    candidate = core_dir / (CORE_BINARY + _EXE_SUFFIX)
    if candidate.is_file():
        return candidate
    return None


def _run(subcommand: str, **kwargs: Any) -> dict[str, Any]:
    binary = _resolve_binary()
    args = [str(binary), subcommand]
    for key, value in kwargs.items():
        flag = f"--{key.replace('_', '-')}"
        if isinstance(value, bool):
            if value:
                args.append(flag)
        elif isinstance(value, list):
            for v in value:
                args.extend([flag, str(v)])
        elif value is not None:
            args.extend([flag, str(value)])

    proc = subprocess.run(
        args,
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
    )
    if proc.returncode != 0:
        try:
            error_data = json.loads(proc.stderr)
            msg = error_data.get("error", proc.stderr.strip())
        except (json.JSONDecodeError, ValueError):
            msg = proc.stderr.strip() or proc.stdout.strip()
        raise EngineError(msg)
    return json.loads(proc.stdout)


def parse_pcc(
    file: Path | str,
    *,
    exports_only: bool = False,
    export_index: int | None = None,
    property_tags: bool = False,
    semantic_props: bool = False,
) -> dict[str, Any]:
    return _run(
        "parse-pcc",
        file=str(file),
        exports_only=exports_only,
        export_index=export_index,
        property_tags=property_tags,
        semantic_props=semantic_props,
    )


def parse_conversations(
    file: Path | str,
    *,
    conv_index: int | None = None,
    resolve_tlk: str | None = None,
    dlc_dir: str | None = None,
    language: str = "INT",
    mode: str = "resilient",
) -> dict[str, Any]:
    return _run(
        "parse-conversations",
        file=str(file),
        conv_index=conv_index,
        resolve_tlk=resolve_tlk,
        dlc_dir=dlc_dir,
        language=language,
        mode=mode,
    )


def layout_graph(
    file: Path | str,
    *,
    conv_index: int | None = None,
    algorithm: str = "sugiyama",
    node_width: int = 240,
    node_height: int = 64,
    x_spacing: int = 80,
    y_spacing: int = 120,
) -> dict[str, Any]:
    return _run(
        "layout-graph",
        file=str(file),
        conv_index=conv_index,
        algorithm=algorithm,
        node_width=node_width,
        node_height=node_height,
        x_spacing=x_spacing,
        y_spacing=y_spacing,
    )


def parse_tlk(
    file: Path | str,
    *,
    search: str | None = None,
    strref: int | None = None,
    dump_all: bool = False,
) -> dict[str, Any]:
    return _run(
        "parse-tlk",
        file=str(file),
        search=search,
        strref=strref,
        dump_all=dump_all,
    )


def resolve_tlk(
    base: Path | str,
    dlc_dir: Path | str,
    strrefs: list[int],
    *,
    language: str = "INT",
) -> dict[str, Any]:
    return _run(
        "resolve-tlk",
        base=str(base),
        dlc_dir=str(dlc_dir),
        strref=strrefs,
        language=language,
    )


def scan_evidence(
    query: str,
    *,
    tlk: Path | str,
    dlc_dir: str | None = None,
    language: str = "INT",
    biogame_root: str | None = None,
    cache: str | None = None,
    workers: int = 0,
) -> dict[str, Any]:
    return _run(
        "scan-evidence",
        query=query,
        tlk=str(tlk),
        dlc_dir=dlc_dir,
        language=language,
        biogame_root=biogame_root,
        cache=cache,
        workers=workers,
    )


def validate(file: Path | str, *, strict: bool = False) -> dict[str, Any]:
    return _run("validate", file=str(file), strict=strict)


def serialize(
    file: Path | str,
    *,
    game: str | None = None,
    resolve_tlk: str | None = None,
    dlc_dir: str | None = None,
    language: str = "INT",
    pretty: bool = False,
) -> dict[str, Any]:
    return _run(
        "serialize",
        file=str(file),
        game=game,
        resolve_tlk=resolve_tlk,
        dlc_dir=dlc_dir,
        language=language,
        pretty=pretty,
    )


def batch_validate(
    dir: Path | str,
    *,
    glob_pattern: str | None = None,
    strict: bool = False,
    output: str | None = None,
) -> dict[str, Any]:
    return _run(
        "batch-validate",
        dir=str(dir),
        glob=glob_pattern,
        strict=strict,
        output=output,
    )


def batch_extract(
    dir: Path | str,
    *,
    glob_pattern: str | None = None,
    output_dir: str | None = None,
    resolve_tlk: str | None = None,
    dlc_dir: str | None = None,
    language: str = "INT",
) -> dict[str, Any]:
    return _run(
        "batch-extract",
        dir=str(dir),
        glob=glob_pattern,
        output_dir=output_dir,
        resolve_tlk=resolve_tlk,
        dlc_dir=dlc_dir,
        language=language,
    )


def dump_lines(
    file: Path | str,
    *,
    resolve_tlk: str | None = None,
    dlc_dir: str | None = None,
    language: str = "INT",
    output_format: str = "json",
    pretty: bool = False,
) -> dict[str, Any]:
    return _run(
        "dump-lines",
        file=str(file),
        resolve_tlk=resolve_tlk,
        dlc_dir=dlc_dir,
        language=language,
        format=output_format,
        pretty=pretty,
    )


def scan_owners(file: Path | str) -> dict[str, Any]:
    return _run("scan-owners", file=str(file))


def edit_conversation(
    file: Path | str,
    *,
    conv_index: int,
    patch_file: Path | str,
    output: Path | str | None = None,
    dry_run: bool = False,
    tlk: Path | str | None = None,
    tlk_output: Path | str | None = None,
) -> dict[str, Any]:
    return _run(
        "edit-conversation",
        file=str(file),
        conv_index=conv_index,
        patch=str(patch_file),
        output=str(output) if output else None,
        dry_run=dry_run,
        tlk=str(tlk) if tlk else None,
        tlk_output=str(tlk_output) if tlk_output else None,
    )
