"""Evidence search tab — query dialogue across game files."""

from imgui_bundle import imgui

from pcc_toolkit_gui.engine import EngineError, scan_evidence
from pcc_toolkit_gui.state import AppState


def render_evidence(state: AppState) -> None:
    imgui.text("TLK path:")
    imgui.same_line()
    imgui.text_disabled(state.tlk_path or "not set")
    imgui.same_line()
    if imgui.button("Set TLK..."):
        state.tlk_path = _open_file_dialog()

    imgui.text("DLC dir:")
    imgui.same_line()
    imgui.text_disabled(state.dlc_dir or "not set")
    imgui.same_line()
    if imgui.button("Set DLC..."):
        state.dlc_dir = _open_dir_dialog()

    imgui.text("BioGame root:")
    imgui.same_line()
    imgui.text_disabled(state.biogame_root or "not set")
    imgui.same_line()
    if imgui.button("Set Root..."):
        state.biogame_root = _open_dir_dialog()

    imgui.separator()
    changed, state.evidence_query = imgui.input_text("##evidence_query", state.evidence_query,
                                                     imgui.InputTextFlags_.enter_returns_true)
    imgui.same_line()
    if imgui.button("Search") or changed:
        _run_search(state)

    if state.is_loading:
        imgui.text("Searching...")
        return

    if state.evidence_results is None:
        imgui.text_disabled("Enter a query and click Search")
        return

    report = state.evidence_results
    imgui.text(f"Files scanned: {report.get('files_scanned', 0)}  "
               f"Hits: {report.get('total_hits', 0)}")
    imgui.separator()

    evidence_list = report.get("evidence", [])
    if not evidence_list:
        imgui.text_disabled("No results found")
        return

    imgui.begin_child("evidence_list", imgui.ImVec2(0, 0), True)
    for ev in evidence_list:
        sid = ev["strref"]
        text = ev.get("text", "")[:100]
        label = f"#{sid}: {text}"

        if imgui.tree_node(label):
            t1 = ev.get("bioconversation", [])
            if t1:
                imgui.text_colored(imgui.ImVec4(0.3, 0.9, 0.3, 1.0), f"Tier 1 — BioConversation ({len(t1)})")
                for hit in t1:
                    imgui.bullet_text(hit.get("conversation_id", "?"))

            t2 = ev.get("semantic_container", [])
            if t2:
                imgui.text_colored(imgui.ImVec4(0.9, 0.9, 0.3, 1.0), f"Tier 2 — Semantic ({len(t2)})")
                for hit in t2:
                    imgui.bullet_text(f"{hit.get('export_name', '?')} [{hit.get('class_name', '?')}]")

            t3 = ev.get("container_fallback", [])
            if t3:
                imgui.text_colored(imgui.ImVec4(0.5, 0.5, 0.5, 1.0), f"Tier 3 — Fallback ({len(t3)})")

            imgui.tree_pop()

    imgui.end_child()


def _run_search(state: AppState) -> None:
    if not state.evidence_query or not state.tlk_path:
        return
    state.is_loading = True
    try:
        state.evidence_results = scan_evidence(
            state.evidence_query,
            tlk=state.tlk_path,
            dlc_dir=state.dlc_dir,
            biogame_root=state.biogame_root,
        )
        state.status_message = f"Search complete: {state.evidence_results.get('total_hits', 0)} hits"
    except EngineError as e:
        state.error_message = str(e)
    finally:
        state.is_loading = False


def _open_file_dialog() -> str | None:
    try:
        import tkinter.filedialog
        import tkinter
        root = tkinter.Tk()
        root.withdraw()
        path = tkinter.filedialog.askopenfilename(
            title="Open TLK file",
            filetypes=[("TLK files", "*.tlk"), ("All files", "*.*")],
        )
        root.destroy()
        return path or None
    except Exception:
        return None


def _open_dir_dialog() -> str | None:
    try:
        import tkinter.filedialog
        import tkinter
        root = tkinter.Tk()
        root.withdraw()
        path = tkinter.filedialog.askdirectory(title="Select DLC directory")
        root.destroy()
        return path or None
    except Exception:
        return None
