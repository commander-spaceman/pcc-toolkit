package dumper

import (
	"pcc-toolkit/core/internal/dialogue"
)

type DumpLine struct {
	ConversationID string `json:"conversation_id"`
	ExportIndex    int    `json:"export_index"`
	NodeType       string `json:"node_type"`
	NodeID         int    `json:"node_id"`
	SpeakerTag     string `json:"speaker_tag"`
	StrRef         int    `json:"strref"`
	LineText       string `json:"line_text,omitempty"`
	File           string `json:"file"`
}

type DumpLinesOutput struct {
	File        string     `json:"file"`
	GameProfile string     `json:"game_profile"`
	Lines       []DumpLine `json:"lines"`
	Total       int        `json:"total"`
}

func BuildDumpLines(result *dialogue.ParseResult) *DumpLinesOutput {
	output := &DumpLinesOutput{
		File:        result.File,
		GameProfile: result.GameProfile,
	}

	for _, conv := range result.Conversations {
		for _, entry := range conv.Entries {
			strref := 0
			if entry.LineStrRef != nil {
				strref = *entry.LineStrRef
			}
			if strref <= 0 {
				continue
			}
			output.Lines = append(output.Lines, DumpLine{
				ConversationID: conv.ID,
				ExportIndex:    conv.ExportIndex,
				NodeType:       "entry",
				NodeID:         entry.ID,
				SpeakerTag:     resolveEntrySpeakerTag(entry.SpeakerTag, entry.SpeakerID, conv.Speakers),
				StrRef:         strref,
				LineText:       entry.LineText,
				File:           result.File,
			})
		}

		for _, reply := range conv.Replies {
			strref := 0
			if reply.LineStrRef != nil {
				strref = *reply.LineStrRef
			}
			if strref <= 0 {
				continue
			}
			output.Lines = append(output.Lines, DumpLine{
				ConversationID: conv.ID,
				ExportIndex:    conv.ExportIndex,
				NodeType:       "reply",
				NodeID:         reply.ID,
				SpeakerTag:     "player",
				StrRef:         strref,
				LineText:       reply.LineText,
				File:           result.File,
			})
		}
	}

	output.Total = len(output.Lines)
	return output
}

func resolveEntrySpeakerTag(tag string, speakerID *int, speakers []dialogue.Speaker) string {
	if tag == "" {
		if speakerID != nil {
			for _, s := range speakers {
				if s.ID == *speakerID {
					return displaySpeakerTag(s)
				}
			}
			if *speakerID == -2 {
				return "player"
			}
			if *speakerID == -1 {
				return "owner"
			}
		}
		return "owner"
	}
	if tag == "player" {
		return tag
	}
	for _, s := range speakers {
		if s.Tag == tag && s.FriendlyName != "" {
			return s.FriendlyName
		}
	}
	return tag
}

func displaySpeakerTag(s dialogue.Speaker) string {
	if s.FriendlyName != "" {
		return s.FriendlyName
	}
	if s.Tag != "" {
		return s.Tag
	}
	if s.ID == -2 {
		return "player"
	}
	if s.ID == -1 {
		return "owner"
	}
	return "owner"
}
