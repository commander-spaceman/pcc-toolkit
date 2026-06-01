package pcc

import "fmt"

var PropertyTypeNames = map[string]bool{
	"ArrayProperty":     true,
	"BoolProperty":      true,
	"ByteProperty":      true,
	"ClassProperty":     true,
	"ComponentProperty": true,
	"DelegateProperty":  true,
	"FloatProperty":     true,
	"InterfaceProperty": true,
	"IntProperty":       true,
	"MapProperty":       true,
	"NameProperty":      true,
	"ObjectProperty":    true,
	"StrProperty":       true,
	"StringRefProperty": true,
	"StructProperty":    true,
}

type PropertyTag struct {
	Name        string `json:"name"`
	PropType    string `json:"prop_type"`
	Size        int    `json:"size"`
	ArrayIndex  int    `json:"array_index"`
	ValueOffset int    `json:"value_offset"`
}

type ArrayLayoutInfo struct {
	Count        int  `json:"count"`
	PayloadSize  int  `json:"payload_size"`
	BytesPerItem *int `json:"bytes_per_item,omitempty"`
	Remainder    int  `json:"remainder"`
	IsTightI32   bool `json:"is_tight_i32"`
}

func ParsePropertyTags(data []byte, names []string, startOffset, size int, strict bool) ([]PropertyTag, error) {
	end := startOffset + size
	if end > len(data) {
		return nil, fmt.Errorf("export data out of range")
	}

	var tags []PropertyTag
	cursor := startOffset

	for cursor+16 <= end {
		nameA := readI32(data, cursor)
		nameB := readI32(data, cursor+4)
		typeA := readI32(data, cursor+8)
		typeB := readI32(data, cursor+12)

		typeAName := resolveName(typeA, names)
		typeBName := resolveName(typeB, names)

		var typeIndex, nameIndex int
		if PropertyTypeNames[typeBName] && !PropertyTypeNames[typeAName] {
			typeIndex = typeB
			nameIndex = nameB
		} else {
			typeIndex = typeA
			nameIndex = nameA
		}

		name := resolveName(nameIndex, names)
		if name == "" {
			if strict {
				return nil, fmt.Errorf("invalid name index: %d", nameIndex)
			}
			break
		}
		cursor += 8

		if name == "None" {
			break
		}

		if cursor+16 > end {
			if strict {
				return nil, fmt.Errorf("property tag truncated before type/size/index")
			}
			break
		}

		propType := resolveName(typeIndex, names)
		cursor += 8

		propSize := readI32(data, cursor)
		arrayIndex := readI32(data, cursor+4)
		cursor += 8

		if propType == "ArrayProperty" && propSize < 4 && arrayIndex >= 4 {
			propSize, arrayIndex = arrayIndex, propSize
		}

		if propSize < 0 || propSize > size-(cursor-startOffset) {
			break
		}

		skipTagMeta(&cursor, propType)
		resolveBoolOrArrayMeta(data, &cursor, propType, propSize, names, end)

		if cursor > end {
			if strict {
				return nil, fmt.Errorf("property tag out of range in metadata")
			}
			break
		}

		valueOffset := cursor
		cursor += propSize
		if cursor > end {
			if strict {
				return nil, fmt.Errorf("property value out of range")
			}
			break
		}

		tags = append(tags, PropertyTag{
			Name:        name,
			PropType:    propType,
			Size:        propSize,
			ArrayIndex:  arrayIndex,
			ValueOffset: valueOffset,
		})
	}

	return tags, nil
}

func skipTagMeta(cursor *int, propType string) {
	if propType == "StructProperty" || propType == "ByteProperty" {
		*cursor += 8
	}
}

