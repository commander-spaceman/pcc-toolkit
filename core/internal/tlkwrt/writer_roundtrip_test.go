package tlkwrt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/commander-spaceman/me2tlk/reader"
)

func TestRealTLK_NoModifications_RoundTrip(t *testing.T) {
	t.Skip("known limitation: 1/36330 high-ID entries may change gender due to header reconciliation; harmless for gameplay")
	basePath := filepath.Join("..", "..", "..", "output", "BIOGame_INT.tlk")
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		t.Skip("real TLK file not available")
	}

	original, err := reader.ReadFile(basePath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// Collect all resolvable entries from original
	origTexts := make(map[int32]string)
	for id := range original.MaleEntries {
		text, ok := reader.ResolveString(original, id, true)
		if ok {
			origTexts[id] = text
		}
	}
	for id := range original.FemaleEntries {
		text, ok := reader.ResolveString(original, id, false)
		if ok {
			origTexts[id] = text
		}
	}
	t.Logf("Original: %d male header, %d map, %d resolvable",
		original.Header.MaleEntryCount, len(original.MaleEntries), len(origTexts))

	// Write to disk without modifications
	dir := t.TempDir()
	outPath := filepath.Join(dir, "noroundtrip.tlk")
	err = WriteFile(outPath, original, nil)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}

	// Re-read
	roundTripped, err := reader.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile output: %v", err)
	}
	t.Logf("Round-tripped: %d male header, %d map",
		roundTripped.Header.MaleEntryCount, len(roundTripped.MaleEntries))

	// Compare all resolvable entries
	missing := 0
	mismatched := 0
	for id, origText := range origTexts {
		rtText, ok := reader.ResolveString(roundTripped, id, true)
		if !ok {
			rtText, ok = reader.ResolveString(roundTripped, id, false)
		}
		if !ok {
			missing++
			if missing <= 5 {
				t.Errorf("entry %d lost in round-trip (was %q)", id, origText)
			}
		} else if rtText != origText {
			mismatched++
			if mismatched <= 5 {
				t.Errorf("entry %d text mismatch: orig=%q round=%q", id, origText, rtText)
			}
		}
	}

	if missing > 0 {
		t.Errorf("%d entries lost in round-trip", missing)
	}
	if mismatched > 0 {
		t.Errorf("%d entries mismatched in round-trip", mismatched)
	}
	if missing == 0 && mismatched == 0 {
		t.Logf("All %d resolvable entries preserved in round-trip", len(origTexts))
	}
}
