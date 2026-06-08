package tlk

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestBuildCodeTable_3Node(t *testing.T) {
	nodes := []Node{
		{LeftNodeID: -66, RightNodeID: 1},
		{LeftNodeID: -67, RightNodeID: 2},
		{LeftNodeID: -1, RightNodeID: -68},
	}
	table := BuildCodeTable(nodes)
	if len(table) != 4 {
		t.Fatalf("expected 4 entries, got %d", len(table))
	}
	if code, ok := table[uint16('A')]; !ok {
		t.Error("'A' not in table")
	} else if code != "0" {
		t.Errorf("code for 'A': want %q, got %q", "0", code)
	}
	if code, ok := table[uint16('B')]; !ok {
		t.Error("'B' not in table")
	} else if code != "10" {
		t.Errorf("code for 'B': want %q, got %q", "10", code)
	}
	if code, ok := table[uint16('C')]; !ok {
		t.Error("'C' not in table")
	} else if code != "111" {
		t.Errorf("code for 'C': want %q, got %q", "111", code)
	}
	if code, ok := table[uint16(0)]; !ok {
		t.Error("null terminator not in table")
	} else if code != "110" {
		t.Errorf("code for null: want %q, got %q", "110", code)
	}
}

func TestBuildCodeTable_5Node(t *testing.T) {
	nodes := []Node{
		{LeftNodeID: 1, RightNodeID: 2},
		{LeftNodeID: -66, RightNodeID: -67},
		{LeftNodeID: 3, RightNodeID: 4},
		{LeftNodeID: -68, RightNodeID: -69},
		{LeftNodeID: -1, RightNodeID: -70},
	}
	table := BuildCodeTable(nodes)
	if len(table) != 6 {
		t.Fatalf("expected 6 entries, got %d", len(table))
	}
	if code, ok := table[uint16('A')]; !ok || code != "00" {
		t.Errorf("'A': ok=%v code=%q want 00", ok, code)
	}
	if code, ok := table[uint16('B')]; !ok || code != "01" {
		t.Errorf("'B': ok=%v code=%q want 01", ok, code)
	}
	if code, ok := table[uint16('C')]; !ok || code != "100" {
		t.Errorf("'C': ok=%v code=%q want 100", ok, code)
	}
	if code, ok := table[uint16('D')]; !ok || code != "101" {
		t.Errorf("'D': ok=%v code=%q want 101", ok, code)
	}
	if code, ok := table[uint16('E')]; !ok || code != "111" {
		t.Errorf("'E': ok=%v code=%q want 111", ok, code)
	}
	if code, ok := table[uint16(0)]; !ok || code != "110" {
		t.Errorf("null: ok=%v code=%q want 110", ok, code)
	}
}

func TestBuildCodeTable_Deep(t *testing.T) {
	nodes := []Node{
		{LeftNodeID: -66, RightNodeID: 1},
		{LeftNodeID: -67, RightNodeID: 2},
		{LeftNodeID: -68, RightNodeID: 3},
		{LeftNodeID: -69, RightNodeID: 4},
		{LeftNodeID: -70, RightNodeID: 5},
		{LeftNodeID: -1, RightNodeID: -71},
	}
	table := BuildCodeTable(nodes)
	if len(table) != 7 {
		t.Fatalf("expected 7 entries, got %d", len(table))
	}
	if code, ok := table[uint16('A')]; !ok || code != "0" {
		t.Errorf("'A' depth 1: code=%q", code)
	}
	if code, ok := table[uint16('E')]; !ok || len(code) != 5 {
		t.Errorf("'E' depth 5: code=%q len=%d", code, len(code))
	}
	if code, ok := table[uint16('F')]; !ok || len(code) != 6 {
		t.Errorf("'F' depth 6: code=%q len=%d", code, len(code))
	}
	if _, ok := table[uint16(0)]; !ok {
		t.Error("null terminator not in deep tree")
	}
}

func TestBuildCodeTable_Empty(t *testing.T) {
	table := BuildCodeTable(nil)
	if len(table) != 0 {
		t.Errorf("expected empty table for nil, got %d entries", len(table))
	}
	table = BuildCodeTable([]Node{})
	if len(table) != 0 {
		t.Errorf("expected empty table for empty slice, got %d entries", len(table))
	}
}

