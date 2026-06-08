package editor

import (
	"encoding/binary"
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

	spans, err := scanPropertySpans(original, names)
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

	spans, _ := scanPropertySpans(preserved, names)
	entrySpan := findSpan(spans, "m_EntryList", "EntryList")
	if entrySpan == nil {
		t.Fatal("m_EntryList not found in preserved output")
	}
}

func TestScanPropertySpans_WithPrefix(t *testing.T) {
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

	prefixed := make([]byte, 4+len(original))
	binary.LittleEndian.PutUint32(prefixed[0:4], 0)
	copy(prefixed[4:], original)

	spans, err := scanPropertySpans(prefixed, names)
	if err != nil {
		t.Fatalf("scan with prefix: %v", err)
	}

	entrySpan := findSpan(spans, "m_EntryList", "EntryList")
	if entrySpan == nil {
		t.Fatal("m_EntryList not found with prefix")
	}
	t.Logf("entryList: headerStart=%d valueStart=%d totalEnd=%d",
		entrySpan.headerStart, entrySpan.valueStart, entrySpan.totalEnd)

	modified := conv
	modified.Entries = append(modified.Entries, dialogue.EntryNode{
		ID: 1, SpeakerID: ptrInt(0), LineStrRef: ptrInt(200),
	})

	preserved, err := SerializeConversationPreserving(modified, prefixed, names)
	if err != nil {
		t.Fatalf("preserving with prefix: %v", err)
	}

	if len(preserved) <= len(prefixed) {
		t.Errorf("size didn't grow: was %d, now %d", len(prefixed), len(preserved))
	}

	reparsed, _ := scanPropertySpans(preserved, names)
	t.Logf("preserved spans: %d", len(reparsed))
	for _, s := range reparsed {
		t.Logf("  %s type=%s start=%d end=%d", s.name, s.propType, s.headerStart, s.totalEnd)
	}
}
