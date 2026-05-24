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
	gameProfile := string(summary.GameProfile)

	for _, export := range summary.Exports {
		if export.ClassName != "BioConversation" {
			continue
		}
		if export.SerialSize < 8 || export.SerialOffset < 0 ||
			export.SerialOffset+export.SerialSize > len(rawData) {
			continue
		}

		if mode == "resilient" {
			exportData := rawData[export.SerialOffset : export.SerialOffset+export.SerialSize]
			conv, err := parseOneConversation(exportData, summary.Names, gameProfile, export, schema)
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
			exportData := rawData[export.SerialOffset : export.SerialOffset+export.SerialSize]
			conv, err := parseOneConversation(exportData, summary.Names, gameProfile, export, schema)
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
	if export.SerialSize < 8 || export.SerialOffset < 0 ||
		export.SerialOffset+export.SerialSize > len(rawData) {
		return nil, fmt.Errorf("export %d has invalid serial bounds", export.Index)
	}
	schema := GetSchemaForProfile(string(summary.GameProfile))
	exportData := rawData[export.SerialOffset : export.SerialOffset+export.SerialSize]
	return parseOneConversation(exportData, summary.Names, string(summary.GameProfile), export, schema)
}

func exportName(export pcc.Export) string {
	if export.ObjectName != "" {
		return export.ObjectName
	}
	return fmt.Sprintf("Export_%d", export.Index)
}

func parseOneConversation(data []byte, names []string, gameProfile string, export pcc.Export, schema ConversationListSchema) (*Conversation, error) {

	if len(data) < 8 {
		return nil, fmt.Errorf("export %d has invalid serial bounds", export.Index)
	}

	tags := pcc.ExtractBioConversationKeyProperties(data, names, 0, len(data))
	tagMap := map[string]pcc.PropertyTag{}
	for _, tag := range tags {
		tagMap[tag.Name] = tag
	}

	lookupTag := func(keys ...string) (pcc.PropertyTag, bool) {
		for _, key := range keys {
			if t, ok := tagMap[key]; ok {
				return t, true
			}
		}
		return pcc.PropertyTag{}, false
	}

	entryCount, replyCount, speakerCount := 0, 0, 0
	if t, ok := lookupTag("m_EntryList", "EntryList"); ok {
		entryCount = max(0, pcc.ReadArrayPropertyCount(data, t))
	}
	if t, ok := lookupTag("m_ReplyList", "ReplyList"); ok {
		replyCount = max(0, pcc.ReadArrayPropertyCount(data, t))
	}
	if t, ok := lookupTag("m_SpeakerList", "SpeakerList"); ok {
		speakerCount = max(0, pcc.ReadArrayPropertyCount(data, t))
	}

	var entryValues, replyValues, speakerValues []int
	if t, ok := lookupTag("m_EntryList", "EntryList"); ok {
		entryValues = pcc.ReadArrayPropertyI32Values(data, t)
	}
	if t, ok := lookupTag("m_ReplyList", "ReplyList"); ok {
		replyValues = pcc.ReadArrayPropertyI32Values(data, t)
	}
	if t, ok := lookupTag("m_SpeakerList", "SpeakerList"); ok {
		speakerValues = pcc.ReadArrayPropertyI32Values(data, t)
	}

	var entryRows, replyRows, speakerRows [][]int
	var usedStructHead, usedStructMatrix bool

	// Direct row reads (replaces buggy readRowArrays closure)
	if t, ok := lookupTag("m_EntryList", "EntryList"); ok {
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
	if t, ok := lookupTag("m_ReplyList", "ReplyList"); ok {
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
	if t, ok := lookupTag("m_SpeakerList", "SpeakerList"); ok {
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
	for _, key := range []string{"m_EntryList", "EntryList", "m_ReplyList", "ReplyList", "m_SpeakerList", "SpeakerList"} {
		if _, ok := tagMap[key]; ok {
			missingKeys = nil
			break
		}
	}
	if len(missingKeys) > 0 && !semanticMode {
		warnings = append(warnings, fmt.Sprintf("missing_key_properties:%s", joinStrings(missingKeys, ",")))
	}

	var entryMatrix, replyMatrix, speakerMatrix [][]int
	if t, ok := lookupTag("m_EntryList", "EntryList"); ok {
		entryMatrix = pcc.ReadArrayPropertyStructI32Matrix(data, t)
	}
	if t, ok := lookupTag("m_ReplyList", "ReplyList"); ok {
		replyMatrix = pcc.ReadArrayPropertyStructI32Matrix(data, t)
	}
	if t, ok := lookupTag("m_SpeakerList", "SpeakerList"); ok {
		speakerMatrix = pcc.ReadArrayPropertyStructI32Matrix(data, t)
	}

	for _, key := range []string{"m_EntryList", "EntryList", "m_ReplyList", "ReplyList", "m_SpeakerList", "SpeakerList"} {
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
		entries = semanticNodes.entries
		replies = semanticNodes.replies
		if len(semanticNodes.speakers) > 0 {
			speakers = semanticNodes.speakers
		}
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
			tids := []int{}
			if target >= 0 {
				tids = append(tids, target)
			}
			replies[i] = ReplyNode{ID: i, TargetEntryIDs: tids}
		}
	}

	if len(speakers) == 0 {
		if rowMode {
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
	}

	if semanticMode {
		// Validate entry → reply links from ReplyListNew
		validLinks := true
		for _, entry := range entries {
			for _, rid := range entry.ReplyLinks {
				if rid < 0 || rid >= len(replies) {
					validLinks = false
					break
				}
			}
			if !validLinks {
				break
			}
		}
		if validLinks {
			// Derive Reply → Entry from inverse mapping
			replyTargets := map[int][]int{}
			for _, entry := range entries {
				for _, rid := range entry.ReplyLinks {
					replyTargets[rid] = append(replyTargets[rid], entry.ID)
				}
			}
			for i := range replies {
				replies[i].TargetEntryIDs = replyTargets[replies[i].ID]
			}
		} else if len(replies) > 0 {
			// Links invalid but replies exist — fallback
			warnings = append(warnings, "invalid_ReplyListNew_links_fallback")
			for i := range entries {
				entries[i].ReplyLinks = []int{}
			}
		} else {
			// No replies — clear links silently
			for i := range entries {
				entries[i].ReplyLinks = []int{}
			}
		}
	} else {
		linksByEntry := map[int][]int{}
		knownEntryIDs := map[int]bool{}
		for _, entry := range entries {
			knownEntryIDs[entry.ID] = true
			linksByEntry[entry.ID] = []int{}
		}
		for _, reply := range replies {
			for _, tid := range reply.TargetEntryIDs {
				if knownEntryIDs[tid] {
					linksByEntry[tid] = append(linksByEntry[tid], reply.ID)
				} else if len(entries) > 0 {
					warnings = append(warnings, fmt.Sprintf("reply_target_missing_entry:%d->%d", reply.ID, tid))
				}
			}
		}
		for i := range entries {
			entries[i].ReplyLinks = linksByEntry[entries[i].ID]
		}
	}

	var startValues []int
	if t, ok := lookupTag("m_StartingList", "StartingList"); ok {
		startValues = pcc.ReadArrayPropertyI32Values(data, t)
	}
	starts := make([]StartNode, len(startValues))
	for i, val := range startValues {
		starts[i] = StartNode{ID: i, TargetEntryIDs: []int{val}}
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
