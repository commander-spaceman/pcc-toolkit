package pcc

import (
	"encoding/binary"
	"testing"
)

func writeI32LE(buf []byte, offset int, v int32) {
	binary.LittleEndian.PutUint32(buf[offset:], uint32(v))
}

func writeU32LE(buf []byte, offset int, v uint32) {
	binary.LittleEndian.PutUint32(buf[offset:], v)
}

func unrealStringBytes(s string) []byte {
	encoded := []byte(s)
	encoded = append(encoded, 0)
	hdr := make([]byte, 4)
	binary.LittleEndian.PutUint32(hdr, uint32(len(encoded)))
	return append(hdr, encoded...)
}

func buildMinimalPCC(uv, lv int) []byte {
	names := []string{"Core", "Class", "BioConversation", "Conv_Test"}

	headerSize := 64
	nameTable := []byte{}
	for _, name := range names {
		nameTable = append(nameTable, unrealStringBytes(name)...)
		pad := make([]byte, 4)
		nameTable = append(nameTable, pad...)
	}

	importEntry := make([]byte, 28)
	writeI32LE(importEntry, 0, 0)
	writeI32LE(importEntry, 8, 1)
	writeI32LE(importEntry, 20, 2)

	exportSize := 72
	exportEntry := make([]byte, exportSize)
	writeI32LE(exportEntry, 0, -1)
	writeI32LE(exportEntry, 4, 0)
	writeI32LE(exportEntry, 8, 0)
	writeI32LE(exportEntry, 12, 3)
	writeI32LE(exportEntry, 32, 0)
	writeI32LE(exportEntry, 36, 0)
	writeI32LE(exportEntry, 40, 0)
	writeI32LE(exportEntry, 48, 0)

	nameOffset := headerSize
	importOffset := nameOffset + len(nameTable)
	exportOffset := importOffset + len(importEntry)

	buf := make([]byte, headerSize)
	writeU32LE(buf, 0, 0x9E2A83C1)
	packedVersion := (uint32(lv&0xFFFF) << 16) | uint32(uv&0xFFFF)
	writeU32LE(buf, 4, packedVersion)
	writeI32LE(buf, 20, int32(len(names)))
	writeI32LE(buf, 24, int32(nameOffset))
	writeI32LE(buf, 28, 1)
	writeI32LE(buf, 32, int32(exportOffset))
	writeI32LE(buf, 36, 1)
	writeI32LE(buf, 40, int32(importOffset))

	buf = append(buf, nameTable...)
	buf = append(buf, importEntry...)
	buf = append(buf, exportEntry...)
	return buf
}

func TestParsePccHeader_ME2OT(t *testing.T) {
	data := buildMinimalPCC(512, 130)
	header, err := parsePccHeader(data)
	if err != nil {
		t.Fatalf("parsePccHeader: %v", err)
	}
	if header.UnrealVersion != 512 {
		t.Errorf("UV = %d, want 512", header.UnrealVersion)
	}
	if header.LicenseeVersion != 130 {
		t.Errorf("LV = %d, want 130", header.LicenseeVersion)
	}
	if header.NameCount != 4 {
		t.Errorf("NameCount = %d, want 4", header.NameCount)
	}
	if header.ExportCount != 1 {
		t.Errorf("ExportCount = %d, want 1", header.ExportCount)
	}
	if header.ImportCount != 1 {
		t.Errorf("ImportCount = %d, want 1", header.ImportCount)
	}
}

func TestParsePccHeader_InvalidMagic(t *testing.T) {
	data := buildMinimalPCC(512, 130)
	data[0] = 0x00
	_, err := parsePccHeader(data)
	if err == nil {
		t.Fatal("expected error for invalid magic")
	}
}

func TestParsePccNames(t *testing.T) {
	data := buildMinimalPCC(512, 130)
	header, err := parsePccHeader(data)
	if err != nil {
		t.Fatalf("parsePccHeader: %v", err)
	}
	names, err := parsePccNames(data, header)
	if err != nil {
		t.Fatalf("parsePccNames: %v", err)
	}
	if len(names) != 4 {
		t.Fatalf("got %d names, want 4", len(names))
	}
	expected := []string{"Core", "Class", "BioConversation", "Conv_Test"}
	for i, want := range expected {
		if names[i] != want {
			t.Errorf("names[%d] = %q, want %q", i, names[i], want)
		}
	}
}

func TestParsePccImports(t *testing.T) {
	data := buildMinimalPCC(512, 130)
	header, err := parsePccHeader(data)
	if err != nil {
		t.Fatalf("parsePccHeader: %v", err)
	}
	imports, err := parsePccImports(data, header)
	if err != nil {
		t.Fatalf("parsePccImports: %v", err)
	}
	if len(imports) != 1 {
		t.Fatalf("got %d imports, want 1", len(imports))
	}
	if imports[0].ClassNameIndex != 1 {
		t.Errorf("ClassNameIndex = %d, want 1", imports[0].ClassNameIndex)
	}
	if imports[0].ObjectNameIndex != 2 {
		t.Errorf("ObjectNameIndex = %d, want 2", imports[0].ObjectNameIndex)
	}
}

