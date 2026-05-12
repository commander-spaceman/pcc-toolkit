package dialogue

type EntryNode struct {
	ID          int    `json:"id"`
	SpeakerID   *int   `json:"speaker_id,omitempty"`
	SpeakerTag  string `json:"speaker_tag,omitempty"`
	ListenerTag string `json:"listener_tag,omitempty"`
	LineStrRef  *int   `json:"line_strref,omitempty"`
	LineText    string `json:"line_text,omitempty"`
	ReplyLinks  []int  `json:"reply_links"`
}

type ReplyNode struct {
	ID             int      `json:"id"`
	LineStrRef     *int     `json:"line_strref,omitempty"`
	LineText       string   `json:"line_text,omitempty"`
	TargetEntryIDs []int    `json:"target_entry_ids,omitempty"`
	ConditionRefs  []string `json:"condition_refs,omitempty"`
	Category       string   `json:"category,omitempty"`
}

type Speaker struct {
	ID          int    `json:"id"`
	Tag         string `json:"tag,omitempty"`
	DisplayName string `json:"display_name,omitempty"`
}

type StartNode struct {
	ID             int    `json:"id"`
	TargetEntryIDs []int  `json:"target_entry_ids,omitempty"`
	Label          string `json:"label,omitempty"`
}

type Conversation struct {
	ID              string       `json:"id"`
	ExportIndex     int          `json:"export_index"`
	GameProfile     string       `json:"game_profile"`
	ParseMode       string       `json:"parse_mode"`
	Entries         []EntryNode  `json:"entries"`
	Replies         []ReplyNode  `json:"replies"`
	Speakers        []Speaker    `json:"speakers"`
	Starts          []StartNode  `json:"starts"`
	Warnings        []string     `json:"warnings,omitempty"`
}

type ParseResult struct {
	File          string         `json:"file"`
	GameProfile   string         `json:"game_profile"`
	Conversations []Conversation `json:"conversations"`
	Errors        []ParseError   `json:"errors,omitempty"`
}

type ParseError struct {
	ID          string `json:"id"`
	ExportIndex int    `json:"export_index"`
	Error       string `json:"error"`
}
