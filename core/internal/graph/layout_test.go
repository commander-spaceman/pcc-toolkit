package graph

import (
	"fmt"
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

func TestLayoutConversation_ReplyChoices_Category(t *testing.T) {
	conv := &dialogue.Conversation{
		ID: "Conv_Category",
		Starts: []dialogue.StartNode{
			{ID: 0, TargetEntryIDs: []int{0}},
		},
		Entries: []dialogue.EntryNode{
			{
				ID: 0,
				ReplyChoices: []dialogue.ReplyChoice{
					{ToReplyID: 0, Order: 0, Category: "Paragon"},
					{ToReplyID: 1, Order: 1, Category: "Renegade"},
				},
			},
		},
		Replies: []dialogue.ReplyNode{
			{ID: 0},
			{ID: 1},
		},
	}

	result := LayoutConversation(conv, 240, 64, 80, 120)
	if len(result.Edges) != 3 {
		t.Fatalf("got %d edges, want 3", len(result.Edges))
	}

	categories := map[string]string{}
	for _, e := range result.Edges {
		if e.From.Type == "entry" && e.To.Type == "reply" {
			categories[fmtKey(e.To)] = e.Category
		}
	}

	if categories["reply:0"] != "Paragon" {
		t.Errorf("reply:0 category = %q, want Paragon", categories["reply:0"])
	}
	if categories["reply:1"] != "Renegade" {
		t.Errorf("reply:1 category = %q, want Renegade", categories["reply:1"])
	}
}

func TestLayoutConversation_ReplyChoices_ParaphraseText(t *testing.T) {
	conv := &dialogue.Conversation{
		ID: "Conv_Paraphrase",
		Starts: []dialogue.StartNode{
			{ID: 0, TargetEntryIDs: []int{0}},
		},
		Entries: []dialogue.EntryNode{
			{
				ID: 0,
				ReplyChoices: []dialogue.ReplyChoice{
					{ToReplyID: 0, Order: 0, ParaphraseText: "I should go."},
					{ToReplyID: 1, Order: 1, ParaphraseText: "Let me think."},
				},
			},
		},
		Replies: []dialogue.ReplyNode{
			{ID: 0},
			{ID: 1},
		},
	}

	result := LayoutConversation(conv, 240, 64, 80, 120)
	if len(result.Edges) != 3 {
		t.Fatalf("got %d edges, want 3", len(result.Edges))
	}

	paraphrases := map[string]string{}
	for _, e := range result.Edges {
		if e.From.Type == "entry" && e.To.Type == "reply" {
			paraphrases[fmtKey(e.To)] = e.ParaphraseText
		}
	}

	if paraphrases["reply:0"] != "I should go." {
		t.Errorf("reply:0 paraphrase_text = %q, want \"I should go.\"", paraphrases["reply:0"])
	}
	if paraphrases["reply:1"] != "Let me think." {
		t.Errorf("reply:1 paraphrase_text = %q, want \"Let me think.\"", paraphrases["reply:1"])
	}
}

func TestLayoutConversation_ReplyChoices_InputIndex(t *testing.T) {
	conv := &dialogue.Conversation{
		ID: "Conv_InputIndex",
		Starts: []dialogue.StartNode{
			{ID: 0, TargetEntryIDs: []int{0}},
		},
		Entries: []dialogue.EntryNode{
			{
				ID: 0,
				ReplyChoices: []dialogue.ReplyChoice{
					{ToReplyID: 0, Order: 2},
					{ToReplyID: 1, Order: 5},
				},
			},
		},
		Replies: []dialogue.ReplyNode{
			{ID: 0},
			{ID: 1},
		},
	}

	result := LayoutConversation(conv, 240, 64, 80, 120)
	if len(result.Edges) != 3 {
		t.Fatalf("got %d edges, want 3", len(result.Edges))
	}

	indices := map[string]int{}
	for _, e := range result.Edges {
		if e.From.Type == "entry" && e.To.Type == "reply" && e.InputIndex != nil {
			indices[fmtKey(e.To)] = *e.InputIndex
		}
	}

	if indices["reply:0"] != 2 {
		t.Errorf("reply:0 input_index = %d, want 2", indices["reply:0"])
	}
	if indices["reply:1"] != 5 {
		t.Errorf("reply:1 input_index = %d, want 5", indices["reply:1"])
	}
}

func TestLayoutConversation_ReplyLinks_BareEdges(t *testing.T) {
	conv := &dialogue.Conversation{
		ID: "Conv_Bare",
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
	if len(result.Edges) != 2 {
		t.Fatalf("got %d edges, want 2", len(result.Edges))
	}

	for _, e := range result.Edges {
		if e.From.Type == "entry" && e.To.Type == "reply" {
			if e.Category != "" {
				t.Errorf("expected empty category on bare reply-link edge, got %q", e.Category)
			}
			if e.ParaphraseText != "" {
				t.Errorf("expected empty paraphrase_text on bare reply-link edge, got %q", e.ParaphraseText)
			}
			if e.InputIndex != nil {
				t.Errorf("expected nil input_index on bare reply-link edge, got %d", *e.InputIndex)
			}
		}
	}
}

func fmtKey(nk NodeKey) string {
	return fmt.Sprintf("%s:%d", nk.Type, nk.ID)
}
