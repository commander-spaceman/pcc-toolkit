"""Evidence search tab — query dialogue across game files."""

import signal
import subprocess
import sys
import threading
import time

from imgui_bundle import imgui

from pcc_toolkit_gui.engine import EngineError, run_async
from pcc_toolkit_gui.state import AppState

_active_process: subprocess.Popen | None = None
_process_lock = threading.Lock()


def _copy_clipboard(text: str) -> None:
    try:
        import tkinter
        root = tkinter.Tk()
        root.withdraw()
        root.clipboard_clear()
        root.clipboard_append(text)
        root.update()
        root.destroy()
    except Exception:
        pass


def render_evidence(state: AppState) -> None:
    imgui.text("TLK:")
    imgui.same_line()
    imgui.text_disabled(state.tlk_path or "not set")
    imgui.same_line()
    if imgui.button("Set##tlk_path"):
        state.tlk_path = _open_file_dialog()
    if state.tlk_path:
        imgui.same_line()
        if imgui.button("X##tlk_path"):
            state.tlk_path = None

    imgui.text("DLC:")
    imgui.same_line()
    imgui.text_disabled(state.dlc_dir or "not set")
    imgui.same_line()
    if imgui.button("Set##dlc_dir"):
        state.dlc_dir = _open_dir_dialog()
    if state.dlc_dir:
        imgui.same_line()
        if imgui.button("X##dlc_dir"):
            state.dlc_dir = None

    imgui.text("Root:")
    imgui.same_line()
    imgui.text_disabled(state.biogame_root or "not set")
    imgui.same_line()
    if imgui.button("Set##biogame_root"):
        state.biogame_root = _open_dir_dialog()
    if state.biogame_root:
        imgui.same_line()
        if imgui.button("X##biogame_root"):
            state.biogame_root = None

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
        text = ev.get("text", "")[:120]
        label = f"#{sid}: {text}"

        t1 = ev.get("bioconversation", [])
        t2 = ev.get("semantic_container", [])
        t3 = ev.get("container_fallback", [])
        has_children = t1 or t2 or t3

        if has_children:
            if imgui.tree_node(label):
                if imgui.begin_popup_context_item(f"ev_ctx_{sid}"):
                    if imgui.menu_item("Copy Text", "", False, True)[0]:
                        _copy_clipboard(ev.get("text", ""))
                    if imgui.menu_item("Copy StrRef", "", False, True)[0]:
                        _copy_clipboard(str(sid))
                    imgui.end_popup()
                if t1:
                    imgui.text_colored(imgui.ImVec4(0.3, 0.9, 0.3, 1.0), f"Tier 1 — BioConversation ({len(t1)})")
                    for hit in t1:
                        imgui.bullet_text(hit.get("conversation_id", "?"))
                if t2:
                    imgui.text_colored(imgui.ImVec4(0.9, 0.9, 0.3, 1.0), f"Tier 2 — Semantic ({len(t2)})")
                    for hit in t2:
                        imgui.bullet_text(f"{hit.get('export_name', '?')} [{hit.get('class_name', '?')}]")
                if t3:
                    imgui.text_colored(imgui.ImVec4(0.5, 0.5, 0.5, 1.0), f"Tier 3 — Fallback ({len(t3)})")
                imgui.tree_pop()
        else:
            imgui.selectable(label, False)
            if imgui.begin_popup_context_item(f"ev_ctx_{sid}"):
                if imgui.menu_item("Copy Text", "", False, True)[0]:
                    _copy_clipboard(ev.get("text", ""))
                if imgui.menu_item("Copy StrRef", "", False, True)[0]:
                    _copy_clipboard(str(sid))
                imgui.end_popup()

    imgui.end_child()


