package dialogue

import (
	"fmt"

	"pcc-toolkit/core/internal/pcc"
)

type semanticResult struct {
	entries  []EntryNode
	replies  []ReplyNode
	speakers []Speaker
}

func findReplyChoicesFromEntry(data []byte, names []string, item map[string]pcc.ParsedProperty, entryID int) []ReplyChoice {
	prop, ok := item["ReplyListNew"]
	if !ok {
		return nil
	}
	arrMap, ok := prop.Value.(map[string]interface{})
	if !ok {
		return nil
	}
	count, _ := arrMap["count"].(int)
	ps, _ := arrMap["payload_offset"].(int)
	psz, _ := arrMap["payload_size"].(int)
	if count <= 0 || psz <= 0 {
		return nil
	}
	structItems := pcc.ParseStructArrayItemsAsPropertyCollections(data, names, ps, psz, count)
	var choices []ReplyChoice
	for order, si := range structItems {
		choice := ReplyChoice{
			FromEntryID: entryID,
			Order:       order,
		}
		if nidx, ok := si["nIndex"]; ok {
			if v, ok := nidx.Value.(int); ok && v >= 0 {
				choice.ToReplyID = v
			}
		}
		if srPar, ok := si["srParaphrase"]; ok {
			if v, ok := srPar.Value.(int); ok && v != 0 {
				choice.ParaphraseStrRef = &v
			}
		}
		if sPar, ok := si["sParaphrase"]; ok {
			if v, ok := sPar.Value.(string); ok {
				choice.Paraphrase = v
			}
		}
		if cat, ok := si["Category"]; ok {
			if v, ok := cat.Value.(string); ok {
				choice.Category = v
			}
		}
		choices = append(choices, choice)
	}
	return choices
}

func findEntryIndicesFromReply(data []byte, names []string, itemStart, itemEnd int) []int {
	return pcc.FindInt32ArrayByName(data, names, itemStart, itemEnd, "EntryList")
}

