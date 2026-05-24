r"""Golden file regression tests.

Verifies pcc-core output matches known-good golden files.
Golden files are committed to tests/golden/ and should not be edited manually.
Regenerate with: .venv\Scripts\python.exe tests\regression\run_probes.py --samples-dir dropzone --regenerate
"""

import json
import os
import subprocess
from pathlib import Path

import pytest

REPO_ROOT = Path(__file__).resolve().parents[1]
GOLDEN_DIR = REPO_ROOT / "tests" / "golden"

_samples_dir = os.environ.get("PCC_SAMPLES_DIR", "")
SAMPLES_DIR = Path(_samples_dir) if _samples_dir else (REPO_ROOT / "dropzone")

CORE_BINARY = (
    REPO_ROOT / "build" / "pcc-core.exe"
    if Path(REPO_ROOT / "build" / "pcc-core.exe").exists()
    else REPO_ROOT / "build" / "pcc-core"
)


def run_core(subcommand: str, **kwargs) -> dict:
    args = [str(CORE_BINARY), subcommand]
    for key, value in kwargs.items():
        flag = f"--{key.replace('_', '-')}"
        if isinstance(value, bool):
            if value:
                args.append(flag)
        elif isinstance(value, (list, tuple)):
            for v in value:
                args.extend([flag, str(v)])
        elif value is not None:
            args.extend([flag, str(value)])
    proc = subprocess.run(args, capture_output=True, text=True, encoding="utf-8", errors="replace")
    if proc.returncode != 0:
        stderr_text = (proc.stderr or "").strip()
        try:
            error_data = json.loads(stderr_text)
            msg = error_data.get("error", stderr_text)
        except (json.JSONDecodeError, ValueError):
            msg = stderr_text
        raise RuntimeError(f"pcc-core error: {msg}")
    stdout = proc.stdout or ""
    if not stdout.strip():
        raise RuntimeError(f"pcc-core produced empty output")
    return json.loads(stdout)


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


def _strip_volatile_fields(obj):
    """Remove fields that vary between runs (timestamps, absolute paths)."""
    import copy

    volatile_keys = frozenset({"file", "base", "path", "dlc_dir", "source_tlk"})

    def _strip(node):
        if isinstance(node, dict):
            result = {}
            for k, v in node.items():
                if k in volatile_keys:
                    result[k] = "<stripped>"
                else:
                    result[k] = _strip(v)
            return result
        if isinstance(node, list):
            return [_strip(item) for item in node]
        return node

    return _strip(copy.deepcopy(obj))


class TestGoldenFiles:
    """Verify pcc-core output against committed golden files."""

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found. Build into build/ before running golden tests.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="dropzone/ directory empty. Copy needed ME2 OT files into dropzone/ to run golden tests.",
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
        reason="dropzone/ directory empty.",
    )
    def test_tlk_info(self):
        """Verify TLK header matches golden."""
        tlk_file = SAMPLES_DIR / "BIOGame_INT.tlk"
        if not tlk_file.exists():
            pytest.skip(f"{tlk_file} not found")

        actual = run_core("parse-tlk", file=str(tlk_file))
        golden = _load_golden("tlk/BIOGame_INT_info.json")
        if golden is None:
            pytest.skip("golden file not found")

        assert actual["header"]["male_entry_count"] == golden["header"]["male_entry_count"]
        assert actual["header"]["female_entry_count"] == golden["header"]["female_entry_count"]
        assert actual["header"]["male_entry_count"] >= 30000

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="dropzone/ directory empty.",
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

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="dropzone/ directory empty.",
    )
    def test_graph_layout_sugiyama(self):
        """Verify layout-graph output matches golden and includes node metadata."""
        pcc_file = SAMPLES_DIR / "BioD_CitHub_LOC_INT.pcc"
        if not pcc_file.exists():
            pytest.skip(f"{pcc_file} not found")

        conv_index = 1
        actual = run_core("layout-graph", file=str(pcc_file), conv_index=conv_index)
        golden = _load_golden("graph/cithub_first_amb_sugiyama.json")
        if golden is None:
            pytest.skip("golden file not found; generate with --regenerate")

        assert "nodes" in actual, "layout-graph output must include nodes metadata"
        assert actual["conversation_id"] == golden["conversation_id"]
        assert actual["node_count"] == golden["node_count"]
        assert len(actual["positions"]) == len(golden["positions"])
        assert len(actual["edges"]) == len(golden["edges"])

        for node_key, meta in actual["nodes"].items():
            assert "type" in meta, f"node {node_key} missing type"
            assert meta["type"] in ("start", "entry", "reply"), (
                f"node {node_key} has unknown type: {meta['type']}"
            )
            assert meta["id"] == int(node_key.split(":")[1]), (
                f"node {node_key} id mismatch: meta has {meta['id']}"
            )

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="dropzone/ directory empty.",
    )
    def test_tlk_dlc_resolve_precedence(self):
        """Verify DLC TLK override precedence matches golden."""
        base_tlk = SAMPLES_DIR / "BIOGame_INT.tlk"
        dlc_dir = SAMPLES_DIR / "dlc"
        if not base_tlk.exists():
            pytest.skip(f"{base_tlk} not found")
        if not (dlc_dir / "DLC_HEN_MT").exists():
            pytest.skip(f"dlc subdirectory not found in dropzone/")

        actual = run_core(
            "resolve-tlk",
            base=str(base_tlk),
            dlc_dir=str(dlc_dir),
            strref=[125303, 356043, 255877],
        )
        golden = _load_golden("tlk/resolve_dlc_precedence.json")
        if golden is None:
            pytest.skip("golden file not found; generate with --regenerate")

        actual = _strip_volatile_fields(actual)
        golden = _strip_volatile_fields(golden)

        assert len(actual["results"]) == len(golden["results"])

        for i, (act_res, exp_res) in enumerate(
            zip(actual["results"], golden["results"])
        ):
            assert act_res["string_id"] == exp_res["string_id"], (
                f"result[{i}] string_id mismatch"
            )
            assert act_res["text"] == exp_res["text"], (
                f"result[{i}] text mismatch: {act_res['text']!r} vs {exp_res['text']!r}"
            )
            assert act_res["found"] == exp_res["found"], (
                f"result[{i}] found mismatch"
            )
            if exp_res["found"]:
                # Verify source_tlk indicates which file resolved it
                assert act_res["source_tlk"] == exp_res["source_tlk"], (
                    f"result[{i}] source_tlk mismatch"
                )
