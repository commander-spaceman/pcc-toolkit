"""Terminal output formatting using Rich."""

from rich.console import Console
from rich.table import Table
from rich.text import Text

console = Console()

STATUS_COLORS = {
    "valid": "green",
    "warning": "yellow",
    "invalid": "red",
}

SEVERITY_COLORS = {
    "error": "red",
    "warning": "yellow",
    "info": "blue",
}

SEVERITY_ICONS = {
    "error": "✗",
    "warning": "⚠",
    "info": "ℹ",
}


def pcc_export_table(exports: list[dict]) -> Table:
    table = Table(title="PCC Exports")
    table.add_column("Index", justify="right", style="cyan")
    table.add_column("Class", style="blue")
    table.add_column("Object", style="green")
    table.add_column("Size", justify="right")
    table.add_column("Offset", justify="right")

    for e in exports:
        table.add_row(
            str(e.get("index", "")),
            e.get("class_name", ""),
            e.get("object_name", ""),
            str(e.get("serial_size", "")),
            str(e.get("serial_offset", "")),
        )
    return table


def conversation_table(conversations: list[dict]) -> Table:
    table = Table(title="BioConversations")
    table.add_column("Index", justify="right", style="cyan")
    table.add_column("ID", style="green")
    table.add_column("Mode", style="yellow")
    table.add_column("Entries", justify="right")
    table.add_column("Replies", justify="right")
    table.add_column("Status", justify="center")

    for c in conversations:
        val = c.get("validation", {})
        status = val.get("status", "?") if val else "?"
        sc = STATUS_COLORS.get(status, "white")
        table.add_row(
            str(c.get("export_index", "")),
            c.get("id", ""),
            c.get("parse_mode", ""),
            str(len(c.get("entries", []))),
            str(len(c.get("replies", []))),
            f"[{sc}]{status}[/{sc}]",
        )
    return table


def validation_summary(report: dict) -> None:
    s = report.get("report_summary", {})
    console.print()
    console.print(
        f"  Total: {s.get('total', 0)}  "
        f"[green]Valid: {s.get('valid', 0)}[/green]  "
        f"[yellow]Warning: {s.get('warning', 0)}[/yellow]  "
        f"[red]Invalid: {s.get('invalid', 0)}[/red]"
    )

    for r in report.get("results", []):
        status = r.get("status", "?")
        sc = STATUS_COLORS.get(status, "white")
        icon = {"valid": "✓", "warning": "⚠", "invalid": "✗"}.get(status, "?")
        console.print(
            f"  [{sc}]{icon} {r['conversation_id']} (#{r['export_index']}) — {status}[/{sc}]"
        )
        for issue in r.get("issues", []):
            sev = issue.get("severity", "info")
            ic = SEVERITY_ICONS.get(sev, "?")
            ico = SEVERITY_COLORS.get(sev, "white")
            loc = (
                f" [{issue.get('node_type', '')}#{issue.get('node_id', '')}]"
                if issue.get("node_id") is not None
                else ""
            )
            console.print(f"      [{ico}]{ic} {sev}{loc}: {issue['message']}[/{ico}]")


def evidence_report(report: dict) -> None:
    console.print(f"Query: [bold]{report.get('query', '?')}[/bold]")
    console.print(
        f"Files scanned: {report.get('files_scanned', 0)}  "
        f"Files with hits: {report.get('files_with_hits', 0)}  "
        f"Total hits: {report.get('total_hits', 0)}"
    )

    for ev in report.get("evidence", []):
        console.print(f"\n[bold]--- StrRef #{ev['strref']} ---[/bold]")
        if ev.get("text"):
            console.print(f"  Text: [italic]{ev['text']}[/italic]")

        t1 = ev.get("bioconversation", [])
        if t1:
            console.print(f"  [green]Tier 1 — BioConversation[/green] ({len(t1)}):")
            for hit in t1:
                console.print(
                    f"    {hit.get('conversation_id', '?')} in {hit.get('file_path', '?')}"
                )

        t2 = ev.get("semantic_container", [])
        if t2:
            console.print(
                f"  [yellow]Tier 2 — Semantic container[/yellow] ({len(t2)}):"
            )
            for hit in t2:
                console.print(
                    f"    {hit.get('export_name', '?')} [{hit.get('class_name', '?')}]"
                )

        t3 = ev.get("container_fallback", [])
        if t3:
            console.print(f"  [dim]Tier 3 — Container fallback[/dim] ({len(t3)})")


def batch_summary(report: dict) -> None:
    console.print(f"Directory: [bold]{report.get('dir', '?')}[/bold]")
    console.print(f"Pattern: {report.get('pattern', '?')}")
    console.print(
        f"Files found: {report.get('files_found', 0)}  "
        f"OK: [green]{report.get('files_ok', 0)}[/green]  "
        f"Errors: [red]{report.get('files_error', 0)}[/red]"
    )
    if report.get("total_conversations"):
        console.print(
            f"Conversations: {report.get('total_conversations', 0)}  "
            f"[green]Valid: {report.get('valid', 0)}[/green]  "
            f"[yellow]Warning: {report.get('warning', 0)}[/yellow]  "
            f"[red]Invalid: {report.get('invalid', 0)}[/red]"
        )


def dump_lines_table(lines: list[dict]) -> Table:
    table = Table(title="Dialogue Lines")
    table.add_column("Conv", style="cyan")
    table.add_column("Type", style="yellow")
    table.add_column("ID", justify="right")
    table.add_column("Speaker", style="green")
    table.add_column("StrRef", justify="right")
    table.add_column("Text")

    for line in lines:
        text = line.get("line_text") or ""
        if len(text) > 80:
            text = text[:77] + "..."
        table.add_row(
            line.get("conversation_id", "") or "",
            line.get("node_type", ""),
            str(line.get("node_id", "")),
            line.get("speaker_tag", "") or "",
            str(line.get("strref", "")),
            text.replace("\n", " "),
        )
    return table


def owner_table(owners: dict) -> Table:
    table = Table(title="Conversation Owners")
    table.add_column("Conversation", style="cyan")
    table.add_column("Owner", style="green")

    for conv_name, owner_tag in owners.items():
        table.add_row(conv_name, owner_tag)
    return table
