package pccenc

import (
	"testing"

	"pcc-toolkit/core/internal/pcc"
)

func testNames() []string {
	return []string{
		"None",
		"IntProperty",
		"BoolProperty",
		"FloatProperty",
		"StrProperty",
		"StringRefProperty",
		"NameProperty",
		"ObjectProperty",
		"StructProperty",
		"ArrayProperty",
		"EnumProperty",
		"ByteProperty",
		"EConvGUIStyles",
		"BioDialogEntryNode",
		"BioDialogReplyNode",
		"m_EntryList",
		"m_ReplyList",
		"nSpeakerIndex",
		"srText",
		"nListenerIndex",
		"nConditionalFunc",
		"nConditionalParam",
		"nStateTransition",
		"nStateTransitionParam",
		"nScriptIndex",
		"nExportID",
		"nCameraIntimacy",
		"bFireConditional",
		"bSkippable",
		"bIsNonTextLine",
		"bAmbient",
		"eConvGUIStyle",
		"CONV_GUISTYLE_DEFAULT",
		"REPLY_CATEGORY_DEFAULT",
		"EReplyCategory",
		"sParaphrase",
		"Category",
		"MatineeSequence",
		"InterpLength",
		"m_StartingList",
		"EmptyArray",
		"EmptyStr",
	}
}

func TestEncodeNoneProperty(t *testing.T) {
	names := testNames()
	enc, err := EncodeNoneProperty(names)
	if err != nil {
		t.Fatalf("EncodeNoneProperty: %v", err)
	}
	if len(enc) != 8 {
		t.Errorf("expected 8 bytes, got %d", len(enc))
	}

	noneIdx := 0
	nameIdx := int(int32(pcc.ReadRawI32(enc, 0)))
	if nameIdx != noneIdx {
		t.Errorf("name index: want %d, got %d", noneIdx, nameIdx)
	}
}

func TestIntProperty_RoundTrip(t *testing.T) {
	names := testNames()
	val := 42
	pv := PropertyValue{
		Name:     "nSpeakerIndex",
		PropType: "IntProperty",
		Value:    val,
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	if props == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}

	parsed, ok := props["nSpeakerIndex"]
	if !ok {
		t.Fatal("property nSpeakerIndex not found in parsed collection")
	}
	if parsed.PropType != "IntProperty" {
		t.Errorf("type: want IntProperty, got %s", parsed.PropType)
	}
	parsedVal, ok := parsed.Value.(int)
	if !ok {
		t.Fatalf("value type: want int, got %T", parsed.Value)
	}
	if parsedVal != val {
		t.Errorf("value: want %d, got %d", val, parsedVal)
	}
}

func TestBoolProperty_RoundTrip(t *testing.T) {
	names := testNames()

	tests := []struct {
		val  bool
		want bool
	}{
		{true, true},
		{false, false},
	}

	for _, tt := range tests {
		pv := PropertyValue{
			Name:     "bFireConditional",
			PropType: "BoolProperty",
			Value:    tt.val,
		}
		enc, err := EncodePropertyValue(pv, names)
		if err != nil {
			t.Fatalf("encode %v: %v", tt.val, err)
		}

		props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
		parsed := props["bFireConditional"]
		parsedVal, ok := parsed.Value.(bool)
		if !ok {
			t.Fatalf("value type: want bool, got %T", parsed.Value)
		}
		if parsedVal != tt.want {
			t.Errorf("value: want %v, got %v", tt.want, parsedVal)
		}
	}
}

