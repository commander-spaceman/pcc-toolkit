package graph

type NodeKey struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
}

type Edge struct {
	From NodeKey `json:"from"`
	To   NodeKey `json:"to"`
}

type NodeMeta struct {
	Type          string   `json:"type"`
	ID            int      `json:"id"`
	SpeakerTag    string   `json:"speaker_tag,omitempty"`
	LineText      string   `json:"line_text,omitempty"`
	LineStrRef    *int     `json:"line_strref,omitempty"`
	ConditionRefs []string `json:"condition_refs,omitempty"`
	Category      string   `json:"category,omitempty"`
}

type LayoutResult struct {
	ConversationID string               `json:"conversation_id"`
	NodeCount      int                  `json:"node_count"`
	Positions      map[string][]float64 `json:"positions"`
	Edges          []Edge               `json:"edges"`
	Nodes          map[string]NodeMeta  `json:"nodes,omitempty"`
}
