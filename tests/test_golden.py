r"""Golden file regression tests.

Verifies pcc-core output matches known-good golden files.
Golden files are committed to tests/golden/ and should not be edited manually.
Regenerate with: .venv\Scripts\python.exe tests\regression\run_probes.py --samples-dir output --regenerate
"""

import json
import os
import subprocess
from pathlib import Path

import pytest

REPO_ROOT = Path(__file__).resolve().parents[1]
GOLDEN_DIR = REPO_ROOT / "tests" / "golden"

_samples_dir = os.environ.get("PCC_SAMPLES_DIR", "")
SAMPLES_DIR = Path(_samples_dir) if _samples_dir else (REPO_ROOT / "output")

CORE_BINARY = (
    REPO_ROOT / "build" / "pcc-core.exe"
    if Path(REPO_ROOT / "build" / "pcc-core.exe").exists()
    else REPO_ROOT / "build" / "pcc-core"
)


def run_core(subcommand: str, allow_nonzero: bool = False, **kwargs) -> dict:
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
    proc = subprocess.run(
        args, capture_output=True, text=True, encoding="utf-8", errors="replace"
    )
    stdout = (proc.stdout or "").strip()
    stderr_text = (proc.stderr or "").strip()
    if proc.returncode != 0:
        if stdout.startswith("{"):
            try:
                return json.loads(stdout)
            except json.JSONDecodeError:
                pass
        try:
            error_data = json.loads(stderr_text)
            msg = error_data.get("error", stderr_text)
        except (json.JSONDecodeError, ValueError):
            msg = stderr_text
        if not msg and stdout:
            try:
                return json.loads(stdout)
            except json.JSONDecodeError:
                pass
        raise RuntimeError(f"pcc-core error: {msg}")
    if not stdout:
        raise RuntimeError("pcc-core produced empty output")
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

    volatile_keys = frozenset(
        {
            "file",
            "base",
            "path",
            "dlc_dir",
            "source_tlk",
            "dir",
            "tlk_path",
            "biogame_root",
        }
    )

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
        reason="output/ directory empty. Copy needed ME2 OT files into output/ to run golden tests.",
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
        reason="output/ directory empty.",
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

        assert (
            actual["header"]["male_entry_count"] == golden["header"]["male_entry_count"]
        )
        assert (
            actual["header"]["female_entry_count"]
            == golden["header"]["female_entry_count"]
        )
        assert actual["header"]["male_entry_count"] >= 30000

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="output/ directory empty.",
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
        reason="output/ directory empty.",
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
        reason="output/ directory empty.",
    )
    def test_tlk_dlc_resolve_precedence(self):
        """Verify DLC TLK override precedence matches golden."""
        base_tlk = SAMPLES_DIR / "BIOGame_INT.tlk"
        dlc_dir = SAMPLES_DIR / "dlc"
        if not base_tlk.exists():
            pytest.skip(f"{base_tlk} not found")
        if not (dlc_dir / "DLC_HEN_MT").exists():
            pytest.skip(f"dlc subdirectory not found in output/")

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
            assert act_res["found"] == exp_res["found"], f"result[{i}] found mismatch"
            if exp_res["found"]:
                # Verify source_tlk indicates which file resolved it
                assert act_res["source_tlk"] == exp_res["source_tlk"], (
                    f"result[{i}] source_tlk mismatch"
                )

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="output/ directory empty.",
    )
    def test_dump_lines_BioD_CitHub_LOC_INT(self):
        """Verify dump-lines output matches golden."""
        pcc_file = SAMPLES_DIR / "BioD_CitHub_LOC_INT.pcc"
        if not pcc_file.exists():
            pytest.skip(f"{pcc_file} not found")

        actual = run_core("dump-lines", file=str(pcc_file))
        golden = _load_golden("conversation/BioD_CitHub_LOC_INT_dump_lines.json")
        if golden is None:
            pytest.skip("golden file not found; generate with --regenerate")

        actual = _strip_volatile_fields(actual)
        golden = _strip_volatile_fields(golden)

        assert actual["game_profile"] == golden["game_profile"]
        assert actual["total"] == golden["total"]
        assert len(actual["lines"]) == len(golden["lines"])
        assert len(actual["lines"]) == golden["total"]

        for i, (act_line, exp_line) in enumerate(zip(actual["lines"], golden["lines"])):
            assert act_line["conversation_id"] == exp_line["conversation_id"], (
                f"line[{i}] conversation_id mismatch"
            )
            assert act_line["node_type"] == exp_line["node_type"], (
                f"line[{i}] node_type mismatch"
            )
            assert act_line["strref"] == exp_line["strref"], (
                f"line[{i}] strref mismatch"
            )

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="output/ directory empty.",
    )
    def test_scan_owners_BioD_CitHub_LOC_INT(self):
        """Verify scan-owners output matches golden."""
        pcc_file = SAMPLES_DIR / "BioD_CitHub_LOC_INT.pcc"
        if not pcc_file.exists():
            pytest.skip(f"{pcc_file} not found")

        actual = run_core("scan-owners", file=str(pcc_file))
        golden = _load_golden("conversation/BioD_CitHub_LOC_INT_scan_owners.json")
        if golden is None:
            pytest.skip("golden file not found; generate with --regenerate")

        actual = _strip_volatile_fields(actual)
        golden = _strip_volatile_fields(golden)

        assert "file" in actual
        assert "owners" in actual
        assert actual["owners"] == golden["owners"]

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="output/ directory empty.",
    )
    def test_batch_validate(self):
        """Verify batch-validate output matches golden."""
        pcc_files = list(SAMPLES_DIR.glob("*.pcc"))
        if not pcc_files:
            pytest.skip("no PCC files in output/")

        actual = run_core("batch-validate", dir=str(SAMPLES_DIR), glob="*.pcc")
        golden = _load_golden("batch/validate_output.json")
        if golden is None:
            pytest.skip("golden file not found; generate with --regenerate")

        actual = _strip_volatile_fields(actual)
        golden = _strip_volatile_fields(golden)

        assert actual["dir"] == golden["dir"]
        assert actual["pattern"] == golden["pattern"]
        assert actual["files_found"] == golden["files_found"]
        assert actual["files_ok"] == golden["files_ok"]
        assert actual["files_error"] == golden["files_error"]
        assert actual["total_conversations"] == golden["total_conversations"]

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="output/ directory empty.",
    )
    def test_evidence_scan_tlk_only(self):
        """Verify scan-evidence TLK-only output matches golden."""
        tlk_file = SAMPLES_DIR / "BIOGame_INT.tlk"
        if not tlk_file.exists():
            pytest.skip(f"{tlk_file} not found")

        actual = run_core("scan-evidence", query="quarian", tlk=str(tlk_file))
        golden = _load_golden("evidence/scan_quarian_tlk_only.json")
        if golden is None:
            pytest.skip("golden file not found; generate with --regenerate")

        actual = _strip_volatile_fields(actual)
        golden = _strip_volatile_fields(golden)

        assert actual["query"] == golden["query"]
        assert actual["files_scanned"] == golden["files_scanned"]
        assert actual["total_hits"] == golden["total_hits"]
        assert actual["total_hits"] == 0
        assert len(actual["candidate_strrefs"]) == len(golden["candidate_strrefs"])
        assert len(actual["evidence"]) == len(golden["evidence"])

        # Verify evidence items have text (reason: resolver resolved TLK text)
        for ev in actual.get("evidence", [])[:5]:
            assert "text" in ev, f"strref {ev.get('strref')} missing text"

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="output/ directory empty.",
    )
    def test_evidence_scan_biogame(self):
        """Verify scan-evidence with --biogame-root matches golden."""
        tlk_file = SAMPLES_DIR / "BIOGame_INT.tlk"
        if not tlk_file.exists():
            pytest.skip(f"{tlk_file} not found")

        pcc_files = list(SAMPLES_DIR.glob("BioD_CitHub_*.pcc"))
        if not pcc_files:
            pytest.skip("no CitHub PCC files found for BioGame test")

        # Build temp BioGame root from samples assets
        import tempfile

        with tempfile.TemporaryDirectory() as tmpdir:
            tmp = Path(tmpdir)
            cooked = tmp / "CookedPC"
            cooked.mkdir()
            tlk_dest = tmp / tlk_file.name
            import shutil

            shutil.copy2(tlk_file, tlk_dest)
            for pcc in pcc_files:
                shutil.copy2(pcc, cooked / pcc.name)

            actual = run_core(
                "scan-evidence",
                query="quarian",
                tlk=str(tlk_dest),
                biogame_root=str(tmp),
            )
            golden = _load_golden("evidence/scan_quarian_biogame.json")
            if golden is None:
                pytest.skip("golden file not found; generate with --regenerate")

            actual = _strip_volatile_fields(actual)
            golden = _strip_volatile_fields(golden)

            assert actual["query"] == golden["query"]
            assert actual["files_scanned"] >= 1
            assert actual["total_hits"] >= golden["total_hits"]
            assert len(actual["candidate_strrefs"]) == len(golden["candidate_strrefs"])
            assert len(actual["evidence"]) == len(golden["evidence"])

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="samples directory empty.",
    )
    def test_batch_extract(self):
        """Verify batch-extract output matches golden."""
        pcc_files = list(SAMPLES_DIR.glob("*.pcc"))
        if not pcc_files:
            pytest.skip("no PCC files in samples dir")

        actual = run_core("batch-extract", dir=str(SAMPLES_DIR), glob="*.pcc")
        golden = _load_golden("batch/extract_output.json")
        if golden is None:
            pytest.skip("golden file not found; generate with --regenerate")

        actual = _strip_volatile_fields(actual)
        golden = _strip_volatile_fields(golden)

        assert actual["dir"] == golden["dir"]
        assert actual["pattern"] == golden["pattern"]
        assert actual["files_found"] == golden["files_found"]
        assert actual["files_ok"] == golden["files_ok"]
        assert actual["files_error"] == golden["files_error"]

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="samples directory empty.",
    )
    def test_evidence_scan_no_hits(self):
        """Verify scan-evidence with no matching hits produces empty results."""
        tlk_file = SAMPLES_DIR / "BIOGame_INT.tlk"
        if not tlk_file.exists():
            pytest.skip(f"{tlk_file} not found")

        actual = run_core(
            "scan-evidence", query="xyzzynonexistent12345", tlk=str(tlk_file)
        )
        golden = _load_golden("evidence/scan_no_hits.json")
        if golden is None:
            pytest.skip("golden file not found; generate with --regenerate")

        actual = _strip_volatile_fields(actual)
        golden = _strip_volatile_fields(golden)

        assert actual["total_hits"] == 0
        assert actual["total_hits"] == golden["total_hits"]
        assert len(actual.get("evidence") or []) == 0

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="output/ directory empty.",
    )
    def test_dump_lines_csv_resolve_tlk(self):
        """Verify dump-lines CSV output resolves TLK text correctly."""
        pcc_file = SAMPLES_DIR / "BioD_CitHub_LOC_INT.pcc"
        tlk_file = SAMPLES_DIR / "BIOGame_INT.tlk"
        if not pcc_file.exists() or not tlk_file.exists():
            pytest.skip("required files not found")

        args = [
            str(CORE_BINARY),
            "dump-lines",
            "--file",
            str(pcc_file),
            "--resolve-tlk",
            str(tlk_file),
            "--format",
            "csv",
        ]
        proc = subprocess.run(
            args, capture_output=True, text=True, encoding="utf-8", errors="replace"
        )
        assert proc.returncode == 0, f"stderr: {proc.stderr}"
        stdout = proc.stdout.strip()
        assert stdout, "CSV output is empty"

        lines = stdout.splitlines()
        assert len(lines) >= 2, f"expected header + at least 1 row, got {len(lines)}"
        header = lines[0]
        assert "conversation_id" in header
        assert "node_type" in header
        assert "strref" in header
        assert "line_text" in header
        assert "line_status" in header

        resolved_count = sum(1 for l in lines[1:] if "resolved" in l)
        assert resolved_count > 0, "expected at least 1 resolved line"

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="output/ directory empty.",
    )
    def test_serialize_resolve_tlk(self):
        """Verify serialize --resolve-tlk output matches golden and has resolved text."""
        pcc_file = SAMPLES_DIR / "BioD_CitHub_LOC_INT.pcc"
        tlk_file = SAMPLES_DIR / "BIOGame_INT.tlk"
        if not pcc_file.exists():
            pytest.skip(f"{pcc_file} not found")
        if not tlk_file.exists():
            pytest.skip(f"{tlk_file} not found")

        actual = run_core("serialize", file=str(pcc_file), resolve_tlk=str(tlk_file))
        golden = _load_golden("conversation/serialize_resolved.json")
        if golden is None:
            pytest.skip("golden file not found; generate with --regenerate")

        actual = _strip_volatile_fields(actual)
        golden = _strip_volatile_fields(golden)

        assert actual["game_profile"] == golden["game_profile"]
        assert actual["compressed"] == golden["compressed"]
        assert len(actual["conversations"]) == len(golden["conversations"])

        for i, (act_conv, exp_conv) in enumerate(
            zip(actual["conversations"], golden["conversations"])
        ):
            assert act_conv["id"] == exp_conv["id"], f"conv[{i}] id mismatch"
            assert len(act_conv.get("entries", [])) == len(
                exp_conv.get("entries", [])
            ), f"conv[{i}] entry count mismatch"
            assert len(act_conv.get("replies", [])) == len(
                exp_conv.get("replies", [])
            ), f"conv[{i}] reply count mismatch"

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="output/ directory empty.",
    )
    def test_parse_conversations_resolve_tlk(self):
        """Verify parse-conversations --resolve-tlk resolves text in all nodes."""
        pcc_file = SAMPLES_DIR / "BioD_CitHub_LOC_INT.pcc"
        tlk_file = SAMPLES_DIR / "BIOGame_INT.tlk"
        if not pcc_file.exists() or not tlk_file.exists():
            pytest.skip("required files not found")

        actual = run_core(
            "parse-conversations", file=str(pcc_file), resolve_tlk=str(tlk_file)
        )

        resolved_entries = 0
        unresolved_entries = 0
        for conv in actual.get("conversations", []):
            for entry in conv.get("entries", []):
                if entry.get("line_strref") is not None:
                    if entry.get("line_status") == "resolved":
                        resolved_entries += 1
                        assert entry.get("line_text"), (
                            f"entry {entry['id']} missing line_text despite resolved status"
                        )
                    elif entry.get("line_status") == "unresolved_strref":
                        unresolved_entries += 1
            for reply in conv.get("replies", []):
                if reply.get("line_strref") is not None:
                    if reply.get("line_status") == "resolved":
                        assert reply.get("line_text"), (
                            f"reply {reply['id']} missing line_text despite resolved status"
                        )

        assert resolved_entries > 0, "expected at least 1 resolved entry with TLK text"

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="output/ directory empty.",
    )
    def test_parse_tlk_dump_all(self):
        """Verify parse-tlk --dump-all produces all entries with correct count."""
        tlk_file = SAMPLES_DIR / "BIOGame_INT.tlk"
        if not tlk_file.exists():
            pytest.skip(f"{tlk_file} not found")

        actual = run_core("parse-tlk", file=str(tlk_file), dump_all=True)

        assert "total_entries" in actual
        assert actual["total_entries"] >= 30000, (
            f"expected >= 30000 entries, got {actual['total_entries']}"
        )
        assert "entries" in actual
        assert len(actual["entries"]) == actual["total_entries"]
        assert actual["total_entries"] >= 36000, (
            f"expected >= 36000 entries, got {actual['total_entries']}"
        )

        first = actual["entries"][0]
        assert "string_id" in first
        assert "text" in first
        assert "source" in first

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="output/ directory empty.",
    )
    def test_resolve_tlk_multi_strref(self):
        """Verify resolve-tlk handles multiple strrefs including invalid ones."""
        base_tlk = SAMPLES_DIR / "BIOGame_INT.tlk"
        dlc_dir = SAMPLES_DIR / "dlc"
        if not base_tlk.exists():
            pytest.skip(f"{base_tlk} not found")
        if not (dlc_dir / "DLC_HEN_MT").exists():
            pytest.skip("dlc subdirectory not found in output/")

        actual = run_core(
            "resolve-tlk",
            base=str(base_tlk),
            dlc_dir=str(dlc_dir),
            strref=[125303, 356043, 255877, 1, 999999],
        )
        golden = _load_golden("tlk/resolve_multi.json")
        if golden is None:
            pytest.skip("golden file not found; generate with --regenerate")

        actual_stripped = _strip_volatile_fields(actual)
        golden_stripped = _strip_volatile_fields(golden)

        assert len(actual_stripped["results"]) == len(golden_stripped["results"])
        for i, (ar, gr) in enumerate(
            zip(actual_stripped["results"], golden_stripped["results"])
        ):
            assert ar["string_id"] == gr["string_id"], f"result[{i}] string_id mismatch"
            assert ar["text"] == gr["text"], f"result[{i}] text mismatch"
            assert ar["found"] == gr["found"], f"result[{i}] found mismatch"

        found_count = sum(1 for r in actual["results"] if r["found"])
        assert found_count >= 3, f"expected >= 3 found strrefs, got {found_count}"

        not_found = [r for r in actual["results"] if not r["found"]]
        assert len(not_found) >= 1, "expected at least 1 not-found strref (999999)"

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="output/ directory empty.",
    )
    def test_batch_edit_dry_run(self):
        """Verify batch-edit --dry-run validates without writing output."""
        pcc_file = SAMPLES_DIR / "BioD_CitHub_LOC_INT.pcc"
        if not pcc_file.exists():
            pytest.skip(f"{pcc_file} not found")

        patch_path = SAMPLES_DIR / "batch_edit_patch.json"
        patch = {
            "add_entries": [{"speaker_id": 0, "line_strref": 663399}],
            "add_replies": [{"line_strref": 663399, "target_entry_ids": [0]}],
        }
        patch_path.write_text(json.dumps(patch))

        try:
            args = [
                str(CORE_BINARY),
                "batch-edit",
                "--dir",
                str(SAMPLES_DIR),
                "--glob",
                "BioD_CitHub_LOC_INT.pcc",
                "--patch",
                str(patch_path),
                "--dry-run",
            ]
            proc = subprocess.run(
                args, capture_output=True, text=True, encoding="utf-8", errors="replace"
            )
            assert proc.returncode == 0, f"batch-edit failed: {proc.stderr}"

            results = json.loads(proc.stdout)
            assert isinstance(results, list)
            assert len(results) >= 1

            for r in results:
                assert "file" in r
                assert "status" in r
                assert "validation" in r
                assert r["status"] == "dry_run", (
                    f"expected dry_run status, got {r['status']}"
                )
        finally:
            patch_path.unlink(missing_ok=True)

    @pytest.mark.skipif(
        not CORE_BINARY.exists(),
        reason="pcc-core binary not found.",
    )
    @pytest.mark.skipif(
        not SAMPLES_DIR.exists() or not any(SAMPLES_DIR.iterdir()),
        reason="output/ directory empty.",
    )
    def test_export_detail_semantic_props(self):
        """Verify parse-pcc --export-index --semantic-props matches golden."""
        pcc_file = SAMPLES_DIR / "BioD_CitHub_LOC_INT.pcc"
        if not pcc_file.exists():
            pytest.skip(f"{pcc_file} not found")

        actual = run_core(
            "parse-pcc",
            file=str(pcc_file),
            export_index=1,
            semantic_props=True,
            property_tags=True,
        )
        golden = _load_golden("pcc/export_detail_semantic.json")
        if golden is None:
            pytest.skip("golden file not found; generate with --regenerate")

        assert actual["index"] == golden["index"]
        assert actual["class_name"] == golden["class_name"]
        assert actual["object_name"] == golden["object_name"]

        assert "serial_data" in actual
        assert len(actual["serial_data"]) > 0
        assert "property_tags" in actual
        assert len(actual["property_tags"]) > 0
        assert "semantic_props" in actual
        assert len(actual["semantic_props"]) > 0

    def test_edit_conversation_dry_run(self):
        """Verify edit-conversation dry-run produces expected validation."""
        pcc_name = "BioD_BchLmL_201BeachPath_LOC_INT.pcc"
        pcc_path = SAMPLES_DIR / pcc_name
        if not pcc_path.exists():
            pytest.skip(f"sample PCC not found: {pcc_name}")

        patch_path = SAMPLES_DIR / "edit_patch.json"
        patch = {
            "add_entries": [{"speaker_id": 0, "line_strref": 663399}],
            "add_replies": [{"line_strref": 663399, "target_entry_ids": [3]}],
        }
        patch_path.write_text(json.dumps(patch))

        try:
            actual = run_core(
                "edit-conversation",
                file=str(pcc_path),
                conv_index=0,
                patch=str(patch_path),
                dry_run=True,
            )
        finally:
            patch_path.unlink(missing_ok=True)

        golden = _load_golden("edit/edit_dry_run.json")
        if golden is None:
            pytest.skip("golden file not found; generate with --regenerate")

        actual = _strip_volatile_fields(actual)
        golden = _strip_volatile_fields(golden)
        issues = _compare_keys(actual, golden)
        if issues:
            pytest.fail("\n".join(issues))
