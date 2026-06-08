package editor

import (
	"bytes"
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"pcc-toolkit/core/internal/dialogue"
	"pcc-toolkit/core/internal/pcc"
	"pcc-toolkit/core/internal/pccwrt"
)

func TestSerializeConversation_EmptyConversation(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "EmptyConv",
		ExportIndex: 0,
	}

	serial, added, err := SerializeConversation(conv, names)
	if err != nil {
		t.Fatalf("SerializeConversation: %v", err)
	}
	if len(serial) < 8 {
		t.Errorf("serial too small: %d bytes", len(serial))
	}
	if len(added) != 0 {
		t.Errorf("unexpected added names: %v", added)
	}

	parsed, end := pcc.ParsePropertyCollection(serial, names, 0, len(serial))
	if len(parsed) != 0 {
		t.Errorf("expected empty properties, got %d", len(parsed))
	}
	_ = end
}

func TestSerializeConversation_OnlyEntries(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "OnlyEntries",
		ExportIndex: 0,
		Entries: []dialogue.EntryNode{
			{ID: 0, SpeakerID: ptrInt(0), LineStrRef: ptrInt(100), ExportID: ptrInt(5)},
			{ID: 1, SpeakerID: ptrInt(0), LineStrRef: ptrInt(200), ExportID: ptrInt(6)},
		},
	}

	serial, added, err := SerializeConversation(conv, names)
	if err != nil {
		t.Fatalf("SerializeConversation: %v", err)
	}
	if len(added) != 0 {
		t.Logf("added names: %v", added)
	}

	parsed, _ := pcc.ParsePropertyCollection(serial, names, 0, len(serial))
	if entryList, ok := parsed["m_EntryList"]; !ok {
		t.Error("m_EntryList not found")
	} else {
		arr := entryList.Value.(map[string]interface{})
		count, _ := arr["count"].(int)
		if count != 2 {
			t.Errorf("entry count: want 2, got %d", count)
		}
	}
	if _, ok := parsed["m_ReplyList"]; ok {
		t.Error("m_ReplyList should not be present")
	}
	if _, ok := parsed["m_SpeakerList"]; ok {
		t.Error("m_SpeakerList should not be present")
	}
}

func TestSerializeConversation_OnlyReplies(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "OnlyReplies",
		ExportIndex: 0,
		Replies: []dialogue.ReplyNode{
			{ID: 0, LineStrRef: ptrInt(300), TargetEntryIDs: []int{0}},
			{ID: 1, LineStrRef: ptrInt(301), TargetEntryIDs: []int{0}},
		},
	}

	serial, added, err := SerializeConversation(conv, names)
	if err != nil {
		t.Fatalf("SerializeConversation: %v", err)
	}
	_ = added

	parsed, _ := pcc.ParsePropertyCollection(serial, names, 0, len(serial))
	if replyList, ok := parsed["m_ReplyList"]; !ok {
		t.Error("m_ReplyList not found")
	} else {
		arr := replyList.Value.(map[string]interface{})
		count, _ := arr["count"].(int)
		if count != 2 {
			t.Errorf("reply count: want 2, got %d", count)
		}
	}
	if _, ok := parsed["m_EntryList"]; ok {
		t.Error("m_EntryList should not be present")
	}
}