func resolveBoolOrArrayMeta(data []byte, cursor *int, propType string, propSize int, names []string, end int) {
	if propType == "BoolProperty" {
		oneByteNext := *cursor + 1 + propSize
		fourByteNext := *cursor + 4 + propSize
		fourOk := oneByteNext+3 <= end && isPlausibleTagStart(data, fourByteNext, end, names)
		oneOk := oneByteNext <= end && isPlausibleTagStart(data, oneByteNext, end, names)
		if fourOk && !oneOk {
			*cursor += 4
		} else {
			*cursor += 1
		}
		return
	}
	if propType == "ArrayProperty" {
		if *cursor+8 <= end {
			a := readI32(data, *cursor)
			b := readI32(data, *cursor+4)
			aName := resolveName(a, names)
			bName := resolveName(b, names)
			if PropertyTypeNames[aName] || PropertyTypeNames[bName] {
				*cursor += 8
				return
			}
		}
		noMetaNext := *cursor + propSize
		fnameMetaNext := *cursor + 8 + propSize
		noOk := noMetaNext <= end && isPlausibleTagStart(data, noMetaNext, end, names)
		fnameOk := fnameMetaNext <= end && isPlausibleTagStart(data, fnameMetaNext, end, names)
		if fnameOk && !noOk {
			*cursor += 8
		}
	}
}

func isPlausibleTagStart(data []byte, offset int, end int, names []string) bool {
	if offset+8 > end {
		return false
	}
	nameA := readI32(data, offset)
	nameB := readI32(data, offset+4)
	nameAText := resolveName(nameA, names)
	nameBText := resolveName(nameB, names)
	if nameAText == "None" || nameBText == "None" {
		return true
	}
	if offset+16 > end {
		return false
	}
	typeA := readI32(data, offset+8)
	typeB := readI32(data, offset+12)
	typeAText := resolveName(typeA, names)
	typeBText := resolveName(typeB, names)
	return PropertyTypeNames[typeAText] || PropertyTypeNames[typeBText]
}

func ExtractBioConversationKeyProperties(data []byte, names []string, serialOffset, serialSize int) []PropertyTag {
	keyAliases := map[string]string{
		"EntryList":      "EntryList",
		"m_EntryList":    "EntryList",
		"ReplyList":      "ReplyList",
		"m_ReplyList":    "ReplyList",
		"ReplyListNew":   "ReplyList",
		"SpeakerList":    "SpeakerList",
		"m_SpeakerList":  "SpeakerList",
		"m_StartingList": "StartingList",
	}
	keyProps := map[string]bool{
		"EntryList": true, "ReplyList": true, "SpeakerList": true, "StartingList": true,
	}

	bestKeyTags := map[string]PropertyTag{}
	bestKeyScore := -999999999

	scoreTags := func(tags []PropertyTag) int {
		score := 0
		for _, tag := range tags {
			if _, ok := keyAliases[tag.Name]; ok {
				score += 100
			}
			if tag.PropType == "ArrayProperty" {
				score += 5
				if tag.ValueOffset+4 <= len(data) {
					count := readI32(data, tag.ValueOffset)
					if count > 0 {
						score += 20
					} else if count < 0 {
						score -= 20
					}
				}
			}
			if PropertyTypeNames[tag.PropType] {
				score += 2
			} else {
				score--
			}
		}
		return score
	}

	for _, delta := range []int{0, 4, 8, 12} {
		if serialSize <= delta {
			continue
		}
		tags, err := ParsePropertyTags(data, names, serialOffset+delta, serialSize-delta, false)
		if err != nil {
			continue
		}
		sc := scoreTags(tags)

		keyTags := map[string]PropertyTag{}
		for _, tag := range tags {
			canonical, ok := keyAliases[tag.Name]
			if !ok {
				continue
			}
			if keyProps[canonical] {
				tag.Name = canonical
				keyTags[canonical] = tag
			}
		}
		if len(keyTags) > len(bestKeyTags) || (len(keyTags) == len(bestKeyTags) && sc > bestKeyScore) {
			bestKeyTags = keyTags
			bestKeyScore = sc
		}
	}

	fuzzyTags := scanBioConvKeyPropsFuzzy(data, names, serialOffset, serialSize)
	for _, tag := range fuzzyTags {
		if _, ok := bestKeyTags[tag.Name]; !ok {
			bestKeyTags[tag.Name] = tag
		}
	}

	order := []string{"EntryList", "ReplyList", "SpeakerList", "StartingList"}
	var result []PropertyTag
	for _, key := range order {
		if tag, ok := bestKeyTags[key]; ok {
			result = append(result, tag)
		}
	}
	return result
}

