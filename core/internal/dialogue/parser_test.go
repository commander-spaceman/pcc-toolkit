package dialogue

import "testing"

func TestSemanticEntryReplyLinksAreValid(t *testing.T) {
	entries := []EntryNode{
		{ID: 0, ReplyLinks: []int{0, 1}},
		{ID: 1, ReplyLinks: []int{1}},
	}
	replies := []ReplyNode{
		{ID: 0, TargetEntryIDs: []int{10}},
		{ID: 1, TargetEntryIDs: []int{11}},
	}

	if !semanticEntryReplyLinksAreValid(entries, replies) {
		t.Fatal("expected semantic reply links to validate")
	}

	if replies[0].TargetEntryIDs[0] != 10 || replies[1].TargetEntryIDs[0] != 11 {
		t.Fatal("validation should not mutate semantic reply targets")
	}
}

func TestSemanticEntryReplyLinksAreValid_RejectsOutOfRangeReplyID(t *testing.T) {
	entries := []EntryNode{{ID: 0, ReplyLinks: []int{2}}}
	replies := []ReplyNode{{ID: 0}, {ID: 1}}

	if semanticEntryReplyLinksAreValid(entries, replies) {
		t.Fatal("expected invalid reply link to fail validation")
	}
}

func TestFillMissingSemanticReplyTargetsFromRows(t *testing.T) {
	replies := []ReplyNode{
		{ID: 0},
		{ID: 1, TargetEntryIDs: []int{99}},
		{ID: 2},
	}
	replyRows := [][]int{{0, 10}, {1, 11}, {2, -1}}

	fillMissingSemanticReplyTargetsFromRows(replies, replyRows, 20)

	if len(replies[0].TargetEntryIDs) != 1 || replies[0].TargetEntryIDs[0] != 10 {
		t.Fatalf("reply 0 targets = %v, want [10]", replies[0].TargetEntryIDs)
	}
	if len(replies[1].TargetEntryIDs) != 1 || replies[1].TargetEntryIDs[0] != 99 {
		t.Fatalf("reply 1 targets = %v, want existing [99]", replies[1].TargetEntryIDs)
	}
	if len(replies[2].TargetEntryIDs) != 0 {
		t.Fatalf("reply 2 targets = %v, want empty", replies[2].TargetEntryIDs)
	}
}

func TestFillMissingSemanticReplyTargetsFromRows_IgnoresOutOfRangeTarget(t *testing.T) {
	replies := []ReplyNode{{ID: 0}}
	replyRows := [][]int{{0, 87}}

	fillMissingSemanticReplyTargetsFromRows(replies, replyRows, 10)

	if len(replies[0].TargetEntryIDs) != 0 {
		t.Fatalf("reply 0 targets = %v, want empty", replies[0].TargetEntryIDs)
	}
}
