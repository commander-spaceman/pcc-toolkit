package graph

type NodeKey struct {
	Type string `json:"type"`
	ID   int    `json:"id"`
}

type Edge struct {
	From NodeKey `json:"from"`
	To   NodeKey `json:"to"`
}

type LayoutResult struct {
	ConversationID string               `json:"conversation_id"`
	NodeCount      int                  `json:"node_count"`
	Positions      map[string][]float64 `json:"positions"`
	Edges          []Edge               `json:"edges"`
}
