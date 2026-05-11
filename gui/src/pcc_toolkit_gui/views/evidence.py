"""Evidence search tab — query dialogue across game files."""

import threading
import time

from imgui_bundle import imgui

from pcc_toolkit_gui.engine import EngineError, scan_evidence
from pcc_toolkit_gui.state import AppState


def _path_row(label: str, path: str | None, set_cb, clear_cb) -> None:
    imgui.text(label)
    imgui.same_line()
    imgui.text_disabled(path or "not set")
    imgui.same_line()
    if imgui.button(f"Set##{label}"):
        set_cb()
    if path:
        imgui.same_line()
        if imgui.button(f"X##{label}"):
            clear_cb()


def render_evidence(state: AppState) -> None:
    _path_row("TLK:", state.tlk_path,
              lambda: setattr(state, 'tlk_path', _open_file_dialog()),
              lambda: setattr(state, 'tlk_path', None))
    _path_row("DLC:", state.dlc_dir,
              lambda: setattr(state, 'dlc_dir', _open_dir_dialog()),
              lambda: setattr(state, 'dlc_dir', None))
    _path_row("Root:", state.biogame_root,
              lambda: setattr(state, 'biogame_root', _open_dir_dialog()),
              lambda: setattr(state, 'biogame_root', None))

    imgui.separator()
    changed, state.evidence_query = imgui.input_text("##evidence_query", state.evidence_query,
                                                      imgui.InputTextFlags_.enter_returns_true)
    imgui.same_line()
    if state.is_loading:
        if imgui.button("Cancel"):
            state.search_cancel = True
    else:
        if (imgui.button("Search") or changed) and state.evidence_query and state.tlk_path:
            _start_search(state)

    if state.is_loading:
        _render_loading()
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


def _render_loading() -> None:
    t = int(time.time() * 2) % 8
    spinner = ["|", "/", "-", "\\", "|", "/", "-", "\\"][t]
    imgui.text(f"{spinner} Searching... (this may take a while)")
    imgui.text("")
    imgui.text_disabled("Files are being scanned for matching dialogue.")
    imgui.text_disabled("The first search may take several minutes.")
    imgui.text_disabled("Results will appear automatically when complete.")


def _start_search(state: AppState) -> None:
    if not state.evidence_query or not state.tlk_path:
        return
    state.is_loading = True
    state.error_message = None
    state.evidence_results = None
    state.search_cancel = False
    threading.Thread(target=_run_search_thread, args=(state,), daemon=True).start()


def _run_search_thread(state: AppState) -> None:
    try:
        result = scan_evidence(
            state.evidence_query,
            tlk=state.tlk_path,
            dlc_dir=state.dlc_dir,
            biogame_root=state.biogame_root,
        )
        if not state.search_cancel:
            state.evidence_results = result
            state.status_message = f"Search complete: {result.get('total_hits', 0)} hits"
        else:
            state.status_message = "Search cancelled"
    except EngineError as e:
        if not state.search_cancel:
            state.error_message = str(e)
    except Exception as e:
        if not state.search_cancel:
            state.error_message = f"Search error: {e}"
    finally:
        state.is_loading = False
        state.search_cancel = False


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
