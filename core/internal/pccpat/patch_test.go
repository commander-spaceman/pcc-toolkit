package pccpat

import (
	"encoding/binary"
	"testing"

	"pcc-toolkit/core/internal/pcc"
)

func buildMinimalPCC(exportSerialData [][]byte) ([]byte, *pcc.FileSummary) {
	names := []string{
		"None", "Class", "Core", "Object", "BioConversation",
		"entry_0", "entry_1", "entry_2",
	}
	imports := []pcc.Import{
		{ClassNameIndex: 1, ObjectNameIndex: 2},
	}

	var nameBytes []byte
	for _, n := range names {
		nb := make([]byte, 4+len(n)+1)
		binary.LittleEndian.PutUint32(nb[0:4], uint32(len(n)+1))
		copy(nb[4:], n)
		nb[4+len(n)] = 0
		flags := make([]byte, 4)
		binary.LittleEndian.PutUint32(flags, 0xFFFFFFF2) // ME2 name flags
		nb = append(nb, flags...)
		nameBytes = append(nameBytes, nb...)
	}

	importOffset := 60 + len(nameBytes)
	exportOffset := importOffset + (len(imports) * 28)

	type exportBinEntry struct {
		data   []byte
		offset int
		size   int
	}
	exportEntries := make([]exportBinEntry, len(exportSerialData))

	exportTableSize := 0
	for range exportSerialData {
		entrySize := 40 + 4 + 8 + 16 + 4 // base + component + gen + extra + footer (componentCount=0, generationCount=0)
		exportTableSize += entrySize
	}

	serialStart := exportOffset + exportTableSize
	cursor := serialStart
	for i, data := range exportSerialData {
		exportEntries[i] = exportBinEntry{
			data:   data,
			offset: cursor,
			size:   len(data),
		}
		cursor += len(data)
	}

	headerSize := 60
	buf := make([]byte, cursor)
	binary.LittleEndian.PutUint32(buf[0:], pcc.PackageMagic)
	binary.LittleEndian.PutUint32(buf[4:], uint32(512|(130<<16))) // ME2
	binary.LittleEndian.PutUint32(buf[8:], 0)
	binary.LittleEndian.PutUint32(buf[12:], 0) // folderLen
	binary.LittleEndian.PutUint32(buf[16:], 0) // flags
	binary.LittleEndian.PutUint32(buf[20:], uint32(len(names)))
	binary.LittleEndian.PutUint32(buf[24:], uint32(headerSize))
	binary.LittleEndian.PutUint32(buf[28:], uint32(len(exportSerialData)))
	binary.LittleEndian.PutUint32(buf[32:], uint32(exportOffset))
	binary.LittleEndian.PutUint32(buf[36:], uint32(len(imports)))
	binary.LittleEndian.PutUint32(buf[40:], uint32(importOffset))
	binary.LittleEndian.PutUint32(buf[44:], uint32(len(exportSerialData)))
	binary.LittleEndian.PutUint32(buf[48:], 0)
	binary.LittleEndian.PutUint32(buf[52:], 0)
	binary.LittleEndian.PutUint32(buf[56:], 0)

	copy(buf[headerSize:], nameBytes)

	impCursor := importOffset
	for _, imp := range imports {
		writeI32(buf, impCursor, 0)
		writeI32(buf, impCursor+8, imp.ClassNameIndex)
		writeI32(buf, impCursor+20, imp.ObjectNameIndex)
		impCursor += 28
	}

	expCursor := exportOffset
	for i, ed := range exportEntries {
		writeI32(buf, expCursor, 3)      // classIndex (Object)
		writeI32(buf, expCursor+12, 5+i) // objectNameIndex
		writeI32(buf, expCursor+32, ed.size)
		writeI32(buf, expCursor+36, ed.offset)
		entrySize := 40 + 4 + 8 + 16 + 4 // matches buildExportEntryMap for componentCount=0, generationCount=0
		expCursor += entrySize
	}

	for _, ed := range exportEntries {
		copy(buf[ed.offset:], ed.data)
	}

	exports := make([]pcc.Export, len(exportEntries))
	for i, ed := range exportEntries {
		exports[i] = pcc.Export{
			Index:           i,
			ClassIndex:      3,
			ObjectNameIndex: 5 + i,
			SerialSize:      ed.size,
			SerialOffset:    ed.offset,
			ObjectName:      names[5+i],
			ClassName:       "Object",
		}
	}

	return buf, &pcc.FileSummary{
		Path:        "test.pcc",
		GameProfile: pcc.ProfileME2OT,
		Compressed:  false,
		Header: pcc.Header{
			UnrealVersion:   512,
			LicenseeVersion: 130,
			Flags:           0,
			NameCount:       len(names),
			NameOffset:      headerSize,
			ExportCount:     len(exports),
			ExportOffset:    exportOffset,
			ImportCount:     len(imports),
			ImportOffset:    importOffset,
		},
		Names:   names,
		Imports: imports,
		Exports: exports,
	}
}