func TestParsePccExports(t *testing.T) {
	data := buildMinimalPCC(512, 130)
	header, err := parsePccHeader(data)
	if err != nil {
		t.Fatalf("parsePccHeader: %v", err)
	}
	exports, err := parsePccExports(data, header)
	if err != nil {
		t.Fatalf("parsePccExports: %v", err)
	}
	if len(exports) != 1 {
		t.Fatalf("got %d exports, want 1", len(exports))
	}
	if exports[0].ClassIndex != -1 {
		t.Errorf("ClassIndex = %d, want -1", exports[0].ClassIndex)
	}
	if exports[0].ObjectNameIndex != 3 {
		t.Errorf("ObjectNameIndex = %d, want 3", exports[0].ObjectNameIndex)
	}
	if exports[0].SerialOffset != 0 {
		t.Errorf("SerialOffset = %d, want 0", exports[0].SerialOffset)
	}
}

func TestResolveExportNames(t *testing.T) {
	data := buildMinimalPCC(512, 130)
	header, err := parsePccHeader(data)
	if err != nil {
		t.Fatalf("parsePccHeader: %v", err)
	}
	names, err := parsePccNames(data, header)
	if err != nil {
		t.Fatalf("parsePccNames: %v", err)
	}
	imports, err := parsePccImports(data, header)
	if err != nil {
		t.Fatalf("parsePccImports: %v", err)
	}
	exports, err := parsePccExports(data, header)
	if err != nil {
		t.Fatalf("parsePccExports: %v", err)
	}
	resolveExportNames(exports, imports, names)
	if exports[0].ObjectName != "Conv_Test" {
		t.Errorf("ObjectName = %q, want Conv_Test", exports[0].ObjectName)
	}
	if exports[0].ClassName != "BioConversation" {
		t.Errorf("ClassName = %q, want BioConversation", exports[0].ClassName)
	}
}

func TestResolveExportNames_LE2(t *testing.T) {
	data := buildMinimalPCC(684, 168)
	header, err := parsePccHeader(data)
	if err != nil {
		t.Fatalf("parsePccHeader: %v", err)
	}
	names, err := parsePccNames(data, header)
	if err != nil {
		t.Fatalf("parsePccNames: %v", err)
	}
	imports, err := parsePccImports(data, header)
	if err != nil {
		t.Fatalf("parsePccImports: %v", err)
	}
	exports, err := parsePccExports(data, header)
	if err != nil {
		t.Fatalf("parsePccExports: %v", err)
	}
	resolveExportNames(exports, imports, names)
	if exports[0].ClassName != "BioConversation" {
		t.Errorf("ClassName = %q, want BioConversation", exports[0].ClassName)
	}
}

func TestInferGameProfile(t *testing.T) {
	tests := []struct {
		uv, lv  int
		profile GameProfile
	}{
		{512, 130, ProfileME2OT},
		{684, 168, ProfileLE2},
		{684, 194, ProfileME3OT},
		{685, 205, ProfileLE3},
		{500, 100, ProfileUnknown},
	}
	for _, tc := range tests {
		got := InferGameProfile(tc.uv, tc.lv)
		if got != tc.profile {
			t.Errorf("InferGameProfile(%d, %d) = %q, want %q", tc.uv, tc.lv, got, tc.profile)
		}
	}
}

func TestReadUnrealString(t *testing.T) {
	data := []byte{0x05, 0x00, 0x00, 0x00, 'H', 'e', 'l', 'l', 'o'}
	s, _, err := readUnrealString(data, 0)
	if err != nil {
		t.Fatalf("readUnrealString: %v", err)
	}
	if s != "Hello" {
		t.Errorf("got %q, want Hello", s)
	}
}

func TestReadUnrealString_Empty(t *testing.T) {
	data := []byte{0x00, 0x00, 0x00, 0x00}
	s, _, err := readUnrealString(data, 0)
	if err != nil {
		t.Fatalf("readUnrealString: %v", err)
	}
	if s != "" {
		t.Errorf("got %q, want empty", s)
	}
}

func TestReadUnrealString_OOB(t *testing.T) {
	data := []byte{0xFF, 0xFF}
	_, _, err := readUnrealString(data, 0)
	if err == nil {
		t.Fatal("expected error for out-of-bounds")
	}
}

func TestReadUnrealString_NullTerminated(t *testing.T) {
	data := []byte{0x06, 0x00, 0x00, 0x00, 'H', 'e', 'l', 'l', 'o', 0x00}
	s, _, err := readUnrealString(data, 0)
	if err != nil {
		t.Fatalf("readUnrealString: %v", err)
	}
	if s != "Hello" {
		t.Errorf("got %q, want Hello", s)
	}
}

