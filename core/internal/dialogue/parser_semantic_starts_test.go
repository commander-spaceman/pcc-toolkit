package dialogue

import "testing"

func TestFindStructItemBoundsByRepeatedStarts(t *testing.T) {
	names := []string{"Foo", "IntProperty", "Bar", "None"}
	buf := make([]byte, 128)
	off := 0

	writeFName := func(nameIdx int) {
		putI32LE(buf, off, nameIdx)
		off += 4
		putI32LE(buf, off, 0)
		off += 4
	}
	writeProp := func(nameIdx, value int) {
		writeFName(nameIdx)
		writeFName(1)
		putI32LE(buf, off, 4)
		off += 4
		putI32LE(buf, off, 0)
		off += 4
		putI32LE(buf, off, value)
		off += 4
	}
	writeItem := func(value int, withExtra bool) {
		writeProp(0, value)
		if withExtra {
			writeProp(2, value+100)
		}
		writeFName(3)
	}

	start := off
	writeItem(10, false)
	firstEnd := off
	writeItem(20, true)
	bounds := findStructItemBoundsByRepeatedStarts(buf, names, start, off-start, 2)
	if len(bounds) != 2 {
		t.Fatalf("expected 2 bounds, got %d", len(bounds))
	}
	if bounds[0][0] != start || bounds[0][1] <= bounds[0][0] || bounds[0][1] > firstEnd {
		t.Fatalf("unexpected first bounds: %v with firstEnd %d", bounds[0], firstEnd)
	}
	if bounds[1][0] < bounds[0][1] || bounds[1][1] != off {
		t.Fatalf("unexpected second bounds: %v with final offset %d", bounds[1], off)
	}
}
