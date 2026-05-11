"""Main application — HelloImGui window with docking tabs."""

from imgui_bundle import hello_imgui, imgui, ImVec2

from pcc_toolkit_gui.state import AppState
from pcc_toolkit_gui.views import (
    render_dialogue,
    render_evidence,
    render_package,
    render_tlk,
)

TABS = [
    ("Package", render_package),
    ("TLK", render_tlk),
    ("Dialogue", render_dialogue),
    ("Evidence", render_evidence),
]


def _show_app_menu(state: AppState) -> None:
    if imgui.menu_item("About..."):
        state.show_about = True


def _show_gui(state: AppState) -> None:
    if imgui.begin_tab_bar("MainTabs"):
        for i, (name, render_fn) in enumerate(TABS):
            opened, _selected = imgui.begin_tab_item(name)
            if opened:
                state.active_tab = i
                render_fn(state)
                imgui.end_tab_item()
        imgui.end_tab_bar()

    if state.show_about:
        imgui.set_next_window_size(ImVec2(300, 120), imgui.Cond_.appearing)
        if imgui.begin("About", True):
            imgui.text("PCC Toolkit v2")
            imgui.text("ME2 OT Dialogue Extraction Toolkit")
            imgui.text("Go core + Python GUI (Dear ImGui)")
            if imgui.button("Close"):
                state.show_about = False
        imgui.end()


def _show_menu(state: AppState) -> None:
    if imgui.begin_menu("File"):
        if imgui.menu_item("Load PCC...", "", False, True):
            state.pcc_path = _open_file_dialog()
        if imgui.menu_item("Load TLK...", "", False, True):
            state.tlk_path = _open_tlk_dialog()
        if imgui.menu_item("Exit", "", False, True):
            hello_imgui.get_runner_params().app_shall_exit = True
        imgui.end_menu()
    if imgui.begin_menu("Help"):
        if imgui.menu_item("About", "", False, True):
            state.show_about = True
        imgui.end_menu()


def _show_status(state: AppState) -> None:
    if state.is_loading:
        imgui.text("Loading...")
    elif state.error_message:
        imgui.push_style_color(imgui.Col_.text, imgui.IM_COL32(255, 80, 80, 255))
        imgui.text(state.error_message)
        imgui.pop_style_color()
    else:
        imgui.text(state.status_message)


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


def _open_tlk_dialog() -> str | None:
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


def main() -> None:
    state = AppState()

    params = hello_imgui.RunnerParams()
    params.app_window_params.window_title = "PCC Toolkit v2 — Dialogue Explorer"
    params.app_window_params.window_geometry.size = (1280, 800)
    params.imgui_window_params.show_menu_bar = True
    params.imgui_window_params.show_status_bar = True
    params.callbacks.show_gui = lambda: _show_gui(state)
    params.callbacks.show_menus = lambda: _show_menu(state)
    params.callbacks.show_status = lambda: _show_status(state)
    params.callbacks.show_app_menu_items = lambda: _show_app_menu(state)
    params.fps_idling.enable_idling = False

    hello_imgui.run(params)


if __name__ == "__main__":
    main()
