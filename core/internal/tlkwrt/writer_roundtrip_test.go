package tlkwrt

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/commander-spaceman/me2tlk/reader"
)

type resolvedEntry struct {
	id   int32
	male bool
	text string
}

func TestRealTLK_NoModifications_RoundTrip(t *testing.T) {
	basePath := filepath.Join("..", "..", "..", "output", "BIOGame_INT.tlk")
	if _, err := os.Stat(basePath); os.IsNotExist(err) {
		t.Skip("real TLK file not available")
	}

	original, err := reader.ReadFile(basePath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	// Collect all resolvable entries, keeping male/female separate
	var origEntries []resolvedEntry
	for id := range original.MaleEntries {
		text, ok := reader.ResolveString(original, id, true)
		if ok {
			origEntries = append(origEntries, resolvedEntry{id: id, male: true, text: text})
		}
	}
	for id := range original.FemaleEntries {
		text, ok := reader.ResolveString(original, id, false)
		if ok {
			origEntries = append(origEntries, resolvedEntry{id: id, male: false, text: text})
		}
	}
	t.Logf("Original: %d male header, %d male map, %d resolvable entries",
		original.Header.MaleEntryCount, len(original.MaleEntries), len(origEntries))

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
	t.Logf("Round-tripped: %d male header, %d male map",
		roundTripped.Header.MaleEntryCount, len(roundTripped.MaleEntries))

	// Header counts must match
	if roundTripped.Header.MaleEntryCount != original.Header.MaleEntryCount {
		t.Errorf("MaleEntryCount: want %d, got %d",
			original.Header.MaleEntryCount, roundTripped.Header.MaleEntryCount)
	}
	if roundTripped.Header.FemaleEntryCount != original.Header.FemaleEntryCount {
		t.Errorf("FemaleEntryCount: want %d, got %d",
			original.Header.FemaleEntryCount, roundTripped.Header.FemaleEntryCount)
	}
	if len(roundTripped.MaleEntries) != len(original.MaleEntries) {
		t.Errorf("Male map size: want %d, got %d",
			len(original.MaleEntries), len(roundTripped.MaleEntries))
	}
	if len(roundTripped.FemaleEntries) != len(original.FemaleEntries) {
		t.Errorf("Female map size: want %d, got %d",
			len(original.FemaleEntries), len(roundTripped.FemaleEntries))
	}

	// Compare each resolvable entry preserving gender
	missing := 0
	mismatched := 0
	for _, orig := range origEntries {
		rtText, ok := reader.ResolveString(roundTripped, orig.id, orig.male)
		if !ok {
			missing++
			if missing <= 5 {
				t.Errorf("entry %d (male=%v) lost in round-trip (was %q)",
					orig.id, orig.male, orig.text)
			}
		} else if rtText != orig.text {
			mismatched++
			if mismatched <= 5 {
				t.Errorf("entry %d (male=%v) text mismatch: orig=%q round=%q",
					orig.id, orig.male, orig.text, rtText)
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
		t.Logf("All %d resolvable entries preserved in round-trip (perfect fidelity)",
			len(origEntries))
	}
}
