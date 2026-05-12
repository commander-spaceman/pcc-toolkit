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

func readReplyIndicesFromEntry(data []byte, names []string, item map[string]pcc.ParsedProperty) []int {
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
	var ids []int
	for _, si := range structItems {
		if nidx, ok := si["nIndex"]; ok {
			if v, ok := nidx.Value.(int); ok && v >= 0 {
				ids = append(ids, v)
			}
		}
	}
	return ids
}

func readEntryIndicesFromReply(data []byte, names []string, item map[string]pcc.ParsedProperty) []int {
	el, ok := item["EntryList"]
	if !ok {
		return nil
	}
	arrMap, ok := el.Value.(map[string]interface{})
	if !ok {
		return nil
	}
	count, _ := arrMap["count"].(int)
	ps, _ := arrMap["payload_offset"].(int)
	if count <= 0 {
		return nil
	}
	var ids []int
	for k := 0; k < count; k++ {
		eid := pcc.ReadRawI32(data, ps+(k*4))
		if eid >= 0 {
			ids = append(ids, eid)
		}
	}
	return ids
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
		replyLinks := readReplyIndicesFromEntry(data, names, item)
		entries[idx] = EntryNode{
			ID:         idx,
			SpeakerID:  speakerID,
			LineStrRef: lineStrRef,
			ReplyLinks: replyLinks,
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
		targetEntryIDs = append(targetEntryIDs, readEntryIndicesFromReply(data, names, item)...)
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
			ID:             idx,
			LineStrRef:     lineStrRef,
			TargetEntryIDs: targetEntryIDs,
			ConditionRefs:  condRefs,
			Category:       category,
		}
	}

	var speakerItems []map[string]pcc.ParsedProperty
	if spkTag, ok := tagMap["SpeakerList"]; ok {
		spkCount, spkPayload, spkPayloadSize := pcc.ReadArrayPropertyPayloadInfo(data, spkTag)
		if spkCount > 0 && spkPayloadSize > 0 {
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
