package dialogue

import (
	"fmt"

	"pcc-toolkit/core/internal/pcc"
)

func ParseConversations(summary *pcc.FileSummary, rawData []byte, mode string) *ParseResult {
	result := &ParseResult{
		File:        summary.Path,
		GameProfile: string(summary.GameProfile),
	}

	schema := GetSchemaForProfile(string(summary.GameProfile))

	for _, export := range summary.Exports {
		if export.ClassName != "BioConversation" {
			continue
		}
		if export.SerialSize < 8 || export.SerialOffset <= 0 {
			continue
		}

		if mode == "resilient" {
			conv, err := parseOneConversation(rawData, summary.Names, export, schema)
			if err != nil {
				result.Errors = append(result.Errors, ParseError{
					ID:          exportName(export),
					ExportIndex: export.Index,
					Error:       err.Error(),
				})
				continue
			}
			result.Conversations = append(result.Conversations, *conv)
		} else {
			conv, err := parseOneConversation(rawData, summary.Names, export, schema)
			if err != nil {
				result.Errors = append(result.Errors, ParseError{
					ID:          exportName(export),
					ExportIndex: export.Index,
					Error:       err.Error(),
				})
				continue
			}
			result.Conversations = append(result.Conversations, *conv)
		}
	}

	return result
}

func ParseConversation(summary *pcc.FileSummary, rawData []byte, convIndex int) (*Conversation, error) {
	if convIndex < 0 || convIndex >= len(summary.Exports) {
		return nil, fmt.Errorf("export index %d out of range", convIndex)
	}
	export := summary.Exports[convIndex]
	schema := GetSchemaForProfile(string(summary.GameProfile))
	return parseOneConversation(rawData, summary.Names, export, schema)
}

func exportName(export pcc.Export) string {
	if export.ObjectName != "" {
		return export.ObjectName
	}
	return fmt.Sprintf("Export_%d", export.Index)
}