func TestEncodeString_SingleChar(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	table := BuildCodeTable(tlkFile.Nodes)
	encoded, err := EncodeString("A", table)
	if err != nil {
		t.Fatalf("EncodeString: %v", err)
	}
	if len(encoded) == 0 {
		t.Fatal("encoded output is empty")
	}
	decoded, ok := DecodeString(encoded, tlkFile.Nodes, 0)
	if !ok {
		t.Fatal("DecodeString failed")
	}
	if decoded != "A" {
		t.Errorf("round-trip: want %q, got %q", "A", decoded)
	}
}

func TestEncodeString_MultipleChars(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	table := BuildCodeTable(tlkFile.Nodes)
	input := "BABA"
	encoded, err := EncodeString(input, table)
	if err != nil {
		t.Fatalf("EncodeString: %v", err)
	}
	decoded, ok := DecodeString(encoded, tlkFile.Nodes, 0)
	if !ok {
		t.Fatal("DecodeString failed")
	}
	if decoded != input {
		t.Errorf("round-trip: want %q, got %q", input, decoded)
	}
}

func TestEncodeString_Empty(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	table := BuildCodeTable(tlkFile.Nodes)
	encoded, err := EncodeString("", table)
	if err != nil {
		t.Fatalf("EncodeString: %v", err)
	}
	decoded, ok := DecodeString(encoded, tlkFile.Nodes, 0)
	if !ok {
		t.Fatal("DecodeString failed for empty string")
	}
	if decoded != "" {
		t.Errorf("round-trip: want %q, got %q", "", decoded)
	}
}

func TestEncodeString_UnicodeMultiple(t *testing.T) {
	nodes := []Node{
		{LeftNodeID: -242, RightNodeID: 1},
		{LeftNodeID: -234, RightNodeID: 2},
		{LeftNodeID: -232, RightNodeID: -1},
	}
	table := BuildCodeTable(nodes)
	for ch, code := range table {
		t.Logf("U+%04X (%q) -> %q", ch, rune(ch), code)
	}
	input := "ñéç"
	encoded, err := EncodeString(input, table)
	if err != nil {
		t.Fatalf("EncodeString: %v", err)
	}
	decoded, ok := DecodeString(encoded, nodes, 0)
	if !ok {
		t.Fatal("DecodeString failed")
	}
	if decoded != input {
		t.Errorf("round-trip: want %q, got %q", input, decoded)
	}
}

func TestEncodeString_CharNotFound_Extra(t *testing.T) {
	nodes := []Node{
		{LeftNodeID: -66, RightNodeID: -1},
	}
	table := BuildCodeTable(nodes)
	_, err := EncodeString("Z", table)
	if err == nil {
		t.Fatal("expected error for character not in code table")
	}
}

func TestEncodeString_NoNullTerminator_Extra(t *testing.T) {
	nodes := []Node{
		{LeftNodeID: -66, RightNodeID: -67},
	}
	table := BuildCodeTable(nodes)
	_, err := EncodeString("A", table)
	if err == nil {
		t.Fatal("expected error for missing null terminator")
	}
}

func TestEncodeString_RoundTrip(t *testing.T) {
	nodes := []Node{
		{LeftNodeID: -66, RightNodeID: 1},
		{LeftNodeID: -67, RightNodeID: 2},
		{LeftNodeID: -68, RightNodeID: -1},
	}
	table := BuildCodeTable(nodes)
	testCases := []string{"A", "B", "C", "AB", "BA", "ABC", "CBA", "AABBCC"}
	for _, tc := range testCases {
		encoded, err := EncodeString(tc, table)
		if err != nil {
			t.Errorf("EncodeString(%q): %v", tc, err)
			continue
		}
		decoded, ok := DecodeString(encoded, nodes, 0)
		if !ok {
			t.Errorf("DecodeString(%q) failed", tc)
			continue
		}
		if decoded != tc {
			t.Errorf("round-trip %q: got %q", tc, decoded)
		}
	}
}

