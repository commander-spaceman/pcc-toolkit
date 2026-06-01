package dumper

import (
	"testing"

	"pcc-toolkit/core/internal/dialogue"
)

func intPtr(i int) *int { return &i }

func TestBuildDumpLines_Basic(t *testing.T) {
	result := &dialogue.ParseResult{
		File:        "test.pcc",
		GameProfile: "me2_ot",
		Conversations: []dialogue.Conversation{
			{
				ID:          "TestConv_01",
				ExportIndex: 3,
				GameProfile: "me2_ot",
				ParseMode:   "struct_property_semantic",
				Entries: []dialogue.EntryNode{
					{ID: 0, SpeakerTag: "garrus", LineStrRef: intPtr(100), LineText: "Calibrating.", LineStatus: "resolved"},
					{ID: 1, SpeakerTag: "garrus", LineStrRef: intPtr(101), LineText: "Done.", LineStatus: "resolved"},
					{ID: 2, SpeakerTag: "", LineStrRef: intPtr(0), LineText: "", LineStatus: "no_line_text"}, // no strref, should be skipped
				},
				Replies: []dialogue.ReplyNode{
					{ID: 0, LineStrRef: intPtr(200), LineText: "Keep at it.", LineStatus: "resolved"},
					{ID: 1, LineStrRef: intPtr(0), LineStatus: "no_line_text"}, // no strref, should be skipped
				},
				Speakers: []dialogue.Speaker{
					{ID: 0, Tag: "garrus", FriendlyName: "Garrus"},
					{ID: -2, Tag: "player", FriendlyName: "Shepard"},
					{ID: -1, Tag: "owner", FriendlyName: "No data"},
				},
			},
		},
	}

	output := BuildDumpLines(result)

	if output.File != "test.pcc" {
		t.Errorf("file = %q, want test.pcc", output.File)
	}
	if output.Total != 3 {
		t.Errorf("total = %d, want 3", output.Total)
	}
	if len(output.Lines) != 3 {
		t.Fatalf("lines = %d, want 3", len(output.Lines))
	}

	// Entry 0
	if output.Lines[0].NodeType != "entry" {
		t.Errorf("lines[0].NodeType = %q, want entry", output.Lines[0].NodeType)
	}
	if output.Lines[0].StrRef != 100 {
		t.Errorf("lines[0].StrRef = %d, want 100", output.Lines[0].StrRef)
	}
	if output.Lines[0].LineStatus != "resolved" {
		t.Errorf("lines[0].LineStatus = %q, want resolved", output.Lines[0].LineStatus)
	}
	if output.Lines[0].SpeakerTag != "Garrus" {
		t.Errorf("lines[0].SpeakerTag = %q, want Garrus", output.Lines[0].SpeakerTag)
	}

	// Entry 1
	if output.Lines[1].NodeType != "entry" {
		t.Errorf("lines[1].NodeType = %q, want entry", output.Lines[1].NodeType)
	}
	if output.Lines[1].StrRef != 101 {
		t.Errorf("lines[1].StrRef = %d, want 101", output.Lines[1].StrRef)
	}

	// Reply 0
	if output.Lines[2].NodeType != "reply" {
		t.Errorf("lines[2].NodeType = %q, want reply", output.Lines[2].NodeType)
	}
	if output.Lines[2].SpeakerTag != "player" {
		t.Errorf("lines[2].SpeakerTag = %q, want player", output.Lines[2].SpeakerTag)
	}
}

func TestBuildDumpLines_EmptySpeakerTagFallback(t *testing.T) {
	result := &dialogue.ParseResult{
		File:        "test.pcc",
		GameProfile: "me2_ot",
		Conversations: []dialogue.Conversation{
			{
				ID:          "TestConv_02",
				ExportIndex: 1,
				GameProfile: "me2_ot",
				ParseMode:   "row_payload_struct_matrix",
				Entries: []dialogue.EntryNode{
					{ID: 0, SpeakerTag: "", LineStrRef: intPtr(50), LineText: "...", LineStatus: "resolved"},
				},
				Replies: []dialogue.ReplyNode{},
				Speakers: []dialogue.Speaker{
					{ID: -2, Tag: "player"},
				},
			},
		},
	}

	output := BuildDumpLines(result)

	if output.Lines[0].SpeakerTag != "owner" {
		t.Errorf("SpeakerTag = %q, want owner", output.Lines[0].SpeakerTag)
	}
}

func TestBuildDumpLines_EmptySpeakerTagUsesSpeakerID(t *testing.T) {
	playerID := -2
	result := &dialogue.ParseResult{
		File:        "test.pcc",
		GameProfile: "me2_ot",
		Conversations: []dialogue.Conversation{
			{
				ID:          "TestConv_02",
				ExportIndex: 1,
				GameProfile: "me2_ot",
				ParseMode:   "row_payload_struct_matrix",
				Entries: []dialogue.EntryNode{
					{ID: 0, SpeakerID: &playerID, LineStrRef: intPtr(50), LineText: "...", LineStatus: "resolved"},
				},
				Speakers: []dialogue.Speaker{
					{ID: -2, Tag: "player"},
				},
			},
		},
	}

	output := BuildDumpLines(result)

	if output.Lines[0].SpeakerTag != "player" {
		t.Errorf("SpeakerTag = %q, want player", output.Lines[0].SpeakerTag)
	}
}

func TestBuildDumpLines_ReplyAlwaysPlayer(t *testing.T) {
	result := &dialogue.ParseResult{
		File:        "test.pcc",
		GameProfile: "me2_ot",
		Conversations: []dialogue.Conversation{
			{
				ID:          "TestConv_03",
				ExportIndex: 2,
				GameProfile: "me2_ot",
				ParseMode:   "struct_property_semantic",
				Entries:     []dialogue.EntryNode{},
				Replies: []dialogue.ReplyNode{
					{ID: 0, LineStrRef: intPtr(300), LineText: "Let's go.", LineStatus: "resolved"},
				},
			},
		},
	}

	output := BuildDumpLines(result)

	if output.Lines[0].SpeakerTag != "player" {
		t.Errorf("SpeakerTag = %q, want player", output.Lines[0].SpeakerTag)
	}
}

func TestResolveSpeakerTag_FindsFriendlyName(t *testing.T) {
	speakers := []dialogue.Speaker{
		{ID: 0, Tag: "garrus", FriendlyName: "Garrus Vakarian"},
		{ID: 1, Tag: "tali", FriendlyName: "Tali'Zorah"},
	}

	got := resolveEntrySpeakerTag("garrus", nil, speakers)
	if got != "Garrus Vakarian" {
		t.Errorf("resolveSpeakerTag = %q, want Garrus Vakarian", got)
	}
}

func TestResolveSpeakerTag_ReturnsTagWhenNoMatch(t *testing.T) {
	speakers := []dialogue.Speaker{}
	got := resolveEntrySpeakerTag("unknown_tag", nil, speakers)
	if got != "unknown_tag" {
		t.Errorf("resolveSpeakerTag = %q, want unknown_tag", got)
	}
}
