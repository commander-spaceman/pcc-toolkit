package editor

import (
	"encoding/binary"
	"fmt"

	pcc "github.com/commander-spaceman/me2pcc"
	"pcc-toolkit/core/internal/dialenc"
	"pcc-toolkit/core/internal/dialogue"
	"pcc-toolkit/core/internal/pccenc"
)

type propertySpan struct {
	name          string
	headerStart   int
	headerEnd     int
	valueStart    int
	valueEnd      int
	totalEnd      int
	propType      string
	structType    string
	arrayElemType string
	hasMeta       bool
}

func scanPropertySpans(data []byte, names []string) ([]propertySpan, error) {
	for _, delta := range []int{0, 4, 8, 12} {
		if delta >= len(data) {
			continue
		}
		tags, err := pcc.ParsePropertyTags(data, names, delta, len(data)-delta, false)
		if err == nil && len(tags) > 0 {
			return tagsToSpans(data, names, delta, tags), nil
		}
	}
	return nil, fmt.Errorf("no valid property spans found")
}

func tagsToSpans(data []byte, names []string, delta int, tags []pcc.PropertyTag) []propertySpan {
	var spans []propertySpan
	cursor := delta

	for _, tag := range tags {
		metaSize := computeMetaSize(data, names, cursor, tag)

		spans = append(spans, propertySpan{
			name:        tag.Name,
			propType:    tag.PropType,
			headerStart: cursor,
			headerEnd:   cursor + 24 + metaSize,
			valueStart:  tag.ValueOffset,
			valueEnd:    tag.ValueOffset + tag.Size,
			totalEnd:    tag.ValueOffset + tag.Size,
		})

		cursor = tag.ValueOffset + tag.Size
	}

	return spans
}

func computeMetaSize(data []byte, names []string, cursor int, tag pcc.PropertyTag) int {
	switch tag.PropType {
	case "StructProperty", "ByteProperty":
		return 8
	case "BoolProperty":
		if cursor+28 <= len(data) {
			name, _, _, _, _, m := pcc.ParsePropertyHeader(data, names, cursor, len(data))
			_ = name
			return m
		}
	case "ArrayProperty":
		if cursor+32 <= len(data) {
			a := int(int32(binary.LittleEndian.Uint32(data[cursor+24:])))
			b := int(int32(binary.LittleEndian.Uint32(data[cursor+28:])))
			aName := pcc.ResolveName(data, cursor+24, names)
			bName := pcc.ResolveName(data, cursor+28, names)
			if aName != "" && bName != "" {
				_ = a
				_ = b
			}
			if pcc.PropertyTypeNames[aName] || pcc.PropertyTypeNames[bName] {
				return 8
			}
		}
	}
	return 0
}

func SerializeConversationPreserving(conv dialogue.Conversation, originalData []byte, names []string) ([]byte, error) {
	hasContent := conv.ExportIndex >= 0
	if !hasContent {
		return SerializeConversationSimple(conv, names)
	}

	spans, err := scanPropertySpans(originalData, names)
	if err != nil {
		return nil, fmt.Errorf("scan spans: %w", err)
	}

	var modifiedSpans []propertySpan
	var replacementBytes [][]byte

	if len(conv.Entries) > 0 || hasEntryListInSpans(spans) {
		entryBytes, err := encodeEntryList(conv.Entries, names)
		if err != nil {
			return nil, fmt.Errorf("entry list: %w", err)
		}
		candidate := findSpan(spans, "m_EntryList", "EntryList")
		if candidate != nil {
			modifiedSpans = append(modifiedSpans, *candidate)
			replacementBytes = append(replacementBytes, entryBytes)
		}
	}

	if len(conv.Replies) > 0 || hasSpan(spans, "m_ReplyList", "ReplyList") {
		replyBytes, err := encodeReplyList(conv.Replies, names)
		if err != nil {
			return nil, fmt.Errorf("reply list: %w", err)
		}
		candidate := findSpan(spans, "m_ReplyList", "ReplyList")
		if candidate != nil {
			modifiedSpans = append(modifiedSpans, *candidate)
			replacementBytes = append(replacementBytes, replyBytes)
		}
	}

	if len(conv.Speakers) > 0 {
		speakerBytes, err := encodeSpeakerList(conv.Speakers, names)
		if err != nil {
			return nil, fmt.Errorf("speaker list: %w", err)
		}
		candidate := findSpan(spans, "m_SpeakerList", "SpeakerList")
		if candidate != nil {
			modifiedSpans = append(modifiedSpans, *candidate)
			replacementBytes = append(replacementBytes, speakerBytes)
		}
	}

	if len(conv.Starts) > 0 {
		startBytes, err := encodeStartList(conv.Starts, names)
		if err != nil {
			return nil, fmt.Errorf("start list: %w", err)
		}
		candidate := findSpan(spans, "m_StartingList", "StartingList")
		if candidate != nil {
			modifiedSpans = append(modifiedSpans, *candidate)
			replacementBytes = append(replacementBytes, startBytes)
		}
	}

	if len(modifiedSpans) == 0 {
		return originalData, nil
	}

	sortSpans(&modifiedSpans, &replacementBytes)

	return spliceProperties(originalData, modifiedSpans, replacementBytes)
}