func scanBioConvKeyPropsFuzzy(data []byte, names []string, serialOffset, serialSize int) []PropertyTag {
	nameToIdx := map[string]int{}
	for i, n := range names {
		nameToIdx[n] = i
	}
	arrayPropIdx, ok := nameToIdx["ArrayProperty"]
	if !ok {
		return nil
	}

	aliasToCanonical := map[string]string{
		"EntryList":      "EntryList",
		"m_EntryList":    "EntryList",
		"ReplyList":      "ReplyList",
		"m_ReplyList":    "ReplyList",
		"ReplyListNew":   "ReplyList",
		"SpeakerList":    "SpeakerList",
		"m_SpeakerList":  "SpeakerList",
		"StartingList":   "StartingList",
		"m_StartingList": "StartingList",
	}

	keyNameIndices := map[int]string{}
	for alias, canonical := range aliasToCanonical {
		if idx, ok := nameToIdx[alias]; ok {
			keyNameIndices[idx] = canonical
		}
	}
	if len(keyNameIndices) == 0 {
		return nil
	}

	start := serialOffset
	end := serialOffset + serialSize
	if start < 0 || end > len(data) || start >= end {
		return nil
	}

	found := map[string]PropertyTag{}

	resolvePair := func(a, b int, wanted map[int]bool) int {
		if _, ok := wanted[a]; ok && !wanted[b] {
			return a
		}
		if _, ok := wanted[b]; ok && !wanted[a] {
			return b
		}
		if _, ok := wanted[a]; ok {
			return a
		}
		if _, ok := wanted[b]; ok {
			return b
		}
		return -1
	}

	wantedNames := map[int]bool{}
	for idx := range keyNameIndices {
		wantedNames[idx] = true
	}
	wantedTypes := map[int]bool{arrayPropIdx: true}

	limit := end - 24
	if limit < start {
		limit = start
	}
	for pos := start; pos < limit; pos += 4 {
		if pos+16 > end {
			break
		}
		nameA := readI32(data, pos)
		nameB := readI32(data, pos+4)
		typeA := readI32(data, pos+8)
		typeB := readI32(data, pos+12)

		nameIdx := resolvePair(nameA, nameB, wantedNames)
		if nameIdx < 0 {
			continue
		}
		typeIdx := resolvePair(typeA, typeB, wantedTypes)
		if typeIdx < 0 {
			continue
		}

		propName := keyNameIndices[nameIdx]
		if _, ok := found[propName]; ok {
			continue
		}

		sizePos := pos + 16
		if sizePos+8 > end {
			continue
		}
		rawSize := readI32(data, sizePos)
		rawArrayIdx := readI32(data, sizePos+4)

		propSize := rawSize
		arrayIdx := rawArrayIdx
		if propSize < 4 && arrayIdx >= 4 {
			propSize, arrayIdx = arrayIdx, propSize
		}
		if propSize < 4 {
			continue
		}

		var chosenValueOffset int
		for _, vo := range []int{sizePos + 8, sizePos + 16} {
			if vo+propSize > end {
				continue
			}
			if vo+4 > len(data) {
				continue
			}
			count := readI32(data, vo)
			if count >= 0 && count <= 200000 {
				chosenValueOffset = vo
				break
			}
		}
		if chosenValueOffset == 0 {
			continue
		}

		found[propName] = PropertyTag{
			Name:        propName,
			PropType:    "ArrayProperty",
			Size:        propSize,
			ArrayIndex:  arrayIdx,
			ValueOffset: chosenValueOffset,
		}

		if len(found) == 4 {
			break
		}
	}

	order := []string{"EntryList", "ReplyList", "SpeakerList", "StartingList"}
	var result []PropertyTag
	for _, key := range order {
		if tag, ok := found[key]; ok {
			result = append(result, tag)
		}
	}
	return result
}

func ReadArrayPropertyCount(data []byte, tag PropertyTag) int {
	if tag.PropType != "ArrayProperty" {
		return -1
	}
	if tag.Size < 4 {
		return -1
	}
	count, _ := resolveArrayCountAndPayloadStart(data, tag)
	return count
}

func ReadArrayPropertyPayloadInfo(data []byte, tag PropertyTag) (int, int, int) {
	count, payloadStart := resolveArrayCountAndPayloadStart(data, tag)
	payloadSize := tag.Size - (payloadStart - tag.ValueOffset)
	if payloadSize < 0 {
		payloadSize = 0
	}
	return count, payloadStart, payloadSize
}

