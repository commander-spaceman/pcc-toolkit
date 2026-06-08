package pcc

type ParsedProperty struct {
	Name     string      `json:"name"`
	PropType string      `json:"prop_type"`
	Value    interface{} `json:"value"`
}

func readFName(data []byte, names []string, offset int) (string, int) {
	a := readI32(data, offset)
	b := readI32(data, offset+4)
	aName := resolveName(a, names)
	bName := resolveName(b, names)
	if PropertyTypeNames[bName] && !PropertyTypeNames[aName] {
		return resolveName(b, names), 8
	}
	return resolveName(a, names), 8
}

func parsePropertyHeader(data []byte, names []string, cursor, end int) (string, string, int, int, int, int) {
	if cursor+24 > end {
		return "", "", 0, 0, cursor, 0
	}
	name, nameLen := readFName(data, names, cursor)
	cursor += nameLen
	if name == "None" {
		return name, "", 0, 0, cursor, 0
	}
	propType, typeLen := readFName(data, names, cursor)
	cursor += typeLen
	propSize := readI32(data, cursor)
	arrayIndex := readI32(data, cursor+4)
	cursor += 8

	metaSize := 0
	switch propType {
	case "StructProperty", "ByteProperty":
		metaSize = 8
	case "BoolProperty":
		oneByteNext := cursor + 1 + propSize
		fourByteNext := cursor + 4 + propSize
		if isPlausibleTagStart(data, fourByteNext, end, names) && !isPlausibleTagStart(data, oneByteNext, end, names) {
			metaSize = 4
		} else {
			metaSize = 1
		}
	case "ArrayProperty":
		if cursor+8 <= end {
			a := readI32(data, cursor)
			b := readI32(data, cursor+4)
			aName := resolveName(a, names)
			bName := resolveName(b, names)
			if PropertyTypeNames[aName] || PropertyTypeNames[bName] {
				metaSize = 8
				break
			}
		}
		noMetaNext := cursor + propSize
		fnameMetaNext := cursor + 8 + propSize
		if isPlausibleTagStart(data, fnameMetaNext, end, names) && !isPlausibleTagStart(data, noMetaNext, end, names) {
			metaSize = 8
		}
	}
	valueOffset := cursor + metaSize
	return name, propType, propSize, arrayIndex, valueOffset, metaSize
}

func ParsePropertyCollection(data []byte, names []string, startOffset, maxSize int) (map[string]ParsedProperty, int) {
	end := startOffset + maxSize
	if end > len(data) {
		end = len(data)
	}
	cursor := startOffset
	props := map[string]ParsedProperty{}

	for cursor+8 <= end {
		name, propType, propSize, _, valueOffset, metaSize := parsePropertyHeader(data, names, cursor, end)
		if name == "" || name == "None" {
			if len(props) == 0 {
				return nil, valueOffset
			}
			return props, valueOffset
		}
		cursor += 16 + 8 + metaSize

		if propSize < 0 {
			return props, cursor
		}
		if valueOffset+propSize > end {
			return props, cursor
		}

		var value interface{}
		switch propType {
		case "IntProperty", "ObjectProperty", "StringRefProperty", "NameProperty", "EnumProperty":
			if propSize >= 4 && valueOffset+4 <= len(data) {
				raw := readI32(data, valueOffset)
				if (propType == "NameProperty" || propType == "EnumProperty") && raw >= 0 && raw < len(names) {
					value = names[raw]
				} else {
					value = raw
				}
			}
		case "StrProperty":
			if propSize >= 4 && valueOffset+4 <= len(data) {
				str, _, err := readUnrealString(data, valueOffset)
				if err == nil {
					value = str
				}
			}
		case "BoolProperty":
			if metaSize >= 4 {
				value = readI32(data, valueOffset-metaSize) != 0
			} else if valueOffset-1 >= 0 {
				value = data[valueOffset-1] != 0
			}
		case "ArrayProperty":
			if valueOffset+4 <= len(data) {
				count := readI32(data, valueOffset)
				pSize := propSize - 4 - metaSize
				if pSize < 0 {
					pSize = 0
				}
				value = map[string]interface{}{
					"count":          count,
					"payload_offset": valueOffset + 4,
					"payload_size":   pSize,
				}
			}
		case "StructProperty":
			nested, _ := ParsePropertyCollection(data, names, valueOffset, propSize)
			value = nested
		}

		props[name] = ParsedProperty{Name: name, PropType: propType, Value: value}
		cursor = valueOffset + propSize
	}

	return props, cursor
}

