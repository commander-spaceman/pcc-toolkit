"""Smoke tests for the pcc-toolkit CLI entry point."""

import subprocess
import sys
from pathlib import Path

import pytest

REPO_ROOT = Path(__file__).resolve().parents[1]
_EXE_SUFFIX = ".exe" if sys.platform == "win32" else ""
ENTRY_POINT = REPO_ROOT / ".venv" / "Scripts" / f"pcc-toolkit{_EXE_SUFFIX}"


def _entry_point_exists() -> bool:
    return ENTRY_POINT.is_file()


requires_entry_point = pytest.mark.skipif(
    not _entry_point_exists(),
    reason="pcc-toolkit entry point not found in .venv",
)


@requires_entry_point
def test_help_succeeds() -> None:
    """pcc-toolkit --help exits 0 and prints usage."""
    proc = subprocess.run(
        [str(ENTRY_POINT), "--help"],
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
    )
    assert proc.returncode == 0, f"stderr:\n{proc.stderr}"
    assert "Usage:" in proc.stdout or "Usage:" in proc.stderr


@requires_entry_point
def test_dev_subcommand_visible() -> None:
    """pcc-toolkit --help shows dev subcommand group."""
    proc = subprocess.run(
        [str(ENTRY_POINT), "--help"],
        capture_output=True,
        text=True,
        encoding="utf-8",
        errors="replace",
    )
    assert proc.returncode == 0, f"stderr:\n{proc.stderr}"
    output = proc.stdout + proc.stderr
    assert "dev" in output, f"dev subcommand not found in:\n{output}"
