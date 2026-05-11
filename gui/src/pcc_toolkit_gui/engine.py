"""Go core subprocess interface — shared with CLI."""

import json
import subprocess
import sys
from pathlib import Path
from typing import Any


CORE_BINARY = "pcc-core"
_EXE_SUFFIX = ".exe" if sys.platform == "win32" else ""


class EngineError(Exception):
    pass


def _resolve_binary() -> Path:
    binary = Path(CORE_BINARY + _EXE_SUFFIX)
    if binary.is_absolute():
        return binary
    core_dir = Path(__file__).resolve().parents[3] / "core"
    candidate = core_dir / (CORE_BINARY + _EXE_SUFFIX)
    if candidate.is_file():
        return candidate
    return binary


def _build_args(subcommand: str, **kwargs: Any) -> list[str]:
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
    return args


def _run(subcommand: str, **kwargs: Any) -> dict[str, Any]:
    args = _build_args(subcommand, **kwargs)
    proc = subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace")
    if proc.returncode != 0:
        raise EngineError(proc.stderr.strip() or proc.stdout.strip())
    return json.loads(proc.stdout)


def run_async(subcommand: str, **kwargs: Any) -> subprocess.Popen:
    """Launch the Go core as a cancellable subprocess. Returns a Popen handle."""
    args = _build_args(subcommand, **kwargs)
    return subprocess.Popen(
        args,
        stdout=subprocess.PIPE,
        stderr=subprocess.PIPE,
        text=True,
        encoding="utf-8",
        errors="replace",
    )


def version() -> dict[str, Any]:
    return _run("version")


def parse_pcc(file: Path | str, *, exports_only: bool = False, export_index: int | None = None,
              property_tags: bool = False, semantic_props: bool = False) -> dict[str, Any]:
    return _run("parse-pcc", file=str(file), exports_only=exports_only,
                export_index=export_index, property_tags=property_tags, semantic_props=semantic_props)


def parse_conversations(file: Path | str, *, conv_index: int | None = None,
                        resolve_tlk: str | None = None, dlc_dir: str | None = None,
                        mode: str = "resilient") -> dict[str, Any]:
    return _run("parse-conversations", file=str(file), conv_index=conv_index,
                resolve_tlk=resolve_tlk, dlc_dir=dlc_dir, mode=mode)


def layout_graph(file: Path | str, *, conv_index: int | None = None,
                 algorithm: str = "sugiyama", node_width: int = 240,
                 node_height: int = 64, x_spacing: int = 80,
                 y_spacing: int = 120) -> dict[str, Any]:
    return _run("layout-graph", file=str(file), conv_index=conv_index,
                algorithm=algorithm, node_width=node_width,
                node_height=node_height, x_spacing=x_spacing, y_spacing=y_spacing)


def parse_tlk(file: Path | str, *, search: str | None = None,
              strref: int | None = None, dump_all: bool = False) -> dict[str, Any]:
    return _run("parse-tlk", file=str(file), search=search, strref=strref, dump_all=dump_all)


def scan_evidence(query: str, *, tlk: Path | str, dlc_dir: str | None = None,
                  biogame_root: str | None = None, workers: int = 0) -> dict[str, Any]:
    return _run("scan-evidence", query=query, tlk=str(tlk), dlc_dir=dlc_dir,
                biogame_root=biogame_root, workers=workers)
