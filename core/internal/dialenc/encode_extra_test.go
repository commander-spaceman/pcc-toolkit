package dialenc

import (
	"testing"

	"pcc-toolkit/core/internal/dialogue"
	"pcc-toolkit/core/internal/pcc"
	"pcc-toolkit/core/internal/pccenc"
)

func assertPropAbsent(t *testing.T, props map[string]pcc.ParsedProperty, name string) {
	t.Helper()
	if _, ok := props[name]; ok {
		t.Errorf("property %q should be absent but was found", name)
	}
}

// 1. EntryNode with ALL fields populated
func TestEncodeExtra_EntryNodeAllFields(t *testing.T) {
	names := testEntryNames()
	entry := dialogue.EntryNode{
		ID:                   7,
		SpeakerID:            ptrInt(2),
		SpeakerTag:           "commander",
		ListenerIndex:        ptrInt(3),
		ListenerTag:          "henchman",
		LineStrRef:           ptrInt(700001),
		LineText:             "I should go.",
		LineStatus:           "Complete",
		ConditionalFunc:      ptrInt(501),
		ConditionalParam:     ptrInt(1),
		StateTransition:      ptrInt(2),
		StateTransitionParam: ptrInt(3),
		ScriptIndex:          ptrInt(99),
		ScriptName:           "SomeScript",
		FiresConditional:     ptrBool(true),
		ExportID:             ptrInt(42),
		Skippable:            ptrBool(true),
		NonTextLine:          ptrBool(false),
		Ambient:              ptrBool(false),
		CameraIntimacy:       ptrInt(1),
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

	assertIntProp(t, parsed, "nIndex", 7)
	assertIntProp(t, parsed, "nSpeakerIndex", 2)
	assertIntProp(t, parsed, "srText", 700001)
	assertIntProp(t, parsed, "nListenerIndex", 3)
	assertIntProp(t, parsed, "nConditionalFunc", 501)
	assertIntProp(t, parsed, "nConditionalParam", 1)
	assertIntProp(t, parsed, "nStateTransition", 2)
	assertIntProp(t, parsed, "nStateTransitionParam", 3)
	assertIntProp(t, parsed, "nScriptIndex", 99)
	assertIntProp(t, parsed, "nExportID", 42)
	assertIntProp(t, parsed, "nCameraIntimacy", 1)
	assertBoolProp(t, parsed, "bFireConditional", true)
	assertBoolProp(t, parsed, "bSkippable", true)
	assertBoolProp(t, parsed, "bIsNonTextLine", false)
	assertBoolProp(t, parsed, "bAmbient", false)
	assertStrProp(t, parsed, "eConvGUIStyle", "CONV_GUISTYLE_NEUTRAL")
}

// 2. EntryNode with only required fields (minimal)
func TestEncodeExtra_EntryNodeMinimal(t *testing.T) {
	names := testEntryNames()
	entry := dialogue.EntryNode{
		ID: 0,
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
	assertIntProp(t, parsed, "nSpeakerIndex", unusedInt)
	assertIntProp(t, parsed, "srText", unusedInt)

	assertPropAbsent(t, parsed, "nListenerIndex")
	assertPropAbsent(t, parsed, "nConditionalFunc")
	assertPropAbsent(t, parsed, "nScriptIndex")
	assertPropAbsent(t, parsed, "nExportID")
	assertPropAbsent(t, parsed, "bSkippable")
	assertPropAbsent(t, parsed, "eConvGUIStyle")
}

// 3. EntryNode with nil pointer fields should be omitted
func TestEncodeExtra_EntryNodeNilPointerOmission(t *testing.T) {
	names := testEntryNames()
	entry := dialogue.EntryNode{
		ID:         1,
		SpeakerID:  ptrInt(0),
		LineStrRef: ptrInt(100),
	}

	bytes, err := EncodeEntryNodeBytes(entry, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertPropAbsent(t, parsed, "nListenerIndex")
	assertPropAbsent(t, parsed, "nConditionalFunc")
	assertPropAbsent(t, parsed, "nConditionalParam")
	assertPropAbsent(t, parsed, "nStateTransition")
	assertPropAbsent(t, parsed, "nStateTransitionParam")
	assertPropAbsent(t, parsed, "nScriptIndex")
	assertPropAbsent(t, parsed, "nExportID")
	assertPropAbsent(t, parsed, "nCameraIntimacy")
	assertPropAbsent(t, parsed, "bFireConditional")
	assertPropAbsent(t, parsed, "bSkippable")
	assertPropAbsent(t, parsed, "bIsNonTextLine")
	assertPropAbsent(t, parsed, "bAmbient")
	assertPropAbsent(t, parsed, "eConvGUIStyle")
	assertPropAbsent(t, parsed, "ReplyListNew")
}

// 4. EntryNode with multiple reply choices
func TestEncodeExtra_EntryNodeMultipleReplyChoices(t *testing.T) {
	names := testEntryNames()
	entry := dialogue.EntryNode{
		ID:         10,
		SpeakerID:  ptrInt(1),
		LineStrRef: ptrInt(500),
		ReplyChoices: []dialogue.ReplyChoice{
			{ToReplyID: 0, ParaphraseStrRef: ptrInt(600), Paraphrase: "Option A", Category: "REPLY_CATEGORY_DEFAULT"},
			{ToReplyID: 1, ParaphraseStrRef: ptrInt(601), Paraphrase: "Option B", Category: "REPLY_CATEGORY_RENEGADE"},
			{ToReplyID: 2, ParaphraseStrRef: ptrInt(602), Paraphrase: "Option C", Category: "REPLY_CATEGORY_DEFAULT"},
		},
		ReplyLinks: []int{0, 1, 2},
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
	arr, ok := replyList.Value.(map[string]interface{})
	if !ok {
		t.Fatal("ReplyListNew value not a map")
	}
	count, _ := arr["count"].(int)
	if count != 3 {
		t.Fatalf("ReplyListNew count: want 3, got %d", count)
	}

	payloadOff, _ := arr["payload_offset"].(int)
	payloadSize, _ := arr["payload_size"].(int)
	items := pcc.ParseStructArrayItemsAsPropertyCollections(bytes, names, payloadOff, payloadSize, 3)
	if len(items) != 3 {
		t.Fatalf("struct items: want 3, got %d", len(items))
	}

	if v, ok := items[0]["nIndex"]; ok {
		idx, _ := v.Value.(int)
		if idx != 0 {
			t.Errorf("item 0 nIndex: want 0, got %d", idx)
		}
	}
	if v, ok := items[1]["nIndex"]; ok {
		idx, _ := v.Value.(int)
		if idx != 1 {
			t.Errorf("item 1 nIndex: want 1, got %d", idx)
		}
	}
	if v, ok := items[2]["nIndex"]; ok {
		idx, _ := v.Value.(int)
		if idx != 2 {
			t.Errorf("item 2 nIndex: want 2, got %d", idx)
		}
	}
}

// 5. EntryNode with reply links but no reply choices
func TestEncodeExtra_EntryNodeReplyLinksNoChoices(t *testing.T) {
	names := testEntryNames()
	entry := dialogue.EntryNode{
		ID:         3,
		SpeakerID:  ptrInt(1),
		LineStrRef: ptrInt(300),
		ReplyLinks: []int{5, 6, 7},
	}

	bytes, err := EncodeEntryNodeBytes(entry, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertPropAbsent(t, parsed, "ReplyListNew")
	assertIntProp(t, parsed, "nIndex", 3)
	assertIntProp(t, parsed, "srText", 300)
}

// 6. EntryNode byte round-trip
func TestEncodeExtra_EntryNodeBytesRoundTrip(t *testing.T) {
	names := testEntryNames()
	entry := dialogue.EntryNode{
		ID:               15,
		SpeakerID:        ptrInt(4),
		ListenerIndex:    ptrInt(-1),
		LineStrRef:       ptrInt(987654),
		ConditionalFunc:  ptrInt(200),
		StateTransition:  ptrInt(10),
		ScriptIndex:      ptrInt(5),
		ExportID:         ptrInt(77),
		CameraIntimacy:   ptrInt(2),
		FiresConditional: ptrBool(true),
		Skippable:        ptrBool(false),
		NonTextLine:      ptrBool(true),
		Ambient:          ptrBool(true),
		GUIStyle:         "CONV_GUISTYLE_DEFAULT",
	}

	bytes, err := EncodeEntryNodeBytes(entry, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertIntProp(t, parsed, "nIndex", 15)
	assertIntProp(t, parsed, "nSpeakerIndex", 4)
	assertIntProp(t, parsed, "srText", 987654)
	assertIntProp(t, parsed, "nListenerIndex", -1)
	assertIntProp(t, parsed, "nConditionalFunc", 200)
	assertIntProp(t, parsed, "nStateTransition", 10)
	assertIntProp(t, parsed, "nScriptIndex", 5)
	assertIntProp(t, parsed, "nExportID", 77)
	assertIntProp(t, parsed, "nCameraIntimacy", 2)
	assertBoolProp(t, parsed, "bFireConditional", true)
	assertBoolProp(t, parsed, "bSkippable", false)
	assertBoolProp(t, parsed, "bIsNonTextLine", true)
	assertBoolProp(t, parsed, "bAmbient", true)
	assertStrProp(t, parsed, "eConvGUIStyle", "CONV_GUISTYLE_DEFAULT")
}

// 7. ReplyNode with ALL fields populated
func TestEncodeExtra_ReplyNodeAllFields(t *testing.T) {
	names := testReplyNames()
	reply := dialogue.ReplyNode{
		ID:                   5,
		LineStrRef:           ptrInt(888888),
		LineText:             "What do you want?",
		LineStatus:           "Complete",
		TargetEntryIDs:       []int{20, 21},
		ConditionRefs:        []string{"cond_a", "cond_b"},
		Category:             "REPLY_CATEGORY_RENEGADE",
		ReplyType:            "REPLY_TYPE_DEFAULT",
		ConditionalFunc:      ptrInt(300),
		ConditionalParam:     ptrInt(1),
		StateTransition:      ptrInt(4),
		StateTransitionParam: ptrInt(0),
		ScriptIndex:          ptrInt(10),
		ScriptName:           "ReplyScript",
		FiresConditional:     ptrBool(false),
		ExportID:             ptrInt(55),
		Unskippable:          ptrBool(true),
		NonTextLine:          ptrBool(false),
		Ambient:              ptrBool(true),
		CameraIntimacy:       ptrInt(3),
		GUIStyle:             "CONV_GUISTYLE_NEUTRAL",
	}

	bytes, err := EncodeReplyNodeBytes(reply, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertIntProp(t, parsed, "nIndex", 5)
	assertIntProp(t, parsed, "srText", 888888)
	assertIntProp(t, parsed, "nConditionalFunc", 300)
	assertIntProp(t, parsed, "nConditionalParam", 1)
	assertIntProp(t, parsed, "nStateTransition", 4)
	assertIntProp(t, parsed, "nStateTransitionParam", 0)
	assertIntProp(t, parsed, "nScriptIndex", 10)
	assertIntProp(t, parsed, "nExportID", 55)
	assertIntProp(t, parsed, "nCameraIntimacy", 3)
	assertBoolProp(t, parsed, "bFireConditional", false)
	assertBoolProp(t, parsed, "bUnskippable", true)
	assertBoolProp(t, parsed, "bIsNonTextLine", false)
	assertBoolProp(t, parsed, "bAmbient", true)
	assertStrProp(t, parsed, "eConvGUIStyle", "CONV_GUISTYLE_NEUTRAL")
	assertStrProp(t, parsed, "ReplyType", "REPLY_TYPE_DEFAULT")
	assertStrProp(t, parsed, "Category", "REPLY_CATEGORY_RENEGADE")

	entryList, ok := parsed["EntryList"]
	if !ok {
		t.Fatal("EntryList not found")
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

// 8. ReplyNode minimal fields
func TestEncodeExtra_ReplyNodeMinimal(t *testing.T) {
	names := testReplyNames()
	reply := dialogue.ReplyNode{
		ID: 0,
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
	assertIntProp(t, parsed, "srText", unusedInt)

	assertPropAbsent(t, parsed, "nConditionalFunc")
	assertPropAbsent(t, parsed, "nExportID")
	assertPropAbsent(t, parsed, "bUnskippable")
	assertPropAbsent(t, parsed, "eConvGUIStyle")
	assertPropAbsent(t, parsed, "ReplyType")
	assertPropAbsent(t, parsed, "Category")
	assertPropAbsent(t, parsed, "EntryList")
}

// 9. ReplyNode with multiple target entry IDs
func TestEncodeExtra_ReplyNodeTargetEntryIDs(t *testing.T) {
	names := testReplyNames()
	reply := dialogue.ReplyNode{
		ID:             12,
		LineStrRef:     ptrInt(400),
		TargetEntryIDs: []int{100, 200, 300, 400},
	}

	bytes, err := EncodeReplyNodeBytes(reply, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	entryList, ok := parsed["EntryList"]
	if !ok {
		t.Fatal("EntryList not found")
	}
	arr, ok := entryList.Value.(map[string]interface{})
	if !ok {
		t.Fatal("EntryList value not a map")
	}
	count, _ := arr["count"].(int)
	if count != 4 {
		t.Fatalf("EntryList count: want 4, got %d", count)
	}
}

// 10. ReplyNode nil pointer fields omitted
func TestEncodeExtra_ReplyNodeNilPointerOmission(t *testing.T) {
	names := testReplyNames()
	reply := dialogue.ReplyNode{
		ID:         3,
		LineStrRef: ptrInt(777),
	}

	bytes, err := EncodeReplyNodeBytes(reply, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertPropAbsent(t, parsed, "nConditionalFunc")
	assertPropAbsent(t, parsed, "nConditionalParam")
	assertPropAbsent(t, parsed, "nStateTransition")
	assertPropAbsent(t, parsed, "nStateTransitionParam")
	assertPropAbsent(t, parsed, "nScriptIndex")
	assertPropAbsent(t, parsed, "nExportID")
	assertPropAbsent(t, parsed, "nCameraIntimacy")
	assertPropAbsent(t, parsed, "bFireConditional")
	assertPropAbsent(t, parsed, "bUnskippable")
	assertPropAbsent(t, parsed, "bIsNonTextLine")
	assertPropAbsent(t, parsed, "bAmbient")
	assertPropAbsent(t, parsed, "eConvGUIStyle")
	assertPropAbsent(t, parsed, "ReplyType")
	assertPropAbsent(t, parsed, "Category")
	assertPropAbsent(t, parsed, "EntryList")
}

// 11. ReplyNode byte round-trip
func TestEncodeExtra_ReplyNodeBytesRoundTrip(t *testing.T) {
	names := testReplyNames()
	reply := dialogue.ReplyNode{
		ID:               99,
		LineStrRef:       ptrInt(111111),
		TargetEntryIDs:   []int{50, 51, 52},
		Category:         "REPLY_CATEGORY_DEFAULT",
		ReplyType:        "REPLY_TYPE_DEFAULT",
		ConditionalFunc:  ptrInt(900),
		StateTransition:  ptrInt(7),
		ScriptIndex:      ptrInt(3),
		ExportID:         ptrInt(88),
		CameraIntimacy:   ptrInt(0),
		FiresConditional: ptrBool(true),
		Unskippable:      ptrBool(false),
		NonTextLine:      ptrBool(false),
		Ambient:          ptrBool(false),
		GUIStyle:         "CONV_GUISTYLE_DEFAULT",
	}

	bytes, err := EncodeReplyNodeBytes(reply, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertIntProp(t, parsed, "nIndex", 99)
	assertIntProp(t, parsed, "srText", 111111)
	assertIntProp(t, parsed, "nConditionalFunc", 900)
	assertIntProp(t, parsed, "nStateTransition", 7)
	assertIntProp(t, parsed, "nScriptIndex", 3)
	assertIntProp(t, parsed, "nExportID", 88)
	assertIntProp(t, parsed, "nCameraIntimacy", 0)
	assertBoolProp(t, parsed, "bFireConditional", true)
	assertBoolProp(t, parsed, "bUnskippable", false)
	assertBoolProp(t, parsed, "bIsNonTextLine", false)
	assertBoolProp(t, parsed, "bAmbient", false)
	assertStrProp(t, parsed, "eConvGUIStyle", "CONV_GUISTYLE_DEFAULT")
	assertStrProp(t, parsed, "ReplyType", "REPLY_TYPE_DEFAULT")
	assertStrProp(t, parsed, "Category", "REPLY_CATEGORY_DEFAULT")
}

// 12. Speaker with all fields
func TestEncodeExtra_SpeakerAllFields(t *testing.T) {
	names := testSpeakerNames()
	speaker := dialogue.Speaker{
		ID:                3,
		Tag:               "henchman",
		DisplayName:       "Garrus Vakarian",
		StrRefID:          ptrInt(550012),
		FriendlyName:      "Garrus",
		FaceFXMaleAnimSet: ptrInt(1),
		FaceFXFemAnimSet:  ptrInt(2),
	}

	bytes, err := EncodeSpeakerBytes(speaker, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertIntProp(t, parsed, "nIndex", 3)
	assertStrProp(t, parsed, "sSpeakerTag", "henchman")
	assertIntProp(t, parsed, "nDisplayNameStrRef", 550012)
}

// 13. Speaker with no tag defaults to "None"
func TestEncodeExtra_SpeakerNoTagDefault(t *testing.T) {
	names := testSpeakerNames()
	speaker := dialogue.Speaker{
		ID:       7,
		StrRefID: ptrInt(100),
	}

	bytes, err := EncodeSpeakerBytes(speaker, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertIntProp(t, parsed, "nIndex", 7)
	assertStrProp(t, parsed, "sSpeakerTag", "None")
	assertIntProp(t, parsed, "nDisplayNameStrRef", 100)
}

// 14. Speaker with nil StrRefID
func TestEncodeExtra_SpeakerNilStrRefID(t *testing.T) {
	names := testSpeakerNames()
	speaker := dialogue.Speaker{
		ID:  4,
		Tag: "henchman",
	}

	bytes, err := EncodeSpeakerBytes(speaker, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertIntProp(t, parsed, "nIndex", 4)
	assertStrProp(t, parsed, "sSpeakerTag", "henchman")
	assertIntProp(t, parsed, "nDisplayNameStrRef", unusedInt)
}

// 15. Speaker byte round-trip
func TestEncodeExtra_SpeakerBytesRoundTrip(t *testing.T) {
	names := testSpeakerNames()
	speaker := dialogue.Speaker{
		ID:           6,
		Tag:          "henchman",
		DisplayName:  "Tali'Zorah",
		StrRefID:     ptrInt(660066),
		FriendlyName: "Tali",
	}

	bytes, err := EncodeSpeakerBytes(speaker, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertIntProp(t, parsed, "nIndex", 6)
	assertStrProp(t, parsed, "sSpeakerTag", "henchman")
	assertIntProp(t, parsed, "nDisplayNameStrRef", 660066)
}

// 16. ReplyChoice with all fields
func TestEncodeExtra_ReplyChoiceAllFields(t *testing.T) {
	names := testEntryNames()
	choice := dialogue.ReplyChoice{
		FromEntryID:      5,
		ToReplyID:        8,
		Order:            1,
		Paraphrase:       "I'll handle this.",
		ParaphraseStrRef: ptrInt(999001),
		ParaphraseText:   "I'll handle this.",
		Category:         "REPLY_CATEGORY_RENEGADE",
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

	assertIntProp(t, parsed, "nIndex", 8)
	assertIntProp(t, parsed, "srParaphrase", 999001)
	assertStrProp(t, parsed, "sParaphrase", "I'll handle this.")
	assertStrProp(t, parsed, "Category", "REPLY_CATEGORY_RENEGADE")
}

// 17. ReplyChoice with only required fields
func TestEncodeExtra_ReplyChoiceMinimal(t *testing.T) {
	names := testEntryNames()
	choice := dialogue.ReplyChoice{
		ToReplyID: 3,
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

	assertIntProp(t, parsed, "nIndex", 3)
	assertStrProp(t, parsed, "Category", "REPLY_CATEGORY_DEFAULT")
	assertIntProp(t, parsed, "srParaphrase", unusedInt)
}

// 18. EntryNode byte size comparison (minimal vs all fields)
func TestEncodeExtra_EntryNodeByteSizeComparison(t *testing.T) {
	names := testEntryNames()

	minimal := dialogue.EntryNode{ID: 0}
	minBytes, err := EncodeEntryNodeBytes(minimal, names)
	if err != nil {
		t.Fatalf("minimal encode: %v", err)
	}

	full := dialogue.EntryNode{
		ID:                   1,
		SpeakerID:            ptrInt(2),
		ListenerIndex:        ptrInt(3),
		LineStrRef:           ptrInt(999),
		ConditionalFunc:      ptrInt(100),
		ConditionalParam:     ptrInt(1),
		StateTransition:      ptrInt(2),
		StateTransitionParam: ptrInt(0),
		ScriptIndex:          ptrInt(10),
		ExportID:             ptrInt(50),
		CameraIntimacy:       ptrInt(1),
		FiresConditional:     ptrBool(true),
		Skippable:            ptrBool(true),
		NonTextLine:          ptrBool(false),
		Ambient:              ptrBool(false),
		GUIStyle:             "CONV_GUISTYLE_NEUTRAL",
	}
	fullBytes, err := EncodeEntryNodeBytes(full, names)
	if err != nil {
		t.Fatalf("full encode: %v", err)
	}

	if len(fullBytes) <= len(minBytes) {
		t.Errorf("full entry (%d bytes) should be larger than minimal (%d bytes)", len(fullBytes), len(minBytes))
	}
	t.Logf("minimal entry: %d bytes, full entry: %d bytes", len(minBytes), len(fullBytes))
}

// 19. Multiple entries encode correctly (3+ entries)
func TestEncodeExtra_MultipleEntries(t *testing.T) {
	names := testEntryNames()
	entries := []dialogue.EntryNode{
		{ID: 0, SpeakerID: ptrInt(1), LineStrRef: ptrInt(100)},
		{ID: 1, SpeakerID: ptrInt(1), LineStrRef: ptrInt(101)},
		{ID: 2, SpeakerID: ptrInt(2), LineStrRef: ptrInt(102)},
	}

	for i, entry := range entries {
		bytes, err := EncodeEntryNodeBytes(entry, names)
		if err != nil {
			t.Fatalf("entry %d encode: %v", i, err)
		}
		parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
		if parsed == nil {
			t.Fatalf("entry %d ParsePropertyCollection returned nil", i)
		}
		assertIntProp(t, parsed, "nIndex", i)
		assertIntProp(t, parsed, "srText", 100+i)
	}
}

// 20. Multiple replies encode correctly
func TestEncodeExtra_MultipleReplies(t *testing.T) {
	names := testReplyNames()
	replies := []dialogue.ReplyNode{
		{ID: 0, LineStrRef: ptrInt(200)},
		{ID: 1, LineStrRef: ptrInt(201)},
		{ID: 2, LineStrRef: ptrInt(202)},
	}

	for i, reply := range replies {
		bytes, err := EncodeReplyNodeBytes(reply, names)
		if err != nil {
			t.Fatalf("reply %d encode: %v", i, err)
		}
		parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
		if parsed == nil {
			t.Fatalf("reply %d ParsePropertyCollection returned nil", i)
		}
		assertIntProp(t, parsed, "nIndex", i)
		assertIntProp(t, parsed, "srText", 200+i)
	}
}

// 21. Mixed entry and reply in collection
func TestEncodeExtra_MixedEntryReplyInCollection(t *testing.T) {
	names := testEntryNames()
	entry := dialogue.EntryNode{
		ID:         10,
		SpeakerID:  ptrInt(0),
		LineStrRef: ptrInt(500),
	}
	reply := dialogue.ReplyNode{
		ID:         20,
		LineStrRef: ptrInt(501),
	}

	entryProps, _ := EncodeEntryNode(entry, names)
	replyProps, _ := EncodeReplyNode(reply, names)

	allProps := make([]pccenc.PropertyValue, 0, len(entryProps)+len(replyProps))
	allProps = append(allProps, entryProps...)
	allProps = append(allProps, replyProps...)

	bytes, err := pccenc.EncodePropertyCollection(allProps, names)
	if err != nil {
		t.Fatalf("encode collection: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertIntProp(t, parsed, "nIndex", 20)
	assertIntProp(t, parsed, "srText", 501)

	if len(parsed) == 0 {
		t.Error("expected parsed properties in mixed collection")
	}
}

// 22. EntryNode with GUIStyle set to non-default
func TestEncodeExtra_EntryNodeGUIStyleNonDefault(t *testing.T) {
	names := testEntryNames()
	entry := dialogue.EntryNode{
		ID:       1,
		GUIStyle: "CONV_GUISTYLE_NEUTRAL",
	}

	bytes, err := EncodeEntryNodeBytes(entry, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertStrProp(t, parsed, "eConvGUIStyle", "CONV_GUISTYLE_NEUTRAL")
}

// 23. ReplyNode with Category and ReplyType
func TestEncodeExtra_ReplyNodeCategoryReplyType(t *testing.T) {
	names := testReplyNames()
	reply := dialogue.ReplyNode{
		ID:         8,
		LineStrRef: ptrInt(800),
		Category:   "REPLY_CATEGORY_RENEGADE",
		ReplyType:  "REPLY_TYPE_DEFAULT",
	}

	bytes, err := EncodeReplyNodeBytes(reply, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertIntProp(t, parsed, "nIndex", 8)
	assertStrProp(t, parsed, "Category", "REPLY_CATEGORY_RENEGADE")
	assertStrProp(t, parsed, "ReplyType", "REPLY_TYPE_DEFAULT")
}

// 24. EntryNode ExportID, Skippable, CameraIntimacy non-default
func TestEncodeExtra_EntryNodeExportSkippableCamera(t *testing.T) {
	names := testEntryNames()
	entry := dialogue.EntryNode{
		ID:             9,
		ExportID:       ptrInt(123),
		Skippable:      ptrBool(true),
		CameraIntimacy: ptrInt(5),
	}

	bytes, err := EncodeEntryNodeBytes(entry, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertIntProp(t, parsed, "nIndex", 9)
	assertIntProp(t, parsed, "nExportID", 123)
	assertBoolProp(t, parsed, "bSkippable", true)
	assertIntProp(t, parsed, "nCameraIntimacy", 5)
}

// 25. EntryNode with ConditionalFunc and StateTransition set
func TestEncodeExtra_EntryNodeConditionalStateTransition(t *testing.T) {
	names := testEntryNames()
	entry := dialogue.EntryNode{
		ID:                   11,
		ConditionalFunc:      ptrInt(777),
		ConditionalParam:     ptrInt(3),
		StateTransition:      ptrInt(99),
		StateTransitionParam: ptrInt(1),
		FiresConditional:     ptrBool(true),
	}

	bytes, err := EncodeEntryNodeBytes(entry, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertIntProp(t, parsed, "nIndex", 11)
	assertIntProp(t, parsed, "nConditionalFunc", 777)
	assertIntProp(t, parsed, "nConditionalParam", 3)
	assertIntProp(t, parsed, "nStateTransition", 99)
	assertIntProp(t, parsed, "nStateTransitionParam", 1)
	assertBoolProp(t, parsed, "bFireConditional", true)
}

// 26. EncodeMemberCollection tests

func TestEncodeExtra_EntryNodeBoolVariant(t *testing.T) {
	entryTrue := dialogue.EntryNode{
		ID:        1,
		Ambient:   ptrBool(true),
		Skippable: ptrBool(true),
	}
	bytes, _ := EncodeEntryNodeBytes(entryTrue, testEntryNames())
	parsedTrue, _ := pcc.ParsePropertyCollection(bytes, testEntryNames(), 0, len(bytes))
	assertBoolProp(t, parsedTrue, "bAmbient", true)
	assertBoolProp(t, parsedTrue, "bSkippable", true)

	entryFalse := dialogue.EntryNode{
		ID:        2,
		Ambient:   ptrBool(false),
		Skippable: ptrBool(false),
	}
	bytes2, _ := EncodeEntryNodeBytes(entryFalse, testEntryNames())
	parsedFalse, _ := pcc.ParsePropertyCollection(bytes2, testEntryNames(), 0, len(bytes2))
	assertBoolProp(t, parsedFalse, "bAmbient", false)
	assertBoolProp(t, parsedFalse, "bSkippable", false)
}

func TestEncodeExtra_ReplyNodeCategoryDefaultOmitted(t *testing.T) {
	names := testReplyNames()
	reply := dialogue.ReplyNode{
		ID:         1,
		LineStrRef: ptrInt(100),
		Category:   "",
	}

	bytes, err := EncodeReplyNodeBytes(reply, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertPropAbsent(t, parsed, "Category")
}

func TestEncodeExtra_ReplyNodeGUIStyleDefaultOmitted(t *testing.T) {
	names := testReplyNames()
	reply := dialogue.ReplyNode{
		ID:         2,
		LineStrRef: ptrInt(200),
		GUIStyle:   "",
	}

	bytes, err := EncodeReplyNodeBytes(reply, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertPropAbsent(t, parsed, "eConvGUIStyle")
}

func TestEncodeExtra_EntryNodeEmptyGUIStyleOmitted(t *testing.T) {
	names := testEntryNames()
	entry := dialogue.EntryNode{
		ID:       4,
		GUIStyle: "",
	}

	bytes, err := EncodeEntryNodeBytes(entry, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(bytes, names, 0, len(bytes))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	assertPropAbsent(t, parsed, "eConvGUIStyle")
}

func TestEncodeExtra_ReplyChoiceEmptyCategoryDefaults(t *testing.T) {
	names := testEntryNames()
	choice := dialogue.ReplyChoice{
		ToReplyID: 5,
		Category:  "",
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

	assertStrProp(t, parsed, "Category", "REPLY_CATEGORY_DEFAULT")
}
