"""Golden file regression tests.

Verifies pcc-core output matches known-good golden files.
Golden files are committed to tests/golden/ and should not be edited manually.
Regenerate with: python tests/regression/run_probes.py --regenerate
"""

import json
import subprocess
from pathlib import Path

import pytest

REPO_ROOT = Path(__file__).resolve().parents[1]
GOLDEN_DIR = REPO_ROOT / "tests" / "golden"
SAMPLES_DIR = REPO_ROOT / "samples"

CORE_BINARY = (
    REPO_ROOT / "core" / "pcc-core.exe"
    if Path(REPO_ROOT / "core" / "pcc-core.exe").exists()
    else REPO_ROOT / "core" / "pcc-core"
)


def run_core(subcommand: str, **kwargs) -> dict:
    args = [str(CORE_BINARY), subcommand]
    for key, value in kwargs.items():
        flag = f"--{key.replace('_', '-')}"
        if isinstance(value, bool):
            if value:
                args.append(flag)
        elif value is not None:
            args.extend([flag, str(value)])
    proc = subprocess.run(args, capture_output=True, text=True)
    if proc.returncode != 0:
        raise RuntimeError(f"pcc-core error: {proc.stderr.strip()}")
    return json.loads(proc.stdout)


def _load_golden(rel_path: str) -> dict | None:
    p = GOLDEN_DIR / rel_path
    if not p.exists():
        return None
    return json.loads(p.read_text(encoding="utf-8"))


def _compare_keys(actual: dict, expected: dict, prefix: str = "") -> list[str]:
    """Compare structural keys and stable values between actual and expected."""
    issues = []
    for key in expected:
        path = f"{prefix}.{key}" if prefix else key
        if key not in actual:
            issues.append(f"missing key: {path}")
            continue
        exp_val = expected[key]
        act_val = actual[key]
        if isinstance(exp_val, dict) and isinstance(act_val, dict):
            issues.extend(_compare_keys(act_val, exp_val, path))
        elif isinstance(exp_val, (int, float, str, bool, type(None))):
            if exp_val != act_val:
                issues.append(
                    f"value mismatch at {path}: expected {exp_val!r}, got {act_val!r}"
                )
    return issues


def _strip_volatile_fields(obj: dict) -> dict:
    """Remove fields that vary between runs (timestamps, absolute paths)."""
    # Deep copy simple approach
    import copy

    cleaned = copy.deepcopy(obj)
    # Remove file path fields that are absolute
    for key in ("file", "base", "path"):
        if key in cleaned:
            cleaned[key] = "<stripped>"
    return cleaned


class TestGoldenFiles:
    """Verify pcc-core output against committed golden files."""

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found. Build with: cd core && go build ./cmd/pcc-core/",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="samples/ directory empty. Place ME2 OT files in samples/ to run golden tests.",
    )
    def test_conversation_BioD_CitHub_LOC_INT(self):
        """Verify parse-conversations output matches golden."""
        pcc_file = SAMPLES_DIR / "BioD_CitHub_LOC_INT.pcc"
        if not pcc_file.exists():
            pytest.skip(f"{pcc_file} not found")

        actual = run_core("parse-conversations", file=str(pcc_file), mode="resilient")
        golden = _load_golden("conversation/BioD_CitHub_LOC_INT.json")
        if golden is None:
            pytest.skip("golden file not found; generate with --regenerate")

        actual = _strip_volatile_fields(actual)
        golden = _strip_volatile_fields(golden)

        # Structural comparison
        assert actual["game_profile"] == golden["game_profile"]
        assert len(actual["conversations"]) == len(golden["conversations"])

        for i, (act_conv, exp_conv) in enumerate(
            zip(actual["conversations"], golden["conversations"])
        ):
            assert act_conv["id"] == exp_conv["id"], f"conv[{i}] id mismatch"
            assert act_conv["parse_mode"] == exp_conv["parse_mode"], (
                f"conv[{i}] parse_mode mismatch: "
                f"{act_conv['parse_mode']} vs {exp_conv['parse_mode']}"
            )
            assert len(act_conv.get("entries", [])) == len(
                exp_conv.get("entries", [])
            ), f"conv[{i}] entry count mismatch"
            assert len(act_conv.get("replies", [])) == len(
                exp_conv.get("replies", [])
            ), f"conv[{i}] reply count mismatch"
            assert len(act_conv.get("speakers", [])) == len(
                exp_conv.get("speakers", [])
            ), f"conv[{i}] speaker count mismatch"

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="samples/ directory empty.",
    )
    def test_tlk_dump(self):
        """Verify TLK dump matches golden."""
        tlk_file = SAMPLES_DIR / "BIOGame_INT.tlk"
        if not tlk_file.exists():
            pytest.skip(f"{tlk_file} not found")

        actual = run_core("parse-tlk", file=str(tlk_file), dump_all=True)
        golden = _load_golden("tlk/dump_first_50.json")
        if golden is None:
            pytest.skip("golden file not found")

        assert actual["total_entries"] >= 30000, "TLK should have 30k+ entries"
        assert actual["total_entries"] == golden["total_entries"], (
            "TLK entry count mismatch"
        )

        # Compare first 50 entries
        for i in range(min(50, len(actual["entries"]), len(golden["entries"]))):
            assert actual["entries"][i]["StringID"] == golden["entries"][i]["StringID"]
            assert actual["entries"][i]["Text"] == golden["entries"][i]["Text"]

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="samples/ directory empty.",
    )
    def test_pcc_header(self):
        """Verify PCC header extraction matches golden."""
        pcc_file = SAMPLES_DIR / "BioD_CitHub_LOC_INT.pcc"
        if not pcc_file.exists():
            pytest.skip(f"{pcc_file} not found")

        actual = run_core("parse-pcc", file=str(pcc_file), exports_only=True)
        golden = _load_golden("pcc/BioD_CitHub_LOC_INT_exports.json")
        if golden is None:
            pytest.skip("golden file not found")

        assert actual["game_profile"] == golden["game_profile"]
        assert actual["compressed"] == golden["compressed"]
        assert len(actual["exports"]) == len(golden["exports"])
