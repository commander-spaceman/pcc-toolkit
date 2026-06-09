package tlkwrt

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/commander-spaceman/me2tlk/reader"
)

func TestRealTLK_StressAddManyEntries(t *testing.T) {
	basePath := filepath.Join("..", "..", "..", "output", "BIOGame_INT.tlk")
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		t.Skip("real TLK file not available")
	}

	tlkFile, err := reader.ReadFile(basePath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	maxID := int32(0)
	for id := range tlkFile.MaleEntries {
		if id > maxID {
			maxID = id
		}
	}

	entryCount := 100
	entries := make([]StringEntry, entryCount)
	testText := "STRESS TEST ENTRY"
	for i := 0; i < entryCount; i++ {
		entries[i] = StringEntry{
			StringID: maxID + 1 + int32(i),
			Text:     fmt.Sprintf("%s %03d", testText, i),
			Male:     true,
		}
	}

	err = AddEntries(tlkFile, entries)
	if err != nil {
		t.Fatalf("AddEntries: %v", err)
	}

	// Verify in-memory
	for _, e := range entries {
		text, ok := reader.ResolveString(tlkFile, e.StringID, true)
		if !ok {
			t.Errorf("entry %d not found in-memory", e.StringID)
		} else if text != e.Text {
			t.Errorf("entry %d text: want %q, got %q", e.StringID, e.Text, text)
		}
	}

	// Write to disk
	dir := t.TempDir()
	outPath := filepath.Join(dir, "stress_test.tlk")
	err = WriteFile(outPath, tlkFile, nil)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Re-parse
	readBack, err := reader.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// Verify round-trip
	failures := 0
	for _, e := range entries {
		text, ok := reader.ResolveString(readBack, e.StringID, true)
		if !ok {
			t.Errorf("entry %d lost in round-trip", e.StringID)
			failures++
		} else if text != e.Text {
			t.Errorf("entry %d text mismatch: want %q, got %q", e.StringID, e.Text, text)
			failures++
		}
		if failures >= 5 {
			t.Fatalf("too many failures, aborting")
		}
	}

	if failures == 0 {
		t.Logf("All %d stress entries round-tripped successfully", entryCount)
	}
}
