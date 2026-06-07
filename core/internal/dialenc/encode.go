package dialenc

import (
	"pcc-toolkit/core/internal/dialogue"
	"pcc-toolkit/core/internal/pccenc"
)

const defaultInt = 0
const unusedInt = -1

func intVal(p *int) int {
	if p == nil {
		return unusedInt
	}
	return *p
}

func boolVal(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

func EncodeEntryNode(entry dialogue.EntryNode, names []string) ([]pccenc.PropertyValue, error) {
	return EncodeEntryNodeWithAdditions(entry, names, nil)
}

func EncodeEntryNodeWithAdditions(entry dialogue.EntryNode, names []string, added *[]string) ([]pccenc.PropertyValue, error) {
	props := []pccenc.PropertyValue{
		{Name: "nIndex", PropType: "IntProperty", Value: entry.ID},
		{Name: "nSpeakerIndex", PropType: "IntProperty", Value: intVal(entry.SpeakerID)},
		{Name: "srText", PropType: "StringRefProperty", Value: intVal(entry.LineStrRef)},
	}

	if entry.ListenerIndex != nil {
		props = append(props, pccenc.PropertyValue{Name: "nListenerIndex", PropType: "IntProperty", Value: *entry.ListenerIndex})
	}
	if entry.ConditionalFunc != nil {
		props = append(props, pccenc.PropertyValue{Name: "nConditionalFunc", PropType: "IntProperty", Value: *entry.ConditionalFunc})
	}
	if entry.ConditionalParam != nil {
		props = append(props, pccenc.PropertyValue{Name: "nConditionalParam", PropType: "IntProperty", Value: *entry.ConditionalParam})
	}
	if entry.StateTransition != nil {
		props = append(props, pccenc.PropertyValue{Name: "nStateTransition", PropType: "IntProperty", Value: *entry.StateTransition})
	}
	if entry.StateTransitionParam != nil {
		props = append(props, pccenc.PropertyValue{Name: "nStateTransitionParam", PropType: "IntProperty", Value: *entry.StateTransitionParam})
	}
	if entry.ScriptIndex != nil {
		props = append(props, pccenc.PropertyValue{Name: "nScriptIndex", PropType: "IntProperty", Value: *entry.ScriptIndex})
	}
	if entry.ExportID != nil {
		props = append(props, pccenc.PropertyValue{Name: "nExportID", PropType: "IntProperty", Value: *entry.ExportID})
	}
	if entry.CameraIntimacy != nil {
		props = append(props, pccenc.PropertyValue{Name: "nCameraIntimacy", PropType: "IntProperty", Value: *entry.CameraIntimacy})
	}
	if entry.FiresConditional != nil {
		props = append(props, pccenc.PropertyValue{Name: "bFireConditional", PropType: "BoolProperty", Value: *entry.FiresConditional})
	}
	if entry.Skippable != nil {
		props = append(props, pccenc.PropertyValue{Name: "bSkippable", PropType: "BoolProperty", Value: *entry.Skippable})
	}
	if entry.NonTextLine != nil {
		props = append(props, pccenc.PropertyValue{Name: "bIsNonTextLine", PropType: "BoolProperty", Value: *entry.NonTextLine})
	}
	if entry.Ambient != nil {
		props = append(props, pccenc.PropertyValue{Name: "bAmbient", PropType: "BoolProperty", Value: *entry.Ambient})
	}
	if entry.GUIStyle != "" {
		props = append(props, pccenc.PropertyValue{Name: "eConvGUIStyle", PropType: "EnumProperty", Value: coalesceGUI(entry.GUIStyle)})
	}

	if len(entry.ReplyChoices) > 0 {
		items := make([]pccenc.PropertyValue, len(entry.ReplyChoices))
		for i, rc := range entry.ReplyChoices {
			items[i] = EncodeReplyChoice(rc, names)
		}
		props = append(props, pccenc.PropertyValue{
			Name:             "ReplyListNew",
			PropType:         "ArrayProperty",
			ArrayElementType: "StructProperty",
			Items:            items,
		})
	}

	return props, nil
}

func EncodeReplyNode(reply dialogue.ReplyNode, names []string) ([]pccenc.PropertyValue, error) {
	return EncodeReplyNodeWithAdditions(reply, names, nil)
}

func EncodeReplyNodeWithAdditions(reply dialogue.ReplyNode, names []string, added *[]string) ([]pccenc.PropertyValue, error) {
	props := []pccenc.PropertyValue{
		{Name: "nIndex", PropType: "IntProperty", Value: reply.ID},
		{Name: "srText", PropType: "StringRefProperty", Value: intVal(reply.LineStrRef)},
	}

	if reply.ConditionalFunc != nil {
		props = append(props, pccenc.PropertyValue{Name: "nConditionalFunc", PropType: "IntProperty", Value: *reply.ConditionalFunc})
	}
	if reply.ConditionalParam != nil {
		props = append(props, pccenc.PropertyValue{Name: "nConditionalParam", PropType: "IntProperty", Value: *reply.ConditionalParam})
	}
	if reply.StateTransition != nil {
		props = append(props, pccenc.PropertyValue{Name: "nStateTransition", PropType: "IntProperty", Value: *reply.StateTransition})
	}
	if reply.StateTransitionParam != nil {
		props = append(props, pccenc.PropertyValue{Name: "nStateTransitionParam", PropType: "IntProperty", Value: *reply.StateTransitionParam})
	}
	if reply.ScriptIndex != nil {
		props = append(props, pccenc.PropertyValue{Name: "nScriptIndex", PropType: "IntProperty", Value: *reply.ScriptIndex})
	}
	if reply.ExportID != nil {
		props = append(props, pccenc.PropertyValue{Name: "nExportID", PropType: "IntProperty", Value: *reply.ExportID})
	}
	if reply.CameraIntimacy != nil {
		props = append(props, pccenc.PropertyValue{Name: "nCameraIntimacy", PropType: "IntProperty", Value: *reply.CameraIntimacy})
	}
	if reply.FiresConditional != nil {
		props = append(props, pccenc.PropertyValue{Name: "bFireConditional", PropType: "BoolProperty", Value: *reply.FiresConditional})
	}
	if reply.Unskippable != nil {
		props = append(props, pccenc.PropertyValue{Name: "bUnskippable", PropType: "BoolProperty", Value: *reply.Unskippable})
	}
	if reply.NonTextLine != nil {
		props = append(props, pccenc.PropertyValue{Name: "bIsNonTextLine", PropType: "BoolProperty", Value: *reply.NonTextLine})
	}
	if reply.Ambient != nil {
		props = append(props, pccenc.PropertyValue{Name: "bAmbient", PropType: "BoolProperty", Value: *reply.Ambient})
	}
	if reply.GUIStyle != "" {
		props = append(props, pccenc.PropertyValue{Name: "eConvGUIStyle", PropType: "EnumProperty", Value: coalesceGUI(reply.GUIStyle)})
	}
	if reply.ReplyType != "" {
		props = append(props, pccenc.PropertyValue{Name: "ReplyType", PropType: "EnumProperty", Value: coalesceReplyType(reply.ReplyType)})
	}
	if reply.Category != "" {
		props = append(props, pccenc.PropertyValue{Name: "Category", PropType: "NameProperty", Value: coalesceCategory(reply.Category)})
	}

	if len(reply.TargetEntryIDs) > 0 {
		items := make([]pccenc.PropertyValue, len(reply.TargetEntryIDs))
		for i, eid := range reply.TargetEntryIDs {
			items[i] = pccenc.PropertyValue{PropType: "IntProperty", Value: eid}
		}
		props = append(props, pccenc.PropertyValue{
			Name:             "EntryList",
			PropType:         "ArrayProperty",
			ArrayElementType: "IntProperty",
			Items:            items,
		})
	}

	return props, nil
}

func EncodeReplyChoice(choice dialogue.ReplyChoice, names []string) pccenc.PropertyValue {
	paraphraseStrRef := unusedInt
	if choice.ParaphraseStrRef != nil {
		paraphraseStrRef = *choice.ParaphraseStrRef
	}
	category := choice.Category
	if category == "" {
		category = "REPLY_CATEGORY_DEFAULT"
	}
	return pccenc.PropertyValue{
		Properties: []pccenc.PropertyValue{
			{Name: "nIndex", PropType: "IntProperty", Value: choice.ToReplyID},
			{Name: "srParaphrase", PropType: "StringRefProperty", Value: paraphraseStrRef},
			{Name: "sParaphrase", PropType: "StrProperty", Value: choice.Paraphrase},
			{Name: "Category", PropType: "NameProperty", Value: category},
		},
	}
}

func EncodeSpeaker(speaker dialogue.Speaker, names []string) ([]pccenc.PropertyValue, error) {
	strRefID := unusedInt
	if speaker.StrRefID != nil {
		strRefID = *speaker.StrRefID
	}
	if speaker.Tag == "" {
		speaker.Tag = "None"
	}
	return []pccenc.PropertyValue{
		{Name: "nIndex", PropType: "IntProperty", Value: speaker.ID},
		{Name: "sSpeakerTag", PropType: "NameProperty", Value: speaker.Tag},
		{Name: "nDisplayNameStrRef", PropType: "StringRefProperty", Value: strRefID},
	}, nil
}

func EncodeEntryNodeBytes(entry dialogue.EntryNode, names []string) ([]byte, error) {
	props, err := EncodeEntryNode(entry, names)
	if err != nil {
		return nil, err
	}
	return pccenc.EncodePropertyCollection(props, names)
}

func EncodeReplyNodeBytes(reply dialogue.ReplyNode, names []string) ([]byte, error) {
	props, err := EncodeReplyNode(reply, names)
	if err != nil {
		return nil, err
	}
	return pccenc.EncodePropertyCollection(props, names)
}

func EncodeSpeakerBytes(speaker dialogue.Speaker, names []string) ([]byte, error) {
	props, err := EncodeSpeaker(speaker, names)
	if err != nil {
		return nil, err
	}
	return pccenc.EncodePropertyCollection(props, names)
}

func coalesceGUI(s string) string {
	if s == "" {
		return "CONV_GUISTYLE_DEFAULT"
	}
	return s
}

func coalesceReplyType(s string) string {
	if s == "" {
		return "REPLY_TYPE_DEFAULT"
	}
	return s
}

func coalesceCategory(s string) string {
	if s == "" {
		return "REPLY_CATEGORY_DEFAULT"
	}
	return s
}
