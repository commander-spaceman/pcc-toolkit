package tlk

const TLKMagic = 0x006B6C54

type Header struct {
	Magic            int32 `json:"magic"`
	Version          int32 `json:"version"`
	MinVersion       int32 `json:"min_version"`
	MaleEntryCount   int32 `json:"male_entry_count"`
	FemaleEntryCount int32 `json:"female_entry_count"`
	TreeNodeCount    int32 `json:"tree_node_count"`
	DataLen          int32 `json:"data_len"`
}

type Tuple struct {
	StringID  int32 `json:"string_id"`
	BitOffset int32 `json:"bit_offset"`
}

type Node struct {
	LeftNodeID  int32 `json:"left_node_id"`
	RightNodeID int32 `json:"right_node_id"`
}

type File struct {
	Path          string          `json:"path"`
	Header        Header          `json:"header"`
	MaleEntries   map[int32]int32 `json:"-"`
	FemaleEntries map[int32]int32 `json:"-"`
	Nodes         []Node          `json:"-"`
	Bits          []byte          `json:"-"`
	TotalEntries  int             `json:"total_entries"`
}

type Entry struct {
	StringID int32  `json:"string_id"`
	Text     string `json:"text"`
	Source   string `json:"source,omitempty"`
}

type ResolveResult struct {
	StringID  int32  `json:"string_id"`
	Text      string `json:"text"`
	SourceTLK string `json:"source_tlk,omitempty"`
	Found     bool   `json:"found"`
}