func trySemanticStructNodes(data []byte, names []string, tagMap map[string]pcc.PropertyTag) *semanticResult {
	lookupTag := func(keys ...string) (pcc.PropertyTag, bool) {
		for _, key := range keys {
			if t, ok := tagMap[key]; ok {
				return t, true
			}
		}
		return pcc.PropertyTag{}, false
	}

	entryCount := 0
	entryPayload := 0
	entryPayloadSize := 0
	if tag, ok := lookupTag("m_EntryList", "EntryList"); ok {
		entryCount, entryPayload, entryPayloadSize = pcc.ReadArrayPropertyPayloadInfo(data, tag)
	}

	replyCount := 0
	replyPayload := 0
	replyPayloadSize := 0
	if tag, ok := lookupTag("m_ReplyList", "ReplyList"); ok {
		replyCount, replyPayload, replyPayloadSize = pcc.ReadArrayPropertyPayloadInfo(data, tag)
	}

	// Try parsing entries schema-guided first, fall back to generic
	var entryItems []map[string]pcc.ParsedProperty
	if entryCount > 0 && entryPayloadSize > 0 {
		entryItems = parseStructArraySchemaGuided(data, names, "BioDialogEntryNode", entryPayload, entryPayloadSize, entryCount)
		if len(entryItems) == 0 {
			entryItems = pcc.ParseStructArrayItemsAsPropertyCollections(data, names, entryPayload, entryPayloadSize, entryCount)
		}
	}

	// Try parsing replies schema-guided first, fall back to generic
	var replyItems []map[string]pcc.ParsedProperty
	if replyCount > 0 && replyPayloadSize > 0 {
		replyItems = parseStructArraySchemaGuided(data, names, "BioDialogReplyNode", replyPayload, replyPayloadSize, max(0, replyCount))
		if len(replyItems) == 0 {
			replyItems = pcc.ParseStructArrayItemsAsPropertyCollections(data, names, replyPayload, replyPayloadSize, max(0, replyCount))
		}
	}

	// Need at least entries or replies to proceed
	if len(entryItems) == 0 && len(replyItems) == 0 {
		if replyCount > 0 || entryCount > 0 {
			return nil
		}
		return nil
	}

	replyStride := 0
	if len(replyItems) > 0 {
		replyStride = replyPayloadSize / len(replyItems)
	}

	entries := make([]EntryNode, len(entryItems))
	for idx, item := range entryItems {
		var speakerID *int
		if spk, ok := item["nSpeakerIndex"]; ok {
			if v, ok := spk.Value.(int); ok {
				speakerID = &v
			}
		}
		var lineStrRef *int
		if sr, ok := item["srText"]; ok {
			if v, ok := sr.Value.(int); ok {
				lineStrRef = &v
			}
		}
		var listenerIndex *int
		if li, ok := item["nListenerIndex"]; ok {
			if v, ok := li.Value.(int); ok && v >= 0 {
				listenerIndex = &v
			}
		}
		var conditionalFunc *int
		if cf, ok := item["nConditionalFunc"]; ok {
			if v, ok := cf.Value.(int); ok && v >= 0 {
				conditionalFunc = &v
			}
		}
		var conditionalParam *int
		if cp, ok := item["nConditionalParam"]; ok {
			if v, ok := cp.Value.(int); ok && v != 0 {
				conditionalParam = &v
			}
		}
		var stateTransition *int
		if st, ok := item["nStateTransition"]; ok {
			if v, ok := st.Value.(int); ok && v >= 0 {
				stateTransition = &v
			}
		}
		var stateTransitionParam *int
		if stp, ok := item["nStateTransitionParam"]; ok {
			if v, ok := stp.Value.(int); ok && v != 0 {
				stateTransitionParam = &v
			}
		}
		var scriptIndex *int
		if si, ok := item["nScriptIndex"]; ok {
			if v, ok := si.Value.(int); ok && v >= 0 {
				scriptIndex = &v
			}
		}
		var firesConditional *bool
		if fc, ok := item["bFireConditional"]; ok {
			if v, ok := fc.Value.(bool); ok {
				firesConditional = &v
			}
		}
		var exportID *int
		if eid, ok := item["nExportID"]; ok {
			if v, ok := eid.Value.(int); ok && v != 0 {
				exportID = &v
			}
		}
		var skippable *bool
		if s, ok := item["bSkippable"]; ok {
			if v, ok := s.Value.(bool); ok {
				skippable = &v
			}
		}
		var nonTextLine *bool
		if ntl, ok := item["bIsNonTextLine"]; ok {
			if v, ok := ntl.Value.(bool); ok {
				nonTextLine = &v
			}
		}
		var ambient *bool
		if a, ok := item["bAmbient"]; ok {
			if v, ok := a.Value.(bool); ok {
				ambient = &v
			}
		}
		var cameraIntimacy *int
		if ci, ok := item["nCameraIntimacy"]; ok {
			if v, ok := ci.Value.(int); ok && v >= 0 {
				cameraIntimacy = &v
			}
		}
		var guiStyle string
		if gs, ok := item["eConvGUIStyle"]; ok {
			if s, ok := gs.Value.(string); ok {
				guiStyle = s
			}
		}
		replyChoices := findReplyChoicesFromEntry(data, names, item, idx)
		var replyLinks []int
		for _, rc := range replyChoices {
			replyLinks = append(replyLinks, rc.ToReplyID)
		}
		entries[idx] = EntryNode{
			ID:                   idx,
			SpeakerID:            speakerID,
			ListenerIndex:        listenerIndex,
			LineStrRef:           lineStrRef,
			ReplyLinks:           replyLinks,
			ReplyChoices:         replyChoices,
			ConditionalFunc:      conditionalFunc,
			ConditionalParam:     conditionalParam,
			StateTransition:      stateTransition,
			StateTransitionParam: stateTransitionParam,
			ScriptIndex:          scriptIndex,
			FiresConditional:     firesConditional,
			ExportID:             exportID,
			Skippable:            skippable,
			NonTextLine:          nonTextLine,
			Ambient:              ambient,
			CameraIntimacy:       cameraIntimacy,
			GUIStyle:             guiStyle,
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
		var targetEntryIDs []int
		if t, ok := item["nIndex"]; ok {
			if v, ok := t.Value.(int); ok {
				targetEntryIDs = append(targetEntryIDs, v)
			}
		} else if t, ok := item["nEntryIndex"]; ok {
			if v, ok := t.Value.(int); ok {
				targetEntryIDs = append(targetEntryIDs, v)
			}
		}
		itemStart := replyPayload + (idx * replyStride)
		itemEnd := replyPayload + ((idx + 2) * replyStride) // expand into next item range
		if itemEnd > replyPayload+replyPayloadSize {
			itemEnd = replyPayload + replyPayloadSize
		}
		if idx >= replyCount-1 {
			itemEnd = replyPayload + replyPayloadSize
		}
		targetEntryIDs = append(targetEntryIDs, findEntryIndicesFromReply(data, names, itemStart, itemEnd)...)
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
		var replyType string
		if rt, ok := item["ReplyType"]; ok {
			if s, ok := rt.Value.(string); ok {
				replyType = s
			}
		}
		var conditionalFunc *int
		if cf2, ok := item["nConditionalFunc"]; ok {
			if v, ok := cf2.Value.(int); ok && v >= 0 {
				conditionalFunc = &v
			}
		}
		var conditionalParam *int
		if cp2, ok := item["nConditionalParam"]; ok {
			if v, ok := cp2.Value.(int); ok && v != 0 {
				conditionalParam = &v
			}
		}
		var stateTransition *int
		if st, ok := item["nStateTransition"]; ok {
			if v, ok := st.Value.(int); ok && v >= 0 {
				stateTransition = &v
			}
		}
		var stateTransitionParam *int
		if stp, ok := item["nStateTransitionParam"]; ok {
			if v, ok := stp.Value.(int); ok && v != 0 {
				stateTransitionParam = &v
			}
		}
		var scriptIndex *int
		if si, ok := item["nScriptIndex"]; ok {
			if v, ok := si.Value.(int); ok && v >= 0 {
				scriptIndex = &v
			}
		}
		var firesConditional *bool
		if fc, ok := item["bFireConditional"]; ok {
			if v, ok := fc.Value.(bool); ok {
				firesConditional = &v
			}
		}
		var exportID *int
		if eid, ok := item["nExportID"]; ok {
			if v, ok := eid.Value.(int); ok && v != 0 {
				exportID = &v
			}
		}
		var unskippable *bool
		if u, ok := item["bUnskippable"]; ok {
			if v, ok := u.Value.(bool); ok {
				unskippable = &v
			}
		}
		var nonTextLine *bool
		if ntl, ok := item["bIsNonTextLine"]; ok {
			if v, ok := ntl.Value.(bool); ok {
				nonTextLine = &v
			}
		}
		var ambient *bool
		if a, ok := item["bAmbient"]; ok {
			if v, ok := a.Value.(bool); ok {
				ambient = &v
			}
		}
		var cameraIntimacy *int
		if ci, ok := item["nCameraIntimacy"]; ok {
			if v, ok := ci.Value.(int); ok && v >= 0 {
				cameraIntimacy = &v
			}
		}
		var guiStyle string
		if gs, ok := item["eConvGUIStyle"]; ok {
			if s, ok := gs.Value.(string); ok {
				guiStyle = s
			}
		}
		replies[idx] = ReplyNode{
			ID:                   idx,
			LineStrRef:           lineStrRef,
			TargetEntryIDs:       targetEntryIDs,
			ConditionRefs:        condRefs,
			Category:             category,
			ReplyType:            replyType,
			ConditionalFunc:      conditionalFunc,
			ConditionalParam:     conditionalParam,
			StateTransition:      stateTransition,
			StateTransitionParam: stateTransitionParam,
			ScriptIndex:          scriptIndex,
			FiresConditional:     firesConditional,
			ExportID:             exportID,
			Unskippable:          unskippable,
			NonTextLine:          nonTextLine,
			Ambient:              ambient,
			CameraIntimacy:       cameraIntimacy,
			GUIStyle:             guiStyle,
		}
	}

	var speakers []Speaker
	speakers = append(speakers,
		Speaker{ID: -2, Tag: "player", StrRefID: newInt(125303), FriendlyName: "\"Shepard\""},
		Speaker{ID: -1, Tag: "owner", StrRefID: newInt(0), FriendlyName: "No data"},
	)
	if tag, ok := lookupTag("m_SpeakerList", "SpeakerList"); ok {
		spkCount, spkPayload, spkPayloadSize := pcc.ReadArrayPropertyPayloadInfo(data, tag)
		if spkCount > 0 && spkPayloadSize > 0 {
			speakerItems := parseStructArraySchemaGuided(data, names, "BioDialogSpeaker", spkPayload, spkPayloadSize, spkCount)
			if len(speakerItems) == 0 {
				speakerItems = pcc.ParseStructArrayItemsAsPropertyCollections(data, names, spkPayload, spkPayloadSize, spkCount)
			}
			if len(speakerItems) > 0 {
				for idx, item := range speakerItems {
					var tag string
					if tp, ok := item["sSpeakerTag"]; ok {
						if s, ok := tp.Value.(string); ok {
							tag = s
						}
					}
					var strRefID *int
					if sr, ok := item["nDisplayNameStrRef"]; ok {
						if v, ok := sr.Value.(int); ok && v != 0 {
							strRefID = &v
						}
					}
					speakers = append(speakers, Speaker{ID: idx, Tag: tag, StrRefID: strRefID})
				}
			} else {
				directSpeakers := parseSpeakersDirect(data, names, spkPayload, spkPayloadSize, spkCount)
				speakers = append(speakers, directSpeakers...)
			}
		}
	}

	return &semanticResult{entries, replies, speakers}
}

func parseSpeakersDirect(data []byte, names []string, payloadStart, payloadSize, count int) []Speaker {
	// Find the NameProperty value for sSpeakerTag in each item.
	// Each BioDialogSpeaker item: FName(sSpeakerTag) + FName(NameProperty) + 4(size) + 4(idx) + 4(value) = 28 bytes.
	// If items are uniform size, compute stride; otherwise scan for sSpeakerTag occurrences.
	stride := 0
	if count > 0 && payloadSize%count == 0 {
		stride = payloadSize / count
	}
	if stride < 24 {
		stride = 28
	}
	speakers := make([]Speaker, count)
	for i := 0; i < count; i++ {
		itemStart := payloadStart + (i * stride)
		valueOffset := itemStart + 24
		if valueOffset+4 > len(data) {
			break
		}
		nameIdx := pcc.ReadRawI32(data, valueOffset)
		var tag string
		if nameIdx >= 0 && nameIdx < len(names) {
			tag = names[nameIdx]
		}
		speakers[i] = Speaker{ID: i, Tag: tag}
	}
	return speakers
}

// parseStructArraySchemaGuided parses an ArrayProperty<StructProperty> payload
// using the known ME2 OT struct layout from structdb. Returns nil if schema-guided
// parsing fails, so the caller can fall back to generic heuristic parsing.
func parseStructArraySchemaGuided(data []byte, names []string, structType string, payloadOffset, payloadSize, count int) []map[string]pcc.ParsedProperty {
	if count <= 0 || payloadSize <= 0 {
		return nil
	}
	structProps, ok := ME2StructPropInfo[structType]
	if !ok {
		return nil
	}

	end := payloadOffset + payloadSize
	if end > len(data) {
		end = len(data)
	}

	expectedFirstProp := ""
	for _, propName := range []string{"nIndex", "srText", "sSpeakerTag", "nSpeakerIndex"} {
		if _, ok := structProps[propName]; ok {
			expectedFirstProp = propName
			break
		}
	}
	if expectedFirstProp == "" {
		return nil
	}

	firstPropIdx := -1
	for i, n := range names {
		if n == expectedFirstProp {
			firstPropIdx = i
			break
		}
	}
	if firstPropIdx < 0 {
		return nil
	}

	expectedStr := map[string]bool{}
	for name := range structProps {
		expectedStr[name] = true
	}

	findItemStart := func(start int) int {
		for pos := start; pos < end-24; pos += 4 {
			a := pcc.ReadRawI32(data, pos)
			b := pcc.ReadRawI32(data, pos+4)
			if (a == firstPropIdx && b == 0) || (a == 0 && b == firstPropIdx) {
				ta := pcc.ReadRawI32(data, pos+8)
				tb := pcc.ReadRawI32(data, pos+12)
				taName := ""
				if ta >= 0 && ta < len(names) {
					taName = names[ta]
				}
				tbName := ""
				if tb >= 0 && tb < len(names) {
					tbName = names[tb]
				}
				if pcc.PropertyTypeNames[taName] || pcc.PropertyTypeNames[tbName] {
					return pos
				}
			}
		}
		return -1
	}

	firstStart := findItemStart(payloadOffset)
	if firstStart < 0 {
		return nil
	}

	var items []map[string]pcc.ParsedProperty
	cursor := firstStart
	for i := 0; i < count && cursor < end; i++ {
		item, nextCursor := pcc.ParsePropertyCollection(data, names, cursor, end-cursor)
		if item == nil {
			break
		}

		hasExpected := false
		for name := range item {
			if expectedStr[name] {
				hasExpected = true
				break
			}
		}
		if !hasExpected {
			break
		}

		items = append(items, item)
		if nextCursor <= cursor {
			break
		}
		cursor = nextCursor
	}

	if len(items) == count {
		return items
	}

	return nil
}

func newInt(i int) *int { v := i; return &v }
