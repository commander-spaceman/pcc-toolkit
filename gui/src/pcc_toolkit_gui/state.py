"""UI state — selection, zoom, pan, paths. No domain logic."""

from dataclasses import dataclass, field


@dataclass
class AppState:
    pcc_path: str | None = None
    tlk_path: str | None = None
    dlc_dir: str | None = None
    biogame_root: str | None = None

    pcc_exports: dict | None = None
    tlk_entries: dict | None = None
    tlk_dump: dict | None = None
    conversations: dict | None = None
    graph_layout: dict | None = None

    selected_export_index: int | None = None
    selected_conversation_index: int | None = None
    selected_node_key: str | None = None

    graph_view_offset: tuple[float, float] = (0.0, 0.0)
    graph_view_zoom: float = 1.0

    status_message: str = "Ready"
    is_loading: bool = False
    error_message: str | None = None
    conv_filter: str = ""
    tlk_search: str = ""
    selected_tlk_entry: str = ""
    evidence_query: str = ""
    evidence_results: dict | None = None
    search_cancel: bool = False
    cache_path: str = ""

    active_tab: int = 0
    show_about: bool = False