def _render_loading() -> None:
    t = int(time.time() * 3) % 8
    spinner = ["|", "/", "-", "\\", "|", "/", "-", "\\"][t]

    avail = imgui.get_content_region_avail()
    center_y = avail.y * 0.35

    imgui.set_cursor_pos_y(center_y)
    imgui.text("")
    imgui.text("")

    text_w = imgui.calc_text_size(spinner).x
    imgui.set_cursor_pos_x((avail.x - text_w) * 0.5)
    imgui.push_style_color(imgui.Col_.text, imgui.IM_COL32(100, 180, 255, 255))
    imgui.text(spinner)
    imgui.pop_style_color()

    msg = "Scanning PCC files for dialogue matches..."
    msg_w = imgui.calc_text_size(msg).x
    imgui.set_cursor_pos_x((avail.x - msg_w) * 0.5)
    imgui.text(msg)

    sub = "This may take several minutes on first run."
    sub_w = imgui.calc_text_size(sub).x
    imgui.set_cursor_pos_x((avail.x - sub_w) * 0.5)
    imgui.text_disabled(sub)

    cancel = "Press Cancel to abort"
    cancel_w = imgui.calc_text_size(cancel).x
    imgui.set_cursor_pos_x((avail.x - cancel_w) * 0.5)
    imgui.text_disabled(cancel)


def _start_search(state: AppState) -> None:
    if not state.evidence_query or not state.tlk_path:
        return
    state.is_loading = True
    state.error_message = None
    state.evidence_results = None
    state.search_cancel = False

    if state.biogame_root:
        threading.Thread(target=_run_scan_thread, args=(state,), daemon=True).start()
    else:
        threading.Thread(target=_run_tlk_only_thread, args=(state,), daemon=True).start()


def _run_tlk_only_thread(state: AppState) -> None:
    try:
        from pcc_toolkit_gui.engine import scan_evidence
        result = scan_evidence(
            state.evidence_query,
            tlk=state.tlk_path,
            dlc_dir=state.dlc_dir,
        )
        state.evidence_results = result
        state.status_message = f"Search complete: {result.get('total_hits', 0)} hits"
    except EngineError as e:
        state.error_message = str(e)
    except Exception as e:
        state.error_message = f"Search error: {e}"
    finally:
        state.is_loading = False


def _run_scan_thread(state: AppState) -> None:
    global _active_process
    try:
        import tempfile
        import os as _os
        cache_dir = _os.path.join(tempfile.gettempdir(), "pcc-toolkit")
        _os.makedirs(cache_dir, exist_ok=True)
        cache_file = _os.path.join(cache_dir, "file_cache.json")

        with _process_lock:
            _active_process = run_async(
                "scan-evidence",
                query=state.evidence_query,
                tlk=str(state.tlk_path),
                dlc_dir=state.dlc_dir,
                biogame_root=state.biogame_root,
                cache=cache_file,
            )

        while _active_process.poll() is None:
            if state.search_cancel:
                _kill_process()
                state.status_message = "Search cancelled"
                return
            time.sleep(0.1)

        stdout, stderr = _active_process.communicate(timeout=5)
        if state.search_cancel:
            state.status_message = "Search cancelled"
            return

        if _active_process.returncode != 0:
            raise EngineError((stderr or stdout or "").strip())

        import json
        result = json.loads(stdout)
        state.evidence_results = result
        cached_msg = ""
        if result.get("files_scanned", 0) < 50 and state.biogame_root:
            cached_msg = " (mostly from cache)"
        state.status_message = f"Search complete: {result.get('total_hits', 0)} hits{cached_msg}"

    except EngineError as e:
        if not state.search_cancel:
            state.error_message = str(e)
    except Exception as e:
        if not state.search_cancel:
            state.error_message = f"Search error: {e}"
    finally:
        with _process_lock:
            _active_process = None
        state.is_loading = False
        state.search_cancel = False


def _kill_process() -> None:
    global _active_process
    if _active_process is None:
        return
    try:
        if sys.platform == "win32":
            _active_process.kill()
        else:
            _active_process.send_signal(signal.SIGTERM)
    except Exception:
        pass
    _active_process = None


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