func TestEncodeString_LongText(t *testing.T) {
	nodes := []Node{
		{LeftNodeID: -66, RightNodeID: 1},
		{LeftNodeID: -67, RightNodeID: -1},
	}
	table := BuildCodeTable(nodes)
	var sb strings.Builder
	for i := 0; i < 120; i++ {
		if i%2 == 0 {
			sb.WriteByte('A')
		} else {
			sb.WriteByte('B')
		}
	}
	longText := sb.String()
	encoded, err := EncodeString(longText, table)
	if err != nil {
		t.Fatalf("EncodeString long text: %v", err)
	}
	decoded, ok := DecodeString(encoded, nodes, 0)
	if !ok {
		t.Fatal("DecodeString long text failed")
	}
	if decoded != longText {
		t.Errorf("long text round-trip mismatch")
	}
	t.Logf("long text %d chars encoded to %d bytes", len(longText), len(encoded))
}

func TestEncodeString_BitBoundary(t *testing.T) {
	nodes := []Node{
		{LeftNodeID: -66, RightNodeID: 1},
		{LeftNodeID: -67, RightNodeID: -1},
	}
	table := BuildCodeTable(nodes)
	encoded, err := EncodeString("AB", table)
	if err != nil {
		t.Fatalf("EncodeString: %v", err)
	}
	if len(encoded) == 0 {
		t.Fatal("encoded output is empty")
	}
	lastByte := encoded[len(encoded)-1]
	for i := 0; i < 8; i++ {
		t.Logf("  byte %d bit %d = %d", len(encoded)-1, i, (lastByte>>i)&1)
	}
	decoded, ok := DecodeString(encoded, nodes, 0)
	if !ok {
		t.Fatal("DecodeString failed")
	}
	if decoded != "AB" {
		t.Errorf("round-trip: want %q, got %q", "AB", decoded)
	}
}

func TestAddEntries_SingleMale(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	err = AddEntries(tlkFile, []StringEntry{
		{StringID: 42, Text: "AB", Male: true},
	})
	if err != nil {
		t.Fatalf("AddEntries: %v", err)
	}
	text, ok := ResolveString(tlkFile, 42, true)
	if !ok {
		t.Fatal("new male entry not found")
	}
	if text != "AB" {
		t.Errorf("text: want %q, got %q", "AB", text)
	}
}

func TestAddEntries_SingleFemale(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	err = AddEntries(tlkFile, []StringEntry{
		{StringID: 77, Text: "BA", Male: false},
	})
	if err != nil {
		t.Fatalf("AddEntries: %v", err)
	}
	text, ok := ResolveString(tlkFile, 77, false)
	if !ok {
		t.Fatal("new female entry not found")
	}
	if text != "BA" {
		t.Errorf("text: want %q, got %q", "BA", text)
	}
}

func TestAddEntries_Multiple(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	entries := make([]StringEntry, 12)
	for i := 0; i < 12; i++ {
		entries[i] = StringEntry{
			StringID: int32(100 + i),
			Text:     "AB",
			Male:     i%2 == 0,
		}
	}
	err = AddEntries(tlkFile, entries)
	if err != nil {
		t.Fatalf("AddEntries: %v", err)
	}
	for _, e := range entries {
		text, ok := ResolveString(tlkFile, e.StringID, e.Male)
		if !ok {
			t.Errorf("entry %d (male=%v) not found", e.StringID, e.Male)
		} else if text != "AB" {
			t.Errorf("entry %d text: want %q, got %q", e.StringID, "AB", text)
		}
	}
}

func TestAddEntries_ExistingID(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	err = AddEntries(tlkFile, []StringEntry{
		{StringID: 1, Text: "SHOULD_NOT_OVERWRITE", Male: true},
	})
	if err != nil {
		t.Fatalf("AddEntries: %v", err)
	}
	text, ok := ResolveString(tlkFile, 1, true)
	if !ok {
		t.Fatal("original entry not found")
	}
	if text != "AB" {
		t.Errorf("original text was overwritten: want %q, got %q", "AB", text)
	}
}

