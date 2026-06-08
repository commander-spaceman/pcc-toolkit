package editor

import (
	"testing"

	"pcc-toolkit/core/internal/dialogue"
)

func TestScanPropertySpans_FindsEntryList(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "TestConv",
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

	spans, err := scanPropertySpans(original, names, 0, len(original))
	if err != nil {
		t.Fatalf("scanPropertySpans: %v", err)
	}

	found := false
	for _, s := range spans {
		t.Logf("span: name=%s type=%s start=%d end=%d", s.name, s.propType, s.headerStart, s.totalEnd)
		if s.name == "m_EntryList" {
			found = true
		}
	}
	if !found {
		t.Error("m_EntryList span not found")
	}
}

func TestSerializeConversationPreserving_AddEntry(t *testing.T) {
	names := convNames()
	conv := dialogue.Conversation{
		ID:          "TestConv",
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
		t.Fatalf("SerializeConversationPreserving: %v", err)
	}

	if len(preserved) <= len(original) {
		t.Errorf("preserved size (%d) not larger than original (%d)", len(preserved), len(original))
	}

	spans, _ := scanPropertySpans(preserved, names, 0, len(preserved))
	entrySpan := findSpan(spans, "m_EntryList", "EntryList")
	if entrySpan == nil {
		t.Fatal("m_EntryList not found in preserved output")
	}

	_ = preserved
}
