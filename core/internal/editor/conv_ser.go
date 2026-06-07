package editor

import (
	"fmt"

	"pcc-toolkit/core/internal/dialenc"
	"pcc-toolkit/core/internal/dialogue"
	"pcc-toolkit/core/internal/pccenc"
)

func SerializeConversation(conv dialogue.Conversation, names []string) ([]byte, []string, error) {
	var added []string
	props := []pccenc.PropertyValue{}

	if len(conv.Entries) > 0 {
		items := make([]pccenc.PropertyValue, len(conv.Entries))
		for i, entry := range conv.Entries {
			eProps, err := dialenc.EncodeEntryNodeWithAdditions(entry, names, &added)
			if err != nil {
				return nil, nil, fmt.Errorf("entry %d: %w", i, err)
			}
			items[i] = pccenc.PropertyValue{
				Properties: eProps,
			}
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
			rProps, err := dialenc.EncodeReplyNodeWithAdditions(reply, names, &added)
			if err != nil {
				return nil, nil, fmt.Errorf("reply %d: %w", i, err)
			}
			items[i] = pccenc.PropertyValue{
				Properties: rProps,
			}
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
				return nil, nil, fmt.Errorf("speaker %d: %w", i, err)
			}
			items[i] = pccenc.PropertyValue{
				Properties: sProps,
			}
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

	if len(conv.ScriptList) > 0 {
		scriptItems := make([]pccenc.PropertyValue, len(conv.ScriptList))
		for i, se := range conv.ScriptList {
			scriptItems[i] = pccenc.PropertyValue{
				Properties: []pccenc.PropertyValue{
					{Name: "sScriptTag", PropType: "NameProperty", Value: se.Tag},
				},
			}
		}
		props = append(props, pccenc.PropertyValue{
			Name:             "m_ScriptList",
			PropType:         "ArrayProperty",
			ArrayElementType: "StructProperty",
			Items:            scriptItems,
		})
	}

	if conv.MatineeSequenceExportID != nil {
		props = append(props, pccenc.PropertyValue{
			Name:     "MatineeSequence",
			PropType: "ObjectProperty",
			Value:    *conv.MatineeSequenceExportID,
		})
	}

	allNames := make([]string, len(names)+len(added))
	copy(allNames, names)
	copy(allNames[len(names):], added)

	serial, err := pccenc.EncodePropertyCollectionWithAdditions(props, names, &added)
	if err != nil {
		return nil, nil, err
	}

	return serial, added, nil
}