func TestAddEntries_IncrementsCounts(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	origMale := tlkFile.Header.MaleEntryCount
	origFemale := tlkFile.Header.FemaleEntryCount
	origTotal := tlkFile.TotalEntries

	err = AddEntries(tlkFile, []StringEntry{
		{StringID: 10, Text: "AB", Male: true},
		{StringID: 20, Text: "AB", Male: true},
		{StringID: 30, Text: "AB", Male: false},
	})
	if err != nil {
		t.Fatalf("AddEntries: %v", err)
	}
	if tlkFile.Header.MaleEntryCount != origMale+2 {
		t.Errorf("MaleEntryCount: want %d, got %d", origMale+2, tlkFile.Header.MaleEntryCount)
	}
	if tlkFile.Header.FemaleEntryCount != origFemale+1 {
		t.Errorf("FemaleEntryCount: want %d, got %d", origFemale+1, tlkFile.Header.FemaleEntryCount)
	}
	if tlkFile.TotalEntries != origTotal+3 {
		t.Errorf("TotalEntries: want %d, got %d", origTotal+3, tlkFile.TotalEntries)
	}
}

func TestWriteFile_RoundTripSingle(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "test.tlk")
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	err = WriteFile(outPath, tlkFile, []StringEntry{
		{StringID: 99, Text: "BA", Male: true},
	})
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	readBack, err := ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	text, ok := ResolveString(readBack, 99, true)
	if !ok {
		t.Fatal("new entry not found after round-trip")
	}
	if text != "BA" {
		t.Errorf("text: want %q, got %q", "BA", text)
	}
}

func TestWriteFile_RoundTripMultiple(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "test.tlk")
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	entries := make([]StringEntry, 15)
	for i := 0; i < 15; i++ {
		entries[i] = StringEntry{
			StringID: int32(200 + i),
			Text:     "AB",
			Male:     i%2 == 0,
		}
	}
	err = WriteFile(outPath, tlkFile, entries)
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	readBack, err := ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	for _, e := range entries {
		text, ok := ResolveString(readBack, e.StringID, e.Male)
		if !ok {
			t.Errorf("entry %d (male=%v) not found after round-trip", e.StringID, e.Male)
		} else if text != "AB" {
			t.Errorf("entry %d text: want %q, got %q", e.StringID, "AB", text)
		}
	}
}

func TestWriteFile_PreservesOriginal(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "test.tlk")
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	err = WriteFile(outPath, tlkFile, []StringEntry{
		{StringID: 50, Text: "AB", Male: true},
	})
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	readBack, err := ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	text, ok := ResolveString(readBack, 1, true)
	if !ok {
		t.Fatal("original entry lost after round-trip")
	}
	if text != "AB" {
		t.Errorf("original text: want %q, got %q", "AB", text)
	}
}

func TestWriteFile_HeaderFields(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "test.tlk")
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	err = WriteFile(outPath, tlkFile, []StringEntry{
		{StringID: 10, Text: "AB", Male: true},
		{StringID: 20, Text: "AB", Male: true},
		{StringID: 30, Text: "AB", Male: false},
	})
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	readBack, err := ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if readBack.Header.Magic != TLKMagic {
		t.Errorf("Magic: want 0x%08X, got 0x%08X", TLKMagic, readBack.Header.Magic)
	}
	if readBack.Header.Version != 3 {
		t.Errorf("Version: want 3, got %d", readBack.Header.Version)
	}
	if readBack.Header.MinVersion != 2 {
		t.Errorf("MinVersion: want 2, got %d", readBack.Header.MinVersion)
	}
	expectedMale := tlkFile.Header.MaleEntryCount + 2
	expectedFemale := tlkFile.Header.FemaleEntryCount + 1
	if readBack.Header.MaleEntryCount != expectedMale {
		t.Errorf("MaleEntryCount: want %d, got %d", expectedMale, readBack.Header.MaleEntryCount)
	}
	if readBack.Header.FemaleEntryCount != expectedFemale {
		t.Errorf("FemaleEntryCount: want %d, got %d", expectedFemale, readBack.Header.FemaleEntryCount)
	}
	if readBack.Header.TreeNodeCount != tlkFile.Header.TreeNodeCount {
		t.Errorf("TreeNodeCount: want %d, got %d", tlkFile.Header.TreeNodeCount, readBack.Header.TreeNodeCount)
	}
}