func ReadArrayPropertyPayloadInfoWithStructMeta(data []byte, names []string, tag PropertyTag) (int, int, int) {
	vo := tag.ValueOffset
	if vo+4 > len(data) || tag.PropType != "ArrayProperty" {
		return 0, vo, tag.Size
	}

	firstAtVo := readI32(data, vo)
	if firstAtVo > 0 && firstAtVo <= 200000 && firstAtVo*4 <= tag.Size {
		return ReadArrayPropertyPayloadInfo(data, tag)
	}

	if vo+8 <= len(data) {
		count := readI32(data, vo+4)
		if count > 0 && count <= 200000 {
			payloadStart := vo + 8
			payloadSize := tag.Size - 8
			if payloadSize < 0 {
				payloadSize = 0
			}
			return count, payloadStart, payloadSize
		}
	}

	return ReadArrayPropertyPayloadInfo(data, tag)
}

func resolveArrayCountAndPayloadStart(data []byte, tag PropertyTag) (int, int) {
	if tag.ValueOffset+4 > len(data) {
		return 0, tag.ValueOffset
	}
	count := readI32(data, tag.ValueOffset)
	if count >= 0 {
		return count, tag.ValueOffset + 4
	}

	firstNonNeg := -1
	for _, delta := range []int{4, 8, 12, 16} {
		if delta+4 > tag.Size {
			break
		}
		if tag.ValueOffset+delta+4 > len(data) {
			break
		}
		candidate := readI32(data, tag.ValueOffset+delta)
		if candidate >= 0 {
			if candidate > 0 {
				return candidate, tag.ValueOffset + delta + 4
			}
			if firstNonNeg < 0 {
				firstNonNeg = delta
			}
		}
	}
	if firstNonNeg >= 0 {
		return 0, tag.ValueOffset + firstNonNeg + 4
	}
	return count, tag.ValueOffset + 4
}

func ReadArrayPropertyI32Values(data []byte, tag PropertyTag) []int {
	count, payloadStart := resolveArrayCountAndPayloadStart(data, tag)
	payloadSize := tag.Size - (payloadStart - tag.ValueOffset)
	if payloadSize < 0 {
		payloadSize = 0
	}
	expectedSize := count * 4
	if count < 0 || payloadSize != expectedSize {
		return nil
	}
	values := make([]int, count)
	cursor := payloadStart
	for i := 0; i < count; i++ {
		values[i] = readI32(data, cursor)
		cursor += 4
	}
	return values
}

func ReadObjectPropertyValue(data []byte, tag PropertyTag) int {
	if tag.PropType != "ObjectProperty" {
		return -1
	}
	if tag.ValueOffset+4 > len(data) {
		return -1
	}
	return readI32(data, tag.ValueOffset)
}

func ReadArrayPropertyObjectValues(data []byte, tag PropertyTag) []int {
	return ReadArrayPropertyI32Values(data, tag)
}

func FindExtraPropertyTags(data []byte, names []string, serialOffset, serialSize int, wantedNames []string) map[string]PropertyTag {
	nameToIdx := map[string]int{}
	for i, n := range names {
		nameToIdx[n] = i
	}

	keyNameIndices := map[int]string{}
	for _, n := range wantedNames {
		if idx, ok := nameToIdx[n]; ok {
			keyNameIndices[idx] = n
		}
	}
	if len(keyNameIndices) == 0 {
		return nil
	}

	start := serialOffset
	end := serialOffset + serialSize
	if start < 0 || end > len(data) || start >= end {
		return nil
	}

	found := map[string]PropertyTag{}
	wantedNamesIdx := map[int]bool{}
	for idx := range keyNameIndices {
		wantedNamesIdx[idx] = true
	}

	resolveIdx := func(a, b int) int {
		if _, ok := keyNameIndices[a]; ok {
			return a
		}
		if _, ok := keyNameIndices[b]; ok {
			return b
		}
		return -1
	}

	limit := end - 24
	if limit < start {
		limit = start
	}
	for pos := start; pos < limit; pos += 4 {
		if pos+24 > end {
			break
		}
		nameA := readI32(data, pos)
		nameB := readI32(data, pos+4)
		typeA := readI32(data, pos+8)
		typeB := readI32(data, pos+12)

		nameIdx := resolveIdx(nameA, nameB)
		if nameIdx < 0 {
			continue
		}

		typeIdx := resolveIdx(typeA, typeB)
		typeAName := resolveName(typeA, names)
		typeBName := resolveName(typeB, names)
		if !PropertyTypeNames[typeAName] && !PropertyTypeNames[typeBName] {
			continue
		}
		_ = typeIdx

		propName := keyNameIndices[nameIdx]
		if _, ok := found[propName]; ok {
			continue
		}

		sizePos := pos + 8 + 8
		if sizePos+8 > end {
			continue
		}
		propSize := readI32(data, sizePos)
		arrayIdx := readI32(data, sizePos+4)
		if propSize < 0 || propSize > serialSize-(pos-start) {
			continue
		}

		var propType string
		if PropertyTypeNames[typeAName] {
			propType = typeAName
		} else {
			propType = typeBName
		}

		valueOffset := sizePos + 8
		if propType == "StructProperty" || propType == "ByteProperty" {
			valueOffset += 8
		} else if propType == "ArrayProperty" {
			valueOffset += 16
		}
		if valueOffset+propSize > end {
			continue
		}

		found[propName] = PropertyTag{
			Name:        propName,
			PropType:    propType,
			Size:        propSize,
			ArrayIndex:  arrayIdx,
			ValueOffset: valueOffset,
		}

		if len(found) == len(wantedNames) {
			break
		}
	}
	return found
}

