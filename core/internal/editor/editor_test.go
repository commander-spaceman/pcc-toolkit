package editor

import (
	"testing"

	pcc "github.com/commander-spaceman/me2pcc"
	"pcc-toolkit/core/internal/dialogue"
)

func convNames() []string {
	return []string{
		"None",
		"IntProperty", "BoolProperty", "FloatProperty", "StrProperty",
		"StringRefProperty", "NameProperty", "ObjectProperty",
		"StructProperty", "ArrayProperty", "EnumProperty", "ByteProperty",
		"BioDialogEntryNode", "BioDialogReplyNode", "BioDialogSpeaker",
		"BioDialogReplyListDetails",
		"nIndex", "nSpeakerIndex", "srText", "nListenerIndex",
		"nConditionalFunc", "nConditionalParam",
		"nStateTransition", "nStateTransitionParam",
		"nScriptIndex", "nExportID", "nCameraIntimacy",
		"bFireConditional", "bSkippable", "bIsNonTextLine",
		"bAmbient", "eConvGUIStyle",
		"CONV_GUISTYLE_DEFAULT", "CONV_GUISTYLE_NEUTRAL",
		"ReplyListNew", "srParaphrase", "sParaphrase",
		"Category", "REPLY_CATEGORY_DEFAULT",
		"bUnskippable", "ReplyType", "REPLY_TYPE_DEFAULT",
		"EntryList", "nEntryIndex",
		"sSpeakerTag", "nDisplayNameStrRef",
		"EConvGUIStyles", "EReplyCategory", "EReplyTypes",
		"henchman", "owner", "player",
		"m_EntryList", "m_ReplyList", "m_SpeakerList",
		"m_StartingList", "m_ScriptList", "MatineeSequence",
		"BioDialogScript", "sScriptTag",
	}
}

func ptrInt(v int) *int    { return &v }
func ptrBool(v bool) *bool { return &v }

