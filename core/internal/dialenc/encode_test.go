package dialenc

import (
	"testing"

	pcc "github.com/commander-spaceman/me2pcc"
	"pcc-toolkit/core/internal/dialogue"
	"pcc-toolkit/core/internal/pccenc"
)

func testEntryNames() []string {
	return []string{
		"None",
		"IntProperty", "BoolProperty", "FloatProperty", "StrProperty",
		"StringRefProperty", "NameProperty", "ObjectProperty",
		"StructProperty", "ArrayProperty", "EnumProperty", "ByteProperty",
		"BioDialogEntryNode", "BioDialogReplyNode", "BioDialogSpeaker",
		"BioDialogReplyListDetails",
		"nIndex", "nSpeakerIndex", "srText", "nListenerIndex",
		"nConditionalFunc", "nConditionalParam",
		"nStateTransition", "nStateTransitionParam",
		"nScriptIndex", "nExportID", "nCameraIntimacy",
		"bFireConditional", "bSkippable", "bIsNonTextLine",
		"bAmbient", "eConvGUIStyle",
		"CONV_GUISTYLE_DEFAULT", "CONV_GUISTYLE_NEUTRAL",
		"ReplyListNew", "srParaphrase", "sParaphrase",
		"Category", "REPLY_CATEGORY_DEFAULT", "REPLY_CATEGORY_RENEGADE",
		"bUnskippable", "ReplyType", "REPLY_TYPE_DEFAULT",
		"EntryList", "nEntryIndex",
		"sSpeakerTag", "nDisplayNameStrRef",
		"EConvGUIStyles", "EReplyCategory", "EReplyTypes",
		"henchman",
	}
}

func testReplyNames() []string {
	return testEntryNames()
}

func testSpeakerNames() []string {
	return testEntryNames()
}

func ptrInt(v int) *int    { return &v }
func ptrBool(v bool) *bool { return &v }

