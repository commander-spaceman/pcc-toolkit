"""Regression probe runner for Phase 10 QA.

Runs pcc-core against sample files, validates output shape,
and optionally regenerates golden files.

Usage:
    python tests/regression/run_probes.py --samples-dir samples/
    python tests/regression/run_probes.py --samples-dir samples/ --regenerate
"""

import argparse
import json
import os
import subprocess
import sys
from pathlib import Path
from typing import Any

REPO_ROOT = Path(__file__).resolve().parents[2]
GOLDEN_DIR = REPO_ROOT / "tests" / "golden"
SAMPLES_DIR = REPO_ROOT / "samples"

PROBES = [
    {
        "id": "pcc_header_BioD_CitHub_LOC_INT",
        "description": "PCC header + export table for CitHub dialogue file",
        "command": "parse-pcc",
        "args": {"file": "BioD_CitHub_LOC_INT.pcc", "exports_only": True},
        "golden": "pcc/BioD_CitHub_LOC_INT_exports.json",
        "checks": [
            {"path": "game_profile", "equals": "me2_ot"},
            {"path": "exports", "min_count": 100},
            {"path": "compressed", "equals": True},
        ],
    },
    {
        "id": "conv_BioD_CitHub_LOC_INT",
        "description": "All conversations in CitHub dialogue file",
        "command": "parse-conversations",
        "args": {"file": "BioD_CitHub_LOC_INT.pcc", "mode": "resilient"},
        "golden": "conversation/BioD_CitHub_LOC_INT.json",
        "checks": [
            {"path": "conversations", "min_count": 1},
            {"path": "game_profile", "equals": "me2_ot"},
        ],
    },
    {
        "id": "conv_nor_globalnews",
        "description": "Single conversation by index",
        "command": "parse-conversations",
        "args": {"file": "BioD_CitHub_LOC_INT.pcc", "conv_index": 3},
        "golden": "conversation/BioD_CitHub_LOC_INT_conv_3.json",
        "checks": [
            {"path": "conversations", "min_count": 1},
        ],
    },
    {
        "id": "graph_cithub_first_amb",
        "description": "Sugiyama layout for known conversation",
        "command": "layout-graph",
        "args": {"file": "BioD_CitHub_LOC_INT.pcc", "conv_index": 1},
        "golden": "graph/cithub_first_amb_sugiyama.json",
        "checks": [
            {"path": "conversation_id", "not_empty": True},
            {"path": "positions", "min_count": 2},
        ],
    },
    {
        "id": "tlk_info_BIOGame_INT",
        "description": "TLK header info validates entry count",
        "command": "parse-tlk",
        "args": {"file": "BIOGame_INT.tlk"},
        "golden": "tlk/BIOGame_INT_info.json",
        "checks": [
            {"path": "header.male_entry_count", "min": 30000},
        ],
    },
    {
        "id": "tlk_search_quarian",
        "description": "TLK search for quarian",
        "command": "parse-tlk",
        "args": {"file": "BIOGame_INT.tlk", "search": "quarian"},
        "golden": "tlk/search_quarian.json",
        "checks": [
            {"path": "results", "min_count": 1},
        ],
    },
    {
        "id": "tlk_resolve_strref",
        "description": "TLK resolves a specific StrRef",
        "command": "parse-tlk",
        "args": {"file": "BIOGame_INT.tlk", "strref": 125303},
        "golden": "tlk/resolve_125303.json",
        "checks": [
            {"path": "entries", "min_count": 1},
        ],
    },
    {
        "id": "validate_BioD_CitHub",
        "description": "Validate conversations in CitHub file",
        "command": "validate",
        "args": {"file": "BioD_CitHub_LOC_INT.pcc"},
        "golden": "conversation/BioD_CitHub_LOC_INT_validate.json",
        "checks": [
            {"path": "summary.total", "min": 1},
        ],
    },
    {
        "id": "serialize_BioD_CitHub",
        "description": "Full serialization of CitHub file",
        "command": "serialize",
        "args": {"file": "BioD_CitHub_LOC_INT.pcc"},
        "golden": "conversation/BioD_CitHub_LOC_INT_serialize.json",
        "checks": [
            {"path": "conversations", "min_count": 1},
            {"path": "compressed", "equals": True},
        ],
    },
]


def find_core_binary() -> Path:
    """Locate the pcc-core binary."""
    candidates = [
        REPO_ROOT / "core" / "pcc-core.exe",
        REPO_ROOT / "core" / "pcc-core",
    ]
    for c in candidates:
        if c.exists():
            return c
    for p in os.environ.get("PATH", "").split(os.pathsep):
        for name in ("pcc-core.exe", "pcc-core"):
            candidate = Path(p) / name
            if candidate.exists():
                return candidate
    raise FileNotFoundError(
        "pcc-core not found. Build with: cd core && go build ./cmd/pcc-core/"
    )


