"""Dialogue explorer tab — conversation graph and detail."""

import math

from imgui_bundle import imgui, ImVec2, ImVec4

from pcc_toolkit_gui.engine import (
    EngineError,
    layout_graph,
    parse_conversations,
    parse_pcc,
)
from pcc_toolkit_gui.state import AppState


NODE_COLORS = {
    "start": imgui.IM_COL32(76, 175, 80, 255),
    "entry": imgui.IM_COL32(33, 150, 243, 255),
    "reply": imgui.IM_COL32(255, 152, 0, 255),
}

NODE_WIDTH = 240
NODE_HEIGHT = 64


def render_dialogue(state: AppState) -> None:
    if imgui.button("Load PCC..."):
        state.pcc_path = _open_file_dialog()
        if state.pcc_path:
            _load_dialogue_file(state)
    imgui.same_line()
    imgui.text_disabled(state.pcc_path or "No file loaded")

    if state.conversations is None:
        return

    convs = state.conversations.get("conversations", [])
    if not convs:
        imgui.text_disabled("No BioConversations found")
        return

    imgui.columns(2, "dialogue_columns", True)

    imgui.begin_child("conv_list", ImVec2(0, -imgui.get_frame_height_with_spacing()), True)
    imgui.text(f"Conversations: {len(convs)}")
    imgui.separator()
    for c in convs:
        label = f"[{c['export_index']}] {c['id']}  ({len(c['entries'])}e/{len(c['replies'])}r)"
        if imgui.selectable(label, c["export_index"] == state.selected_conversation_index):
            state.selected_conversation_index = c["export_index"]
            _load_graph(state)
    imgui.end_child()

    imgui.next_column()
    _render_graph(state)
    imgui.columns(1)


def _load_dialogue_file(state: AppState) -> None:
    state.is_loading = True
    try:
        state.conversations = parse_conversations(state.pcc_path)
        state.selected_conversation_index = None
        state.graph_layout = None
        state.status_message = f"Loaded {len(state.conversations.get('conversations', []))} conversations"
    except EngineError as e:
        state.error_message = str(e)
    finally:
        state.is_loading = False


def _load_graph(state: AppState) -> None:
    if state.selected_conversation_index is None:
        return
    try:
        state.graph_layout = layout_graph(
            state.pcc_path,
            conv_index=state.selected_conversation_index,
        )
    except EngineError as e:
        state.error_message = str(e)


def _render_graph(state: AppState) -> None:
    if state.graph_layout is None:
        imgui.text_disabled("Select a conversation to view graph")
        return

    layout = state.graph_layout
    imgui.text(f"Conversation: {layout.get('conversation_id', '?')}")
    imgui.text(f"Nodes: {layout.get('node_count', 0)}  "
               f"Edges: {len(layout.get('edges', []))}")

    imgui.separator()

    draw_list = imgui.get_window_draw_list()
    canvas_pos = imgui.get_cursor_screen_pos()
    canvas_size = imgui.get_content_region_avail()
    canvas_size.y -= imgui.get_frame_height_with_spacing()

    if canvas_size.x < 50 or canvas_size.y < 50:
        return

    imgui.invisible_button("graph_canvas", canvas_size)
    canvas_max = ImVec2(canvas_pos.x + canvas_size.x, canvas_pos.y + canvas_size.y)

    draw_list.add_rect_filled(canvas_pos, canvas_max, imgui.IM_COL32(30, 30, 35, 255))
    draw_list.add_rect(canvas_pos, canvas_max, imgui.IM_COL32(60, 60, 65, 255))

    positions = layout.get("positions", {})
    edges = layout.get("edges", [])

    zoom = state.graph_view_zoom
    offset_x = state.graph_view_offset[0]
    offset_y = state.graph_view_offset[1]
    center_x = canvas_pos.x + canvas_size.x / 2 + offset_x
    center_y = canvas_pos.y + canvas_size.y / 2 + offset_y

    _render_edges(draw_list, edges, positions, zoom, center_x, center_y, canvas_pos, canvas_max)
    _render_nodes(draw_list, position_to_node_type(layout), positions, zoom, center_x, center_y, canvas_pos, canvas_max)
    _render_legend(draw_list, canvas_pos)


