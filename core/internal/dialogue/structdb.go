package dialogue

type StructPropInfo struct {
	PropType      string
	ArrayElemType string
}

var ME2StructPropInfo = map[string]map[string]StructPropInfo{
	"BioConversation": {
		"EntryList":      {PropType: "ArrayProperty", ArrayElemType: "BioDialogEntryNode"},
		"m_EntryList":    {PropType: "ArrayProperty", ArrayElemType: "BioDialogEntryNode"},
		"ReplyList":      {PropType: "ArrayProperty", ArrayElemType: "BioDialogReplyNode"},
		"m_ReplyList":    {PropType: "ArrayProperty", ArrayElemType: "BioDialogReplyNode"},
		"ReplyListNew":   {PropType: "ArrayProperty", ArrayElemType: "BioDialogReplyNode"},
		"SpeakerList":    {PropType: "ArrayProperty", ArrayElemType: "BioDialogSpeaker"},
		"m_SpeakerList":  {PropType: "ArrayProperty", ArrayElemType: "BioDialogSpeaker"},
		"m_StartingList": {PropType: "ArrayProperty", ArrayElemType: "IntProperty"},
		"StartingList":   {PropType: "ArrayProperty", ArrayElemType: "IntProperty"},
		"m_ScriptList":   {PropType: "ArrayProperty", ArrayElemType: "BioDialogScript"},
		"ScriptList":     {PropType: "ArrayProperty", ArrayElemType: "BioDialogScript"},
	},
	"BioDialogEntryNode": {
		"nIndex":                {PropType: "IntProperty"},
		"nSpeakerIndex":         {PropType: "IntProperty"},
		"srText":                {PropType: "StringRefProperty"},
		"nListenerIndex":        {PropType: "IntProperty"},
		"nConditionalFunc":      {PropType: "IntProperty"},
		"nConditionalParam":     {PropType: "IntProperty"},
		"nStateTransition":      {PropType: "IntProperty"},
		"nStateTransitionParam": {PropType: "IntProperty"},
		"nScriptIndex":          {PropType: "IntProperty"},
		"nExportID":             {PropType: "IntProperty"},
		"nCameraIntimacy":       {PropType: "IntProperty"},
		"bFireConditional":      {PropType: "BoolProperty"},
		"bSkippable":            {PropType: "BoolProperty"},
		"bIsNonTextLine":        {PropType: "BoolProperty"},
		"bAmbient":              {PropType: "BoolProperty"},
		"eConvGUIStyle":         {PropType: "EnumProperty"},
		"ReplyListNew":          {PropType: "ArrayProperty", ArrayElemType: "BioDialogReplyListDetails"},
	},
	"BioDialogReplyNode": {
		"srText":                {PropType: "StringRefProperty"},
		"nConditionalFunc":      {PropType: "IntProperty"},
		"nConditionalParam":     {PropType: "IntProperty"},
		"nStateTransition":      {PropType: "IntProperty"},
		"nStateTransitionParam": {PropType: "IntProperty"},
		"nScriptIndex":          {PropType: "IntProperty"},
		"nExportID":             {PropType: "IntProperty"},
		"nCameraIntimacy":       {PropType: "IntProperty"},
		"bFireConditional":      {PropType: "BoolProperty"},
		"bUnskippable":          {PropType: "BoolProperty"},
		"bIsNonTextLine":        {PropType: "BoolProperty"},
		"bAmbient":              {PropType: "BoolProperty"},
		"eConvGUIStyle":         {PropType: "EnumProperty"},
		"ReplyType":             {PropType: "EnumProperty"},
		"EntryList":             {PropType: "ArrayProperty", ArrayElemType: "IntProperty"},
		"Category":              {PropType: "NameProperty"},
		"nIndex":                {PropType: "IntProperty"},
		"nEntryIndex":           {PropType: "IntProperty"},
	},
	"BioDialogReplyListDetails": {
		"nIndex": {PropType: "IntProperty"},
	},
	"BioDialogSpeaker": {
		"sSpeakerTag":        {PropType: "NameProperty"},
		"nIndex":             {PropType: "IntProperty"},
		"nDisplayNameStrRef": {PropType: "StringRefProperty"},
	},
	"BioDialogScript": {
		"sScriptTag": {PropType: "NameProperty"},
	},
}

func GetStructArrayElementType(structType, propName string) string {
	if props, ok := ME2StructPropInfo[structType]; ok {
		if info, ok := props[propName]; ok {
			return info.ArrayElemType
		}
	}
	return ""
}

func IsKnownStructType(typeName string) bool {
	_, ok := ME2StructPropInfo[typeName]
	return ok
}

func GetStructPropertyType(structType, propName string) (string, bool) {
	if props, ok := ME2StructPropInfo[structType]; ok {
		if info, ok := props[propName]; ok {
			return info.PropType, true
		}
	}
	return "", false
}