func TestSerializeConversation_RoundTrip(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "Global_TestConv",
		ExportIndex: 0,
		GameProfile: "me2_ot",
		ParseMode:   "struct_property_semantic",
		Entries: []dialogue.EntryNode{
			{ID: 0, SpeakerID: ptrInt(0), LineStrRef: ptrInt(100), ExportID: ptrInt(10)},
			{ID: 1, SpeakerID: ptrInt(1), LineStrRef: ptrInt(200), ExportID: ptrInt(11)},
		},
		Replies: []dialogue.ReplyNode{
			{ID: 0, LineStrRef: ptrInt(300), TargetEntryIDs: []int{1}},
			{ID: 1, LineStrRef: ptrInt(301), TargetEntryIDs: []int{1}},
		},
		Speakers: []dialogue.Speaker{
			{ID: 0, Tag: "henchman", StrRefID: ptrInt(500)},
			{ID: 1, Tag: "player", StrRefID: ptrInt(0)},
		},
		Starts: []dialogue.StartNode{
			{ID: 0, TargetEntryIDs: []int{0}},
		},
	}

	serial, _, err := SerializeConversation(conv, names)
	if err != nil {
		t.Fatalf("SerializeConversation: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(serial, names, 0, len(serial))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	entryList, ok := parsed["m_EntryList"]
	if !ok {
		t.Error("m_EntryList not found")
	} else {
		arr := entryList.Value.(map[string]interface{})
		count, _ := arr["count"].(int)
		if count != 2 {
			t.Errorf("entry count: want 2, got %d", count)
		}

		payloadOff, _ := arr["payload_offset"].(int)
		payloadSize, _ := arr["payload_size"].(int)
		items := pcc.ParseStructArrayItemsAsPropertyCollections(serial, names, payloadOff, payloadSize, 2)
		if len(items) == 2 {
			if v, ok := items[0]["nSpeakerIndex"]; ok {
				if v.Value.(int) != 0 {
					t.Errorf("entry 0 speaker: want 0, got %d", v.Value)
				}
			}
			if v, ok := items[1]["nSpeakerIndex"]; ok {
				if v.Value.(int) != 1 {
					t.Errorf("entry 1 speaker: want 1, got %d", v.Value)
				}
			}
		}
	}

	replyList, ok := parsed["m_ReplyList"]
	if !ok {
		t.Error("m_ReplyList not found")
	} else {
		arr := replyList.Value.(map[string]interface{})
		count, _ := arr["count"].(int)
		if count != 2 {
			t.Errorf("reply count: want 2, got %d", count)
		}
	}

	speakerList, ok := parsed["m_SpeakerList"]
	if !ok {
		t.Error("m_SpeakerList not found")
	} else {
		arr := speakerList.Value.(map[string]interface{})
		count, _ := arr["count"].(int)
		if count != 2 {
			t.Errorf("speaker count: want 2, got %d", count)
		}
	}

	startList, ok := parsed["m_StartingList"]
	if !ok {
		t.Error("m_StartingList not found")
	} else {
		arr := startList.Value.(map[string]interface{})
		count, _ := arr["count"].(int)
		if count != 1 {
			t.Errorf("start count: want 1, got %d", count)
		}
	}
}

func TestSerializeConversation_WithReplyChoices(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "Test_WithChoices",
		ExportIndex: 0,
		Entries: []dialogue.EntryNode{
			{
				ID:         0,
				SpeakerID:  ptrInt(0),
				LineStrRef: ptrInt(100),
				ExportID:   ptrInt(5),
				ReplyChoices: []dialogue.ReplyChoice{
					{ToReplyID: 0, ParaphraseStrRef: ptrInt(200), Category: "REPLY_CATEGORY_DEFAULT"},
				},
			},
		},
		Replies: []dialogue.ReplyNode{
			{ID: 0, LineStrRef: ptrInt(300), TargetEntryIDs: []int{0}},
		},
		Speakers: []dialogue.Speaker{
			{ID: 0, Tag: "henchman", StrRefID: ptrInt(0)},
		},
	}

	serial, _, err := SerializeConversation(conv, names)
	if err != nil {
		t.Fatalf("SerializeConversation: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(serial, names, 0, len(serial))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	entryList, ok := parsed["m_EntryList"]
	if !ok {
		t.Fatal("m_EntryList not found")
	}

	arr := entryList.Value.(map[string]interface{})
	payloadOff, _ := arr["payload_offset"].(int)
	payloadSize, _ := arr["payload_size"].(int)
	items := pcc.ParseStructArrayItemsAsPropertyCollections(serial, names, payloadOff, payloadSize, 1)
	if len(items) != 1 {
		t.Fatalf("items: want 1, got %d", len(items))
	}

	replyListNew, ok := items[0]["ReplyListNew"]
	if !ok {
		t.Error("ReplyListNew not found in entry")
	} else {
		arr := replyListNew.Value.(map[string]interface{})
		count, _ := arr["count"].(int)
		if count != 1 {
			t.Errorf("ReplyListNew count: want 1, got %d", count)
		}
	}
}

func TestSerializeConversation_Empty(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "EmptyConv",
		ExportIndex: 0,
	}

	serial, _, err := SerializeConversation(conv, names)
	if err != nil {
		t.Fatalf("SerializeConversation: %v", err)
	}

	if len(serial) < 8 {
		t.Errorf("serial too small: %d bytes", len(serial))
	}
}

func TestFirstError_NoErrors(t *testing.T) {
	result := dialogue.ParseResult{
		Conversations: []dialogue.Conversation{},
		Errors:        nil,
	}
	if err := firstError(result); err != nil {
		t.Errorf("expected nil, got %v", err)
	}
}

func TestFirstError_WithErrors(t *testing.T) {
	result := dialogue.ParseResult{
		Errors: []dialogue.ParseError{
			{ID: "conv1", ExportIndex: 0, Error: "parse failed"},
		},
	}
	if err := firstError(result); err == nil {
		t.Error("expected error for parse error")
	}
}
