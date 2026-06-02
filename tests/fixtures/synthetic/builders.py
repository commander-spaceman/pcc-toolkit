"""Synthetic test fixture builders for tests that do not require game files."""

import json
from pathlib import Path


class ConversationBuilder:
    """Builds synthetic conversation dicts for testing CLI formatters."""

    def __init__(self, conv_id: str = "test_conv", export_index: int = 1):
        self.data = {
            "id": conv_id,
            "export_index": export_index,
            "game_profile": "me2_ot",
            "parse_mode": "struct_property_semantic",
            "entries": [],
            "replies": [],
            "speakers": [],
            "starts": [],
        }

    def add_entry(
        self,
        node_id: int,
        speaker_tag: str = "No data",
        strref: int = 0,
        text: str = "",
    ):
        entry = {
            "id": node_id,
            "speaker_tag": speaker_tag,
            "line_strref": strref,
        }
        if text:
            entry["line_text"] = text
            entry["line_status"] = "resolved"
        self.data.setdefault("entries", []).append(entry)
        return self

    def add_reply(self, node_id: int, strref: int = 0, text: str = ""):
        reply = {
            "id": node_id,
            "line_strref": strref,
            "target_entry_ids": [node_id + 1],
        }
        if text:
            reply["line_text"] = text
            reply["line_status"] = "resolved"
        self.data.setdefault("replies", []).append(reply)
        return self

    def add_speaker(self, speaker_id: int, tag: str, display_name: str = ""):
        speaker = {"id": speaker_id, "tag": tag}
        if display_name:
            speaker["display_name"] = display_name
        self.data.setdefault("speakers", []).append(speaker)
        return self

    def add_validation(self, status: str = "valid"):
        self.data["validation"] = {"status": status, "issues": []}
        return self

    def build(self) -> dict:
        return self.data


class ParseResultBuilder:
    """Builds synthetic parse-result dict matching pcc-core JSON output."""

    def __init__(self, file_name: str = "test.pcc", game_profile: str = "me2_ot"):
        self.data = {
            "file": file_name,
            "game_profile": game_profile,
            "conversations": [],
            "errors": [],
        }

    def add_conversation(self, conv: dict):
        self.data["conversations"].append(conv)
        return self

    def build(self) -> dict:
        return self.data


class DumpLinesBuilder:
    """Builds synthetic dump-lines output dict."""

    def __init__(self, file_name: str = "test.pcc", game_profile: str = "me2_ot"):
        self.data = {
            "file": file_name,
            "game_profile": game_profile,
            "lines": [],
            "total": 0,
        }

    def add_line(
        self,
        conversation_id: str,
        node_type: str,
        node_id: int,
        speaker_tag: str,
        strref: int,
        text: str = "",
    ):
        line = {
            "conversation_id": conversation_id,
            "export_index": 1,
            "node_type": node_type,
            "node_id": node_id,
            "speaker_tag": speaker_tag,
            "strref": strref,
            "line_status": "resolved" if text else "unresolved_strref",
            "file": self.data["file"],
        }
        if text:
            line["line_text"] = text
        self.data["lines"].append(line)
        self.data["total"] = len(self.data["lines"])
        return self

    def build(self) -> dict:
        return self.data


class EvidenceBuilder:
    """Builds synthetic evidence report dict."""

    def __init__(
        self,
        query: str = "test query",
        tlk_path: str = "test.tlk",
    ):
        self.data = {
            "query": query,
            "tlk_path": tlk_path,
            "candidate_strrefs": [],
            "files_scanned": 0,
            "files_with_hits": 0,
            "total_hits": 0,
            "evidence": [],
        }

    def add_strref_evidence(self, strref: int, text: str = ""):
        ev = {"strref": strref}
        if text:
            ev["text"] = text
        self.data["evidence"].append(ev)
        self.data["candidate_strrefs"].append(strref)
        return self

    def add_narrative_profiles(self, profiles: list[dict]):
        self.data["narrative_profiles"] = profiles
        return self

    def build(self) -> dict:
        return self.data


def make_minimal_conversation() -> dict:
    return (
        ConversationBuilder("minimal_conv")
        .add_entry(0, "Tali", 12345, "Hello, Commander.")
        .add_reply(0, 12346, "Let's go.")
        .add_speaker(0, "Tali", "Tali'Zorah nar Rayya")
        .build()
    )


def make_valid_parse_result() -> dict:
    return (
        ParseResultBuilder()
        .add_conversation(
            ConversationBuilder("conv_a", 1)
            .add_entry(0, "Garrus", 100, "Calibrations.")
            .add_reply(0, 101, "Okay.")
            .add_speaker(0, "Garrus", "Garrus Vakarian")
            .add_validation("valid")
            .build(),
        )
        .add_conversation(
            ConversationBuilder("conv_b", 3)
            .add_entry(0, "Mordin", 200, "Fascinating.")
            .add_reply(0, 201, "Indeed.")
            .add_speaker(0, "Mordin", "Mordin Solus")
            .add_validation("valid")
            .build(),
        )
        .build()
    )