func TestBuildBytes_EmptyNewEntries_Extra(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	buf, err := BuildBytes(tlkFile, nil)
	if err != nil {
		t.Fatalf("BuildBytes: %v", err)
	}
	readBack, err := Parse(buf, "test.tlk")
	if err != nil {
		t.Fatalf("Parse round-trip: %v", err)
	}
	if readBack.TotalEntries != tlkFile.TotalEntries {
		t.Errorf("entries: want %d, got %d", tlkFile.TotalEntries, readBack.TotalEntries)
	}
}

func TestBuildBytes_WithNewEntries(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	entries := []StringEntry{
		{StringID: 50, Text: "BA", Male: true},
		{StringID: 60, Text: "AB", Male: false},
	}
	buf, err := BuildBytes(tlkFile, entries)
	if err != nil {
		t.Fatalf("BuildBytes: %v", err)
	}
	readBack, err := Parse(buf, "test.tlk")
	if err != nil {
		t.Fatalf("Parse round-trip: %v", err)
	}
	text, ok := ResolveString(readBack, 50, true)
	if !ok || text != "BA" {
		t.Errorf("male entry: ok=%v text=%q", ok, text)
	}
	text, ok = ResolveString(readBack, 60, false)
	if !ok || text != "AB" {
		t.Errorf("female entry: ok=%v text=%q", ok, text)
	}
	text, ok = ResolveString(readBack, 1, true)
	if !ok || text != "AB" {
		t.Errorf("original entry: ok=%v text=%q", ok, text)
	}
}

