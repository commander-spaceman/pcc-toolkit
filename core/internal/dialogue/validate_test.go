package dialogue

import (
	"testing"
)

func makeTestConv() *Conversation {
	return &Conversation{
		ID:          "Test_Conv",
		ExportIndex: 5,
		GameProfile: "me2_ot",
		ParseMode:   "struct_property_semantic",
		Entries: []EntryNode{
			{ID: 0, SpeakerID: intPtr(0), LineStrRef: intPtr(100), ReplyLinks: []int{0, 1}},
			{ID: 1, SpeakerID: intPtr(0), LineStrRef: intPtr(101), ReplyLinks: []int{2}},
		},
		Replies: []ReplyNode{
			{ID: 0, TargetEntryID: intPtr(1), LineStrRef: intPtr(200)},
			{ID: 1, TargetEntryID: intPtr(0), LineStrRef: intPtr(201)},
			{ID: 2, TargetEntryID: intPtr(0), LineStrRef: intPtr(202)},
		},
		Speakers: []Speaker{
			{ID: 0, Tag: "Shepard"},
		},
		Starts: []StartNode{
			{ID: 0, TargetEntryID: intPtr(0)},
		},
	}
}

func intPtr(i int) *int { return &i }

func TestValidateConversation_Valid(t *testing.T) {
	conv := makeTestConv()
	result := ValidateConversation(conv)

	if result.Status != "valid" {
		t.Errorf("expected valid, got %s. Issues: %v", result.Status, result.Issues)
	}
	if result.Summary.EntryCount != 2 {
		t.Errorf("entry_count = %d, want 2", result.Summary.EntryCount)
	}
	if result.Summary.ReplyCount != 3 {
		t.Errorf("reply_count = %d, want 3", result.Summary.ReplyCount)
	}
}

func TestValidateConversation_DanglingLink(t *testing.T) {
	conv := makeTestConv()
	conv.Entries[0].ReplyLinks = append(conv.Entries[0].ReplyLinks, 99)
	result := ValidateConversation(conv)

	if result.Status != "invalid" {
		t.Errorf("expected invalid, got %s", result.Status)
	}
	if result.Summary.DanglingLinks == 0 {
		t.Error("expected dangling links")
	}
}

func TestValidateConversation_BadTargetEntry(t *testing.T) {
	conv := makeTestConv()
	conv.Replies[0].TargetEntryID = intPtr(99)
	result := ValidateConversation(conv)

	if result.Status != "invalid" {
		t.Errorf("expected invalid, got %s", result.Status)
	}
}

func TestValidateConversation_MissingSpeaker(t *testing.T) {
	conv := makeTestConv()
	unknownSpeaker := 99
	conv.Entries[0].SpeakerID = &unknownSpeaker
	result := ValidateConversation(conv)

	hasWarning := false
	for _, issue := range result.Issues {
		if issue.Severity == "warning" {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("expected warning for missing speaker")
	}
}

func TestValidateConversation_UnreachableEntry(t *testing.T) {
	conv := makeTestConv()
	conv.Entries = append(conv.Entries, EntryNode{ID: 2, LineStrRef: intPtr(300)})
	result := ValidateConversation(conv)

	if result.Summary.OrphanedEntries == 0 {
		t.Error("expected orphaned entries")
	}
}

func TestValidateConversation_OrphanedReply(t *testing.T) {
	conv := makeTestConv()
	conv.Replies[0].TargetEntryID = nil
	result := ValidateConversation(conv)

	if result.Summary.OrphanedReplies == 0 {
		t.Error("expected orphaned replies")
	}
}

func TestValidateConversation_ZeroStrRef(t *testing.T) {
	conv := makeTestConv()
	zero := 0
	conv.Entries[0].LineStrRef = &zero
	result := ValidateConversation(conv)

	hasWarning := false
	for _, issue := range result.Issues {
		if issue.Severity == "warning" {
			hasWarning = true
			break
		}
	}
	if !hasWarning {
		t.Error("expected warning for zero strref")
	}
}

func TestValidateConversation_BadStartTarget(t *testing.T) {
	conv := makeTestConv()
	conv.Starts[0].TargetEntryID = intPtr(99)
	result := ValidateConversation(conv)

	if result.Status != "invalid" {
		t.Errorf("expected invalid, got %s", result.Status)
	}
}

func TestValidateConversation_FallbackMode(t *testing.T) {
	conv := makeTestConv()
	conv.ParseMode = "count_or_value_fallback"
	result := ValidateConversation(conv)

	if result.Status != "warning" {
		t.Errorf("expected warning, got %s", result.Status)
	}
}

func TestBuildValidationReport(t *testing.T) {
	result := &ParseResult{
		File: "test.pcc",
		Conversations: []Conversation{
			*makeTestConv(),
		},
	}
	report := BuildValidationReport(result)
	if report.Summary.Total != 1 {
		t.Errorf("total = %d, want 1", report.Summary.Total)
	}
	if report.Summary.Valid != 1 {
		t.Errorf("valid = %d, want 1", report.Summary.Valid)
	}
}

func TestValidateConversation_NullStartTarget(t *testing.T) {
	conv := makeTestConv()
	conv.Starts[0].TargetEntryID = nil
	result := ValidateConversation(conv)

	if result.Status != "invalid" {
		t.Errorf("expected invalid for start with nil target, got %s", result.Status)
	}
}

func TestItoa(t *testing.T) {
	tests := []struct {
		n    int
		want string
	}{
		{0, "0"},
		{1, "1"},
		{-1, "-1"},
		{42, "42"},
		{-99, "-99"},
		{12345, "12345"},
	}

	for _, tt := range tests {
		got := itoa(tt.n)
		if got != tt.want {
			t.Errorf("itoa(%d) = %q, want %q", tt.n, got, tt.want)
		}
	}
}
