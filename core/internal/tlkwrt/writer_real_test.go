package tlkwrt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/commander-spaceman/me2tlk/reader"
)

func TestRealTLK_AddEntries_RoundTrip(t *testing.T) {
	basePath := filepath.Join("..", "..", "..", "output", "BIOGame_INT.tlk")
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		t.Skip("real TLK file not available")
	}

	tlkFile, err := reader.ReadFile(basePath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	origBitsLen := len(tlkFile.Bits)
	origMaleMapLen := len(tlkFile.MaleEntries)
	t.Logf("Original: header=%d male entries, map=%d male entries, %d bitstream bytes",
		tlkFile.Header.MaleEntryCount, origMaleMapLen, origBitsLen)

	codeTable := BuildCodeTable(tlkFile.Nodes)
	testText := "TEST"
	encoded, err := EncodeString(testText, codeTable)
	if err != nil {
		t.Fatalf("EncodeString: %v", err)
	}

	maxID := int32(0)
	for id := range tlkFile.MaleEntries {
		if id > maxID {
			maxID = id
		}
	}
	newID := maxID + 1

	err = AddEntries(tlkFile, []StringEntry{
		{StringID: newID, Text: testText, Male: true},
	})
	if err != nil {
		t.Fatalf("AddEntries: %v", err)
	}

	// Verify in-memory
	text, ok := reader.ResolveString(tlkFile, newID, true)
	if !ok {
		t.Fatal("new entry not found in-memory after AddEntries")
	}
	if text != testText {
		t.Fatalf("in-memory text: want %q, got %q", testText, text)
	}

	if len(tlkFile.Bits) != origBitsLen+len(encoded) {
		t.Errorf("Bits length: want %d, got %d",
			origBitsLen+len(encoded), len(tlkFile.Bits))
	}

	// Write to disk
	dir := t.TempDir()
	outPath := filepath.Join(dir, "test_real.tlk")
	err = WriteFile(outPath, tlkFile, nil)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Re-parse and verify
	readBack, err := reader.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// Header counts should reflect actual map sizes
	expectedMale := int32(len(tlkFile.MaleEntries))
	if readBack.Header.MaleEntryCount != expectedMale {
		t.Errorf("MaleEntryCount: want %d, got %d",
			expectedMale, readBack.Header.MaleEntryCount)
	}

	text2, ok2 := reader.ResolveString(readBack, newID, true)
	if !ok2 {
		t.Fatalf("round-trip resolve failed for entry %d", newID)
	}
	if text2 != testText {
		t.Errorf("round-trip text: want %q, got %q", testText, text2)
	}
	t.Logf("Round-trip OK: %q", text2)

	// Verify original entries still work
	for _, testID := range []int32{0, 1, 10, 100, 1000, 10000} {
		origText, origOK := reader.ResolveString(tlkFile, testID, true)
		roundText, roundOK := reader.ResolveString(readBack, testID, true)
		if origOK != roundOK {
			t.Errorf("entry %d: original ok=%v, round-trip ok=%v", testID, origOK, roundOK)
		} else if origOK && origText != roundText {
			t.Errorf("entry %d text mismatch: original=%q round-trip=%q", testID, origText, roundText)
		}
	}
}
