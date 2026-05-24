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
				SpeakerTag:     resolveSpeakerTag(entry.SpeakerTag, conv.Speakers),
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

func resolveSpeakerTag(tag string, speakers []dialogue.Speaker) string {
	if tag == "" {
		for _, s := range speakers {
			if s.ID == -2 {
				return "player"
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
