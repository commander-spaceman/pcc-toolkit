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
	props := []pccenc.PropertyValue{
		{Name: "nIndex", PropType: "IntProperty", Value: entry.ID},
		{Name: "nSpeakerIndex", PropType: "IntProperty", Value: intVal(entry.SpeakerID)},
		{Name: "srText", PropType: "StringRefProperty", Value: intVal(entry.LineStrRef)},
		{Name: "nListenerIndex", PropType: "IntProperty", Value: intVal(entry.ListenerIndex)},
		{Name: "nConditionalFunc", PropType: "IntProperty", Value: intVal(entry.ConditionalFunc)},
		{Name: "nConditionalParam", PropType: "IntProperty", Value: intVal(entry.ConditionalParam)},
		{Name: "nStateTransition", PropType: "IntProperty", Value: intVal(entry.StateTransition)},
		{Name: "nStateTransitionParam", PropType: "IntProperty", Value: intVal(entry.StateTransitionParam)},
		{Name: "nScriptIndex", PropType: "IntProperty", Value: intVal(entry.ScriptIndex)},
		{Name: "nExportID", PropType: "IntProperty", Value: intVal(entry.ExportID)},
		{Name: "nCameraIntimacy", PropType: "IntProperty", Value: intVal(entry.CameraIntimacy)},
		{Name: "bFireConditional", PropType: "BoolProperty", Value: boolVal(entry.FiresConditional)},
		{Name: "bSkippable", PropType: "BoolProperty", Value: boolVal(entry.Skippable)},
		{Name: "bIsNonTextLine", PropType: "BoolProperty", Value: boolVal(entry.NonTextLine)},
		{Name: "bAmbient", PropType: "BoolProperty", Value: boolVal(entry.Ambient)},
		{Name: "eConvGUIStyle", PropType: "EnumProperty", Value: coalesceGUI(entry.GUIStyle)},
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
	props := []pccenc.PropertyValue{
		{Name: "nIndex", PropType: "IntProperty", Value: reply.ID},
		{Name: "srText", PropType: "StringRefProperty", Value: intVal(reply.LineStrRef)},
		{Name: "nConditionalFunc", PropType: "IntProperty", Value: intVal(reply.ConditionalFunc)},
		{Name: "nConditionalParam", PropType: "IntProperty", Value: intVal(reply.ConditionalParam)},
		{Name: "nStateTransition", PropType: "IntProperty", Value: intVal(reply.StateTransition)},
		{Name: "nStateTransitionParam", PropType: "IntProperty", Value: intVal(reply.StateTransitionParam)},
		{Name: "nScriptIndex", PropType: "IntProperty", Value: intVal(reply.ScriptIndex)},
		{Name: "nExportID", PropType: "IntProperty", Value: intVal(reply.ExportID)},
		{Name: "nCameraIntimacy", PropType: "IntProperty", Value: intVal(reply.CameraIntimacy)},
		{Name: "bFireConditional", PropType: "BoolProperty", Value: boolVal(reply.FiresConditional)},
		{Name: "bUnskippable", PropType: "BoolProperty", Value: boolVal(reply.Unskippable)},
		{Name: "bIsNonTextLine", PropType: "BoolProperty", Value: boolVal(reply.NonTextLine)},
		{Name: "bAmbient", PropType: "BoolProperty", Value: boolVal(reply.Ambient)},
		{Name: "eConvGUIStyle", PropType: "EnumProperty", Value: coalesceGUI(reply.GUIStyle)},
		{Name: "ReplyType", PropType: "EnumProperty", Value: coalesceReplyType(reply.ReplyType)},
		{Name: "Category", PropType: "NameProperty", Value: coalesceCategory(reply.Category)},
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