func ReadArrayPropertyI32Rows(data []byte, tag PropertyTag, itemWidth int) [][]int {
	if itemWidth <= 0 {
		return nil
	}
	values := ReadArrayPropertyI32Values(data, tag)
	if len(values)%itemWidth != 0 {
		return nil
	}
	rows := make([][]int, 0, len(values)/itemWidth)
	for i := 0; i < len(values); i += itemWidth {
		rows = append(rows, values[i:i+itemWidth])
	}
	return rows
}

func AnalyzeArrayPropertyLayout(data []byte, tag PropertyTag) ArrayLayoutInfo {
	count, payloadStart := resolveArrayCountAndPayloadStart(data, tag)
	payloadSize := tag.Size - (payloadStart - tag.ValueOffset)
	if payloadSize < 0 {
		payloadSize = 0
	}
	if count <= 0 {
		return ArrayLayoutInfo{
			Count:       max(0, count),
			PayloadSize: payloadSize,
			Remainder:   payloadSize,
			IsTightI32:  payloadSize == 0,
		}
	}
	bpi := payloadSize / count
	rem := payloadSize % count
	return ArrayLayoutInfo{
		Count:        count,
		PayloadSize:  payloadSize,
		BytesPerItem: &bpi,
		Remainder:    rem,
		IsTightI32:   rem == 0 && bpi == 4,
	}
}

func ReadArrayPropertyStructHeadI32(data []byte, tag PropertyTag, headI32 int) [][]int {
	if headI32 <= 0 {
		return nil
	}
	info := AnalyzeArrayPropertyLayout(data, tag)
	if info.Count <= 0 || info.BytesPerItem == nil {
		return nil
	}
	stride := *info.BytesPerItem
	headSize := headI32 * 4
	if stride < headSize {
		return nil
	}
	_, payloadStart := resolveArrayCountAndPayloadStart(data, tag)
	rows := make([][]int, info.Count)
	for i := 0; i < info.Count; i++ {
		itemStart := payloadStart + (i * stride)
		row := make([]int, headI32)
		for j := 0; j < headI32; j++ {
			row[j] = readI32(data, itemStart+(j*4))
		}
		rows[i] = row
	}
	return rows
}

func ReadArrayPropertyStructI32Matrix(data []byte, tag PropertyTag) [][]int {
	info := AnalyzeArrayPropertyLayout(data, tag)
	if info.Count <= 0 || info.BytesPerItem == nil {
		return nil
	}
	if *info.BytesPerItem%4 != 0 {
		return nil
	}
	width := *info.BytesPerItem / 4
	if width <= 0 {
		return nil
	}
	_, payloadStart := resolveArrayCountAndPayloadStart(data, tag)
	end := payloadStart + (info.Count * *info.BytesPerItem)
	if end > len(data) {
		return nil
	}
	rows := make([][]int, info.Count)
	for i := 0; i < info.Count; i++ {
		itemStart := payloadStart + (i * *info.BytesPerItem)
		row := make([]int, width)
		for j := 0; j < width; j++ {
			row[j] = readI32(data, itemStart+(j*4))
		}
		rows[i] = row
	}
	return rows
}