func sortSpans(spans *[]propertySpan, replacements *[][]byte) {
	for i := 0; i < len(*spans); i++ {
		for j := i + 1; j < len(*spans); j++ {
			if (*spans)[i].headerStart > (*spans)[j].headerStart {
				(*spans)[i], (*spans)[j] = (*spans)[j], (*spans)[i]
				(*replacements)[i], (*replacements)[j] = (*replacements)[j], (*replacements)[i]
			}
		}
	}
}

func SerializeConversationSimple(conv dialogue.Conversation, names []string) ([]byte, error) {
	props := []pccenc.PropertyValue{}

	if len(conv.Entries) > 0 {
		items := make([]pccenc.PropertyValue, len(conv.Entries))
		for i, entry := range conv.Entries {
			eProps, err := dialenc.EncodeEntryNode(entry, names)
			if err != nil {
				return nil, fmt.Errorf("entry %d: %w", i, err)
			}
			items[i] = pccenc.PropertyValue{Properties: eProps}
		}
		props = append(props, pccenc.PropertyValue{
			Name:             "m_EntryList",
			PropType:         "ArrayProperty",
			ArrayElementType: "StructProperty",
			Items:            items,
		})
	}

	if len(conv.Replies) > 0 {
		items := make([]pccenc.PropertyValue, len(conv.Replies))
		for i, reply := range conv.Replies {
			rProps, err := dialenc.EncodeReplyNode(reply, names)
			if err != nil {
				return nil, fmt.Errorf("reply %d: %w", i, err)
			}
			items[i] = pccenc.PropertyValue{Properties: rProps}
		}
		props = append(props, pccenc.PropertyValue{
			Name:             "m_ReplyList",
			PropType:         "ArrayProperty",
			ArrayElementType: "StructProperty",
			Items:            items,
		})
	}

	if len(conv.Speakers) > 0 {
		items := make([]pccenc.PropertyValue, len(conv.Speakers))
		for i, speaker := range conv.Speakers {
			sProps, err := dialenc.EncodeSpeaker(speaker, names)
			if err != nil {
				return nil, fmt.Errorf("speaker %d: %w", i, err)
			}
			items[i] = pccenc.PropertyValue{Properties: sProps}
		}
		props = append(props, pccenc.PropertyValue{
			Name:             "m_SpeakerList",
			PropType:         "ArrayProperty",
			ArrayElementType: "StructProperty",
			Items:            items,
		})
	}

	if len(conv.Starts) > 0 {
		startItems := make([]pccenc.PropertyValue, 0, len(conv.Starts))
		for _, start := range conv.Starts {
			for _, targetID := range start.TargetEntryIDs {
				startItems = append(startItems, pccenc.PropertyValue{
					PropType: "IntProperty",
					Value:    targetID,
				})
			}
		}
		props = append(props, pccenc.PropertyValue{
			Name:             "m_StartingList",
			PropType:         "ArrayProperty",
			ArrayElementType: "IntProperty",
			Items:            startItems,
		})
	}

	return pccenc.EncodePropertyCollection(props, names)
}