func TestPatchExport_SameSize(t *testing.T) {
	original := []byte("hello world")
	data, summary := buildMinimalPCC([][]byte{original})

	newData := []byte("HELLO WORLD")
	result, newSummary, err := PatchExport(data, summary, 0, newData)
	if err != nil {
		t.Fatalf("PatchExport: %v", err)
	}

	exp := newSummary.Exports[0]
	if exp.SerialSize != len(newData) {
		t.Errorf("SerialSize: want %d, got %d", len(newData), exp.SerialSize)
	}
	if exp.SerialOffset < len(data)-100 {
		t.Error("SerialOffset should be valid")
	}

	extracted := result[exp.SerialOffset : exp.SerialOffset+exp.SerialSize]
	if string(extracted) != string(newData) {
		t.Errorf("data: want %q, got %q", string(newData), string(extracted))
	}
}

func TestPatchExport_Grow(t *testing.T) {
	original := []byte("short")
	data, summary := buildMinimalPCC([][]byte{original})

	newData := []byte("this is much longer data")
	result, newSummary, err := PatchExport(data, summary, 0, newData)
	if err != nil {
		t.Fatalf("PatchExport: %v", err)
	}

	exp := newSummary.Exports[0]
	if exp.SerialSize != len(newData) {
		t.Errorf("SerialSize: want %d, got %d", len(newData), exp.SerialSize)
	}

	extracted := result[exp.SerialOffset : exp.SerialOffset+exp.SerialSize]
	if string(extracted) != string(newData) {
		t.Errorf("data: want %q, got %q", string(newData), string(extracted))
	}
}

func TestPatchExport_Shrink(t *testing.T) {
	original := []byte("this is long text")
	data, summary := buildMinimalPCC([][]byte{original})

	newData := []byte("short")
	result, newSummary, err := PatchExport(data, summary, 0, newData)
	if err != nil {
		t.Fatalf("PatchExport: %v", err)
	}

	exp := newSummary.Exports[0]
	if exp.SerialSize != len(newData) {
		t.Errorf("SerialSize: want %d, got %d", len(newData), exp.SerialSize)
	}

	extracted := result[exp.SerialOffset : exp.SerialOffset+exp.SerialSize]
	if string(extracted) != string(newData) {
		t.Errorf("data: want %q, got %q", string(newData), string(extracted))
	}
}

func TestPatchExport_ShiftSubsequent(t *testing.T) {
	original1 := []byte("AAAA")
	original2 := []byte("BBBB")
	data, summary := buildMinimalPCC([][]byte{original1, original2})

	newData1 := []byte("AAAAAAAAAAAA") // 12 bytes, was 4 → delta +8
	result, newSummary, err := PatchExport(data, summary, 0, newData1)
	if err != nil {
		t.Fatalf("PatchExport: %v", err)
	}

	exp0 := newSummary.Exports[0]
	if exp0.SerialSize != 12 {
		t.Errorf("export 0 SerialSize: want 12, got %d", exp0.SerialSize)
	}

	exp1 := newSummary.Exports[1]
	if exp1.SerialSize != 4 {
		t.Errorf("export 1 SerialSize: want 4, got %d", exp1.SerialSize)
	}
	if exp1.SerialOffset != summary.Exports[1].SerialOffset+8 {
		t.Errorf("export 1 SerialOffset: want %d, got %d",
			summary.Exports[1].SerialOffset+8, exp1.SerialOffset)
	}

	extracted0 := result[exp0.SerialOffset : exp0.SerialOffset+exp0.SerialSize]
	extracted1 := result[exp1.SerialOffset : exp1.SerialOffset+exp1.SerialSize]

	if string(extracted0) != string(newData1) {
		t.Errorf("export 0 data: want %q, got %q", string(newData1), string(extracted0))
	}
	if string(extracted1) != string(original2) {
		t.Errorf("export 1 data: want %q, got %q", string(original2), string(extracted1))
	}
}

func TestPatchExport_InvalidIndex(t *testing.T) {
	data, summary := buildMinimalPCC([][]byte{[]byte("test")})

	_, _, err := PatchExport(data, summary, -1, []byte("x"))
	if err == nil {
		t.Fatal("expected error for negative index")
	}

	_, _, err = PatchExport(data, summary, 1, []byte("x"))
	if err == nil {
		t.Fatal("expected error for out-of-range index")
	}
}

func TestPatchExport_NilInput(t *testing.T) {
	_, _, err := PatchExport(nil, nil, 0, []byte("x"))
	if err == nil {
		t.Fatal("expected error for nil input")
	}
}

func TestPatchExport_EmptyNewData(t *testing.T) {
	original := []byte("some data")
	data, summary := buildMinimalPCC([][]byte{original})

	result, newSummary, err := PatchExport(data, summary, 0, []byte{})
	if err != nil {
		t.Fatalf("PatchExport: %v", err)
	}

	if newSummary.Exports[0].SerialSize != 0 {
		t.Errorf("SerialSize: want 0, got %d", newSummary.Exports[0].SerialSize)
	}

	_ = result
}

func TestBuildExportEntryMap(t *testing.T) {
	_, summary := buildMinimalPCC([][]byte{[]byte("a"), []byte("b"), []byte("c")})

	buf, _ := buildMinimalPCC([][]byte{[]byte("a"), []byte("b"), []byte("c")})
	entries, err := buildExportEntryMap(buf, summary.Header)
	if err != nil {
		t.Fatalf("buildExportEntryMap: %v", err)
	}
	if len(entries) != 3 {
		t.Fatalf("expected 3 entries, got %d", len(entries))
	}
	for i, e := range entries {
		if e.entryOffset < 0 || e.entrySize <= 0 {
			t.Errorf("entry %d: invalid offset=%d size=%d", i, e.entryOffset, e.entrySize)
		}
	}
}