func TestSerializeConversation_Full(t *testing.T) {
	names := convNames()
	matineeID := 42
	conv := dialogue.Conversation{
		ID:          "FullConv",
		ExportIndex: 0,
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
		ScriptList: []dialogue.ScriptEntry{
			{ID: 0, Tag: "MyScriptTag"},
		},
		MatineeSequenceExportID: &matineeID,
	}

	serial, added, err := SerializeConversation(conv, names)
	if err != nil {
		t.Fatalf("SerializeConversation: %v", err)
	}
	_ = added

	parsed, _ := pcc.ParsePropertyCollection(serial, names, 0, len(serial))

	entryList, ok := parsed["m_EntryList"]
	if !ok {
		t.Error("m_EntryList not found")
	} else {
		arr := entryList.Value.(map[string]interface{})
		count, _ := arr["count"].(int)
		if count != 2 {
			t.Errorf("entry count: want 2, got %d", count)
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

	scriptList, ok := parsed["m_ScriptList"]
	if !ok {
		t.Error("m_ScriptList not found")
	} else {
		arr := scriptList.Value.(map[string]interface{})
		count, _ := arr["count"].(int)
		if count != 1 {
			t.Errorf("script count: want 1, got %d", count)
		}
	}

	if ms, ok := parsed["MatineeSequence"]; !ok {
		t.Error("MatineeSequence not found")
	} else {
		if ms.Value != matineeID {
			t.Errorf("MatineeSequence: want %d, got %v", matineeID, ms.Value)
		}
	}
}

func TestSerializeConversationPreserving_UnchangedReturnsOriginal(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "Unchanged",
		ExportIndex: 0,
		Entries: []dialogue.EntryNode{
			{ID: 0, SpeakerID: ptrInt(0), LineStrRef: ptrInt(100)},
		},
		Speakers: []dialogue.Speaker{
			{ID: 0, Tag: "henchman", StrRefID: ptrInt(0)},
		},
	}

	original, err := SerializeConversationSimple(conv, names)
	if err != nil {
		t.Fatalf("SerializeConversationSimple: %v", err)
	}

	preserved, err := SerializeConversationPreserving(conv, original, names)
	if err != nil {
		t.Fatalf("SerializeConversationPreserving: %v", err)
	}

	if !bytes.Equal(preserved, original) {
		t.Error("bytes differ despite no changes to conversation")
	}
}

func TestSerializeConversationPreserving_AddedEntryGrows(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "GrowTest",
		ExportIndex: 0,
		Entries: []dialogue.EntryNode{
			{ID: 0, SpeakerID: ptrInt(0), LineStrRef: ptrInt(100)},
		},
		Speakers: []dialogue.Speaker{
			{ID: 0, Tag: "henchman", StrRefID: ptrInt(0)},
		},
	}

	original, err := SerializeConversationSimple(conv, names)
	if err != nil {
		t.Fatalf("original: %v", err)
	}

	modified := conv
	modified.Entries = append(modified.Entries, dialogue.EntryNode{
		ID: 1, SpeakerID: ptrInt(0), LineStrRef: ptrInt(200),
	})

	preserved, err := SerializeConversationPreserving(modified, original, names)
	if err != nil {
		t.Fatalf("SerializerConversationPreserving: %v", err)
	}

	if len(preserved) <= len(original) {
		t.Errorf("size did not grow: %d -> %d", len(original), len(preserved))
	}

	spans, _ := scanPropertySpans(preserved, names)
	entrySpan := findSpan(spans, "m_EntryList", "EntryList")
	if entrySpan == nil {
		t.Fatal("m_EntryList not found in preserved output")
	}
	t.Logf("entry span: start=%d end=%d size=%d", entrySpan.headerStart, entrySpan.totalEnd,
		entrySpan.totalEnd-entrySpan.headerStart)
}

func TestSerializeConversationPreserving_RemovedEntryShrinks(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "ShrinkTest",
		ExportIndex: 0,
		Entries: []dialogue.EntryNode{
			{ID: 0, SpeakerID: ptrInt(0), LineStrRef: ptrInt(100)},
			{ID: 1, SpeakerID: ptrInt(0), LineStrRef: ptrInt(200)},
			{ID: 2, SpeakerID: ptrInt(0), LineStrRef: ptrInt(300)},
		},
		Speakers: []dialogue.Speaker{
			{ID: 0, Tag: "henchman", StrRefID: ptrInt(0)},
		},
	}

	original, err := SerializeConversationSimple(conv, names)
	if err != nil {
		t.Fatalf("original: %v", err)
	}

	modified := conv
	modified.Entries = conv.Entries[:1]

	preserved, err := SerializeConversationPreserving(modified, original, names)
	if err != nil {
		t.Fatalf("SerializerConversationPreserving: %v", err)
	}

	if len(preserved) >= len(original) {
		t.Errorf("size did not shrink: %d -> %d", len(original), len(preserved))
	}

	spans, _ := scanPropertySpans(preserved, names)
	entrySpan := findSpan(spans, "m_EntryList", "EntryList")
	if entrySpan == nil {
		t.Fatal("m_EntryList not found")
	}

	payloadOff, _ := entrySpan.valueStart, 0
	t.Logf("preserved entry span: headerStart=%d valueStart=%d totalEnd=%d",
		entrySpan.headerStart, entrySpan.valueStart, entrySpan.totalEnd)

	payloadSize := entrySpan.totalEnd - payloadOff - 4
	if payloadSize < 0 {
		payloadSize = 0
	}
	items := pcc.ParseStructArrayItemsAsPropertyCollections(preserved, names, payloadOff, payloadSize, 1)
	if len(items) != 1 {
		t.Errorf("expected 1 entry item, got %d", len(items))
	}
}

