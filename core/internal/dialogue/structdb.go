package dialogue

type StructField struct {
	Name     string
	PropType string
	RefType  string
}

type StructLayout struct {
	TypeName string
	Fields   []StructField
}

var ME2StructLayouts = map[string]StructLayout{
	"BioDialogEntryNode": {
		TypeName: "BioDialogEntryNode",
		Fields: []StructField{
			{Name: "nIndex", PropType: "IntProperty"},
			{Name: "nSpeakerIndex", PropType: "IntProperty"},
			{Name: "srText", PropType: "StringRefProperty"},
			{Name: "ReplyListNew", PropType: "ArrayProperty", RefType: "BioDialogReplyListDetails"},
		},
	},
	"BioDialogReplyNode": {
		TypeName: "BioDialogReplyNode",
		Fields: []StructField{
			{Name: "srText", PropType: "StringRefProperty"},
			{Name: "nConditionalFunc", PropType: "IntProperty"},
			{Name: "nConditionalParam", PropType: "IntProperty"},
			{Name: "nStateTransition", PropType: "IntProperty"},
			{Name: "bFireConditional", PropType: "BoolProperty"},
			{Name: "ReplyType", PropType: "EnumProperty"},
			{Name: "EntryList", PropType: "ArrayProperty", RefType: "IntProperty"},
		},
	},
	"BioDialogReplyListDetails": {
		TypeName: "BioDialogReplyListDetails",
		Fields: []StructField{
			{Name: "nIndex", PropType: "IntProperty"},
		},
	},
	"BioDialogSpeaker": {
		TypeName: "BioDialogSpeaker",
		Fields: []StructField{
			{Name: "sSpeakerTag", PropType: "NameProperty"},
		},
	},
}

func GetStructLayout(typeName string) (StructLayout, bool) {
	layout, ok := ME2StructLayouts[typeName]
	return layout, ok
}