def position_to_node_type(layout: dict) -> dict[str, str]:
    result: dict[str, str] = {}
    for key in layout.get("positions", {}):
        parts = key.split(":")
        if len(parts) >= 2:
            result[key] = parts[0]
    return result


def _render_nodes(draw_list, node_types, positions, zoom, cx, cy, clip_min, clip_max):
    for key, pos in positions.items():
        x = cx + pos[0] * zoom
        y = cy + pos[1] * zoom
        w = NODE_WIDTH * zoom
        h = NODE_HEIGHT * zoom

        if x + w < clip_min.x or x > clip_max.x or y + h < clip_min.y or y > clip_max.y:
            continue

        ntype = node_types.get(key, "entry")
        color = NODE_COLORS.get(ntype, NODE_COLORS["entry"])

        min_pos = ImVec2(x, y)
        max_pos = ImVec2(x + w, y + h)
        draw_list.add_rect_filled(min_pos, max_pos, color, 4.0 * zoom)
        draw_list.add_rect(min_pos, max_pos, imgui.IM_COL32(255, 255, 255, 100), 4.0 * zoom)

        label = key.replace(":", " #").upper()
        text_size = imgui.calc_text_size(label)
        tx = x + (w - text_size.x) / 2
        ty = y + h * 0.2
        draw_list.add_text(ImVec2(tx, ty), imgui.IM_COL32(255, 255, 255, 255), label)


def _render_edges(draw_list, edges, positions, zoom, cx, cy, clip_min, clip_max):
    for edge in edges:
        from_key = f"{edge['from']['type']}:{edge['from']['id']}"
        to_key = f"{edge['to']['type']}:{edge['to']['id']}"
        if from_key not in positions or to_key not in positions:
            continue

        fx = cx + positions[from_key][0] * zoom + NODE_WIDTH * zoom / 2
        fy = cy + positions[from_key][1] * zoom + NODE_HEIGHT * zoom
        tx = cx + positions[to_key][0] * zoom + NODE_WIDTH * zoom / 2
        ty = cy + positions[to_key][1] * zoom

        if (fx < clip_min.x and tx < clip_min.x) or (fy < clip_min.y and ty < clip_min.y):
            continue

        mid_y = (fy + ty) / 2
        offset = min(abs(tx - fx) * 0.3, 80) * zoom
        if tx > fx:
            cp1 = ImVec2(fx + offset, fy)
            cp2 = ImVec2(tx - offset, ty)
        else:
            cp1 = ImVec2(fx - offset, fy)
            cp2 = ImVec2(tx + offset, ty)

        color = imgui.IM_COL32(150, 150, 150, 200)
        draw_list.add_bezier_cubic(cp1, cp2, cp1, cp2, color, 2.0 * zoom)

        mid_x = (fx + tx) / 2
        arrow_size = 8.0 * zoom
        angle = math.atan2(ty - mid_y, tx - mid_x)
        ax = tx - math.cos(angle) * arrow_size
        ay = ty - math.sin(angle) * arrow_size
        draw_list.add_triangle_filled(
            ImVec2(tx, ty),
            ImVec2(ax - math.sin(angle) * arrow_size * 0.5, ay + math.cos(angle) * arrow_size * 0.5),
            ImVec2(ax + math.sin(angle) * arrow_size * 0.5, ay - math.cos(angle) * arrow_size * 0.5),
            imgui.IM_COL32(200, 200, 200, 255),
        )


def _render_legend(draw_list, canvas_pos):
    x = canvas_pos.x + 10
    y = canvas_pos.y + 10
    items = [("Start", NODE_COLORS["start"]),
             ("Entry", NODE_COLORS["entry"]),
             ("Reply", NODE_COLORS["reply"])]
    for label, color in items:
        draw_list.add_rect_filled(ImVec2(x, y), ImVec2(x + 12, y + 12), color)
        draw_list.add_text(ImVec2(x + 16, y), imgui.IM_COL32(200, 200, 200, 255), label)
        y += 16


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