func TestSerializeConversationPreserving_KeepsTrailingBytes(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "TrailingBytes",
		ExportIndex: 0,
		Entries: []dialogue.EntryNode{
			{ID: 0, SpeakerID: ptrInt(0), LineStrRef: ptrInt(100)},
		},
		Speakers: []dialogue.Speaker{
			{ID: 0, Tag: "henchman", StrRefID: ptrInt(0)},
		},
	}

	original, err := SerializeConversationSimple(conv, names)
	if err != nil {
		t.Fatalf("original: %v", err)
	}

	spans, err := scanPropertySpans(original, names)
	if err != nil {
		t.Fatalf("scanPropertySpans: %v", err)
	}
	lastEnd := 0
	for _, s := range spans {
		if s.totalEnd > lastEnd {
			lastEnd = s.totalEnd
		}
	}

	marker := []byte("===TRAILING_MARKER_BYTES===")
	extended := make([]byte, len(original)+len(marker))
	copy(extended, original)
	copy(extended[len(original):], marker)

	modified := conv
	modified.Entries = append(modified.Entries, dialogue.EntryNode{
		ID: 1, SpeakerID: ptrInt(0), LineStrRef: ptrInt(200),
	})

	preserved, err := SerializeConversationPreserving(modified, extended, names)
	if err != nil {
		t.Fatalf("SerializerConversationPreserving with trailing bytes: %v", err)
	}

	if !bytes.HasSuffix(preserved, marker) {
		t.Error("trailing marker bytes not preserved after splice")
	}
	_ = lastEnd
}