// FindInt32ArrayByName scans raw bytes in range [start, end) for an ArrayProperty<IntProperty>
// with the given property name. Returns the int32 element values.
func FindInt32ArrayByName(data []byte, names []string, start, end int, propName string) []int {
	for pos := start; pos < end-24; pos += 4 {
		name, propType, propSize, arrayIndex, valueOffset, _ := parsePropertyHeader(data, names, pos, end)
		if name != propName || propType != "ArrayProperty" || arrayIndex != 0 {
			continue
		}
		tag := PropertyTag{
			Name:        name,
			PropType:    propType,
			Size:        propSize,
			ArrayIndex:  arrayIndex,
			ValueOffset: valueOffset,
		}
		values := ReadArrayPropertyI32Values(data, tag)
		if len(values) > 0 || (tag.ValueOffset+4 <= len(data) && readI32(data, tag.ValueOffset) == 0) {
			return values
		}
	}
	return nil
}

// FindIntPropertyByName scans raw bytes in range [start, end) for a scalar int-like
// property with the given name and returns its value.
func FindIntPropertyByName(data []byte, names []string, start, end int, propName string) (int, bool) {
	for pos := start; pos < end-24; pos += 4 {
		name, propType, propSize, _, valueOffset, _ := parsePropertyHeader(data, names, pos, end)
		if name != propName {
			continue
		}
		switch propType {
		case "IntProperty", "ObjectProperty", "StringRefProperty":
			if propSize >= 4 && valueOffset+4 <= end {
				return readI32(data, valueOffset), true
			}
		}
	}
	return 0, false
}

// FindStructArrayItemStarts scans raw bytes for an ArrayProperty<StructProperty>
// and returns the absolute payload start, item count, and stride hint.
// Returns (-1, 0, 0) if not found.
func FindStructArrayItemStarts(data []byte, names []string, start, end int, propName string) (int, int, int) {
	nameIdx := -1
	arrayIdx := -1
	structIdx := -1
	for i, n := range names {
		if n == propName {
			nameIdx = i
		}
		if n == "ArrayProperty" {
			arrayIdx = i
		}
		if n == "StructProperty" {
			structIdx = i
		}
	}
	if nameIdx < 0 || arrayIdx < 0 || structIdx < 0 {
		return -1, 0, 0
	}
	for pos := start; pos < end-32; pos += 4 {
		nA := readI32(data, pos)
		nB := readI32(data, pos+4)
		tA := readI32(data, pos+8)
		tB := readI32(data, pos+12)
		isName := (nA == nameIdx && nB == 0) || (nA == 0 && nB == nameIdx)
		isType := (tA == arrayIdx && tB == 0) || (tA == 0 && tB == arrayIdx)
		if isName && isType {
			countPos := pos + 16 + 8 + 8
			if countPos+4 > end {
				continue
			}
			count := readI32(data, countPos)
			if count > 0 && count <= 200000 {
				return countPos + 4, count, 0
			}
		}
	}
	return -1, 0, 0
}

type ExportProperties struct {
	Tags          []PropertyTag             `json:"property_tags,omitempty"`
	SemanticProps map[string]ParsedProperty `json:"semantic_props,omitempty"`
}

func ComputeExportProperties(rawData []byte, summary *FileSummary, includeTags, includeSemantic bool) map[int]ExportProperties {
	result := map[int]ExportProperties{}
	if rawData == nil || summary == nil {
		return result
	}
	for _, exp := range summary.Exports {
		if exp.SerialSize <= 0 || exp.SerialOffset < 0 ||
			exp.SerialOffset+exp.SerialSize > len(rawData) {
			continue
		}
		ep := ExportProperties{}
		if includeTags {
			tags, err := ParsePropertyTags(rawData, summary.Names, exp.SerialOffset, exp.SerialSize, false)
			if err == nil {
				ep.Tags = tags
			}
		}
		if includeSemantic {
			props, _ := ParsePropertyCollection(rawData, summary.Names, exp.SerialOffset, exp.SerialSize)
			if props != nil {
				ep.SemanticProps = props
			}
		}
		if includeTags || includeSemantic {
			result[exp.Index] = ep
		}
	}
	return result
}
