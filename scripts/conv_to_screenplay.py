"""Convert a conversation JSON to a readable screenplay format."""

import json, sys
from collections import defaultdict

SPEAKER_ALIASES = {
    "player": "SHEPARD",
    "owner": "LIA'VAEL",
    "cithub_rp2_csec": "C-SEC OFFICER",
    "cithub_rp2_lostwallet": "KOR TUN",
    "hench_tali": "TALI",
    "hench_garrus": "GARRUS",
    "hench_vixen": "MIRANDA",
    "hench_leading": "JACOB",
    "hench_convict": "JACK",
    "hench_thief": "KASUMI",
    "hench_veteran": "ZAEED",
}


def speaker_display(tag):
    return SPEAKER_ALIASES.get(tag, tag.upper())


def main():
    path = sys.argv[1]
    with open(path, encoding="utf-8") as f:
        data = json.load(f)

    for conv in data.get("conversations", []):
        conv_id = conv["id"]
        entries = {e["id"]: e for e in conv.get("entries", [])}
        replies = {r["id"]: r for r in conv.get("replies", [])}

        # Collect node types
        entry_lines = defaultdict(list)  # speaker -> [(id, text, reply_links)]
        reply_nodes = {}  # id -> (text, targets)

        for e in entries.values():
            tag = e.get("speaker_tag", "?")
            txt = e.get("line_text") or f"[strref {e.get('line_strref', '?')}]"
            links = e.get("reply_links") or []
            entry_lines[tag].append((e["id"], txt, links))

        for r in replies.values():
            txt = r.get("line_text") or f"[strref {r.get('line_strref', '?')}]"
            targets = r.get("target_entry_ids") or []
            reply_nodes[r["id"]] = (txt, targets)

        print(f"# {conv_id} (export {conv['export_index']})")
        print(f"# Parse mode: {conv.get('parse_mode', '?')}")
        print()

        # Speakers
        speakers = conv.get("speakers", [])
        if speakers:
            print("## Speakers")
            for s in speakers:
                tag = s.get("tag", "")
                friendly = s.get("friendly_name", "") or tag
                alias = speaker_display(tag)
                print(f"#   {alias:20s} = {friendly}")
            print()

        # Start nodes
        starts = conv.get("starts", [])
        if starts:
            print("## Entry Points")
            for s in starts:
                targets = s.get("target_entry_ids", [])
                print(f"#   Start {s['id']} -> entries {targets}")
            print()

        # Print conversation as screenplay
        print("## Dialogue")
        print()

        # Sort entries by ID for linear-ish reading
        for tag, lines in sorted(entry_lines.items()):
            if tag == "owner":
                continue  # handle owner last
            alias = speaker_display(tag)
            for eid, txt, links in lines:
                print(f"**{alias}** [{eid}]")
                print(f"  {txt}")
                if links:
                    print(f"  -> replies: {links}")
                print()

        # Print owner lines (Lia'Vael) last for emphasis
        if "owner" in entry_lines:
            alias = speaker_display("owner")
            for eid, txt, links in entry_lines["owner"]:
                print(f"**{alias}** [{eid}]")
                print(f"  {txt}")
                if links:
                    print(f"  -> replies: {links}")
                print()

        # Replies
        print("## Player Choices (Replies)")
        print()
        for rid, (txt, targets) in sorted(reply_nodes.items()):
            print(f"**SHEPARD** [reply {rid}]")
            print(f"  {txt}")
            if targets:
                print(f"  -> entries: {targets}")
            print()

        # Graph summary
        print("---")
        print(
            f"Total: {len(entries)} entries, {len(replies)} replies, {len(starts)} starts"
        )
        print()


if __name__ == "__main__":
    main()