func findSpan(spans []propertySpan, primary, secondary string) *propertySpan {
	for i := range spans {
		if spans[i].name == primary || spans[i].name == secondary {
			return &spans[i]
		}
	}
	return nil
}

func hasSpan(spans []propertySpan, primary, secondary string) bool {
	return findSpan(spans, primary, secondary) != nil
}

func hasEntryListInSpans(spans []propertySpan) bool {
	return hasSpan(spans, "m_EntryList", "EntryList")
}

func spliceProperties(data []byte, spans []propertySpan, replacements [][]byte) ([]byte, error) {
	type edit struct {
		oldStart int
		oldEnd   int
		newBytes []byte
	}

	var edits []edit
	for i, span := range spans {
		if i < len(replacements) {
			edits = append(edits, edit{
				oldStart: span.headerStart,
				oldEnd:   span.totalEnd,
				newBytes: replacements[i],
			})
		}
	}

	delta := 0
	for i := range edits {
		delta += len(edits[i].newBytes) - (edits[i].oldEnd - edits[i].oldStart)
	}

	result := make([]byte, len(data)+delta)

	pos := 0
	cursor := 0
	for _, ed := range edits {
		copy(result[pos:], data[cursor:ed.oldStart])
		pos += ed.oldStart - cursor
		cursor = ed.oldEnd

		copy(result[pos:], ed.newBytes)
		pos += len(ed.newBytes)
	}
	copy(result[pos:], data[cursor:])

	return result, nil
}

func encodeEntryList(entries []dialogue.EntryNode, names []string) ([]byte, error) {
	var added []string
	items := make([]pccenc.PropertyValue, len(entries))
	for i, entry := range entries {
		eProps, err := dialenc.EncodeEntryNodeWithAdditions(entry, names, &added)
		if err != nil {
			return nil, fmt.Errorf("entry %d: %w", i, err)
		}
		items[i] = pccenc.PropertyValue{Properties: eProps}
	}
	return pccenc.EncodePropertyValueWithAdditions(pccenc.PropertyValue{
		Name:             "m_EntryList",
		PropType:         "ArrayProperty",
		ArrayElementType: "StructProperty",
		Items:            items,
	}, names, &added)
}

func encodeReplyList(replies []dialogue.ReplyNode, names []string) ([]byte, error) {
	var added []string
	items := make([]pccenc.PropertyValue, len(replies))
	for i, reply := range replies {
		rProps, err := dialenc.EncodeReplyNodeWithAdditions(reply, names, &added)
		if err != nil {
			return nil, fmt.Errorf("reply %d: %w", i, err)
		}
		items[i] = pccenc.PropertyValue{Properties: rProps}
	}
	return pccenc.EncodePropertyValueWithAdditions(pccenc.PropertyValue{
		Name:             "m_ReplyList",
		PropType:         "ArrayProperty",
		ArrayElementType: "StructProperty",
		Items:            items,
	}, names, &added)
}

func encodeSpeakerList(speakers []dialogue.Speaker, names []string) ([]byte, error) {
	var added []string
	items := make([]pccenc.PropertyValue, len(speakers))
	for i, speaker := range speakers {
		sProps, err := dialenc.EncodeSpeaker(speaker, names)
		if err != nil {
			return nil, fmt.Errorf("speaker %d: %w", i, err)
		}
		items[i] = pccenc.PropertyValue{Properties: sProps}
	}
	return pccenc.EncodePropertyValueWithAdditions(pccenc.PropertyValue{
		Name:             "m_SpeakerList",
		PropType:         "ArrayProperty",
		ArrayElementType: "StructProperty",
		Items:            items,
	}, names, &added)
}

func encodeStartList(starts []dialogue.StartNode, names []string) ([]byte, error) {
	startItems := make([]pccenc.PropertyValue, 0, len(starts))
	for _, start := range starts {
		for _, targetID := range start.TargetEntryIDs {
			startItems = append(startItems, pccenc.PropertyValue{
				PropType: "IntProperty",
				Value:    targetID,
			})
		}
	}
	return pccenc.EncodePropertyValue(pccenc.PropertyValue{
		Name:             "m_StartingList",
		PropType:         "ArrayProperty",
		ArrayElementType: "IntProperty",
		Items:            startItems,
	}, names)
}