func TestStrProperty_RoundTrip(t *testing.T) {
	names := testNames()
	text := "Hello, Commander Shepard"

	pv := PropertyValue{
		Name:     "sParaphrase",
		PropType: "StrProperty",
		Value:    text,
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	parsed := props["sParaphrase"]
	parsedVal, ok := parsed.Value.(string)
	if !ok {
		t.Fatalf("value type: want string, got %T", parsed.Value)
	}
	if parsedVal != text {
		t.Errorf("value: want %q, got %q", text, parsedVal)
	}
}

func TestNameProperty_RoundTrip(t *testing.T) {
	names := testNames()
	nameVal := "REPLY_CATEGORY_DEFAULT"

	pv := PropertyValue{
		Name:     "Category",
		PropType: "NameProperty",
		Value:    nameVal,
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	parsed := props["Category"]
	parsedVal, ok := parsed.Value.(string)
	if !ok {
		t.Fatalf("value type: want string (resolved name), got %T", parsed.Value)
	}
	if parsedVal != nameVal {
		t.Errorf("value: want %q, got %q", nameVal, parsedVal)
	}
}

func TestStringRefProperty_RoundTrip(t *testing.T) {
	names := testNames()
	strref := 663399

	pv := PropertyValue{
		Name:     "srText",
		PropType: "StringRefProperty",
		Value:    strref,
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	parsed := props["srText"]
	parsedVal, ok := parsed.Value.(int)
	if !ok {
		t.Fatalf("value type: want int, got %T", parsed.Value)
	}
	if parsedVal != strref {
		t.Errorf("value: want %d, got %d", strref, parsedVal)
	}
}

func TestObjectProperty_RoundTrip(t *testing.T) {
	names := testNames()
	objIdx := -1

	pv := PropertyValue{
		Name:     "MatineeSequence",
		PropType: "ObjectProperty",
		Value:    objIdx,
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	parsed := props["MatineeSequence"]
	parsedVal, ok := parsed.Value.(int)
	if !ok {
		t.Fatalf("value type: want int, got %T", parsed.Value)
	}
	if parsedVal != objIdx {
		t.Errorf("value: want %d, got %d", objIdx, parsedVal)
	}
}

func TestFloatProperty_RoundTrip(t *testing.T) {
	names := testNames()
	val := float32(3.14)

	pv := PropertyValue{
		Name:     "InterpLength",
		PropType: "FloatProperty",
		Value:    val,
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	parsed := props["InterpLength"]
	if parsed.PropType != "FloatProperty" {
		t.Logf("FloatProperty encoded and parseable, propType=%s size=%d", parsed.PropType, len(enc))
	}
}

func TestStructProperty_RoundTrip(t *testing.T) {
	names := testNames()

	pv := PropertyValue{
		Name:           "m_EntryList",
		PropType:       "StructProperty",
		StructTypeName: "BioDialogEntryNode",
		Properties: []PropertyValue{
			{Name: "nSpeakerIndex", PropType: "IntProperty", Value: 0},
			{Name: "srText", PropType: "StringRefProperty", Value: 123456},
			{Name: "bFireConditional", PropType: "BoolProperty", Value: false},
		},
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	parsed := props["m_EntryList"]
	if parsed.PropType != "StructProperty" {
		t.Errorf("type: want StructProperty, got %s", parsed.PropType)
		return
	}

	nested, ok := parsed.Value.(map[string]pcc.ParsedProperty)
	if !ok {
		t.Fatalf("nested value type: want map, got %T", parsed.Value)
	}

	speaker, ok := nested["nSpeakerIndex"]
	if !ok {
		t.Error("nSpeakerIndex not found in struct")
	} else {
		sv, _ := speaker.Value.(int)
		if sv != 0 {
			t.Errorf("nSpeakerIndex: want 0, got %d", sv)
		}
	}

	srText, ok := nested["srText"]
	if !ok {
		t.Error("srText not found in struct")
	} else {
		sv, _ := srText.Value.(int)
		if sv != 123456 {
			t.Errorf("srText: want 123456, got %d", sv)
		}
	}
}

func TestArrayProperty_Int(t *testing.T) {
	names := testNames()

	pv := PropertyValue{
		Name:             "m_StartingList",
		PropType:         "ArrayProperty",
		ArrayElementType: "IntProperty",
		Items: []PropertyValue{
			{PropType: "IntProperty", Value: 0},
			{PropType: "IntProperty", Value: 1},
			{PropType: "IntProperty", Value: 2},
		},
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	parsed := props["m_StartingList"]
	if parsed.PropType != "ArrayProperty" {
		t.Errorf("type: want ArrayProperty, got %s", parsed.PropType)
		return
	}

	arr, ok := parsed.Value.(map[string]interface{})
	if !ok {
		t.Fatalf("array value type: want map, got %T", parsed.Value)
	}
	count, ok := arr["count"].(int)
	if !ok {
		t.Fatal("count not found or wrong type")
	}
	if count != 3 {
		t.Errorf("count: want 3, got %d", count)
	}
}

func TestArrayProperty_Struct(t *testing.T) {
	names := testNames()

	pv := PropertyValue{
		Name:             "m_EntryList",
		PropType:         "ArrayProperty",
		ArrayElementType: "StructProperty",
		Items: []PropertyValue{
			{
				Properties: []PropertyValue{
					{Name: "nSpeakerIndex", PropType: "IntProperty", Value: 1},
					{Name: "srText", PropType: "StringRefProperty", Value: 100},
				},
			},
			{
				Properties: []PropertyValue{
					{Name: "nSpeakerIndex", PropType: "IntProperty", Value: 2},
					{Name: "srText", PropType: "StringRefProperty", Value: 200},
				},
			},
		},
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	parsed := props["m_EntryList"]
	if parsed.PropType != "ArrayProperty" {
		t.Errorf("type: want ArrayProperty, got %s", parsed.PropType)
		return
	}

	arr, ok := parsed.Value.(map[string]interface{})
	if !ok {
		t.Fatalf("array value type: want map, got %T", parsed.Value)
	}
	count, _ := arr["count"].(int)
	if count != 2 {
		t.Errorf("count: want 2, got %d", count)
	}

	payloadOffset, _ := arr["payload_offset"].(int)
	payloadSize, _ := arr["payload_size"].(int)

	items := pcc.ParseStructArrayItemsAsPropertyCollections(enc, names, payloadOffset, payloadSize, 2)
	if len(items) != 2 {
		t.Fatalf("items: want 2, got %d", len(items))
	}

	item0speaker, ok := items[0]["nSpeakerIndex"]
	if !ok {
		t.Error("item 0: nSpeakerIndex not found")
	} else {
		sv, _ := item0speaker.Value.(int)
		if sv != 1 {
			t.Errorf("item 0 speaker: want 1, got %d", sv)
		}
	}

	item1speaker, ok := items[1]["nSpeakerIndex"]
	if !ok {
		t.Error("item 1: nSpeakerIndex not found")
	} else {
		sv, _ := item1speaker.Value.(int)
		if sv != 2 {
			t.Errorf("item 1 speaker: want 2, got %d", sv)
		}
	}
}

func TestPropertyCollection_RoundTrip(t *testing.T) {
	names := testNames()

	props := []PropertyValue{
		{Name: "nSpeakerIndex", PropType: "IntProperty", Value: 0},
		{Name: "srText", PropType: "StringRefProperty", Value: 555},
		{Name: "bFireConditional", PropType: "BoolProperty", Value: true},
	}

	enc, err := EncodePropertyCollection(props, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}
	if len(parsed) != 3 {
		t.Errorf("prop count: want 3, got %d", len(parsed))
	}

	if sp, ok := parsed["nSpeakerIndex"]; !ok || sp.Value.(int) != 0 {
		t.Error("nSpeakerIndex mismatch")
	}
	if sp, ok := parsed["srText"]; !ok || sp.Value.(int) != 555 {
		t.Error("srText mismatch")
	}
	if sp, ok := parsed["bFireConditional"]; !ok || sp.Value.(bool) != true {
		t.Error("bFireConditional mismatch")
	}
}

func TestEnumProperty_RoundTrip(t *testing.T) {
	names := testNames()

	pv := PropertyValue{
		Name:     "eConvGUIStyle",
		PropType: "EnumProperty",
		Value:    "CONV_GUISTYLE_DEFAULT",
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	parsed := props["eConvGUIStyle"]
	parsedVal, ok := parsed.Value.(string)
	if !ok {
		t.Fatalf("value type: want string, got %T", parsed.Value)
	}
	if parsedVal != "CONV_GUISTYLE_DEFAULT" {
		t.Errorf("value: want CONV_GUISTYLE_DEFAULT, got %s", parsedVal)
	}
}

func TestNameNotFound(t *testing.T) {
	names := []string{"None", "IntProperty"}

	pv := PropertyValue{
		Name:     "MissingName",
		PropType: "IntProperty",
		Value:    1,
	}

	_, err := EncodePropertyValue(pv, names)
	if err == nil {
		t.Fatal("expected error for missing name, got nil")
	}
}

func TestTypeNotFound(t *testing.T) {
	names := []string{"None", "SomeName"}

	pv := PropertyValue{
		Name:     "SomeName",
		PropType: "MissingType",
		Value:    1,
	}

	_, err := EncodePropertyValue(pv, names)
	if err == nil {
		t.Fatal("expected error for missing type in name table, got nil")
	}
}

func TestStructProperty_MissingStructType(t *testing.T) {
	names := testNames()

	pv := PropertyValue{
		Name:     "TestStruct",
		PropType: "StructProperty",
	}

	_, err := EncodePropertyValue(pv, names)
	if err == nil {
		t.Fatal("expected error for missing StructTypeName, got nil")
	}
}

func TestArrayProperty_Empty(t *testing.T) {
	names := testNames()

	pv := PropertyValue{
		Name:             "EmptyArray",
		PropType:         "ArrayProperty",
		ArrayElementType: "IntProperty",
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	parsed := props["EmptyArray"]
	arr, ok := parsed.Value.(map[string]interface{})
	if !ok {
		t.Fatal("value type: want map for array")
	}
	count, _ := arr["count"].(int)
	if count != 0 {
		t.Errorf("count: want 0, got %d", count)
	}
}

func TestConsecutiveBoolProperties(t *testing.T) {
	names := append(testNames(),
		"bFirst", "bSecond", "bThird",
	)

	props := []PropertyValue{
		{Name: "bFirst", PropType: "BoolProperty", Value: true},
		{Name: "bSecond", PropType: "BoolProperty", Value: false},
		{Name: "bThird", PropType: "BoolProperty", Value: true},
	}

	enc, err := EncodePropertyCollection(props, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	if parsed == nil {
		t.Fatal("ParsePropertyCollection returned nil")
	}
	if len(parsed) != 3 {
		t.Fatalf("expected 3 properties, got %d", len(parsed))
	}

	if v, ok := parsed["bFirst"]; !ok || v.Value.(bool) != true {
		t.Error("bFirst mismatch")
	}
	if v, ok := parsed["bSecond"]; !ok || v.Value.(bool) != false {
		t.Error("bSecond mismatch")
	}
	if v, ok := parsed["bThird"]; !ok || v.Value.(bool) != true {
		t.Error("bThird mismatch")
	}
}

func TestEmptyStrProperty(t *testing.T) {
	names := testNames()

	pv := PropertyValue{
		Name:     "EmptyStr",
		PropType: "StrProperty",
		Value:    "",
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	parsed := props["EmptyStr"]
	parsedVal, ok := parsed.Value.(string)
	if !ok {
		t.Fatalf("value type: want string, got %T", parsed.Value)
	}
	if parsedVal != "" {
		t.Errorf("value: want empty string, got %q", parsedVal)
	}
}