func ParseStructArrayItemsAsPropertyCollections(
	data []byte, names []string, payloadOffset, payloadSize, count int,
) []map[string]ParsedProperty {
	if count <= 0 || payloadSize <= 0 {
		return nil
	}
	end := payloadOffset + payloadSize
	if end > len(data) {
		end = len(data)
	}

	starts := findRepeatedStructItemStarts(data, names, payloadOffset, end, count)
	if len(starts) > 1 {
		bounds := append(starts, end)
		var items []map[string]ParsedProperty
		for i, itemStart := range starts {
			itemEnd := bounds[i+1]
			item, _ := ParsePropertyCollection(data, names, itemStart, itemEnd-itemStart)
			if item != nil {
				items = append(items, item)
			}
		}
		return items
	}

	if payloadSize%count == 0 {
		stride := payloadSize / count
		if stride > 0 {
			var items []map[string]ParsedProperty
			for i := 0; i < count; i++ {
				itemStart := payloadOffset + (i * stride)
				item, _ := ParsePropertyCollection(data, names, itemStart, stride)
				if item == nil {
					break
				}
				items = append(items, item)
			}
			if len(items) == count {
				return items
			}
		}
	}

	// Strategy 4: scan for plausible tag starts at any position
	if count > 0 {
		candidates := findPlausibleStructStarts(data, names, payloadOffset, end)
		if len(candidates) >= count {
			var items []map[string]ParsedProperty
			bounds := append(candidates, end)
			for i := 0; i < count && i < len(candidates); i++ {
				itemStart := candidates[i]
				itemEnd := bounds[i+1]
				item, _ := ParsePropertyCollection(data, names, itemStart, itemEnd-itemStart)
				if item != nil {
					items = append(items, item)
				}
			}
			if len(items) == count {
				return items
			}
		}
	}

	// Sequential parsing fallback (original strategy 3)

	var items []map[string]ParsedProperty
	cursor := payloadOffset
	for i := 0; i < count; i++ {
		if cursor+8 > end {
			break
		}
		item, nextCursor := ParsePropertyCollection(data, names, cursor, end-cursor)
		if item == nil {
			break
		}
		items = append(items, item)
		if nextCursor <= cursor {
			break
		}
		cursor = nextCursor
	}
	return items
}

func findRepeatedStructItemStarts(data []byte, names []string, payloadOffset, end, count int) []int {
	if count <= 1 || payloadOffset+16 > end {
		return nil
	}
	firstA := readI32(data, payloadOffset)
	firstB := readI32(data, payloadOffset+4)
	if !isPlausibleTagStart(data, payloadOffset, end, names) {
		return nil
	}
	var starts []int
	for pos := payloadOffset; pos < end-16; pos += 4 {
		if readI32(data, pos) != firstA || readI32(data, pos+4) != firstB {
			continue
		}
		if !isPlausibleTagStart(data, pos, end, names) {
			continue
		}
		starts = append(starts, pos)
		if len(starts) == count {
			break
		}
	}
	if len(starts) > 1 {
		return starts
	}
	return nil
}

func HasStructSignature(data []byte, names []string, payloadOffset, payloadSize int) bool {
	if payloadSize < 24 {
		return false
	}
	end := payloadOffset + payloadSize
	if payloadOffset < 0 || end > len(data) {
		return false
	}
	a := readI32(data, payloadOffset)
	b := readI32(data, payloadOffset+4)
	ta := readI32(data, payloadOffset+8)
	tb := readI32(data, payloadOffset+12)
	aOk := a >= 0 && a < len(names)
	bOk := b >= 0 && b < len(names)
	if !aOk && !bOk {
		return false
	}
	taName := resolveName(ta, names)
	tbName := resolveName(tb, names)
	return PropertyTypeNames[taName] || PropertyTypeNames[tbName]
}

func findPlausibleStructStarts(data []byte, names []string, payloadOffset, end int) []int {
	var starts []int
	for pos := payloadOffset; pos < end-16; pos += 4 {
		if isPlausibleTagStart(data, pos, end, names) {
			starts = append(starts, pos)
		}
	}
	return starts
}

func ParsePropertyHeader(data []byte, names []string, cursor, end int) (string, string, int, int, int, int) {
	return parsePropertyHeader(data, names, cursor, end)
}
