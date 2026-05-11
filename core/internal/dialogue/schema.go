package dialogue

type ConversationListSchema struct {
	EntryHeadI32                int
	ReplyHeadI32                int
	SpeakerHeadI32              int
	EntryListenerTagNameIdxCol  int
	ReplyConditionStartCol      int
	SpeakerDisplayNameStrRefCol int
}

var SchemaByProfile = map[string]ConversationListSchema{
	"me2_ot": {
		EntryHeadI32:                3,
		ReplyHeadI32:                3,
		SpeakerHeadI32:              3,
		EntryListenerTagNameIdxCol:  3,
		ReplyConditionStartCol:      3,
		SpeakerDisplayNameStrRefCol: 2,
	},
	"le2": {
		EntryHeadI32:                3,
		ReplyHeadI32:                3,
		SpeakerHeadI32:              3,
		EntryListenerTagNameIdxCol:  3,
		ReplyConditionStartCol:      3,
		SpeakerDisplayNameStrRefCol: 2,
	},
	"me3_ot": {
		EntryHeadI32:   3,
		ReplyHeadI32:   3,
		SpeakerHeadI32: 3,
	},
	"le3": {
		EntryHeadI32:   3,
		ReplyHeadI32:   3,
		SpeakerHeadI32: 3,
	},
}

func GetSchemaForProfile(profile string) ConversationListSchema {
	if s, ok := SchemaByProfile[profile]; ok {
		return s
	}
	return ConversationListSchema{EntryHeadI32: 3, ReplyHeadI32: 3, SpeakerHeadI32: 3}
}
