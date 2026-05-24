package dialogue

type ReplyChoice struct {
	FromEntryID      int    `json:"from_entry_id"`
	ToReplyID        int    `json:"to_reply_id"`
	Order            int    `json:"order"`
	Paraphrase       string `json:"paraphrase,omitempty"`
	ParaphraseStrRef *int   `json:"paraphrase_strref,omitempty"`
	ParaphraseText   string `json:"paraphrase_text,omitempty"`
	Category         string `json:"category,omitempty"`
}

type EntryNode struct {
	ID                   int           `json:"id"`
	SpeakerID            *int          `json:"speaker_id,omitempty"`
	SpeakerTag           string        `json:"speaker_tag,omitempty"`
	ListenerIndex        *int          `json:"listener_index,omitempty"`
	ListenerTag          string        `json:"listener_tag,omitempty"`
	LineStrRef           *int          `json:"line_strref,omitempty"`
	LineText             string        `json:"line_text,omitempty"`
	ReplyLinks           []int         `json:"reply_links"`
	ReplyChoices         []ReplyChoice `json:"reply_choices,omitempty"`
	ConditionalFunc      *int          `json:"conditional_func,omitempty"`
	ConditionalParam     *int          `json:"conditional_param,omitempty"`
	StateTransition      *int          `json:"state_transition,omitempty"`
	StateTransitionParam *int          `json:"state_transition_param,omitempty"`
	ScriptIndex          *int          `json:"script_index,omitempty"`
	FiresConditional     *bool         `json:"fires_conditional,omitempty"`
	ExportID             *int          `json:"export_id,omitempty"`
	Skippable            *bool         `json:"skippable,omitempty"`
	NonTextLine          *bool         `json:"non_text_line,omitempty"`
	Ambient              *bool         `json:"ambient,omitempty"`
	CameraIntimacy       *int          `json:"camera_intimacy,omitempty"`
	GUIStyle             string        `json:"gui_style,omitempty"`
}

type ReplyNode struct {
	ID                   int      `json:"id"`
	LineStrRef           *int     `json:"line_strref,omitempty"`
	LineText             string   `json:"line_text,omitempty"`
	TargetEntryIDs       []int    `json:"target_entry_ids,omitempty"`
	ConditionRefs        []string `json:"condition_refs,omitempty"`
	Category             string   `json:"category,omitempty"`
	ReplyType            string   `json:"reply_type,omitempty"`
	ConditionalFunc      *int     `json:"conditional_func,omitempty"`
	ConditionalParam     *int     `json:"conditional_param,omitempty"`
	StateTransition      *int     `json:"state_transition,omitempty"`
	StateTransitionParam *int     `json:"state_transition_param,omitempty"`
	ScriptIndex          *int     `json:"script_index,omitempty"`
	FiresConditional     *bool    `json:"fires_conditional,omitempty"`
	ExportID             *int     `json:"export_id,omitempty"`
	Unskippable          *bool    `json:"unskippable,omitempty"`
	NonTextLine          *bool    `json:"non_text_line,omitempty"`
	Ambient              *bool    `json:"ambient,omitempty"`
	CameraIntimacy       *int     `json:"camera_intimacy,omitempty"`
	GUIStyle             string   `json:"gui_style,omitempty"`
}

type Speaker struct {
	ID           int    `json:"id"`
	Tag          string `json:"tag,omitempty"`
	DisplayName  string `json:"display_name,omitempty"`
	StrRefID     *int   `json:"strref_id,omitempty"`
	FriendlyName string `json:"friendly_name,omitempty"`
}

type StartNode struct {
	ID             int    `json:"id"`
	TargetEntryIDs []int  `json:"target_entry_ids,omitempty"`
	Label          string `json:"label,omitempty"`
}

type Conversation struct {
	ID          string      `json:"id"`
	ExportIndex int         `json:"export_index"`
	GameProfile string      `json:"game_profile"`
	ParseMode   string      `json:"parse_mode"`
	Entries     []EntryNode `json:"entries"`
	Replies     []ReplyNode `json:"replies"`
	Speakers    []Speaker   `json:"speakers"`
	Starts      []StartNode `json:"starts"`
	Warnings    []string    `json:"warnings,omitempty"`
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