func parseOneConversation(data []byte, names []string, export pcc.Export, schema ConversationListSchema) (*Conversation, error) {
	gameProfile := string(pcc.InferGameProfile(
		len(names), len(names), 
	))

	if export.SerialOffset < 0 || export.SerialSize < 8 {
		return nil, fmt.Errorf("export %d has invalid serial bounds", export.Index)
	}
	if export.SerialOffset+export.SerialSize > len(data) {
		return nil, fmt.Errorf("export %d serial data out of range", export.Index)
	}

	tags := pcc.ExtractBioConversationKeyProperties(data, names, export.SerialOffset, export.SerialSize)
	tagMap := map[string]pcc.PropertyTag{}
	for _, tag := range tags {
		tagMap[tag.Name] = tag
	}

	entryCount, replyCount, speakerCount := 0, 0, 0
	if t, ok := tagMap["EntryList"]; ok {
		entryCount = max(0, pcc.ReadArrayPropertyCount(data, t))
	}
	if t, ok := tagMap["ReplyList"]; ok {
		replyCount = max(0, pcc.ReadArrayPropertyCount(data, t))
	}
	if t, ok := tagMap["SpeakerList"]; ok {
		speakerCount = max(0, pcc.ReadArrayPropertyCount(data, t))
	}

	var entryValues, replyValues, speakerValues []int
	if t, ok := tagMap["EntryList"]; ok {
		entryValues = pcc.ReadArrayPropertyI32Values(data, t)
	}
	if t, ok := tagMap["ReplyList"]; ok {
		replyValues = pcc.ReadArrayPropertyI32Values(data, t)
	}
	if t, ok := tagMap["SpeakerList"]; ok {
		speakerValues = pcc.ReadArrayPropertyI32Values(data, t)
	}

	var entryRows, replyRows, speakerRows [][]int
	var usedStructHead, usedStructMatrix bool

	// Direct row reads (replaces buggy readRowArrays closure)
	if t, ok := tagMap["EntryList"]; ok {
		if m := pcc.ReadArrayPropertyStructI32Matrix(data, t); len(m) > 0 && len(m[0]) >= schema.EntryHeadI32 {
			usedStructMatrix = true
			entryRows = make([][]int, len(m))
			for i, row := range m {
				entryRows[i] = row[:schema.EntryHeadI32]
			}
		} else if s := pcc.ReadArrayPropertyStructHeadI32(data, t, schema.EntryHeadI32); len(s) > 0 {
			usedStructHead = true
			entryRows = s
		}
	}
	if t, ok := tagMap["ReplyList"]; ok {
		if m := pcc.ReadArrayPropertyStructI32Matrix(data, t); len(m) > 0 && len(m[0]) >= schema.ReplyHeadI32 {
			usedStructMatrix = true
			replyRows = make([][]int, len(m))
			for i, row := range m {
				replyRows[i] = row[:schema.ReplyHeadI32]
			}
		} else if s := pcc.ReadArrayPropertyStructHeadI32(data, t, schema.ReplyHeadI32); len(s) > 0 {
			usedStructHead = true
			replyRows = s
		}
	}
	if t, ok := tagMap["SpeakerList"]; ok {
		if m := pcc.ReadArrayPropertyStructI32Matrix(data, t); len(m) > 0 && len(m[0]) >= schema.SpeakerHeadI32 {
			usedStructMatrix = true
			speakerRows = make([][]int, len(m))
			for i, row := range m {
				speakerRows[i] = row[:schema.SpeakerHeadI32]
			}
		} else if s := pcc.ReadArrayPropertyStructHeadI32(data, t, schema.SpeakerHeadI32); len(s) > 0 {
			usedStructHead = true
			speakerRows = s
		}
	}

	var warnings []string
	if entryCount == 0 && replyCount == 0 && speakerCount == 0 {
		warnings = append(warnings, "empty_key_arrays")
	}

	rowMode := len(entryRows) > 0 && len(replyRows) > 0 && len(speakerRows) > 0
	rowPayloadRejected := false
	if rowMode && !rowPayloadIsCoherent(entryRows, replyRows, speakerRows) {
		rowMode = false
		rowPayloadRejected = true
		usedStructHead = false
		usedStructMatrix = false
	}

	semanticMode := false
	semanticNodes := trySemanticStructNodes(data, names, tagMap)
	if semanticNodes != nil {
		semanticMode = true
	}

	missingKeys := []string{}
	for _, key := range []string{"EntryList", "ReplyList", "SpeakerList"} {
		if _, ok := tagMap[key]; !ok {
			missingKeys = append(missingKeys, key)
		}
	}
	if len(missingKeys) > 0 && !semanticMode {
		warnings = append(warnings, fmt.Sprintf("missing_key_properties:%s", joinStrings(missingKeys, ",")))
	}

	var entryMatrix, replyMatrix, speakerMatrix [][]int
	if t, ok := tagMap["EntryList"]; ok {
		entryMatrix = pcc.ReadArrayPropertyStructI32Matrix(data, t)
	}
	if t, ok := tagMap["ReplyList"]; ok {
		replyMatrix = pcc.ReadArrayPropertyStructI32Matrix(data, t)
	}
	if t, ok := tagMap["SpeakerList"]; ok {
		speakerMatrix = pcc.ReadArrayPropertyStructI32Matrix(data, t)
	}

	for _, key := range []string{"EntryList", "ReplyList", "SpeakerList"} {
		tag, ok := tagMap[key]
		if !ok {
			continue
		}
		layout := pcc.AnalyzeArrayPropertyLayout(data, tag)
		if !layout.IsTightI32 && layout.Count > 0 && !semanticMode && !(usedStructHead || usedStructMatrix) {
			warnings = append(warnings, fmt.Sprintf(
				"non_tight_i32_array:%s:count=%d:bytes_per_item=%d:remainder=%d",
				key, layout.Count, ptrOr(layout.BytesPerItem, 0), layout.Remainder,
			))
		}
	}

	entryIDs := entryValues
	if len(entryValues) != entryCount {
		entryIDs = make([]int, entryCount)
		for i := range entryIDs {
			entryIDs[i] = i
		}
	}
	replyTargets := replyValues
	if len(replyValues) != replyCount {
		replyTargets = make([]int, replyCount)
		for i := range replyTargets {
			if i < entryCount {
				replyTargets[i] = i
			} else {
				replyTargets[i] = -1
			}
		}
	}
	speakerIDs := speakerValues
	if len(speakerValues) != speakerCount {
		speakerIDs = make([]int, speakerCount)
		for i := range speakerIDs {
			speakerIDs[i] = i
		}
	}

	var entries []EntryNode
	var replies []ReplyNode
	var speakers []Speaker

	if semanticMode && semanticNodes != nil {
		entries, replies, speakers = semanticNodes.entries, semanticNodes.replies, semanticNodes.speakers
	} else if rowMode {
		entries = buildEntriesRowMode(entryRows, entryMatrix, schema, names, usedStructMatrix)
	} else if len(entryRows) > 0 {
		entries = buildEntriesRowMode(entryRows, entryMatrix, schema, names, usedStructMatrix)
		warnings = append(warnings, "partial_row_payload_entries")
	} else {
		if rowPayloadRejected {
			warnings = append(warnings, "row_payload_incoherent_fallback_applied")
		} else if len(replyRows) > 0 || len(speakerRows) > 0 {
			warnings = append(warnings, "partial_row_payload_detected_fallback_applied")
		}
		entries = make([]EntryNode, len(entryIDs))
		for i, id := range entryIDs {
			entries[i] = EntryNode{ID: id, ReplyLinks: []int{}}
		}
	}

	if semanticMode {
	} else if rowMode {
		replies = buildRepliesRowMode(replyRows, replyMatrix, schema, usedStructMatrix)
	} else if len(replyRows) > 0 {
		replies = buildRepliesRowMode(replyRows, replyMatrix, schema, usedStructMatrix)
		warnings = append(warnings, "partial_row_payload_replies")
	} else {
		replies = make([]ReplyNode, len(replyTargets))
		for i, target := range replyTargets {
			t := target
			if target < 0 {
				t = 0
			}
			replies[i] = ReplyNode{ID: i, TargetEntryID: &t}
		}
	}

	if semanticMode {
	} else if rowMode {
		speakers = buildSpeakersRowMode(speakerRows, speakerMatrix, schema, names, usedStructMatrix)
	} else if len(speakerRows) > 0 {
		speakers = buildSpeakersRowMode(speakerRows, speakerMatrix, schema, names, usedStructMatrix)
		warnings = append(warnings, "partial_row_payload_speakers")
	} else {
		speakers = make([]Speaker, len(speakerIDs))
		for i, id := range speakerIDs {
			speakers[i] = Speaker{ID: id}
		}
	}

	linksByEntry := map[int][]int{}
	knownEntryIDs := map[int]bool{}
	for _, entry := range entries {
		knownEntryIDs[entry.ID] = true
		linksByEntry[entry.ID] = []int{}
	}
	for _, reply := range replies {
		if reply.TargetEntryID == nil {
			continue
		}
		tid := *reply.TargetEntryID
		if knownEntryIDs[tid] {
			linksByEntry[tid] = append(linksByEntry[tid], reply.ID)
		} else if len(entries) > 0 {
			warnings = append(warnings, fmt.Sprintf("reply_target_missing_entry:%d->%d", reply.ID, tid))
		}
	}
	for i := range entries {
		entries[i].ReplyLinks = linksByEntry[entries[i].ID]
	}

	var startValues []int
	if t, ok := tagMap["StartingList"]; ok {
		startValues = pcc.ReadArrayPropertyI32Values(data, t)
	}
	starts := make([]StartNode, len(startValues))
	for i, val := range startValues {
		starts[i] = StartNode{ID: i, TargetEntryID: &val}
	}

	parseMode := "count_or_value_fallback"
	if semanticMode {
		parseMode = "struct_property_semantic"
	} else if rowMode {
		if usedStructMatrix {
			parseMode = "row_payload_struct_matrix"
		} else if usedStructHead {
			parseMode = "row_payload_struct_head"
		} else {
			parseMode = "row_payload"
		}
	}

	return &Conversation{
		ID:          exportName(export),
		ExportIndex: export.Index,
		GameProfile: gameProfile,
		ParseMode:   parseMode,
		Entries:     entries,
		Replies:     replies,
		Speakers:    speakers,
		Starts:      starts,
		Warnings:    warnings,
	}, nil
}

