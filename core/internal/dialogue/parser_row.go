package dialogue

import "fmt"

func buildEntriesRowMode(entryRows, entryMatrix [][]int, schema ConversationListSchema, names []string, usedStructMatrix bool) []EntryNode {
	entries := make([]EntryNode, len(entryRows))
	for idx, row := range entryRows {
		var speakerID *int
		if len(row) > 1 {
			speakerID = &row[1]
		}
		var lineStrRef *int
		if row[2] >= 0 {
			lineStrRef = &row[2]
		}
		var listenerTag string
		if usedStructMatrix && schema.EntryListenerTagNameIdxCol > 0 &&
			idx < len(entryMatrix) && len(entryMatrix[idx]) > schema.EntryListenerTagNameIdxCol {
			nameIdx := entryMatrix[idx][schema.EntryListenerTagNameIdxCol]
			if nameIdx >= 0 && nameIdx < len(names) {
				listenerTag = names[nameIdx]
			}
		}
		entries[idx] = EntryNode{
			ID:          row[0],
			SpeakerID:   speakerID,
			LineStrRef:  lineStrRef,
			ListenerTag: listenerTag,
			ReplyLinks:  []int{},
		}
	}
	return entries
}

func buildRepliesRowMode(replyRows, replyMatrix [][]int, schema ConversationListSchema, usedStructMatrix bool) []ReplyNode {
	replies := make([]ReplyNode, len(replyRows))
	for idx, row := range replyRows {
		var lineStrRef *int
		if row[2] >= 0 {
			lineStrRef = &row[2]
		}
		var targetEntryIDs []int
		if row[1] >= 0 {
			targetEntryIDs = append(targetEntryIDs, row[1])
		}
		var condRefs []string
		if usedStructMatrix && schema.ReplyConditionStartCol > 0 &&
			idx < len(replyMatrix) && len(replyMatrix[idx]) > schema.ReplyConditionStartCol {
			for _, val := range replyMatrix[idx][schema.ReplyConditionStartCol:] {
				if val >= 0 {
					condRefs = append(condRefs, fmt.Sprintf("cond_i32:%d", val))
				}
			}
		}
		replies[idx] = ReplyNode{
			ID:             row[0],
			LineStrRef:     lineStrRef,
			TargetEntryIDs: targetEntryIDs,
			ConditionRefs:  condRefs,
		}
	}
	return replies
}

func buildSpeakersRowMode(speakerRows, speakerMatrix [][]int, schema ConversationListSchema, names []string, usedStructMatrix bool) []Speaker {
	speakers := make([]Speaker, len(speakerRows))
	for idx, row := range speakerRows {
		var tag string
		if row[1] >= 0 && row[1] < len(names) {
			tag = names[row[1]]
		}
		var displayName string
		if usedStructMatrix && schema.SpeakerDisplayNameStrRefCol > 0 &&
			idx < len(speakerMatrix) && len(speakerMatrix[idx]) > schema.SpeakerDisplayNameStrRefCol {
			strref := speakerMatrix[idx][schema.SpeakerDisplayNameStrRefCol]
			if strref >= 0 {
				displayName = fmt.Sprintf("strref:%d", strref)
			}
		}
		speakers[idx] = Speaker{
			ID:          row[0],
			Tag:         tag,
			DisplayName: displayName,
		}
	}
	return speakers
}