func TestBuildBytes_HeaderIntegrity(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	buf, err := BuildBytes(tlkFile, []StringEntry{
		{StringID: 100, Text: "AB", Male: true},
	})
	if err != nil {
		t.Fatalf("BuildBytes: %v", err)
	}
	readBack, err := Parse(buf, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if readBack.Header.Magic != TLKMagic {
		t.Errorf("Magic: want 0x%08X, got 0x%08X", TLKMagic, readBack.Header.Magic)
	}
	if readBack.Header.Version != 3 {
		t.Errorf("Version: want 3, got %d", readBack.Header.Version)
	}
	if readBack.Header.MinVersion != 2 {
		t.Errorf("MinVersion: want 2, got %d", readBack.Header.MinVersion)
	}
	if readBack.Header.TreeNodeCount != tlkFile.Header.TreeNodeCount {
		t.Errorf("TreeNodeCount changed: was %d, now %d", tlkFile.Header.TreeNodeCount, readBack.Header.TreeNodeCount)
	}
}

func TestAddEntries_NilFile_Extra(t *testing.T) {
	err := AddEntries(nil, []StringEntry{{StringID: 1, Text: "X", Male: true}})
	if err == nil {
		t.Fatal("expected error for nil file")
	}
}

func TestBuildBytes_NilFile(t *testing.T) {
	_, err := BuildBytes(nil, nil)
	if err == nil {
		t.Fatal("expected error for nil file")
	}
}

func TestMultipleRoundTrips(t *testing.T) {
	dir := t.TempDir()
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	path1 := filepath.Join(dir, "step1.tlk")
	err = WriteFile(path1, tlkFile, []StringEntry{
		{StringID: 50, Text: "AB", Male: true},
	})
	if err != nil {
		t.Fatalf("WriteFile step1: %v", err)
	}
	step1, err := ReadFile(path1)
	if err != nil {
		t.Fatalf("ReadFile step1: %v", err)
	}
	err = AddEntries(step1, []StringEntry{
		{StringID: 60, Text: "BA", Male: false},
	})
	if err != nil {
		t.Fatalf("AddEntries step2: %v", err)
	}
	path2 := filepath.Join(dir, "step2.tlk")
	err = WriteFile(path2, step1, nil)
	if err != nil {
		t.Fatalf("WriteFile step2: %v", err)
	}
	final, err := ReadFile(path2)
	if err != nil {
		t.Fatalf("ReadFile step2: %v", err)
	}
	text, ok := ResolveString(final, 1, true)
	if !ok || text != "AB" {
		t.Errorf("original entry: ok=%v text=%q", ok, text)
	}
	text, ok = ResolveString(final, 50, true)
	if !ok || text != "AB" {
		t.Errorf("step1 male entry: ok=%v text=%q", ok, text)
	}
	text, ok = ResolveString(final, 60, false)
	if !ok || text != "BA" {
		t.Errorf("step2 female entry: ok=%v text=%q", ok, text)
	}
}

func TestAddEntries_Unicode(t *testing.T) {
	nodes := []Node{
		{LeftNodeID: -242, RightNodeID: 1},
		{LeftNodeID: -234, RightNodeID: -1},
	}
	var buf []byte
	writeI32 := func(v int32) {
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32(v))
		buf = append(buf, b...)
	}
	writeI32(int32(TLKMagic))
	writeI32(3)
	writeI32(2)
	writeI32(1)
	writeI32(0)
	writeI32(int32(len(nodes)))
	writeI32(2)
	writeI32(1)
	writeI32(0)
	for _, n := range nodes {
		writeI32(n.LeftNodeID)
		writeI32(n.RightNodeID)
	}
	buf = append(buf, 0b00011010, 0)
	tlkFile, err := Parse(buf, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	err = AddEntries(tlkFile, []StringEntry{
		{StringID: 10, Text: "ñé", Male: true},
	})
	if err != nil {
		t.Fatalf("AddEntries: %v", err)
	}
	text, ok := ResolveString(tlkFile, 10, true)
	if !ok {
		t.Fatal("unicode entry not found")
	}
	if text != "ñé" {
		t.Errorf("text: want %q, got %q", "ñé", text)
	}
}

func TestBuildBytes_LargeEntryTable(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	entries := make([]StringEntry, 150)
	for i := 0; i < 150; i++ {
		entries[i] = StringEntry{
			StringID: int32(1000 + i),
			Text:     "AB",
			Male:     true,
		}
	}
	buf, err := BuildBytes(tlkFile, entries)
	if err != nil {
		t.Fatalf("BuildBytes: %v", err)
	}
	readBack, err := Parse(buf, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if readBack.Header.MaleEntryCount != tlkFile.Header.MaleEntryCount+150 {
		t.Errorf("MaleEntryCount: want %d, got %d",
			tlkFile.Header.MaleEntryCount+150, readBack.Header.MaleEntryCount)
	}
	if readBack.TotalEntries != tlkFile.TotalEntries+150 {
		t.Errorf("TotalEntries: want %d, got %d",
			tlkFile.TotalEntries+150, readBack.TotalEntries)
	}
	for i := 0; i < 150; i++ {
		text, ok := ResolveString(readBack, int32(1000+i), true)
		if !ok {
			t.Errorf("entry %d not found", 1000+i)
			break
		}
		if text != "AB" {
			t.Errorf("entry %d text: want %q, got %q", 1000+i, "AB", text)
		}
	}
}

func TestEncodeString_ExtraRoundTrip(t *testing.T) {
	data := buildLargeTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Skipf("buildLargeTLK parse not yet valid: %v", err)
	}
	table := BuildCodeTable(tlkFile.Nodes)
	encoded, err := EncodeString("A", table)
	if err != nil {
		t.Fatalf("EncodeString: %v", err)
	}
	decoded, ok := DecodeString(encoded, tlkFile.Nodes, 0)
	if !ok {
		t.Fatal("DecodeString failed")
	}
	if decoded != "A" {
		t.Errorf("round-trip: want %q, got %q", "A", decoded)
	}
}

func TestWriteFile_ExistingIDNotDuplicated(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "test.tlk")
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	err = WriteFile(outPath, tlkFile, []StringEntry{
		{StringID: 1, Text: "SHOULD_NOT_OVERWRITE", Male: true},
	})
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	readBack, err := ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	text, ok := ResolveString(readBack, 1, true)
	if !ok || text != "AB" {
		t.Errorf("original entry overwritten: ok=%v text=%q", ok, text)
	}
	if readBack.TotalEntries != tlkFile.TotalEntries {
		t.Errorf("TotalEntries changed: was %d, now %d", tlkFile.TotalEntries, readBack.TotalEntries)
	}
}

