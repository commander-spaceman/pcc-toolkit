package graph

import (
	"fmt"
	"sort"

	"pcc-toolkit/core/internal/dialogue"
)

func LayoutConversation(
	conv *dialogue.Conversation,
	nodeWidth, nodeHeight, xSpacing, ySpacing float64,
) *LayoutResult {
	starts := conv.Starts
	entries := conv.Entries
	replies := conv.Replies

	nStart := len(starts)
	nEntry := len(entries)
	nReply := len(replies)
	total := nStart + nEntry + nReply

	if total == 0 {
		return &LayoutResult{
			ConversationID: conv.ID,
			NodeCount:      0,
			Positions:      map[string][]float64{},
			Edges:          []Edge{},
		}
	}

	adj := make([][]int, total)
	for i := range adj {
		adj[i] = []int{}
	}

	var allEdges []Edge

	for _, s := range starts {
		for _, tid := range s.TargetEntryIDs {
			dst := nStart + tid
			if dst >= 0 && dst < total {
				adj[nStart+s.ID] = append(adj[nStart+s.ID], dst)
				allEdges = append(allEdges, Edge{
					From: NodeKey{Type: "start", ID: s.ID},
					To:   NodeKey{Type: "entry", ID: tid},
				})
			}
		}
	}

	for _, e := range entries {
		src := nStart + e.ID
		if len(e.ReplyChoices) > 0 {
			for _, rc := range e.ReplyChoices {
				rid := rc.ToReplyID
				dst := nStart + nEntry + rid
				if src >= 0 && src < total && dst >= 0 && dst < total {
					adj[src] = append(adj[src], dst)
					order := rc.Order
					allEdges = append(allEdges, Edge{
						From:           NodeKey{Type: "entry", ID: e.ID},
						To:             NodeKey{Type: "reply", ID: rid},
						Category:       rc.Category,
						ParaphraseText: rc.Paraphrase,
						InputIndex:     &order,
					})
				}
			}
		} else {
			for _, rid := range e.ReplyLinks {
				dst := nStart + nEntry + rid
				if src >= 0 && src < total && dst >= 0 && dst < total {
					adj[src] = append(adj[src], dst)
					allEdges = append(allEdges, Edge{
						From: NodeKey{Type: "entry", ID: e.ID},
						To:   NodeKey{Type: "reply", ID: rid},
					})
				}
			}
		}
	}

	for _, r := range replies {
		for _, tid := range r.TargetEntryIDs {
			src := nStart + nEntry + r.ID
			dst := nStart + tid
			if src >= 0 && src < total && dst >= 0 && dst < total {
				adj[src] = append(adj[src], dst)
				allEdges = append(allEdges, Edge{
					From: NodeKey{Type: "reply", ID: r.ID},
					To:   NodeKey{Type: "entry", ID: tid},
				})
			}
		}
	}

	layers := assignLayers(adj, total, nStart)
	positions := computePositions(layers, adj, nodeWidth, nodeHeight, xSpacing, ySpacing)

	result := &LayoutResult{
		ConversationID: conv.ID,
		NodeCount:      total,
		Positions:      map[string][]float64{},
		Edges:          allEdges,
		Nodes:          map[string]NodeMeta{},
	}

	for i := 0; i < nStart; i++ {
		key := fmt.Sprintf("start:%d", starts[i].ID)
		result.Positions[key] = positions[i]
		result.Nodes[key] = NodeMeta{
			Type: "start",
			ID:   starts[i].ID,
		}
	}
	for i, e := range entries {
		key := fmt.Sprintf("entry:%d", e.ID)
		result.Positions[key] = positions[nStart+i]
		meta := NodeMeta{
			Type:       "entry",
			ID:         e.ID,
			SpeakerTag: e.SpeakerTag,
			LineText:   truncateText(e.LineText, 120),
			LineStrRef: e.LineStrRef,
		}
		result.Nodes[key] = meta
	}
	for i, r := range replies {
		key := fmt.Sprintf("reply:%d", r.ID)
		result.Positions[key] = positions[nStart+nEntry+i]
		meta := NodeMeta{
			Type:          "reply",
			ID:            r.ID,
			LineText:      truncateText(r.LineText, 120),
			LineStrRef:    r.LineStrRef,
			ConditionRefs: r.ConditionRefs,
			Category:      r.Category,
		}
		result.Nodes[key] = meta
	}

	return result
}

