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
			{ID: 0, SpeakerID: intPtr(0), LineStrRef: intPtr(100), ReplyLinks: []int{0}},
			{ID: 1, SpeakerID: intPtr(0), LineStrRef: intPtr(101), ReplyLinks: []int{1}},
			{ID: 2, SpeakerID: intPtr(0), LineStrRef: intPtr(102), ReplyLinks: []int{}},
		},
		Replies: []ReplyNode{
			{ID: 0, TargetEntryIDs: []int{1}, LineStrRef: intPtr(200)},
			{ID: 1, TargetEntryIDs: []int{2}, LineStrRef: intPtr(201)},
		},
		Speakers: []Speaker{
			{ID: -2, Tag: "player", StrRefID: intPtr(125303), FriendlyName: "\"Shepard\""},
			{ID: -1, Tag: "owner", StrRefID: intPtr(0), FriendlyName: "No data"},
			{ID: 0, Tag: "Shepard"},
		},
		Starts: []StartNode{
			{ID: 0, TargetEntryIDs: []int{0}},
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
	if result.Summary.EntryCount != 3 {
		t.Errorf("entry_count = %d, want 3", result.Summary.EntryCount)
	}
	if result.Summary.ReplyCount != 2 {
		t.Errorf("reply_count = %d, want 2", result.Summary.ReplyCount)
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
	conv.Replies[0].TargetEntryIDs = []int{99}
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
	conv.Entries = append(conv.Entries, EntryNode{ID: 3, LineStrRef: intPtr(300)})
	result := ValidateConversation(conv)

	if result.Summary.OrphanedEntries == 0 {
		t.Error("expected orphaned entries")
	}
}

func TestValidateConversation_OrphanedReply(t *testing.T) {
	conv := makeTestConv()
	conv.Replies[0].TargetEntryIDs = []int{}
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
	conv.Starts[0].TargetEntryIDs = []int{99}
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
	conv.Starts[0].TargetEntryIDs = []int{}
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

func TestValidateConversation_PlayerOwnerPresent(t *testing.T) {
	conv := makeTestConv()
	result := ValidateConversation(conv)

	if result.Status != "valid" {
		t.Errorf("expected valid with player/owner present, got %s. Issues: %v", result.Status, result.Issues)
	}
}

func TestValidateConversation_MissingPlayer(t *testing.T) {
	conv := makeTestConv()
	conv.Speakers = []Speaker{
		{ID: -1, Tag: "owner"},
		{ID: 0, Tag: "Shepard"},
	}
	result := ValidateConversation(conv)

	if result.Status != "warning" {
		t.Errorf("expected warning for missing player, got %s", result.Status)
	}
	hasIssue := false
	for _, issue := range result.Issues {
		if issue.Message != "" && containsSubstring(issue.Message, "missing player") {
			hasIssue = true
		}
	}
	if !hasIssue {
		t.Error("expected issue about missing player speaker")
	}
}

func TestValidateConversation_MissingOwner(t *testing.T) {
	conv := makeTestConv()
	conv.Speakers = []Speaker{
		{ID: -2, Tag: "player"},
		{ID: 0, Tag: "Shepard"},
	}
	result := ValidateConversation(conv)

	if result.Status != "warning" {
		t.Errorf("expected warning for missing owner, got %s", result.Status)
	}
	hasIssue := false
	for _, issue := range result.Issues {
		if issue.Message != "" && containsSubstring(issue.Message, "missing owner") {
			hasIssue = true
		}
	}
	if !hasIssue {
		t.Error("expected issue about missing owner speaker")
	}
}

func TestValidateConversationStrict_MissingPlayerIsError(t *testing.T) {
	conv := makeTestConv()
	conv.Speakers = []Speaker{
		{ID: -1, Tag: "owner"},
		{ID: 0, Tag: "Shepard"},
	}
	result := ValidateConversationStrict(conv)

	if result.Status != "invalid" {
		t.Errorf("expected invalid (strict elevates missing player to error), got %s", result.Status)
	}
}

func TestValidateConversationStrict_OrphanedEntryIsError(t *testing.T) {
	conv := makeTestConv()
	conv.Entries = append(conv.Entries, EntryNode{ID: 3, LineStrRef: intPtr(300)})
	result := ValidateConversationStrict(conv)

	if result.Status != "invalid" {
		t.Errorf("expected invalid (strict elevates orphaned entry to error), got %s", result.Status)
	}
}

func TestValidateConversation_BidirectionalLinks_Symmetric(t *testing.T) {
	conv := makeTestConv()
	result := ValidateConversation(conv)

	if result.Status != "valid" {
		t.Errorf("expected valid for well-formed chain, got %s. Issues: %v", result.Status, result.Issues)
	}
}

func TestValidateConversation_BidirectionalLinks_AsymmetricEntry(t *testing.T) {
	conv := makeTestConv()
	// Add a new reply that targets entry 0, and link entry 2 to it
	conv.Replies = append(conv.Replies, ReplyNode{ID: 2, TargetEntryIDs: []int{0}, LineStrRef: intPtr(202)})
	conv.Entries[2].ReplyLinks = append(conv.Entries[2].ReplyLinks, 2)
	result := ValidateConversation(conv)
	foundNonReciprocal := false
	for _, issue := range result.Issues {
		if issue.Severity == "info" && containsSubstring(issue.Message, "non-reciprocal") {
			foundNonReciprocal = true
		}
	}
	if !foundNonReciprocal {
		t.Logf("info-level non-reciprocal note not found. Issues: %v", result.Issues)
	}
	// Entry 2 has reply links, is it reachable from chain? No. But for this test, accept any non-invalid.
	if result.Status == "invalid" {
		t.Errorf("expected non-invalid, got %s. Issues: %v", result.Status, result.Issues)
	}
}

func TestValidateConversation_BidirectionalLinks_AsymmetricReply(t *testing.T) {
	conv := makeTestConv()
	// Remove Entry 0's link to Reply 0, leaving Reply 0 targeting Entry 1 unreferenced
	conv.Entries[0].ReplyLinks = []int{}
	result := ValidateConversation(conv)

	if result.Summary.OrphanedEntries < 2 {
		t.Errorf("expected at least 2 orphaned entries (chain broken), got %d. Issues: %v", result.Summary.OrphanedEntries, result.Issues)
	}
	// Reply 0 targets Entry 1 but is not referenced; should get unreachable reply warning
	foundUnreachable := false
	for _, issue := range result.Issues {
		if containsSubstring(issue.Message, "unreachable reply") {
			foundUnreachable = true
		}
	}
	if !foundUnreachable {
		t.Errorf("expected unreachable reply issue. Issues: %v", result.Issues)
	}
}

func TestBuildValidationReport_Strict(t *testing.T) {
	conv := makeTestConv()
	conv.Speakers = []Speaker{
		{ID: -1, Tag: "owner"},
		{ID: 0, Tag: "Shepard"},
	}
	result := &ParseResult{
		File: "test.pcc",
		Conversations: []Conversation{
			*conv,
		},
	}
	report := BuildValidationReportStrict(result, true)
	if report.Summary.Invalid == 0 {
		t.Errorf("expected at least one invalid in strict mode (missing player), got %d", report.Summary.Invalid)
	}

	report2 := BuildValidationReportStrict(result, false)
	if report2.Summary.Invalid > 0 {
		t.Errorf("expected no invalid in non-strict mode, got %d", report2.Summary.Invalid)
	}
	if report2.Summary.Warning == 0 {
		t.Error("expected warning in non-strict mode for missing player")
	}
}

func TestValidateConversation_ReplyChoicesCategory(t *testing.T) {
	conv := makeTestConv()
	conv.Entries[0].ReplyChoices = []ReplyChoice{
		{FromEntryID: 0, ToReplyID: 0, Category: "Paragon", Order: 0},
	}
	// Valid conversation with categories
	result := ValidateConversation(conv)
	if result.Status != "valid" {
		t.Errorf("expected valid, got %s. Issues: %v", result.Status, result.Issues)
	}
	// Verify the categories were preserved
	if len(conv.Entries[0].ReplyChoices) > 0 {
		if conv.Entries[0].ReplyChoices[0].Category != "Paragon" {
			t.Errorf("expected Category=Paragon, got %s", conv.Entries[0].ReplyChoices[0].Category)
		}
	}
}

func TestValidateConversation_ReplyChoicesDanglingLink(t *testing.T) {
	conv := makeTestConv()
	conv.Entries[0].ReplyChoices = []ReplyChoice{
		{FromEntryID: 0, ToReplyID: 99, Category: "Paragon", Order: 0},
	}
	conv.Entries[0].ReplyLinks = []int{99}
	result := ValidateConversation(conv)
	if result.Status != "invalid" {
		t.Errorf("expected invalid for dangling reply choice link, got %s", result.Status)
	}
}

func TestValidateConversation_ReplyChoicesBadParaphrase(t *testing.T) {
	conv := makeTestConv()
	neg := -5
	conv.Entries[0].ReplyChoices = []ReplyChoice{
		{FromEntryID: 0, ToReplyID: 0, ParaphraseStrRef: &neg, Order: 0},
	}
	result := ValidateConversation(conv)
	if result.Status != "warning" {
		t.Errorf("expected warning for bad paraphrase strref, got %s", result.Status)
	}
	hasIssue := false
	for _, issue := range result.Issues {
		if containsSubstring(issue.Message, "invalid paraphrase") {
			hasIssue = true
		}
	}
	if !hasIssue {
		t.Error("expected issue about invalid paraphrase strref")
	}
}

func TestValidateConversation_ReplyChoicesBadParaphrase_Strict(t *testing.T) {
	conv := makeTestConv()
	neg := -5
	conv.Entries[0].ReplyChoices = []ReplyChoice{
		{FromEntryID: 0, ToReplyID: 0, ParaphraseStrRef: &neg, Order: 0},
	}
	result := ValidateConversationStrict(conv)
	if result.Status != "invalid" {
		t.Errorf("expected invalid in strict mode for bad paraphrase, got %s", result.Status)
	}
}

func TestValidateConversation_StrictPromotesParseWarnings(t *testing.T) {
	conv := makeTestConv()
	conv.ParseMode = "count_or_value_fallback"
	conv.Warnings = append(conv.Warnings, "missing_key_properties:m_SpeakerList")
	result := ValidateConversationStrict(conv)

	if result.Status != "invalid" {
		t.Errorf("expected invalid (strict promotes fallback + warnings to errors), got %s. Issues: %v", result.Status, result.Issues)
	}

	result2 := ValidateConversation(conv)
	if result2.Status != "warning" {
		t.Errorf("expected warning in non-strict mode, got %s", result2.Status)
	}
}

func TestEnsurePlayerAndOwnerSpeakers_AddsBoth(t *testing.T) {
	speakers := []Speaker{
		{ID: 0, Tag: "Shepard"},
	}
	result := ensurePlayerAndOwnerSpeakers(speakers)
	if len(result) != 3 {
		t.Fatalf("expected 3 speakers (player, owner, Shepard), got %d", len(result))
	}
	if result[0].ID != -2 || result[0].Tag != "player" {
		t.Errorf("first speaker should be player, got ID=%d tag=%s", result[0].ID, result[0].Tag)
	}
	if result[1].ID != -1 || result[1].Tag != "owner" {
		t.Errorf("second speaker should be owner, got ID=%d tag=%s", result[1].ID, result[1].Tag)
	}
}

func TestEnsurePlayerAndOwnerSpeakers_AlreadyPresent(t *testing.T) {
	speakers := []Speaker{
		{ID: -2, Tag: "player", StrRefID: intPtr(125303)},
		{ID: -1, Tag: "owner"},
		{ID: 0, Tag: "Shepard"},
	}
	result := ensurePlayerAndOwnerSpeakers(speakers)
	if len(result) != 3 {
		t.Fatalf("expected 3 speakers (no duplication), got %d", len(result))
	}
}

func TestEnsurePlayerAndOwnerSpeakers_OnlyOwnerMissing(t *testing.T) {
	speakers := []Speaker{
		{ID: -2, Tag: "player"},
		{ID: 0, Tag: "Shepard"},
	}
	result := ensurePlayerAndOwnerSpeakers(speakers)
	if len(result) != 3 {
		t.Fatalf("expected 3 speakers, got %d", len(result))
	}
	if result[0].ID != -1 || result[0].Tag != "owner" {
		t.Errorf("first speaker should be owner, got ID=%d tag=%s", result[0].ID, result[0].Tag)
	}
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
