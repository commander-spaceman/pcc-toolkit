package pccenc

import (
	"encoding/binary"
	"errors"
	"unicode/utf16"
)

type writeBuffer struct {
	buf []byte
	pos int
}

func newWriteBuffer(capacity int) *writeBuffer {
	return &writeBuffer{buf: make([]byte, capacity)}
}

func (w *writeBuffer) WriteI32(v int) {
	binary.LittleEndian.PutUint32(w.buf[w.pos:], uint32(int32(v)))
	w.pos += 4
}

func (w *writeBuffer) WriteU32(v uint32) {
	binary.LittleEndian.PutUint32(w.buf[w.pos:], v)
	w.pos += 4
}

func (w *writeBuffer) WriteFName(nameIdx int, nameNumber int) {
	w.WriteI32(nameIdx)
	w.WriteI32(nameNumber)
}

func (w *writeBuffer) WriteBytes(b []byte) {
	copy(w.buf[w.pos:], b)
	w.pos += len(b)
}

func (w *writeBuffer) Bytes() []byte {
	return w.buf[:w.pos]
}

func writeFNameSize(nameIdx, nameNumber int) int {
	return 8
}

func writeUnrealStringLatin1(s string) []byte {
	strLen := len(s) + 1
	buf := make([]byte, 4+strLen)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(strLen))
	copy(buf[4:4+len(s)], s)
	buf[4+len(s)] = 0
	return buf
}

func writeUnrealStringUnicode(s string) ([]byte, error) {
	runes := []rune(s)
	charCount := len(runes) + 1
	buf := make([]byte, 4+charCount*2)
	binary.LittleEndian.PutUint32(buf[0:4], uint32(-charCount))

	units := utf16.Encode(runes)
	for i, u := range units {
		binary.LittleEndian.PutUint16(buf[4+i*2:], u)
	}
	binary.LittleEndian.PutUint16(buf[4+len(units)*2:], 0)
	return buf, nil
}

func buildNameLookup(names []string) map[string]int {
	m := make(map[string]int, len(names))
	for i, n := range names {
		m[n] = i
	}
	return m
}

func resolveNameIdx(name string, names []string, nameLookup map[string]int) (int, error) {
	if name == "" {
		return -1, errors.New("empty name not allowed")
	}
	if idx, ok := nameLookup[name]; ok {
		return idx, nil
	}
	return -1, errors.New("name not found in table: " + name)
}

func resolveTypeIdx(propType string, names []string, nameLookup map[string]int) (int, error) {
	if propType == "" {
		return -1, errors.New("empty property type")
	}
	idx, err := resolveNameIdx(propType, names, nameLookup)
	if err != nil {
		return -1, errors.New("property type not found in name table: " + propType)
	}
	return idx, nil
}