def run_core(subcommand: str, **kwargs) -> dict[str, Any]:
    """Run a pcc-core subcommand and return parsed JSON output."""
    binary = find_core_binary()
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
    proc = subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace")
    if proc.returncode != 0:
        stderr = (proc.stderr or "").strip()
        raise RuntimeError(
            f"pcc-core {subcommand} failed (exit {proc.returncode}):\n"
            f"stderr: {stderr[:500]}\n"
            f"args: {args}"
        )
    stdout = proc.stdout or ""
    if not stdout.strip():
        raise RuntimeError(
            f"pcc-core {subcommand} produced empty output\n"
            f"stderr: {(proc.stderr or '').strip()[:500]}\n"
            f"args: {args}"
        )
    try:
        return json.loads(stdout)
    except json.JSONDecodeError as e:
        raise RuntimeError(
            f"pcc-core {subcommand} returned invalid JSON:\n"
            f"stdout: {stdout[:500]}\n"
            f"error: {e}"
        )


def resolve_file_path(relative_path: str, samples_dir: Path) -> Path:
    """Resolve a file path relative to samples_dir or as absolute."""
    p = Path(relative_path)
    if p.is_absolute() and p.exists():
        return p
    return samples_dir / relative_path


def check_output(output: dict, checks: list[dict]) -> list[str]:
    """Run structural checks on JSON output. Returns list of failures."""
    failures = []

    def _get(obj, path):
        parts = path.split(".")
        for part in parts:
            if isinstance(obj, dict):
                obj = obj.get(part)
            elif isinstance(obj, list):
                try:
                    obj = obj[int(part)]
                except (IndexError, ValueError):
                    return None
            else:
                return None
        return obj

    for check in checks:
        path = check["path"]
        value = _get(output, path)

        if "equals" in check:
            if value != check["equals"]:
                failures.append(
                    f"  {path}: expected {check['equals']!r}, got {value!r}"
                )
        if "not_empty" in check:
            if check["not_empty"] and not value:
                failures.append(f"  {path}: expected non-empty, got {value!r}")
        if "min_count" in check:
            cnt = len(value) if isinstance(value, (list, dict)) else 0
            if cnt < check["min_count"]:
                failures.append(
                    f"  {path}: expected >= {check['min_count']} items, got {cnt}"
                )
        if "min" in check:
            if isinstance(value, (int, float)):
                if value < check["min"]:
                    failures.append(
                        f"  {path}: expected >= {check['min']}, got {value}"
                    )

    return failures


def run_probes(
    probes: list[dict],
    samples_dir: Path,
    regenerate: bool = False,
) -> tuple[int, int, int]:
    """Run all probes. Returns (passed, failed, skipped)."""
    passed = 0
    failed = 0
    skipped = 0

    for probe in probes:
        pid = probe["id"]
        desc = probe.get("description", pid)
        command = probe["command"]
        args_def = probe.get("args", {})
        golden_rel = probe.get("golden", "")
        checks = probe.get("checks", [])

        print(f"\n[{pid}] {desc}")

        resolved = {}
        skip = False
        for key, value in args_def.items():
            if isinstance(value, str):
                if key in ("file", "base", "tlk", "dir"):
                    fp = resolve_file_path(value, samples_dir)
                    if not fp.exists():
                        print(f"  SKIP: {key}={fp} not found")
                        skip = True
                        break
                    resolved[key] = str(fp)
                else:
                    resolved[key] = value
            else:
                resolved[key] = value

        if skip:
            skipped += 1
            continue

        try:
            output = run_core(command, **resolved)
        except RuntimeError as e:
            print(f"  FAIL: {e}")
            failed += 1
            continue

        failures = check_output(output, checks)
        if failures:
            print(f"  FAIL: structural checks failed")
            for f in failures:
                print(f)
            failed += 1
        else:
            print(f"  OK: {len(checks)} checks passed")
            passed += 1

        if regenerate and golden_rel:
            golden_path = GOLDEN_DIR / golden_rel
            golden_path.parent.mkdir(parents=True, exist_ok=True)
            golden_path.write_text(
                json.dumps(output, indent=2, ensure_ascii=False),
                encoding="utf-8",
            )
            print(f"  WROTE: {golden_path}")

    return passed, failed, skipped


def main():
    parser = argparse.ArgumentParser(description="Run regression probes")
    parser.add_argument(
        "--samples-dir",
        type=Path,
        default=SAMPLES_DIR,
        help=f"Directory with ME2 OT sample files (default: {SAMPLES_DIR})",
    )
    parser.add_argument(
        "--regenerate",
        action="store_true",
        help="Regenerate golden files from current output",
    )
    parser.add_argument(
        "--probe",
        type=str,
        action="append",
        help="Run only specific probe(s) by ID (repeatable)",
    )
    args = parser.parse_args()

    probes = PROBES
    if args.probe:
        probes = [p for p in PROBES if p["id"] in args.probe]
        if not probes:
            print(f"No probes match: {args.probe}")
            sys.exit(2)

    print(f"Running {len(probes)} probes against {args.samples_dir}")

    passed, failed, skipped = run_probes(probes, args.samples_dir, args.regenerate)
    print(f"\n{'='*50}")
    print(
        f"Results: {passed} passed, {failed} failed, {skipped} skipped"
    )

    if failed > 0:
        sys.exit(1)


if __name__ == "__main__":
    main()