func TestWriteFile_UnicodeRoundTrip(t *testing.T) {
	nodes := []Node{
		{LeftNodeID: -242, RightNodeID: 1},
		{LeftNodeID: -234, RightNodeID: -1},
	}
	var buf []byte
	writeI32 := func(v int32) {
		b := make([]byte, 4)
		binary.LittleEndian.PutUint32(b, uint32(v))
		buf = append(buf, b...)
	}
	writeI32(int32(TLKMagic))
	writeI32(3)
	writeI32(2)
	writeI32(1)
	writeI32(0)
	writeI32(int32(len(nodes)))
	writeI32(2)
	writeI32(1)
	writeI32(0)
	for _, n := range nodes {
		writeI32(n.LeftNodeID)
		writeI32(n.RightNodeID)
	}
	buf = append(buf, 0b00011010, 0)
	tlkFile, err := Parse(buf, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	dir := t.TempDir()
	outPath := filepath.Join(dir, "test.tlk")
	err = WriteFile(outPath, tlkFile, []StringEntry{
		{StringID: 50, Text: "ñé", Male: true},
	})
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	readBack, err := ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	text, ok := ResolveString(readBack, 50, true)
	if !ok || text != "ñé" {
		t.Errorf("unicode round-trip: ok=%v text=%q", ok, text)
	}
}

func TestEncodeString_RoundTripViaMinimal(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	table := BuildCodeTable(tlkFile.Nodes)
	original, ok := ResolveString(tlkFile, 1, true)
	if !ok {
		t.Fatal("ResolveString original entry failed")
	}
	encoded, err := EncodeString(original, table)
	if err != nil {
		t.Fatalf("EncodeString: %v", err)
	}
	decoded, ok := DecodeString(encoded, tlkFile.Nodes, 0)
	if !ok {
		t.Fatal("DecodeString failed")
	}
	if decoded != original {
		t.Errorf("round-trip: want %q, got %q", original, decoded)
	}
}

func TestBuildBytes_EmptyNewEntriesCreatesValidFile(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	buf, err := BuildBytes(tlkFile, nil)
	if err != nil {
		t.Fatalf("BuildBytes: %v", err)
	}
	dir := t.TempDir()
	outPath := filepath.Join(dir, "test.tlk")
	if err := os.WriteFile(outPath, buf, 0644); err != nil {
		t.Fatalf("os.WriteFile: %v", err)
	}
	readBack, err := ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if readBack.Header.Magic != TLKMagic {
		t.Errorf("Magic mismatch")
	}
	text, ok := ResolveString(readBack, 1, true)
	if !ok || text != "AB" {
		t.Errorf("original entry: ok=%v text=%q", ok, text)
	}
}

func TestWriteFile_MixedEntries(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "test.tlk")
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	err = WriteFile(outPath, tlkFile, []StringEntry{
		{StringID: 201, Text: "AB", Male: true},
		{StringID: 202, Text: "BA", Male: false},
		{StringID: 203, Text: "AA", Male: true},
		{StringID: 204, Text: "BB", Male: false},
		{StringID: 205, Text: "AB", Male: true},
		{StringID: 206, Text: "BA", Male: false},
		{StringID: 207, Text: "AA", Male: true},
		{StringID: 208, Text: "BB", Male: false},
		{StringID: 209, Text: "AB", Male: true},
		{StringID: 210, Text: "BA", Male: false},
	})
	if err != nil {
		t.Fatalf("WriteFile: %v", err)
	}
	readBack, err := ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	maleCount := 0
	femaleCount := 0
	for i := 201; i <= 210; i++ {
		text, ok := ResolveString(readBack, int32(i), true)
		if ok {
			maleCount++
			t.Logf("ID %d (male): %q", i, text)
			continue
		}
		text, ok = ResolveString(readBack, int32(i), false)
		if ok {
			femaleCount++
			t.Logf("ID %d (female): %q", i, text)
		}
	}
	if maleCount != 5 {
		t.Errorf("expected 5 male entries, got %d", maleCount)
	}
	if femaleCount != 5 {
		t.Errorf("expected 5 female entries, got %d", femaleCount)
	}
}
