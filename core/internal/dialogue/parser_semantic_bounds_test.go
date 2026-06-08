package dialogue

import (
	"testing"

	pcc "github.com/commander-spaceman/me2pcc"
)

func TestFindStructItemBoundsSequentially(t *testing.T) {
	names := []string{"Foo", "IntProperty", "None"}
	buf := make([]byte, 128)
	off := 0

	writeFName := func(nameIdx int) {
		putI32LE(buf, off, nameIdx)
		off += 4
		putI32LE(buf, off, 0)
		off += 4
	}
	writeProp := func(value int) {
		writeFName(0)
		writeFName(1)
		putI32LE(buf, off, 4)
		off += 4
		putI32LE(buf, off, 0)
		off += 4
		putI32LE(buf, off, value)
		off += 4
		writeFName(2)
	}

	start := off
	writeProp(10)
	firstEnd := off
	writeProp(20)
	payloadSize := off - start

	bounds := findStructItemBoundsSequentially(buf, names, start, payloadSize, 2)
	if len(bounds) != 2 {
		t.Fatalf("expected 2 bounds, got %d", len(bounds))
	}
	if bounds[0][0] != start || bounds[0][1] != firstEnd {
		t.Fatalf("unexpected first bounds: got %v want [%d %d]", bounds[0], start, firstEnd)
	}
	if bounds[1][0] != firstEnd || bounds[1][1] <= bounds[1][0] || bounds[1][1] > off {
		t.Fatalf("unexpected second bounds: got %v with final offset %d", bounds[1], off)
	}

	item, _ := pcc.ParsePropertyCollection(buf, names, bounds[1][0], bounds[1][1]-bounds[1][0])
	if item == nil {
		t.Fatal("expected second item to parse")
	}
	if got := item["Foo"].Value.(int); got != 20 {
		t.Fatalf("unexpected second item value: got %d want 20", got)
	}
}
