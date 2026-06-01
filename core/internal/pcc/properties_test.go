package pcc

import "testing"

func TestFindInt32ArrayByName_ArrayWithoutExtraMeta(t *testing.T) {
	names := []string{"EntryList", "ArrayProperty", "None"}
	buf := make([]byte, 64)
	off := 0

	writeFName := func(nameIdx int) {
		putI32LE(buf, off, nameIdx)
		off += 4
		putI32LE(buf, off, 0)
		off += 4
	}

	writeFName(0)
	writeFName(1)
	putI32LE(buf, off, 12)
	off += 4
	putI32LE(buf, off, 0)
	off += 4
	putI32LE(buf, off, 2)
	off += 4
	putI32LE(buf, off, 7)
	off += 4
	putI32LE(buf, off, 9)
	off += 4
	writeFName(2)

	got := FindInt32ArrayByName(buf, names, 0, off, "EntryList")
	if len(got) != 2 || got[0] != 7 || got[1] != 9 {
		t.Fatalf("FindInt32ArrayByName() = %v, want [7 9]", got)
	}
}

func TestFindIntPropertyByName_StringRefProperty(t *testing.T) {
	names := []string{"srText", "StringRefProperty", "None"}
	buf := make([]byte, 64)
	off := 0

	writeFName := func(nameIdx int) {
		putI32LE(buf, off, nameIdx)
		off += 4
		putI32LE(buf, off, 0)
		off += 4
	}

	writeFName(0)
	writeFName(1)
	putI32LE(buf, off, 4)
	off += 4
	putI32LE(buf, off, 0)
	off += 4
	putI32LE(buf, off, 250833)
	off += 4
	writeFName(2)

	got, ok := FindIntPropertyByName(buf, names, 0, off, "srText")
	if !ok || got != 250833 {
		t.Fatalf("FindIntPropertyByName() = (%d, %v), want (250833, true)", got, ok)
	}
}

func putI32LE(buf []byte, offset int, v int) {
	u := uint32(v)
	buf[offset] = byte(u)
	buf[offset+1] = byte(u >> 8)
	buf[offset+2] = byte(u >> 16)
	buf[offset+3] = byte(u >> 24)
}
