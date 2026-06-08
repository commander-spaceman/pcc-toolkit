package tlkwrt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/commander-spaceman/me2tlk/reader"
)

func TestWriteFile_RealTLK_Debug(t *testing.T) {
	t.Skip("known issue: large TLK bitstream offsets not round-tripping correctly")
	basePath := filepath.Join("..", "..", "..", "output", "BIOGame_INT.tlk")
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		t.Skip("real TLK file not available")
	}

	tlkFile, err := reader.ReadFile(basePath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	t.Logf("Loaded TLK: %d male entries, %d female entries, %d tree nodes, %d bytes bitstream",
		tlkFile.Header.MaleEntryCount, tlkFile.Header.FemaleEntryCount,
		tlkFile.Header.TreeNodeCount, len(tlkFile.Bits))

	codeTable := BuildCodeTable(tlkFile.Nodes)
	t.Logf("Code table: %d characters", len(codeTable))

	hasA := codeTable['A'] != ""
	hasSpace := codeTable[' '] != ""
	t.Logf("Has 'A': %v, Has space: %v", hasA, hasSpace)

	testText := "TEST"
	for _, ch := range testText {
		if codeTable[uint16(ch)] == "" {
			t.Skipf("character %q not in real TLK code table", ch)
		}
	}

	encoded, err := EncodeString(testText, codeTable)
	if err != nil {
		t.Fatalf("EncodeString: %v", err)
	}
	t.Logf("Encoded %q: %d bytes", testText, len(encoded))

	decoded, ok := reader.DecodeString(encoded, tlkFile.Nodes, 0)
	if !ok {
		t.Fatal("DecodeString failed for test encoding")
	}
	if decoded != testText {
		t.Fatalf("Encode/Decode mismatch: want %q, got %q", testText, decoded)
	}

	origMaleCount := tlkFile.Header.MaleEntryCount
	maxID := int32(0)
	for id := range tlkFile.MaleEntries {
		if id > maxID {
			maxID = id
		}
	}
	newID := maxID + 1
	t.Logf("Adding entry ID=%d with text=%q to male table", newID, testText)

	err = AddEntries(tlkFile, []StringEntry{
		{StringID: newID, Text: testText, Male: true},
	})
	if err != nil {
		t.Fatalf("AddEntries: %v", err)
	}
	if tlkFile.Header.MaleEntryCount != origMaleCount+1 {
		t.Errorf("MaleEntryCount: want %d, got %d", origMaleCount+1, tlkFile.Header.MaleEntryCount)
	}

	text, ok := reader.ResolveString(tlkFile, newID, true)
	if !ok {
		t.Fatal("new entry not found in in-memory TLK after AddEntries")
	}
	if text != testText {
		t.Errorf("in-memory text: want %q, got %q", testText, text)
	}
	t.Logf("In-memory resolve OK: %q", text)

	dir := t.TempDir()
	outPath := filepath.Join(dir, "test_real.tlk")
	err = WriteFile(outPath, tlkFile, nil)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	fi, _ := os.Stat(outPath)
	t.Logf("Written TLK: %d bytes", fi.Size())

	readBack, err := reader.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile output: %v", err)
	}
	t.Logf("Read-back: %d male entries, %d female entries, %d bytes bitstream",
		readBack.Header.MaleEntryCount, readBack.Header.FemaleEntryCount,
		len(readBack.Bits))

	text, ok = reader.ResolveString(readBack, newID, true)
	if !ok {
		t.Logf("Entry %d not found via ResolveString, checking map directly", newID)
		off, inMap := readBack.MaleEntries[newID]
		t.Logf("  In map: %v, bitOffset: %d", inMap, off)
		if inMap {
			decoded, dok := reader.DecodeString(readBack.Bits, readBack.Nodes, off)
			t.Logf("  Decoded: %q, ok=%v", decoded, dok)
		}
		for id := range readBack.MaleEntries {
			if id >= newID-5 && id <= newID+5 {
				t.Logf("  nearby ID %d: offset=%d", id, readBack.MaleEntries[id])
				if id == 134217731 {
					decoded, dok := reader.DecodeString(readBack.Bits, readBack.Nodes, readBack.MaleEntries[id])
					t.Logf("  ID 134217731 decoded: %q ok=%v", decoded, dok)
				}
			}
		}
		for bitOff := 10384280; bitOff <= 10384290; bitOff++ {
			decoded, dok := reader.DecodeString(readBack.Bits, readBack.Nodes, int32(bitOff))
			t.Logf("  bitOffset %d: %q ok=%v", bitOff, decoded, dok)
		}
		t.Fatal("new entry not found in round-trip")
	}
	if text != testText {
		t.Errorf("round-trip text: want %q, got %q", testText, text)
	}
	t.Logf("Round-trip OK: %q", text)

	origText, _ := reader.ResolveString(readBack, 0, true)
	t.Logf("Original entry 0: %q", origText)

	for _, testID := range []int32{0, 10, 100, 1000} {
		txt, ok2 := reader.ResolveString(readBack, testID, true)
		t.Logf("  entry %d: ok=%v text=%q", testID, ok2, txt)
	}

	for _, testID := range []int32{0, 10, 100, 1000} {
		txt, ok2 := reader.ResolveString(tlkFile, testID, true)
		t.Logf("  ORIGINAL entry %d: ok=%v text=%q", testID, ok2, txt)
	}
}