func assignLayers(adj [][]int, total, nStart int) [][]int {
	inDegree := make([]int, total)
	for u := 0; u < total; u++ {
		for _, v := range adj[u] {
			inDegree[v]++
		}
	}

	layers := [][]int{}
	visited := make([]bool, total)
	queue := []int{}

	for u := 0; u < nStart; u++ {
		if inDegree[u] == 0 {
			queue = append(queue, u)
			visited[u] = true
		}
	}

	if len(queue) == 0 {
		for u := 0; u < total; u++ {
			if inDegree[u] == 0 {
				queue = append(queue, u)
				visited[u] = true
			}
		}
	}

	for len(queue) > 0 {
		layer := queue
		layers = append(layers, layer)
		nextQueue := []int{}

		for _, u := range layer {
			for _, v := range adj[u] {
				if visited[v] {
					continue
				}
				inDegree[v]--
				if inDegree[v] == 0 {
					nextQueue = append(nextQueue, v)
					visited[v] = true
				}
			}
		}
		queue = nextQueue
	}

	remaining := []int{}
	for u := 0; u < total; u++ {
		if !visited[u] {
			remaining = append(remaining, u)
		}
	}
	if len(remaining) > 0 {
		layers = append(layers, remaining)
	}

	return layers
}

func computePositions(
	layers [][]int,
	adj [][]int,
	nodeWidth, nodeHeight, xSpacing, ySpacing float64,
) [][]float64 {
	total := 0
	for _, layer := range layers {
		total += len(layer)
	}
	positions := make([][]float64, total)

	barycenterOrdering(layers, adj)

	for layerIdx, layer := range layers {
		layerSize := len(layer)
		if layerSize == 0 {
			continue
		}

		totalWidth := float64(layerSize)*(nodeWidth+xSpacing) - xSpacing
		startX := -totalWidth / 2.0

		for i, u := range layer {
			x := startX + float64(i)*(nodeWidth+xSpacing)
			y := float64(layerIdx) * (nodeHeight + ySpacing)
			positions[u] = []float64{x, y}
		}
	}

	return positions
}

func truncateText(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}

func barycenterOrdering(layers [][]int, adj [][]int) {
	prevPos := map[int]float64{}

	for layerIdx := 1; layerIdx < len(layers); layerIdx++ {
		layer := layers[layerIdx]

		type nodeBary struct {
			node    int
			bary    float64
			origIdx int
		}

		var items []nodeBary
		for origIdx, u := range layer {
			sum := 0.0
			count := 0.0

			for _, v := range adj[u] {
				if pos, ok := prevPos[v]; ok {
					sum += pos
					count++
				}
			}

			for v := 0; v < len(adj); v++ {
				for _, w := range adj[v] {
					if w == u {
						if pos, ok := prevPos[v]; ok {
							sum += pos
							count++
						}
					}
				}
			}

			bary := 0.0
			if count > 0 {
				bary = sum / count
			}

			items = append(items, nodeBary{u, bary, origIdx})
		}

		sort.Slice(items, func(i, j int) bool {
			if items[i].bary != items[j].bary {
				return items[i].bary < items[j].bary
			}
			return items[i].origIdx < items[j].origIdx
		})

		for newIdx, item := range items {
			layer[newIdx] = item.node
			prevPos[item.node] = float64(newIdx)
		}
	}

	for i := 0; i < len(layers[0]); i++ {
		prevPos[layers[0][i]] = float64(i)
	}
}
