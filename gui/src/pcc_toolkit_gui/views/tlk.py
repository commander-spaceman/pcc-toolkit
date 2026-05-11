"""TLK viewer tab — search and browse talk file entries."""

from imgui_bundle import imgui

from pcc_toolkit_gui.engine import EngineError, parse_tlk
from pcc_toolkit_gui.state import AppState


def _copy_to_clipboard(text: str) -> None:
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


def render_tlk(state: AppState) -> None:
    if imgui.button("Load TLK..."):
        state.tlk_path = _open_file_dialog()
        if state.tlk_path:
            _load_tlk(state)

    imgui.same_line()
    imgui.text_disabled(state.tlk_path or "No file loaded")

    if state.tlk_entries is None:
        return

    header = state.tlk_entries.get("header", {})
    entries = state.tlk_entries.get("entries", [])
    results = state.tlk_entries.get("results", [])

    imgui.text(f"Male entries: {header.get('male_entry_count', 0)}  "
               f"Female: {header.get('female_entry_count', 0)}  "
               f"Loaded: {len(entries)}")

    changed, state.tlk_search = imgui.input_text("##tlk_search", state.tlk_search)
    imgui.same_line()
    if imgui.button("Search") or (changed and state.tlk_search):
        _search_tlk(state)

    display = results if results else entries

    imgui.separator()
    if not display:
        imgui.text_disabled("No entries. Load a TLK file or search.")
        return

    imgui.begin_child("tlk_table", imgui.ImVec2(0, 0), True)

    flags = (imgui.TableFlags_.borders_inner_v
             | imgui.TableFlags_.row_bg
             | imgui.TableFlags_.resizable
             | imgui.TableFlags_.scroll_y)

    if imgui.begin_table("tlk_data", 3, flags):
        imgui.table_setup_column("StringID", imgui.TableColumnFlags_.width_fixed, 80)
        imgui.table_setup_column("Text", imgui.TableColumnFlags_.width_stretch)
        imgui.table_setup_column("Source", imgui.TableColumnFlags_.width_fixed, 0)
        imgui.table_headers_row()

        for entry in display[:500]:
            imgui.table_next_row()
            imgui.table_set_column_index(0)
            sid = entry.get("string_id", entry.get("StringID", 0))
            text = entry.get("text", entry.get("Text", ""))
            source = entry.get("source", entry.get("source_tlk", ""))
            imgui.text_disabled(source)

        imgui.end_table()
    imgui.end_child()


def _load_tlk(state: AppState) -> None:
    state.is_loading = True
    state.error_message = None
    try:
        state.tlk_entries = parse_tlk(state.tlk_path, dump_all=True)
        state.status_message = f"Loaded {state.tlk_entries.get('total_entries', 0)} entries"
    except EngineError as e:
        state.error_message = str(e)
        state.tlk_entries = None
    finally:
        state.is_loading = False


def _search_tlk(state: AppState) -> None:
    if not state.tlk_search:
        return
    state.error_message = None
    try:
        state.tlk_entries = parse_tlk(state.tlk_path, search=state.tlk_search)
    except EngineError as e:
        state.error_message = str(e)


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
