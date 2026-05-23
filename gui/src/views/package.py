"""Package viewer tab — export tree and detail."""

from imgui_bundle import imgui

from engine import EngineError, parse_pcc
from state import AppState


def render_package(state: AppState) -> None:
    if imgui.button("Load PCC..."):
        state.pcc_path = _open_file_dialog()
        if state.pcc_path:
            _load_pcc(state)
    imgui.same_line()
    imgui.text_disabled(state.pcc_path or "No file loaded")

    if state.error_message:
        imgui.push_style_color(imgui.Col_.text, imgui.IM_COL32(255, 80, 80, 255))
        imgui.text_wrapped(state.error_message)
        imgui.pop_style_color()
        if imgui.button("Dismiss"):
            state.error_message = None

    if state.pcc_exports is None:
        return

    exports = state.pcc_exports.get("exports", [])
    imgui.text(f"Profile: {state.pcc_exports.get('game_profile', '?')}  "
               f"Exports: {len(exports)}")

    imgui.separator()
    imgui.columns(2, "package_columns", True)

    imgui.text("Exports")
    imgui.separator()
    groups: dict[str, list[dict]] = {}
    for e in exports:
        cls = e.get("class_name", "(unknown)")
        groups.setdefault(cls, []).append(e)

    for cls in sorted(groups):
        if imgui.tree_node(cls):
            for e in groups[cls]:
                label = f"[{e['index']}] {e.get('object_name', '?')}"
                is_selected = e["index"] == state.selected_export_index
                if is_selected:
                    imgui.push_style_color(imgui.Col_.text, imgui.IM_COL32(100, 200, 255, 255))
                if imgui.selectable(label, is_selected):
                    state.selected_export_index = e["index"]
                if is_selected:
                    imgui.pop_style_color()
            imgui.tree_pop()

    imgui.next_column()
    _render_export_detail(state)

    imgui.columns(1)


def _load_pcc(state: AppState) -> None:
    state.is_loading = True
    state.error_message = None
    try:
        state.pcc_exports = parse_pcc(state.pcc_path, exports_only=True)
        state.status_message = f"Loaded {len(state.pcc_exports.get('exports', []))} exports"
    except EngineError as e:
        state.error_message = str(e)
    finally:
        state.is_loading = False


def _render_export_detail(state: AppState) -> None:
    if state.selected_export_index is None:
        imgui.text_disabled("Select an export to view details")
        return

    exports = state.pcc_exports.get("exports", [])
    exp = None
    for e in exports:
        if e["index"] == state.selected_export_index:
            exp = e
            break
    if exp is None:
        return

    imgui.text(f"Export #{exp['index']}")
    imgui.separator()
    imgui.text(f"Class:  {exp.get('class_name', 'N/A')}")
    imgui.text(f"Object: {exp.get('object_name', 'N/A')}")
    imgui.text(f"Serial offset: {exp.get('serial_offset', 0)}")
    imgui.text(f"Serial size:   {exp.get('serial_size', 0)}")


def _open_file_dialog() -> str | None:
    try:
        import tkinter.filedialog
        import tkinter
        root = tkinter.Tk()
        root.withdraw()
        path = tkinter.filedialog.askopenfilename(
            title="Open PCC file",
            filetypes=[("PCC files", "*.pcc"), ("All files", "*.*")],
        )
        root.destroy()
        return path or None
    except Exception:
        return None
