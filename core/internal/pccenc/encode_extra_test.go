package pccenc

import (
	"encoding/binary"
	"math"
	"testing"

	"pcc-toolkit/core/internal/pcc"
)

func extendNames(extra ...string) []string {
	base := testNames()
	out := make([]string, len(base)+len(extra))
	copy(out, base)
	copy(out[len(base):], extra)
	return out
}

// ---------------------------------------------------------------------------
// 1. IntProperty: negative, zero, large (max int32)
// ---------------------------------------------------------------------------

func TestIntProperty_VariousValues(t *testing.T) {
	names := extendNames("myInt")
	cases := []struct {
		label string
		val   int
	}{
		{"negative", -42},
		{"zero", 0},
		{"maxInt32", math.MaxInt32},
		{"minInt32", math.MinInt32},
		{"positive", 99999},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			pv := PropertyValue{
				Name:     "myInt",
				PropType: "IntProperty",
				Value:    tc.val,
			}
			enc, err := EncodePropertyValue(pv, names)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
			parsed := props["myInt"]
			parsedVal, ok := parsed.Value.(int)
			if !ok {
				t.Fatalf("value type: want int, got %T", parsed.Value)
			}
			if parsedVal != tc.val {
				t.Errorf("value: want %d, got %d", tc.val, parsedVal)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 2. FloatProperty: 0.0, -1.5, 3.14159
// ---------------------------------------------------------------------------

func readFloat32At(data []byte, offset int) float32 {
	bits := binary.LittleEndian.Uint32(data[offset:])
	return math.Float32frombits(bits)
}

func TestFloatProperty_VariousValues(t *testing.T) {
	names := extendNames("myFloat")
	cases := []struct {
		label string
		val   float32
	}{
		{"zero", 0.0},
		{"negative", -1.5},
		{"pi", 3.14159},
		{"large", 1e12},
		{"small", 1e-12},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			pv := PropertyValue{
				Name:     "myFloat",
				PropType: "FloatProperty",
				Value:    tc.val,
			}
			enc, err := EncodePropertyValue(pv, names)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}

			// ParsePropertyCollection doesn't extract FloatProperty values,
			// so verify the raw IEEE 754 bits in the encoded payload.
			tags, err := pcc.ParsePropertyTags(enc, names, 0, len(enc), true)
			if err != nil {
				t.Fatalf("ParsePropertyTags: %v", err)
			}
			if len(tags) == 0 {
				t.Fatal("no tags found")
			}
			tag := tags[0]
			if tag.ValueOffset+4 > len(enc) {
				t.Fatal("value offset out of range")
			}
			got := readFloat32At(enc, tag.ValueOffset)
			if got != tc.val {
				t.Errorf("float value: want %v, got %v", tc.val, got)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 3. StrProperty: empty, single char, long text, special chars
// ---------------------------------------------------------------------------

func TestStrProperty_VariousValues(t *testing.T) {
	names := extendNames("myStr")
	cases := []struct {
		label string
		val   string
	}{
		{"empty", ""},
		{"single_char", "A"},
		{"long_text", "The quick brown fox jumps over the lazy dog. This is a longer string for testing purposes."},
		{"special_chars", "Héllo, Wörld! Café — ñoño ﷽"},
		{"newlines", "line1\nline2\r\nline3"},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			pv := PropertyValue{
				Name:     "myStr",
				PropType: "StrProperty",
				Value:    tc.val,
			}
			enc, err := EncodePropertyValue(pv, names)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
			parsed := props["myStr"]
			parsedVal, ok := parsed.Value.(string)
			if !ok {
				t.Fatalf("value type: want string, got %T", parsed.Value)
			}
			if parsedVal != tc.val {
				t.Errorf("value: want %q, got %q", tc.val, parsedVal)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 4. StringRefProperty: 0, negative
// ---------------------------------------------------------------------------

func TestStringRefProperty_VariousValues(t *testing.T) {
	names := extendNames("srRef")
	cases := []struct {
		label string
		val   int
	}{
		{"zero", 0},
		{"negative", -1},
		{"large", 700000},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			pv := PropertyValue{
				Name:     "srRef",
				PropType: "StringRefProperty",
				Value:    tc.val,
			}
			enc, err := EncodePropertyValue(pv, names)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
			parsed := props["srRef"]
			parsedVal, ok := parsed.Value.(int)
			if !ok {
				t.Fatalf("value type: want int, got %T", parsed.Value)
			}
			if parsedVal != tc.val {
				t.Errorf("value: want %d, got %d", tc.val, parsedVal)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 5. NameProperty: various name values
// ---------------------------------------------------------------------------

func TestNameProperty_VariousValues(t *testing.T) {
	names := extendNames("myName", "Alpha", "Beta", "Gamma", "Delta")
	cases := []string{"Alpha", "Beta", "Gamma", "Delta"}

	for _, nameVal := range cases {
		t.Run(nameVal, func(t *testing.T) {
			pv := PropertyValue{
				Name:     "myName",
				PropType: "NameProperty",
				Value:    nameVal,
			}
			enc, err := EncodePropertyValue(pv, names)
			if err != nil {
				t.Fatalf("encode %q: %v", nameVal, err)
			}
			props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
			parsed := props["myName"]
			parsedVal, ok := parsed.Value.(string)
			if !ok {
				t.Fatalf("value type: want string, got %T", parsed.Value)
			}
			if parsedVal != nameVal {
				t.Errorf("value: want %q, got %q", nameVal, parsedVal)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 6. ObjectProperty: -1, 0, large
// ---------------------------------------------------------------------------

func TestObjectProperty_VariousValues(t *testing.T) {
	names := extendNames("myObj")
	cases := []struct {
		label string
		val   int
	}{
		{"neg_one", -1},
		{"zero", 0},
		{"positive", 42},
		{"large", 999999},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			pv := PropertyValue{
				Name:     "myObj",
				PropType: "ObjectProperty",
				Value:    tc.val,
			}
			enc, err := EncodePropertyValue(pv, names)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
			parsed := props["myObj"]
			parsedVal, ok := parsed.Value.(int)
			if !ok {
				t.Fatalf("value type: want int, got %T", parsed.Value)
			}
			if parsedVal != tc.val {
				t.Errorf("value: want %d, got %d", tc.val, parsedVal)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 7. BoolProperty: true/false back-to-back in a collection
// ---------------------------------------------------------------------------

func TestBoolProperty_TrueFalseBackToBack(t *testing.T) {
	names := extendNames("bFirst", "bSecond")
	props := []PropertyValue{
		{Name: "bFirst", PropType: "BoolProperty", Value: true},
		{Name: "bSecond", PropType: "BoolProperty", Value: false},
	}
	enc, err := EncodePropertyCollection(props, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	parsed, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	if len(parsed) != 2 {
		t.Fatalf("expected 2 properties, got %d", len(parsed))
	}
	if v, ok := parsed["bFirst"]; !ok || v.Value.(bool) != true {
		t.Error("bFirst mismatch")
	}
	if v, ok := parsed["bSecond"]; !ok || v.Value.(bool) != false {
		t.Error("bSecond mismatch")
	}
}

// ---------------------------------------------------------------------------
// 8. EnumProperty: various enum values
// ---------------------------------------------------------------------------

func TestEnumProperty_VariousValues(t *testing.T) {
	names := extendNames("eMyEnum", "VALUE_ALPHA", "VALUE_BETA", "VALUE_GAMMA")
	vals := []string{"VALUE_ALPHA", "VALUE_BETA", "VALUE_GAMMA"}

	for _, enumVal := range vals {
		t.Run(enumVal, func(t *testing.T) {
			pv := PropertyValue{
				Name:     "eMyEnum",
				PropType: "EnumProperty",
				Value:    enumVal,
			}
			enc, err := EncodePropertyValue(pv, names)
			if err != nil {
				t.Fatalf("encode %q: %v", enumVal, err)
			}
			props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
			parsed := props["eMyEnum"]
			parsedVal, ok := parsed.Value.(string)
			if !ok {
				t.Fatalf("value type: want string, got %T", parsed.Value)
			}
			if parsedVal != enumVal {
				t.Errorf("value: want %q, got %q", enumVal, parsedVal)
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 9. StructProperty: nested structs 2+ levels deep
// ---------------------------------------------------------------------------

func TestStructProperty_NestedDeep(t *testing.T) {
	names := extendNames("Outer", "Leaf", "Middle", "nVal", "sLabel")

	pv := PropertyValue{
		Name:           "Outer",
		PropType:       "StructProperty",
		StructTypeName: "Middle",
		Properties: []PropertyValue{
			{Name: "nVal", PropType: "IntProperty", Value: 100},
			{
				Name:           "sLabel",
				PropType:       "StructProperty",
				StructTypeName: "Leaf",
				Properties: []PropertyValue{
					{Name: "nVal", PropType: "IntProperty", Value: 200},
				},
			},
		},
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	outer := props["Outer"]
	if outer.PropType != "StructProperty" {
		t.Fatalf("type: want StructProperty, got %s", outer.PropType)
	}
	nestedOuter, ok := outer.Value.(map[string]pcc.ParsedProperty)
	if !ok {
		t.Fatalf("nested value type: want map, got %T", outer.Value)
	}

	nVal, ok := nestedOuter["nVal"]
	if !ok {
		t.Fatal("nVal not found in outer struct")
	} else if nVal.Value.(int) != 100 {
		t.Errorf("outer nVal: want 100, got %v", nVal.Value)
	}

	sLabel, ok := nestedOuter["sLabel"]
	if !ok {
		t.Fatal("sLabel not found in outer struct")
	}
	nestedMiddle, ok := sLabel.Value.(map[string]pcc.ParsedProperty)
	if !ok {
		t.Fatalf("sLabel value type: want map, got %T", sLabel.Value)
	}
	innerNVal, ok := nestedMiddle["nVal"]
	if !ok {
		t.Fatal("nVal not found in inner struct")
	} else if innerNVal.Value.(int) != 200 {
		t.Errorf("inner nVal: want 200, got %v", innerNVal.Value)
	}
}

// ---------------------------------------------------------------------------
// 10. StructProperty: many fields
// ---------------------------------------------------------------------------

func TestStructProperty_ManyFields(t *testing.T) {
	extra := []string{"ManyFields", "fA", "fB", "fC", "fD", "fE", "fF", "fG", "fH"}
	names := extendNames(extra...)

	fields := []PropertyValue{
		{Name: "fA", PropType: "IntProperty", Value: 1},
		{Name: "fB", PropType: "StringRefProperty", Value: 2},
		{Name: "fC", PropType: "IntProperty", Value: 3},
		{Name: "fD", PropType: "BoolProperty", Value: true},
		{Name: "fE", PropType: "IntProperty", Value: 5},
		{Name: "fF", PropType: "StringRefProperty", Value: 6},
		{Name: "fG", PropType: "BoolProperty", Value: false},
		{Name: "fH", PropType: "IntProperty", Value: 8},
	}

	pv := PropertyValue{
		Name:           "ManyFields",
		PropType:       "StructProperty",
		StructTypeName: "ManyFields",
		Properties:     fields,
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	parsed := props["ManyFields"]
	nested, ok := parsed.Value.(map[string]pcc.ParsedProperty)
	if !ok {
		t.Fatalf("nested value type: want map, got %T", parsed.Value)
	}
	if len(nested) != 8 {
		t.Errorf("field count: want 8, got %d", len(nested))
	}
	if v, ok := nested["fA"]; !ok || v.Value.(int) != 1 {
		t.Error("fA mismatch")
	}
	if v, ok := nested["fD"]; !ok || v.Value.(bool) != true {
		t.Error("fD mismatch")
	}
	if v, ok := nested["fG"]; !ok || v.Value.(bool) != false {
		t.Error("fG mismatch")
	}
}

// ---------------------------------------------------------------------------
// 11. ArrayProperty<IntProperty>: zero, one, many items
// ---------------------------------------------------------------------------

func TestArrayPropertyInt_VariousCounts(t *testing.T) {
	names := extendNames("intArray")
	cases := []struct {
		label string
		vals  []int
	}{
		{"zero", []int{}},
		{"one", []int{42}},
		{"many", []int{0, 1, 2, 3, 4, 5, 6, 7, 8, 9}},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			items := make([]PropertyValue, len(tc.vals))
			for i, v := range tc.vals {
				items[i] = PropertyValue{PropType: "IntProperty", Value: v}
			}

			pv := PropertyValue{
				Name:             "intArray",
				PropType:         "ArrayProperty",
				ArrayElementType: "IntProperty",
				Items:            items,
			}
			enc, err := EncodePropertyValue(pv, names)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
			parsed := props["intArray"]
			arr, ok := parsed.Value.(map[string]interface{})
			if !ok {
				t.Fatalf("array value type: want map, got %T", parsed.Value)
			}
			count, _ := arr["count"].(int)
			if count != len(tc.vals) {
				t.Errorf("count: want %d, got %d", len(tc.vals), count)
			}
			if count > 0 {
				values := pcc.FindInt32ArrayByName(enc, names, 0, len(enc), "intArray")
				if len(values) != len(tc.vals) {
					t.Fatalf("i32 values count: want %d, got %d", len(tc.vals), len(values))
				}
				for i, v := range tc.vals {
					if values[i] != v {
						t.Errorf("item[%d]: want %d, got %d", i, v, values[i])
					}
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 12. ArrayProperty<StructProperty>: many items
// ---------------------------------------------------------------------------

func TestArrayPropertyStruct_ManyItems(t *testing.T) {
	names := extendNames("structArray", "nIdx", "srRef")

	items := make([]PropertyValue, 10)
	for i := 0; i < 10; i++ {
		items[i] = PropertyValue{
			Properties: []PropertyValue{
				{Name: "nIdx", PropType: "IntProperty", Value: i},
				{Name: "srRef", PropType: "StringRefProperty", Value: i * 100},
			},
		}
	}

	pv := PropertyValue{
		Name:             "structArray",
		PropType:         "ArrayProperty",
		ArrayElementType: "StructProperty",
		Items:            items,
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	parsed := props["structArray"]
	arr, ok := parsed.Value.(map[string]interface{})
	if !ok {
		t.Fatalf("array value type: want map, got %T", parsed.Value)
	}
	count, _ := arr["count"].(int)
	if count != 10 {
		t.Errorf("count: want 10, got %d", count)
	}

	payloadOffset, _ := arr["payload_offset"].(int)
	payloadSize, _ := arr["payload_size"].(int)

	parsedItems := pcc.ParseStructArrayItemsAsPropertyCollections(enc, names, payloadOffset, payloadSize, 10)
	if len(parsedItems) != 10 {
		t.Fatalf("parsed items: want 10, got %d", len(parsedItems))
	}
	for i, item := range parsedItems {
		nIdx, ok := item["nIdx"]
		if !ok {
			t.Errorf("item %d: nIdx not found", i)
		} else if nIdx.Value.(int) != i {
			t.Errorf("item %d nIdx: want %d, got %v", i, i, nIdx.Value)
		}
	}
}

// ---------------------------------------------------------------------------
// 13. ByteProperty: basic
// ---------------------------------------------------------------------------

func TestByteProperty_Basic(t *testing.T) {
	names := extendNames("bData", "None")
	cases := []struct {
		label string
		val   int
	}{
		{"zero", 0},
		{"one", 1},
		{"max_byte", 255},
		{"large_i32", 65535},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			pv := PropertyValue{
				Name:            "bData",
				PropType:        "ByteProperty",
				ByteSubTypeName: "None",
				Value:           tc.val,
			}
			enc, err := EncodePropertyValue(pv, names)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			// Verify the raw bytes: the value should be a 4-byte LE int32.
			tags, err := pcc.ParsePropertyTags(enc, names, 0, len(enc), true)
			if err != nil {
				t.Fatalf("ParsePropertyTags: %v", err)
			}
			if len(tags) == 0 {
				t.Fatal("no tags found")
			}
			if tags[0].PropType != "ByteProperty" {
				t.Errorf("type: want ByteProperty, got %s", tags[0].PropType)
			}
			if tags[0].Size >= 4 && tags[0].ValueOffset+4 <= len(enc) {
				got := int(binary.LittleEndian.Uint32(enc[tags[0].ValueOffset:]))
				if got != tc.val {
					t.Errorf("value: want %d, got %d", tc.val, got)
				}
			}
		})
	}
}

// ---------------------------------------------------------------------------
// 14. PropertyCollection: mixed types (10+ properties)
// ---------------------------------------------------------------------------

func TestPropertyCollection_MixedTypes(t *testing.T) {
	names := extendNames(
		"pInt", "pFloat", "pStr", "pName", "pBoolT", "pBoolF",
		"pObj", "pStrRef", "pEnum",
		"VALUE_X", "StructType",
	)

	props := []PropertyValue{
		{Name: "pInt", PropType: "IntProperty", Value: -10},
		{Name: "pFloat", PropType: "FloatProperty", Value: float32(2.718)},
		{Name: "pStr", PropType: "StrProperty", Value: "mixed collection"},
		{Name: "pStrRef", PropType: "StringRefProperty", Value: 12345},
		{Name: "pBoolT", PropType: "BoolProperty", Value: true},
		{Name: "pBoolF", PropType: "BoolProperty", Value: false},
		{Name: "pObj", PropType: "ObjectProperty", Value: 0},
		{Name: "pName", PropType: "NameProperty", Value: "StructType"},
		{Name: "pEnum", PropType: "EnumProperty", Value: "VALUE_X"},
	}

	enc, err := EncodePropertyCollection(props, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	if len(parsed) != 9 {
		t.Errorf("prop count: want 9, got %d", len(parsed))
	}

	if v, ok := parsed["pInt"]; !ok || v.Value.(int) != -10 {
		t.Error("pInt mismatch")
	}
	if v, ok := parsed["pStr"]; !ok || v.Value.(string) != "mixed collection" {
		t.Error("pStr mismatch")
	}
	if v, ok := parsed["pStrRef"]; !ok || v.Value.(int) != 12345 {
		t.Error("pStrRef mismatch")
	}
	if v, ok := parsed["pBoolT"]; !ok || v.Value.(bool) != true {
		t.Error("pBoolT mismatch")
	}
	if v, ok := parsed["pBoolF"]; !ok || v.Value.(bool) != false {
		t.Error("pBoolF mismatch")
	}
	if v, ok := parsed["pObj"]; !ok || v.Value.(int) != 0 {
		t.Error("pObj mismatch")
	}
}

// ---------------------------------------------------------------------------
// 15. PropertyCollection: single property
// ---------------------------------------------------------------------------

func TestPropertyCollection_SingleProperty(t *testing.T) {
	names := extendNames("soleProp")
	props := []PropertyValue{
		{Name: "soleProp", PropType: "IntProperty", Value: 77},
	}
	enc, err := EncodePropertyCollection(props, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	parsed, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	if len(parsed) != 1 {
		t.Fatalf("expected 1 property, got %d", len(parsed))
	}
	if v, ok := parsed["soleProp"]; !ok || v.Value.(int) != 77 {
		t.Error("soleProp mismatch")
	}
}

// ---------------------------------------------------------------------------
// 16. StructProperty without StructTypeName (error) – already in encode_test.go
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// 17. ByteProperty without ByteSubType (error)
// ---------------------------------------------------------------------------

func TestByteProperty_MissingSubType(t *testing.T) {
	names := testNames()
	pv := PropertyValue{
		Name:     "bData",
		PropType: "ByteProperty",
		Value:    1,
	}
	_, err := EncodePropertyValue(pv, names)
	if err == nil {
		t.Fatal("expected error for missing ByteSubTypeName, got nil")
	}
}

// ---------------------------------------------------------------------------
// 18. ArrayProperty without ArrayElementType (error)
// ---------------------------------------------------------------------------

func TestArrayProperty_MissingElementType(t *testing.T) {
	names := testNames()
	pv := PropertyValue{
		Name:     "badArray",
		PropType: "ArrayProperty",
	}
	_, err := EncodePropertyValue(pv, names)
	if err == nil {
		t.Fatal("expected error for missing ArrayElementType, got nil")
	}
}

// ---------------------------------------------------------------------------
// 19. Encoding with name not in table (error) – already in encode_test.go
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// 20. Encoding with type not in table (error) – already in encode_test.go
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// 21. Multiple consecutive BoolProperty true/false – already in encode_test.go
// ---------------------------------------------------------------------------

// ---------------------------------------------------------------------------
// 22. Round-trip: encode -> ParsePropertyCollection -> verify all fields
// ---------------------------------------------------------------------------

func TestRoundTrip_FullVerification(t *testing.T) {
	names := extendNames(
		"rtInt", "rtStr", "rtStrRef", "rtBoolT", "rtBoolF",
		"rtObj", "rtName", "rtEnum", "rtStruct", "rtArray",
		"TheName", "TheEnum", "InnerStructType",
	)

	type expectation struct {
		propType string
		value    interface{}
	}

	props := []PropertyValue{
		{Name: "rtInt", PropType: "IntProperty", Value: -500},
		{Name: "rtStr", PropType: "StrProperty", Value: "round-trip test"},
		{Name: "rtStrRef", PropType: "StringRefProperty", Value: 999888},
		{Name: "rtBoolT", PropType: "BoolProperty", Value: true},
		{Name: "rtBoolF", PropType: "BoolProperty", Value: false},
		{Name: "rtObj", PropType: "ObjectProperty", Value: 0},
		{Name: "rtName", PropType: "NameProperty", Value: "TheName"},
		{Name: "rtEnum", PropType: "EnumProperty", Value: "TheEnum"},
		{
			Name:           "rtStruct",
			PropType:       "StructProperty",
			StructTypeName: "InnerStructType",
			Properties: []PropertyValue{
				{Name: "rtInt", PropType: "IntProperty", Value: 300},
				{Name: "rtBoolT", PropType: "BoolProperty", Value: true},
			},
		},
		{
			Name:             "rtArray",
			PropType:         "ArrayProperty",
			ArrayElementType: "IntProperty",
			Items: []PropertyValue{
				{PropType: "IntProperty", Value: 10},
				{PropType: "IntProperty", Value: 20},
				{PropType: "IntProperty", Value: 30},
			},
		},
	}

	enc, err := EncodePropertyCollection(props, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	parsed, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	if len(parsed) != 10 {
		t.Errorf("prop count: want 10, got %d", len(parsed))
	}

	checks := map[string]expectation{
		"rtInt":    {"IntProperty", -500},
		"rtStr":    {"StrProperty", "round-trip test"},
		"rtStrRef": {"StringRefProperty", 999888},
		"rtBoolT":  {"BoolProperty", true},
		"rtBoolF":  {"BoolProperty", false},
		"rtObj":    {"ObjectProperty", 0},
		"rtName":   {"NameProperty", "TheName"},
		"rtEnum":   {"EnumProperty", "TheEnum"},
	}

	for name, want := range checks {
		pp, ok := parsed[name]
		if !ok {
			t.Errorf("%s: not found", name)
			continue
		}
		if pp.PropType != want.propType {
			t.Errorf("%s: type %s, want %s", name, pp.PropType, want.propType)
		}
		if pp.Value != want.value {
			t.Errorf("%s: value %v, want %v", name, pp.Value, want.value)
		}
	}

	// Verify struct
	rtStruct, ok := parsed["rtStruct"]
	if !ok {
		t.Fatal("rtStruct not found")
	}
	if rtStruct.PropType != "StructProperty" {
		t.Errorf("rtStruct type: want StructProperty, got %s", rtStruct.PropType)
	}
	nested, ok := rtStruct.Value.(map[string]pcc.ParsedProperty)
	if !ok {
		t.Fatalf("rtStruct nested value type: want map, got %T", rtStruct.Value)
	}
	if v, ok := nested["rtInt"]; !ok || v.Value.(int) != 300 {
		t.Error("rtStruct.rtInt mismatch")
	}
	if v, ok := nested["rtBoolT"]; !ok || v.Value.(bool) != true {
		t.Error("rtStruct.rtBoolT mismatch")
	}

	// Verify array
	rtArray, ok := parsed["rtArray"]
	if !ok {
		t.Fatal("rtArray not found")
	}
	arr, ok := rtArray.Value.(map[string]interface{})
	if !ok {
		t.Fatalf("rtArray value type: want map, got %T", rtArray.Value)
	}
	count, _ := arr["count"].(int)
	if count != 3 {
		t.Errorf("rtArray count: want 3, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// 23. EncodePropertyCollection: empty list
// ---------------------------------------------------------------------------

func TestPropertyCollection_EmptyList(t *testing.T) {
	names := testNames()
	enc, err := EncodePropertyCollection(nil, names)
	if err != nil {
		t.Fatalf("encode empty: %v", err)
	}
	// An empty collection should still produce the None terminator (8 bytes of FName).
	if len(enc) != 8 {
		t.Errorf("expected 8 bytes for empty collection (None terminator), got %d", len(enc))
	}
}

// ---------------------------------------------------------------------------
// 24. Large struct (20+ properties)
// ---------------------------------------------------------------------------

func TestStructProperty_Large(t *testing.T) {
	fieldNames := make([]string, 22)
	fieldNames[0] = "BigStruct"
	for i := 1; i <= 21; i++ {
		fieldNames[i] = "F" + string(rune('A'+i-1))
	}
	names := extendNames(fieldNames...)

	fields := make([]PropertyValue, 21)
	for i := 0; i < 21; i++ {
		fields[i] = PropertyValue{
			Name:     fieldNames[i+1],
			PropType: "IntProperty",
			Value:    i * 10,
		}
	}

	pv := PropertyValue{
		Name:           "BigStruct",
		PropType:       "StructProperty",
		StructTypeName: "BigStruct",
		Properties:     fields,
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	parsed := props["BigStruct"]
	nested, ok := parsed.Value.(map[string]pcc.ParsedProperty)
	if !ok {
		t.Fatalf("nested value type: want map, got %T", parsed.Value)
	}
	if len(nested) != 21 {
		t.Errorf("field count: want 21, got %d", len(nested))
	}

	// Spot-check first, middle, last
	if v, ok := nested["FA"]; !ok || v.Value.(int) != 0 {
		t.Error("FA mismatch")
	}
	if v, ok := nested["FK"]; !ok || v.Value.(int) != 100 {
		t.Error("FK mismatch")
	}
	// F[U] = 20th field (index 20), value 200
	if v, ok := nested["FU"]; !ok || v.Value.(int) != 200 {
		t.Errorf("FU: got %v", v.Value)
	}
}

// ---------------------------------------------------------------------------
// 25. Nested ArrayProperty inside StructProperty
// ---------------------------------------------------------------------------

func TestNestedArrayInStructProperty(t *testing.T) {
	names := extendNames("Wrapper", "innerArray")

	pv := PropertyValue{
		Name:           "Wrapper",
		PropType:       "StructProperty",
		StructTypeName: "Wrapper",
		Properties: []PropertyValue{
			{Name: "nSpeakerIndex", PropType: "IntProperty", Value: 1},
			{
				Name:             "innerArray",
				PropType:         "ArrayProperty",
				ArrayElementType: "IntProperty",
				Items: []PropertyValue{
					{PropType: "IntProperty", Value: 100},
					{PropType: "IntProperty", Value: 200},
					{PropType: "IntProperty", Value: 300},
				},
			},
		},
	}

	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
	wrapper := props["Wrapper"]
	nested, ok := wrapper.Value.(map[string]pcc.ParsedProperty)
	if !ok {
		t.Fatalf("nested value type: want map, got %T", wrapper.Value)
	}

	// Verify the scalar field
	if sp, ok := nested["nSpeakerIndex"]; !ok || sp.Value.(int) != 1 {
		t.Error("nSpeakerIndex mismatch")
	}

	// Verify the inner array
	innerArr, ok := nested["innerArray"]
	if !ok {
		t.Fatal("innerArray not found")
	}
	if innerArr.PropType != "ArrayProperty" {
		t.Errorf("innerArray type: want ArrayProperty, got %s", innerArr.PropType)
	}
	arrMap, ok := innerArr.Value.(map[string]interface{})
	if !ok {
		t.Fatalf("innerArray value type: want map, got %T", innerArr.Value)
	}
	count, _ := arrMap["count"].(int)
	if count != 3 {
		t.Errorf("innerArray count: want 3, got %d", count)
	}
}

// ---------------------------------------------------------------------------
// 26. FloatProperty with float64 input (should convert to float32)
// ---------------------------------------------------------------------------

func TestFloatProperty_Float64Input(t *testing.T) {
	names := extendNames("f64Prop")
	pv := PropertyValue{
		Name:     "f64Prop",
		PropType: "FloatProperty",
		Value:    float64(3.141592653589793),
	}
	enc, err := EncodePropertyValue(pv, names)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	tags, err := pcc.ParsePropertyTags(enc, names, 0, len(enc), true)
	if err != nil {
		t.Fatalf("ParsePropertyTags: %v", err)
	}
	if len(tags) == 0 {
		t.Fatal("no tags")
	}
	got := readFloat32At(enc, tags[0].ValueOffset)
	expected := float32(3.141592653589793)
	if got != expected {
		t.Errorf("float64->float32: want %v, got %v", expected, got)
	}
}

// ---------------------------------------------------------------------------
// 27. IntProperty with various numeric Go types
// ---------------------------------------------------------------------------

func TestIntProperty_VariousGoTypes(t *testing.T) {
	names := extendNames("nVal")
	cases := []struct {
		label string
		val   interface{}
		want  int
	}{
		{"int32", int32(-100), -100},
		{"int64", int64(200), 200},
		{"float64", float64(42.99), 42},
		{"*int", ptrInt(999), 999},
	}

	for _, tc := range cases {
		t.Run(tc.label, func(t *testing.T) {
			pv := PropertyValue{
				Name:     "nVal",
				PropType: "IntProperty",
				Value:    tc.val,
			}
			enc, err := EncodePropertyValue(pv, names)
			if err != nil {
				t.Fatalf("encode: %v", err)
			}
			props, _ := pcc.ParsePropertyCollection(enc, names, 0, len(enc))
			parsed := props["nVal"]
			if parsed.Value.(int) != tc.want {
				t.Errorf("want %d, got %v", tc.want, parsed.Value)
			}
		})
	}
}

func ptrInt(v int) *int { return &v }

// ---------------------------------------------------------------------------
// 28. ByteProperty with invalid value type (error)
// ---------------------------------------------------------------------------

func TestByteProperty_InvalidValueType(t *testing.T) {
	names := extendNames("bBad")
	pv := PropertyValue{
		Name:            "bBad",
		PropType:        "ByteProperty",
		ByteSubTypeName: "None",
		Value:           true, // bool is not supported
	}
	_, err := EncodePropertyValue(pv, names)
	if err == nil {
		t.Fatal("expected error for unsupported ByteProperty value type, got nil")
	}
}

// ---------------------------------------------------------------------------
// 29. IntProperty with invalid value type (error)
// ---------------------------------------------------------------------------

func TestIntProperty_InvalidValueType(t *testing.T) {
	names := extendNames("nBad")
	pv := PropertyValue{
		Name:     "nBad",
		PropType: "IntProperty",
		Value:    "not_an_int",
	}
	_, err := EncodePropertyValue(pv, names)
	if err == nil {
		t.Fatal("expected error for string as IntProperty value, got nil")
	}
}

// ---------------------------------------------------------------------------
// 30. FloatProperty with invalid value type (error)
// ---------------------------------------------------------------------------

func TestFloatProperty_InvalidValueType(t *testing.T) {
	names := extendNames("fBad")
	pv := PropertyValue{
		Name:     "fBad",
		PropType: "FloatProperty",
		Value:    "not_a_float",
	}
	_, err := EncodePropertyValue(pv, names)
	if err == nil {
		t.Fatal("expected error for string as FloatProperty value, got nil")
	}
}

// ---------------------------------------------------------------------------
// 31. NameProperty with value not in name table (error)
// ---------------------------------------------------------------------------

func TestNameProperty_ValueNotInTable(t *testing.T) {
	names := []string{"None", "NameProperty", "myName"}
	pv := PropertyValue{
		Name:     "myName",
		PropType: "NameProperty",
		Value:    "MissingName",
	}
	_, err := EncodePropertyValue(pv, names)
	if err == nil {
		t.Fatal("expected error for NameProperty value not in name table, got nil")
	}
}

// ---------------------------------------------------------------------------
// 32. EnumProperty with value not in name table (error)
// ---------------------------------------------------------------------------

func TestEnumProperty_ValueNotInTable(t *testing.T) {
	names := []string{"None", "EnumProperty", "eEnum"}
	pv := PropertyValue{
		Name:     "eEnum",
		PropType: "EnumProperty",
		Value:    "MissingEnum",
	}
	_, err := EncodePropertyValue(pv, names)
	if err == nil {
		t.Fatal("expected error for EnumProperty value not in name table, got nil")
	}
}

// ---------------------------------------------------------------------------
// 33. StructProperty struct type name not in table (error)
// ---------------------------------------------------------------------------

func TestStructProperty_StructTypeNotInTable(t *testing.T) {
	names := []string{"None", "StructProperty", "TestStruct"}
	pv := PropertyValue{
		Name:           "TestStruct",
		PropType:       "StructProperty",
		StructTypeName: "MissingStructType",
	}
	_, err := EncodePropertyValue(pv, names)
	if err == nil {
		t.Fatal("expected error for StructTypeName not in name table, got nil")
	}
}

// ---------------------------------------------------------------------------
// 34. ByteProperty byte subtype not in table (error)
// ---------------------------------------------------------------------------

func TestByteProperty_SubTypeNotInTable(t *testing.T) {
	names := []string{"None", "ByteProperty", "bData"}
	pv := PropertyValue{
		Name:            "bData",
		PropType:        "ByteProperty",
		ByteSubTypeName: "MissingSubType",
		Value:           1,
	}
	_, err := EncodePropertyValue(pv, names)
	if err == nil {
		t.Fatal("expected error for ByteSubTypeName not in name table, got nil")
	}
}

// ---------------------------------------------------------------------------
// 35. ArrayProperty element type not in table (error)
// ---------------------------------------------------------------------------

func TestArrayProperty_ElementTypeNotInTable(t *testing.T) {
	names := []string{"None", "ArrayProperty", "arr"}
	pv := PropertyValue{
		Name:             "arr",
		PropType:         "ArrayProperty",
		ArrayElementType: "MissingType",
	}
	_, err := EncodePropertyValue(pv, names)
	if err == nil {
		t.Fatal("expected error for ArrayElementType not in name table, got nil")
	}
}

// ---------------------------------------------------------------------------
// 36. EncodePropertyCollection with nil and empty slice same behavior
// ---------------------------------------------------------------------------

func TestPropertyCollection_NilVsEmpty(t *testing.T) {
	names := testNames()
	encNil, err := EncodePropertyCollection(nil, names)
	if err != nil {
		t.Fatalf("encode nil: %v", err)
	}
	encEmpty, err := EncodePropertyCollection([]PropertyValue{}, names)
	if err != nil {
		t.Fatalf("encode empty slice: %v", err)
	}
	if len(encNil) != len(encEmpty) {
		t.Errorf("nil vs empty: lengths differ: %d vs %d", len(encNil), len(encEmpty))
	}
}

// ---------------------------------------------------------------------------
// 37. ArrayProperty with unsupported element type (error)
// ---------------------------------------------------------------------------

func TestArrayProperty_UnsupportedElementType(t *testing.T) {
	names := extendNames("badArray", "StrProperty")
	pv := PropertyValue{
		Name:             "badArray",
		PropType:         "ArrayProperty",
		ArrayElementType: "StrProperty",
		Items: []PropertyValue{
			{PropType: "StrProperty", Value: "hello"},
		},
	}
	_, err := EncodePropertyValue(pv, names)
	if err == nil {
		t.Fatal("expected error for unsupported array element type, got nil")
	}
}