func TestBestContainingExport(t *testing.T) {
	exports := []Export{
		{Index: 0, SerialOffset: 100, SerialSize: 200, ClassName: "Class"},
		{Index: 1, SerialOffset: 100, SerialSize: 50, ClassName: ""},
		{Index: 2, SerialOffset: 100, SerialSize: 100, ClassName: "Class"},
	}
	best := bestContainingExport(exports, 120, 500)
	if best == nil {
		t.Fatal("expected a containing export")
	}
	if best.Index != 2 {
		t.Errorf("Index = %d, want 2 (named, smallest among named)", best.Index)
	}
}

func TestBestContainingExport_OutOfRange(t *testing.T) {
	exports := []Export{
		{Index: 0, SerialOffset: 100, SerialSize: 50, ClassName: "Class"},
	}
	best := bestContainingExport(exports, 200, 500)
	if best != nil {
		t.Error("expected nil for offset outside export range")
	}
}

func TestBestContainingExport_PrefersNamed(t *testing.T) {
	exports := []Export{
		{Index: 0, SerialOffset: 100, SerialSize: 100, ClassName: ""},
		{Index: 1, SerialOffset: 100, SerialSize: 100, ClassName: "Class"},
	}
	best := bestContainingExport(exports, 150, 500)
	if best == nil {
		t.Fatal("expected a containing export")
	}
	if best.Index != 1 {
		t.Errorf("Index = %d, want 1 (named class)", best.Index)
	}
}

func TestMapOffsetsToContainers(t *testing.T) {
	exports := []Export{
		{Index: 0, ObjectName: "Export0", ClassName: "Class", SerialOffset: 100, SerialSize: 100},
		{Index: 1, ObjectName: "Export1", ClassName: "Other", SerialOffset: 200, SerialSize: 100},
	}
	offsets := map[int][]int{
		42: {120, 220},
	}
	hits := MapOffsetsToContainers(exports, offsets, 500)
	if len(hits) != 2 {
		t.Fatalf("got %d hits, want 2", len(hits))
	}
	if hits[0].ExportIndex != 0 || hits[1].ExportIndex != 1 {
		t.Errorf("wrong export mapping: %v", hits)
	}
}

func TestReadFileFromBytes_ME2OT(t *testing.T) {
	data := buildMinimalPCC(512, 130)
	summary, err := ReadFileFromBytes(data, "test.pcc")
	if err != nil {
		t.Fatalf("ReadFileFromBytes: %v", err)
	}
	if summary.GameProfile != ProfileME2OT {
		t.Errorf("GameProfile = %q, want %q", summary.GameProfile, ProfileME2OT)
	}
	if summary.Compressed {
		t.Error("Compressed should be false")
	}
	if summary.Header.UnrealVersion != 512 {
		t.Errorf("UV = %d, want 512", summary.Header.UnrealVersion)
	}
	if summary.Names == nil {
		t.Error("Names should not be nil for ReadFileFromBytes")
	}
}

func TestRequireME2(t *testing.T) {
	fs := &FileSummary{GameProfile: ProfileME2OT}
	if err := fs.RequireME2(); err != nil {
		t.Errorf("RequireME2 should pass for me2_ot: %v", err)
	}

	fs.GameProfile = ProfileLE2
	if err := fs.RequireME2(); err == nil {
		t.Error("RequireME2 should fail for le2")
	}

	fs.GameProfile = ProfileUnknown
	if err := fs.RequireME2(); err == nil {
		t.Error("RequireME2 should fail for unknown")
	}
}

func TestReadFileExports(t *testing.T) {
	data := buildMinimalPCC(512, 130)
	summary, exportData, err := ReadFileFromBytesAndExports(data, "test.pcc", func(e Export) bool {
		return e.ClassName == "BioConversation"
	})
	if err != nil {
		t.Fatalf("ReadFileFromBytesAndExports: %v", err)
	}
	if summary.GameProfile != ProfileME2OT {
		t.Errorf("GameProfile = %q, want %q", summary.GameProfile, ProfileME2OT)
	}
	if len(exportData) != 1 {
		t.Fatalf("expected 1 export, got %d", len(exportData))
	}
	data0, ok := exportData[0]
	if !ok {
		t.Fatal("expected export data for index 0")
	}
	if len(data0) != 0 {
		t.Errorf("expected empty serial data, got %d bytes", len(data0))
	}
}

func TestReadFileExports_NoMatch(t *testing.T) {
	data := buildMinimalPCC(512, 130)
	_, exportData, err := ReadFileFromBytesAndExports(data, "test.pcc", func(e Export) bool {
		return false
	})
	if err != nil {
		t.Fatalf("ReadFileFromBytesAndExports: %v", err)
	}
	if len(exportData) != 0 {
		t.Errorf("expected 0 exports, got %d", len(exportData))
	}
}
