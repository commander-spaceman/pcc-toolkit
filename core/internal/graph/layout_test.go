package graph

import (
	"testing"

	"pcc-toolkit/core/internal/dialogue"
)

func TestLayoutConversation_Empty(t *testing.T) {
	conv := &dialogue.Conversation{ID: "test"}
	result := LayoutConversation(conv, 240, 64, 80, 120)
	if result.NodeCount != 0 {
		t.Errorf("NodeCount = %d, want 0", result.NodeCount)
	}
	if len(result.Positions) != 0 {
		t.Errorf("got %d positions, want 0", len(result.Positions))
	}
}

func TestLayoutConversation_SimpleChain(t *testing.T) {
	conv := &dialogue.Conversation{
		ID: "Conv_Test",
		Starts: []dialogue.StartNode{
			{ID: 0, TargetEntryIDs: []int{0}},
		},
		Entries: []dialogue.EntryNode{
			{ID: 0, ReplyLinks: []int{0}},
		},
		Replies: []dialogue.ReplyNode{
			{ID: 0},
		},
	}

	result := LayoutConversation(conv, 240, 64, 80, 120)
	if result.NodeCount != 3 {
		t.Errorf("NodeCount = %d, want 3", result.NodeCount)
	}
	if len(result.Positions) != 3 {
		t.Errorf("got %d positions, want 3", len(result.Positions))
	}
	if len(result.Edges) != 2 {
		t.Errorf("got %d edges, want 2", len(result.Edges))
	}

	for _, key := range []string{"start:0", "entry:0", "reply:0"} {
		if _, ok := result.Positions[key]; !ok {
			t.Errorf("missing position for %s", key)
		}
	}
}

func TestLayoutConversation_MultipleReplies(t *testing.T) {
	conv := &dialogue.Conversation{
		ID: "Conv_Multi",
		Starts: []dialogue.StartNode{
			{ID: 0, TargetEntryIDs: []int{0}},
		},
		Entries: []dialogue.EntryNode{
			{ID: 0, ReplyLinks: []int{0, 1}},
		},
		Replies: []dialogue.ReplyNode{
			{ID: 0},
			{ID: 1},
		},
	}

	result := LayoutConversation(conv, 240, 64, 80, 120)
	if result.NodeCount != 4 {
		t.Errorf("NodeCount = %d, want 4", result.NodeCount)
	}
}

func TestLayoutConversation_ReplyLinksBack(t *testing.T) {
	conv := &dialogue.Conversation{
		ID: "Conv_Chain",
		Starts: []dialogue.StartNode{
			{ID: 0, TargetEntryIDs: []int{0}},
		},
		Entries: []dialogue.EntryNode{
			{ID: 0, ReplyLinks: []int{0}},
			{ID: 1, ReplyLinks: []int{}},
		},
		Replies: []dialogue.ReplyNode{
			{ID: 0, TargetEntryIDs: []int{1}},
		},
	}

	result := LayoutConversation(conv, 240, 64, 80, 120)
	if result.NodeCount != 4 {
		t.Errorf("NodeCount = %d, want 4", result.NodeCount)
	}
	if len(result.Edges) != 3 {
		t.Errorf("got %d edges, want 3 (start→e0, e0→r0, r0→e1)", len(result.Edges))
	}
}