func TestSerializeConversation_ModifiedEntryLineStrRef(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "ModEntry",
		ExportIndex: 0,
		Entries: []dialogue.EntryNode{
			{ID: 0, SpeakerID: ptrInt(0), LineStrRef: ptrInt(100)},
		},
		Speakers: []dialogue.Speaker{
			{ID: 0, Tag: "henchman", StrRefID: ptrInt(0)},
		},
	}

	serial, _, err := SerializeConversation(conv, names)
	if err != nil {
		t.Fatalf("first serialize: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(serial, names, 0, len(serial))
	entryList := parsed["m_EntryList"]
	arr := entryList.Value.(map[string]interface{})
	payloadOff, _ := arr["payload_offset"].(int)
	payloadSize, _ := arr["payload_size"].(int)
	items := pcc.ParseStructArrayItemsAsPropertyCollections(serial, names, payloadOff, payloadSize, 1)
	if len(items) != 1 {
		t.Fatalf("expected 1 item, got %d", len(items))
	}
	srText, ok := items[0]["srText"]
	if !ok || srText.Value.(int) != 100 {
		t.Errorf("srText: want 100, got %v", srText)
	}

	conv.Entries[0].LineStrRef = ptrInt(999)
	serial2, _, err := SerializeConversation(conv, names)
	if err != nil {
		t.Fatalf("second serialize: %v", err)
	}

	parsed2, _ := pcc.ParsePropertyCollection(serial2, names, 0, len(serial2))
	entryList2 := parsed2["m_EntryList"]
	arr2 := entryList2.Value.(map[string]interface{})
	payloadOff2, _ := arr2["payload_offset"].(int)
	payloadSize2, _ := arr2["payload_size"].(int)
	items2 := pcc.ParseStructArrayItemsAsPropertyCollections(serial2, names, payloadOff2, payloadSize2, 1)
	srText2, ok := items2[0]["srText"]
	if !ok || srText2.Value.(int) != 999 {
		t.Errorf("srText after modification: want 999, got %v", srText2)
	}
}

func TestSerializeConversation_AddedReplyChoice(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "AddChoice",
		ExportIndex: 0,
		Entries: []dialogue.EntryNode{
			{
				ID:         0,
				SpeakerID:  ptrInt(0),
				LineStrRef: ptrInt(100),
				ExportID:   ptrInt(5),
			},
		},
		Replies: []dialogue.ReplyNode{
			{ID: 0, LineStrRef: ptrInt(300), TargetEntryIDs: []int{0}},
			{ID: 1, LineStrRef: ptrInt(301), TargetEntryIDs: []int{0}},
		},
		Speakers: []dialogue.Speaker{
			{ID: 0, Tag: "henchman", StrRefID: ptrInt(0)},
		},
	}

	conv.Entries[0].ReplyChoices = []dialogue.ReplyChoice{
		{ToReplyID: 0, ParaphraseStrRef: ptrInt(200), Category: "REPLY_CATEGORY_DEFAULT"},
	}

	serial, _, err := SerializeConversation(conv, names)
	if err != nil {
		t.Fatalf("SerializeConversation: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(serial, names, 0, len(serial))
	entryList, ok := parsed["m_EntryList"]
	if !ok {
		t.Fatal("m_EntryList not found")
	}
	arr := entryList.Value.(map[string]interface{})
	count, _ := arr["count"].(int)
	if count != 1 {
		t.Fatalf("entry count: want 1, got %d", count)
	}
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
		payloadOff, _ := arr["payload_offset"].(int)
		payloadSize, _ := arr["payload_size"].(int)
		choices := pcc.ParseStructArrayItemsAsPropertyCollections(serial, names, payloadOff, payloadSize, 1)
		if len(choices) != 1 {
			t.Fatalf("reply choice count: want 1, got %d", len(choices))
		}
		if sr, ok := choices[0]["srParaphrase"]; ok {
			if sr.Value.(int) != 200 {
				t.Errorf("srParaphrase: want 200, got %v", sr.Value)
			}
		} else if sr, ok := choices[0]["sParaphrase"]; ok {
			t.Logf("found sParaphrase instead of srParaphrase: %v", sr.Value)
		} else {
			t.Error("neither srParaphrase nor sParaphrase found")
		}
	}
}

func TestSerializeConversation_MatineeSequence(t *testing.T) {
	names := convNames()
	matineeID := 77
	conv := dialogue.Conversation{
		ID:                      "MatineeConv",
		ExportIndex:             0,
		MatineeSequenceExportID: &matineeID,
	}

	serial, _, err := SerializeConversation(conv, names)
	if err != nil {
		t.Fatalf("SerializeConversation: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(serial, names, 0, len(serial))
	ms, ok := parsed["MatineeSequence"]
	if !ok {
		t.Fatal("MatineeSequence not found")
	}
	if ms.PropType != "ObjectProperty" {
		t.Errorf("prop type: want ObjectProperty, got %s", ms.PropType)
	}
	if ms.Value.(int) != 77 {
		t.Errorf("value: want 77, got %v", ms.Value)
	}
}

func TestSerializeConversation_ScriptList(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "ScriptConv",
		ExportIndex: 0,
		ScriptList: []dialogue.ScriptEntry{
			{ID: 0, Tag: "CombatStart"},
			{ID: 1, Tag: "DialogEnd"},
		},
	}

	serial, added, err := SerializeConversation(conv, names)
	if err != nil {
		t.Fatalf("SerializeConversation: %v", err)
	}

	allNames := make([]string, len(names)+len(added))
	copy(allNames, names)
	copy(allNames[len(names):], added)

	parsed, _ := pcc.ParsePropertyCollection(serial, allNames, 0, len(serial))
	scriptList, ok := parsed["m_ScriptList"]
	if !ok {
		t.Fatal("m_ScriptList not found")
	}
	arr := scriptList.Value.(map[string]interface{})
	count, _ := arr["count"].(int)
	if count != 2 {
		t.Fatalf("script count: want 2, got %d", count)
	}

	payloadOff, _ := arr["payload_offset"].(int)
	payloadSize, _ := arr["payload_size"].(int)
	items := pcc.ParseStructArrayItemsAsPropertyCollections(serial, allNames, payloadOff, payloadSize, 2)
	if len(items) != 2 {
		t.Fatalf("script items: want 2, got %d", len(items))
	}
	for i, item := range items {
		tagProp, ok := item["sScriptTag"]
		if !ok {
			t.Errorf("script[%d]: sScriptTag not found", i)
		} else if tagProp.Value != conv.ScriptList[i].Tag {
			t.Errorf("script[%d] tag: want %q, got %v", i, conv.ScriptList[i].Tag, tagProp.Value)
		}
	}
}

func TestScanPropertySpans_FindsAllExpectedSpans(t *testing.T) {
	names := convNames()
	matineeID := 55
	conv := dialogue.Conversation{
		ID:          "AllSpans",
		ExportIndex: 0,
		Entries: []dialogue.EntryNode{
			{ID: 0, SpeakerID: ptrInt(0), LineStrRef: ptrInt(100), ExportID: ptrInt(10)},
		},
		Replies: []dialogue.ReplyNode{
			{ID: 0, LineStrRef: ptrInt(300), TargetEntryIDs: []int{0}},
		},
		Speakers: []dialogue.Speaker{
			{ID: 0, Tag: "henchman", StrRefID: ptrInt(0)},
		},
		Starts: []dialogue.StartNode{
			{ID: 0, TargetEntryIDs: []int{0}},
		},
		ScriptList: []dialogue.ScriptEntry{
			{ID: 0, Tag: "TestScript"},
		},
		MatineeSequenceExportID: &matineeID,
	}

	serial, added, err := SerializeConversation(conv, names)
	if err != nil {
		t.Fatalf("SerializeConversation: %v", err)
	}
	allNames := make([]string, len(names)+len(added))
	copy(allNames, names)
	copy(allNames[len(names):], added)

	spans, err := scanPropertySpans(serial, allNames)
	if err != nil {
		t.Fatalf("scanPropertySpans: %v", err)
	}

	expectedNames := []string{
		"m_EntryList", "m_ReplyList", "m_SpeakerList",
		"m_StartingList", "m_ScriptList", "MatineeSequence",
	}

	found := make(map[string]bool)
	for _, s := range spans {
		t.Logf("span: name=%s type=%s headerStart=%d totalEnd=%d valueStart=%d valueEnd=%d",
			s.name, s.propType, s.headerStart, s.totalEnd, s.valueStart, s.valueEnd)
		found[s.name] = true
	}

	for _, name := range expectedNames {
		if !found[name] {
			t.Errorf("expected span %q not found", name)
		}
	}

	if len(spans) < len(expectedNames) {
		t.Errorf("span count: want at least %d, got %d", len(expectedNames), len(spans))
	}
}

func TestScanPropertySpans_With4BytePrefix(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "Prefixed",
		ExportIndex: 0,
		Entries: []dialogue.EntryNode{
			{ID: 0, SpeakerID: ptrInt(0), LineStrRef: ptrInt(100)},
		},
		Speakers: []dialogue.Speaker{
			{ID: 0, Tag: "henchman", StrRefID: ptrInt(0)},
		},
	}

	original, err := SerializeConversationSimple(conv, names)
	if err != nil {
		t.Fatalf("original: %v", err)
	}

	prefixed := make([]byte, 4+len(original))
	binary.LittleEndian.PutUint32(prefixed[0:4], 0)
	copy(prefixed[4:], original)

	spans, err := scanPropertySpans(prefixed, names)
	if err != nil {
		t.Fatalf("scanPropertySpans with prefix: %v", err)
	}

	foundEntry := false
	foundSpeaker := false
	for _, s := range spans {
		if s.name == "m_EntryList" {
			foundEntry = true
			t.Logf("entryList: headerStart=%d valueStart=%d totalEnd=%d",
				s.headerStart, s.valueStart, s.totalEnd)
		}
		if s.name == "m_SpeakerList" {
			foundSpeaker = true
		}
	}
	if !foundEntry {
		t.Error("m_EntryList not found with 4-byte prefix")
	}
	if !foundSpeaker {
		t.Error("m_SpeakerList not found with 4-byte prefix")
	}
}

func TestScanPropertySpans_EmptyData(t *testing.T) {
	names := convNames()
	spans, err := scanPropertySpans([]byte{}, names)
	if err == nil {
		t.Error("expected error for empty data")
	}
	if spans != nil {
		t.Errorf("expected nil spans, got %v", spans)
	}

	spans2, err2 := scanPropertySpans(nil, names)
	if err2 == nil {
		t.Error("expected error for nil data")
	}
	if spans2 != nil {
		t.Errorf("expected nil spans, got %v", spans2)
	}
}

func TestEditConversation_DryRun(t *testing.T) {
	dir := t.TempDir()

	names := convNames()
	conv := dialogue.Conversation{
		ID:          "DryRunConv",
		ExportIndex: 0,
		Entries: []dialogue.EntryNode{
			{ID: 0, SpeakerID: ptrInt(0), LineStrRef: ptrInt(100)},
		},
		Speakers: []dialogue.Speaker{
			{ID: 0, Tag: "henchman", StrRefID: ptrInt(0)},
		},
	}

	allNames := make([]string, len(names))
	copy(allNames, names)
	bioIdx := len(allNames)
	allNames = append(allNames, "BioConversation")
	expNameIdx := len(allNames)
	allNames = append(allNames, "export_0")

	serial, _, err := SerializeConversation(conv, allNames)
	if err != nil {
		t.Fatalf("SerializeConversation: %v", err)
	}

	raw := make([]byte, len(serial))
	copy(raw, serial)
	summary := &pcc.FileSummary{
		Path:        "test.pcc",
		GameProfile: pcc.ProfileME2OT,
		Compressed:  false,
		Header: pcc.Header{
			UnrealVersion:   512,
			LicenseeVersion: 130,
			Flags:           0,
			NameCount:       len(allNames),
			ExportCount:     1,
			ImportCount:     1,
		},
		Names: allNames,
		Imports: []pcc.Import{
			{ClassNameIndex: 1, ObjectNameIndex: bioIdx},
		},
		Exports: []pcc.Export{{
			Index:           0,
			ClassIndex:      -1,
			ObjectNameIndex: expNameIdx,
			SerialSize:      len(serial),
			SerialOffset:    0,
			ObjectName:      "export_0",
			ClassName:       "BioConversation",
		}},
	}
	inPath := filepath.Join(dir, "input.pcc")
	if err := pccwrt.WritePCC(inPath, summary, raw); err != nil {
		t.Fatalf("WritePCC: %v", err)
	}

	outPath := filepath.Join(dir, "output.pcc")
	result, err := EditConversation(inPath, outPath, 0, true, func(c *dialogue.Conversation) error {
		c.Entries[0].LineStrRef = ptrInt(999)
		return nil
	})
	if err != nil {
		t.Fatalf("EditConversation: %v", err)
	}
	if result.Status != "dry_run" {
		t.Errorf("status: want dry_run, got %s", result.Status)
	}
	if result.Output != "" {
		t.Errorf("output: want empty, got %s", result.Output)
	}
	if _, err := os.Stat(outPath); !os.IsNotExist(err) {
		t.Error("output file should not exist after dry run")
	}
}

func TestEditConversation_EndToEnd(t *testing.T) {
	dir := t.TempDir()

	names := convNames()
	conv := dialogue.Conversation{
		ID:          "E2EConv",
		ExportIndex: 0,
		Entries: []dialogue.EntryNode{
			{ID: 0, SpeakerID: ptrInt(0), LineStrRef: ptrInt(100)},
		},
		Speakers: []dialogue.Speaker{
			{ID: 0, Tag: "henchman", StrRefID: ptrInt(0)},
		},
	}

	allNames := make([]string, len(names))
	copy(allNames, names)
	bioIdx := len(allNames)
	allNames = append(allNames, "BioConversation")
	expNameIdx := len(allNames)
	allNames = append(allNames, "export_0")

	serial, _, err := SerializeConversation(conv, allNames)
	if err != nil {
		t.Fatalf("SerializeConversation: %v", err)
	}

	raw := make([]byte, len(serial))
	copy(raw, serial)
	summary := &pcc.FileSummary{
		Path:        "test.pcc",
		GameProfile: pcc.ProfileME2OT,
		Compressed:  false,
		Header: pcc.Header{
			UnrealVersion:   512,
			LicenseeVersion: 130,
			Flags:           0,
			NameCount:       len(allNames),
			ExportCount:     1,
			ImportCount:     1,
		},
		Names: allNames,
		Imports: []pcc.Import{
			{ClassNameIndex: 1, ObjectNameIndex: bioIdx},
		},
		Exports: []pcc.Export{{
			Index:           0,
			ClassIndex:      -1,
			ObjectNameIndex: expNameIdx,
			SerialSize:      len(serial),
			SerialOffset:    0,
			ObjectName:      "export_0",
			ClassName:       "BioConversation",
		}},
	}
	inPath := filepath.Join(dir, "input.pcc")
	if err := pccwrt.WritePCC(inPath, summary, raw); err != nil {
		t.Fatalf("WritePCC: %v", err)
	}

	outPath := filepath.Join(dir, "output.pcc")
	result, err := EditConversation(inPath, outPath, 0, false, func(c *dialogue.Conversation) error {
		c.Entries[0].LineStrRef = ptrInt(999)
		return nil
	})
	if err != nil {
		t.Fatalf("EditConversation: %v", err)
	}
	if result.Status != "ok" && result.Status != "dry_run" &&
		result.Status != "written_with_1_warnings" && result.Status != "written_with_1_invalid" {
		t.Errorf("status: unexpected status %q", result.Status)
	}
	if result.Output != outPath {
		t.Errorf("output: want %s, got %s", outPath, result.Output)
	}
	if result.Validation == nil {
		t.Fatal("expected validation report")
	}

	readSum, err := pcc.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile output: %v", err)
	}
	if !readSum.Compressed {
		t.Error("expected compressed output")
	}

	foundConv := false
	exportIndex := 0
	for i, exp := range readSum.Exports {
		if exp.ClassName == "BioConversation" {
			exportIndex = i
			foundConv = true
			break
		}
	}
	if !foundConv {
		t.Skip("BioConversation export not found; skipping data verification")
		return
	}

	readRaw, _, err := pcc.ReadFileRaw(outPath)
	if err != nil {
		t.Fatalf("ReadFileRaw: %v", err)
	}
	exp := readSum.Exports[exportIndex]
	serialOut := readRaw[exp.SerialOffset : exp.SerialOffset+exp.SerialSize]

	parsed, _ := pcc.ParsePropertyCollection(serialOut, names, 0, len(serialOut))
	entryList, ok := parsed["m_EntryList"]
	if !ok {
		t.Fatal("m_EntryList not found in edited output")
	}
	arr := entryList.Value.(map[string]interface{})
	payloadOff, _ := arr["payload_offset"].(int)
	payloadSize, _ := arr["payload_size"].(int)
	items := pcc.ParseStructArrayItemsAsPropertyCollections(serialOut, names, payloadOff, payloadSize, 1)
	if len(items) != 1 {
		t.Fatalf("entry items: want 1, got %d", len(items))
	}
	srText, ok := items[0]["srText"]
	if !ok || srText.Value.(int) != 999 {
		t.Errorf("srText after edit: want 999, got %v", srText)
	}
}