func rowPayloadIsCoherent(entryRows, replyRows, speakerRows [][]int) bool {
	entryIDs := map[int]bool{}
	for _, row := range entryRows {
		if len(row) == 0 {
			continue
		}
		if entryIDs[row[0]] {
			return false
		}
		entryIDs[row[0]] = true
	}
	replyIDs := map[int]bool{}
	for _, row := range replyRows {
		if len(row) == 0 {
			continue
		}
		if replyIDs[row[0]] {
			return false
		}
		replyIDs[row[0]] = true
	}
	speakerIDs := map[int]bool{}
	for _, row := range speakerRows {
		if len(row) == 0 {
			continue
		}
		if speakerIDs[row[0]] {
			return false
		}
		speakerIDs[row[0]] = true
	}
	for _, row := range entryRows {
		if len(row) > 1 && row[1] >= 0 && !speakerIDs[row[1]] {
			return false
		}
	}
	return true
}

type semanticResult struct {
	entries  []EntryNode
	replies  []ReplyNode
	speakers []Speaker
}

func trySemanticStructNodes(data []byte, names []string, tagMap map[string]pcc.PropertyTag) *semanticResult {
	tag, ok := tagMap["EntryList"]
	if !ok {
		return nil
	}
	entryCount, entryPayload, entryPayloadSize := pcc.ReadArrayPropertyPayloadInfo(data, tag)
	if entryCount <= 0 || entryPayloadSize <= 0 {
		return nil
	}
	if !pcc.HasStructSignature(data, names, entryPayload, entryPayloadSize) {
		return nil
	}

	entryItems := pcc.ParseStructArrayItemsAsPropertyCollections(data, names, entryPayload, entryPayloadSize, entryCount)
	if len(entryItems) == 0 {
		return nil
	}

	var replyItems []map[string]pcc.ParsedProperty
	if replyTag, ok := tagMap["ReplyList"]; ok {
		replyCount, replyPayload, replyPayloadSize := pcc.ReadArrayPropertyPayloadInfo(data, replyTag)
		replyItems = pcc.ParseStructArrayItemsAsPropertyCollections(data, names, replyPayload, replyPayloadSize, max(0, replyCount))
	}

	entries := make([]EntryNode, len(entryItems))
	for idx, item := range entryItems {
		var speakerID *int
		if spk, ok := item["nSpeakerIndex"]; ok {
			if v, ok := spk.Value.(int); ok && v >= 0 {
				speakerID = &v
			}
		}
		var lineStrRef *int
		if sr, ok := item["srText"]; ok {
			if v, ok := sr.Value.(int); ok {
				lineStrRef = &v
			}
		}
		entries[idx] = EntryNode{
			ID:         idx,
			SpeakerID:  speakerID,
			LineStrRef: lineStrRef,
			ReplyLinks: []int{},
		}
	}

	replies := make([]ReplyNode, len(replyItems))
	for idx, item := range replyItems {
		var lineStrRef *int
		if sr, ok := item["srText"]; ok {
			if v, ok := sr.Value.(int); ok {
				lineStrRef = &v
			}
		}
		var targetEntryID *int
		if t, ok := item["nIndex"]; ok {
			if v, ok := t.Value.(int); ok {
				targetEntryID = &v
			}
		} else if t, ok := item["nEntryIndex"]; ok {
			if v, ok := t.Value.(int); ok {
				targetEntryID = &v
			}
		}
		var condRefs []string
		if cf, ok := item["nConditionalFunc"]; ok {
			if v, ok := cf.Value.(int); ok && v >= 0 {
				condRefs = append(condRefs, fmt.Sprintf("cond_func:%d", v))
			}
		}
		if cp, ok := item["nConditionalParam"]; ok {
			if v, ok := cp.Value.(int); ok && v != 0 {
				condRefs = append(condRefs, fmt.Sprintf("cond_param:%d", v))
			}
		}
		var category string
		if cat, ok := item["Category"]; ok {
			if s, ok := cat.Value.(string); ok {
				category = s
			}
		}
		replies[idx] = ReplyNode{
			ID:            idx,
			LineStrRef:    lineStrRef,
			TargetEntryID: targetEntryID,
			ConditionRefs: condRefs,
			Category:      category,
		}
	}

	var speakerItems []map[string]pcc.ParsedProperty
	if spkTag, ok := tagMap["SpeakerList"]; ok {
		spkCount, spkPayload, spkPayloadSize := pcc.ReadArrayPropertyPayloadInfo(data, spkTag)
		if spkCount > 0 && spkPayloadSize > 0 && pcc.HasStructSignature(data, names, spkPayload, spkPayloadSize) {
			speakerItems = pcc.ParseStructArrayItemsAsPropertyCollections(data, names, spkPayload, spkPayloadSize, spkCount)
		}
	}
	speakers := make([]Speaker, len(speakerItems))
	for idx, item := range speakerItems {
		var tag string
		if tp, ok := item["sSpeakerTag"]; ok {
			if s, ok := tp.Value.(string); ok {
				tag = s
			}
		}
		speakers[idx] = Speaker{ID: idx, Tag: tag}
	}

	return &semanticResult{entries, replies, speakers}
}

func buildEntriesRowMode(entryRows, entryMatrix [][]int, schema ConversationListSchema, names []string, usedStructMatrix bool) []EntryNode {
	entries := make([]EntryNode, len(entryRows))
	for idx, row := range entryRows {
		var speakerID *int
		if row[1] >= 0 {
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
		var targetEntryID *int
		if row[1] >= 0 {
			targetEntryID = &row[1]
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
			ID:            row[0],
			LineStrRef:    lineStrRef,
			TargetEntryID: targetEntryID,
			ConditionRefs: condRefs,
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

func joinStrings(ss []string, sep string) string {
	result := ""
	for i, s := range ss {
		if i > 0 {
			result += sep
		}
		result += s
	}
	return result
}

func ptrOr(p *int, defaultVal int) int {
	if p == nil {
		return defaultVal
	}
	return *p
}