func TestEncodeEntryNode_AllFields(t *testing.T) {
	names := testEntryNames()
	entry := dialogue.EntryNode{
		ID:                   0,
		SpeakerID:            ptrInt(1),
		SpeakerTag:           "henchman",
		ListenerIndex:        ptrInt(-1),
		LineStrRef:           ptrInt(123456),
		ConditionalFunc:      ptrInt(100),
		ConditionalParam:     ptrInt(0),
		StateTransition:      ptrInt(-1),
		StateTransitionParam: ptrInt(0),
		ScriptIndex:          ptrInt(-1),
		ExportID:             ptrInt(42),
		CameraIntimacy:       ptrInt(0),
		FiresConditional:     ptrBool(false),
		Skippable:            ptrBool(true),
		NonTextLine:          ptrBool(false),
		Ambient:              ptrBool(false),
		GUIStyle:             "CONV_GUISTYLE_NEUTRAL",
	}

	bytes, err := EncodeEntryNodeBytes(entry, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertIntProp(t, parsed, "nIndex", 0)
	assertIntProp(t, parsed, "nSpeakerIndex", 1)
	assertIntProp(t, parsed, "srText", 123456)
	assertIntProp(t, parsed, "nListenerIndex", -1)
	assertIntProp(t, parsed, "nConditionalFunc", 100)
	assertIntProp(t, parsed, "nExportID", 42)
	assertBoolProp(t, parsed, "bSkippable", true)
	assertBoolProp(t, parsed, "bFireConditional", false)
	assertStrProp(t, parsed, "eConvGUIStyle", "CONV_GUISTYLE_NEUTRAL")
}

func TestEncodeEntryNode_NilPointers(t *testing.T) {
	names := testEntryNames()
	entry := dialogue.EntryNode{
		ID: 5,
	}

	bytes, err := EncodeEntryNodeBytes(entry, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	if sp, ok := parsed["nSpeakerIndex"]; ok {
		v, _ := sp.Value.(int)
		if v != -1 {
			t.Errorf("default nSpeakerIndex: want -1, got %d", v)
		}
	}
	if sp, ok := parsed["nConditionalFunc"]; ok {
		v, _ := sp.Value.(int)
		if v != -1 {
			t.Errorf("default nConditionalFunc: want -1, got %d", v)
		}
	}
}

func TestEncodeEntryNode_WithReplyChoices(t *testing.T) {
	names := testEntryNames()
	entry := dialogue.EntryNode{
		ID:         0,
		SpeakerID:  ptrInt(0),
		LineStrRef: ptrInt(100),
		ReplyChoices: []dialogue.ReplyChoice{
			{ToReplyID: 0, ParaphraseStrRef: ptrInt(200), Category: "REPLY_CATEGORY_RENEGADE"},
			{ToReplyID: 1, ParaphraseStrRef: ptrInt(201), Category: "REPLY_CATEGORY_DEFAULT"},
		},
		ReplyLinks: []int{0, 1},
	}

	bytes, err := EncodeEntryNodeBytes(entry, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	replyList, ok := parsed["ReplyListNew"]
	if !ok {
		t.Fatal("ReplyListNew not found")
	}
	if replyList.PropType != "ArrayProperty" {
		t.Fatalf("ReplyListNew type: want ArrayProperty, got %s", replyList.PropType)
	}

	arr, ok := replyList.Value.(map[string]interface{})
	if !ok {
		t.Fatal("ReplyListNew value not a map")
	}
	count, _ := arr["count"].(int)
	if count != 2 {
		t.Fatalf("ReplyListNew count: want 2, got %d", count)
	}

	payloadOff, _ := arr["payload_offset"].(int)
	payloadSize, _ := arr["payload_size"].(int)
	items := pcc.ParseStructArrayItemsAsPropertyCollections(bytes, names, payloadOff, payloadSize, 2)
	if len(items) != 2 {
		t.Fatalf("struct items: want 2, got %d", len(items))
	}

	item0 := items[0]
	if v, ok := item0["nIndex"]; ok {
		idx, _ := v.Value.(int)
		if idx != 0 {
			t.Errorf("item 0 nIndex: want 0, got %d", idx)
		}
	}
	if v, ok := item0["Category"]; ok {
		cat, _ := v.Value.(string)
		if cat != "REPLY_CATEGORY_RENEGADE" {
			t.Errorf("item 0 Category: want REPLAY_CATEGORY_RENEGADE, got %s", cat)
		}
	}

	item1 := items[1]
	if v, ok := item1["nIndex"]; ok {
		idx, _ := v.Value.(int)
		if idx != 1 {
			t.Errorf("item 1 nIndex: want 1, got %d", idx)
		}
	}
}

func TestEncodeReplyNode_AllFields(t *testing.T) {
	names := testReplyNames()
	reply := dialogue.ReplyNode{
		ID:                   0,
		LineStrRef:           ptrInt(654321),
		ConditionalFunc:      ptrInt(200),
		ConditionalParam:     ptrInt(0),
		StateTransition:      ptrInt(-1),
		StateTransitionParam: ptrInt(0),
		ScriptIndex:          ptrInt(-1),
		ExportID:             ptrInt(100),
		CameraIntimacy:       ptrInt(1),
		FiresConditional:     ptrBool(true),
		Unskippable:          ptrBool(false),
		NonTextLine:          ptrBool(false),
		Ambient:              ptrBool(false),
		GUIStyle:             "CONV_GUISTYLE_DEFAULT",
		ReplyType:            "REPLY_TYPE_DEFAULT",
		Category:             "REPLY_CATEGORY_DEFAULT",
		TargetEntryIDs:       []int{0, 1},
	}

	bytes, err := EncodeReplyNodeBytes(reply, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertIntProp(t, parsed, "nIndex", 0)
	assertIntProp(t, parsed, "srText", 654321)
	assertIntProp(t, parsed, "nConditionalFunc", 200)
	assertIntProp(t, parsed, "nExportID", 100)
	assertIntProp(t, parsed, "nCameraIntimacy", 1)
	assertBoolProp(t, parsed, "bFireConditional", true)
	assertBoolProp(t, parsed, "bUnskippable", false)
	assertStrProp(t, parsed, "eConvGUIStyle", "CONV_GUISTYLE_DEFAULT")
	assertStrProp(t, parsed, "ReplyType", "REPLY_TYPE_DEFAULT")
	assertStrProp(t, parsed, "Category", "REPLY_CATEGORY_DEFAULT")

	entryList, ok := parsed["EntryList"]
	if !ok {
		t.Fatal("EntryList not found in reply")
	}
	arr, ok := entryList.Value.(map[string]interface{})
	if !ok {
		t.Fatal("EntryList value not a map")
	}
	count, _ := arr["count"].(int)
	if count != 2 {
		t.Errorf("EntryList count: want 2, got %d", count)
	}
}

func TestEncodeSpeaker_AllFields(t *testing.T) {
	names := testSpeakerNames()
	speaker := dialogue.Speaker{
		ID:           0,
		Tag:          "henchman",
		DisplayName:  "Henchman",
		StrRefID:     ptrInt(789012),
		FriendlyName: "Henchman",
	}

	bytes, err := EncodeSpeakerBytes(speaker, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertIntProp(t, parsed, "nIndex", 0)
	assertStrProp(t, parsed, "sSpeakerTag", "henchman")
	assertIntProp(t, parsed, "nDisplayNameStrRef", 789012)
}

func TestEncodeSpeaker_NoTag(t *testing.T) {
	names := testSpeakerNames()
	speaker := dialogue.Speaker{
		ID:       2,
		StrRefID: ptrInt(0),
	}

	bytes, err := EncodeSpeakerBytes(speaker, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	assertStrProp(t, parsed, "sSpeakerTag", "None")
}

func TestEncodeReplyChoice(t *testing.T) {
	names := testEntryNames()
	choice := dialogue.ReplyChoice{
		ToReplyID:        2,
		ParaphraseStrRef: ptrInt(300),
		Paraphrase:       "Let me think...",
		Category:         "REPLY_CATEGORY_DEFAULT",
	}

	pv := EncodeReplyChoice(choice, names)
	bytes, err := pccenc.EncodePropertyCollection(pv.Properties, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertIntProp(t, parsed, "nIndex", 2)
	assertIntProp(t, parsed, "srParaphrase", 300)
	assertStrProp(t, parsed, "sParaphrase", "Let me think...")
	assertStrProp(t, parsed, "Category", "REPLY_CATEGORY_DEFAULT")
}

func TestEntryNodeByteSize(t *testing.T) {
	names := testEntryNames()
	entry := dialogue.EntryNode{
		ID:         10,
		SpeakerID:  ptrInt(0),
		LineStrRef: ptrInt(999),
		ExportID:   ptrInt(50),
	}

	bytes, err := EncodeEntryNodeBytes(entry, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	if len(bytes) < 100 {
		t.Errorf("entry bytes too small: %d bytes", len(bytes))
	}

	var parsedBytes []byte
	for i := 0; i < 2; i++ {
		entry2 := dialogue.EntryNode{
			ID:         10 + i,
			SpeakerID:  ptrInt(0),
			LineStrRef: ptrInt(999 + i),
			ExportID:   ptrInt(50),
		}
		b2, _ := EncodeEntryNodeBytes(entry2, names)
		parsedBytes = b2
	}
	_ = parsedBytes

	t.Logf("entry node bytes: %d", len(bytes))
}

func assertIntProp(t *testing.T, props map[string]pcc.ParsedProperty, name string, want int) {
	t.Helper()
	p, ok := props[name]
	if !ok {
		t.Errorf("property %q not found", name)
		return
	}
	v, ok := p.Value.(int)
	if !ok {
		t.Errorf("property %q: want int, got %T (%v)", name, p.Value, p.Value)
		return
	}
	if v != want {
		t.Errorf("property %q: want %d, got %d", name, want, v)
	}
}

func assertBoolProp(t *testing.T, props map[string]pcc.ParsedProperty, name string, want bool) {
	t.Helper()
	p, ok := props[name]
	if !ok {
		t.Errorf("property %q not found", name)
		return
	}
	v, ok := p.Value.(bool)
	if !ok {
		t.Errorf("property %q: want bool, got %T", name, p.Value)
		return
	}
	if v != want {
		t.Errorf("property %q: want %v, got %v", name, want, v)
	}
}

func assertStrProp(t *testing.T, props map[string]pcc.ParsedProperty, name string, want string) {
	t.Helper()
	p, ok := props[name]
	if !ok {
		t.Errorf("property %q not found", name)
		return
	}
	v, ok := p.Value.(string)
	if !ok {
		t.Errorf("property %q: want string, got %T (%v)", name, p.Value, p.Value)
		return
	}
	if v != want {
		t.Errorf("property %q: want %q, got %q", name, want, v)
	}
}
