package pcc

import (
	"encoding/binary"
	"errors"
	"unicode/utf16"
)

func readI32(data []byte, offset int) int {
	return int(int32(binary.LittleEndian.Uint32(data[offset : offset+4])))
}

func ReadRawI32(data []byte, offset int) int {
	return readI32(data, offset)
}

func readU32(data []byte, offset int) uint32 {
	return binary.LittleEndian.Uint32(data[offset : offset+4])
}

func readUnrealString(data []byte, offset int) (string, int, error) {
	if offset < 0 || offset+4 > len(data) {
		return "", offset, errors.New("string offset out of range")
	}
	count := readI32(data, offset)
	cursor := offset + 4
	if count == 0 {
		return "", cursor, nil
	}
	if count > 0 {
		end := cursor + count
		if end > len(data) {
			return "", cursor, errors.New("truncated ascii string")
		}
		raw := data[cursor:end]
		if len(raw) > 0 && raw[len(raw)-1] == 0 {
			raw = raw[:len(raw)-1]
		}
		return string(raw), end, nil
	}
	charCount := -count
	end := cursor + (charCount * 2)
	if end > len(data) {
		return "", cursor, errors.New("truncated utf16 string")
	}
	raw := data[cursor:end]
	if len(raw) >= 2 && raw[len(raw)-2] == 0 && raw[len(raw)-1] == 0 {
		raw = raw[:len(raw)-2]
	}
	if len(raw)%2 != 0 {
		return "", cursor, errors.New("invalid utf16 size")
	}
	units := make([]uint16, len(raw)/2)
	for i := 0; i < len(units); i++ {
		units[i] = binary.LittleEndian.Uint16(raw[i*2 : i*2+2])
	}
	return string(utf16.Decode(units)), end, nil
}

func resolveName(index int, names []string) string {
	if index >= 0 && index < len(names) {
		return names[index]
	}
	return ""
}

func ResolveName(data []byte, offset int, names []string) string {
	if offset+4 > len(data) {
		return ""
	}
	idx := readI32(data, offset)
	return resolveName(idx, names)
}
