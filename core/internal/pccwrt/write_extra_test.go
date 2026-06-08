package pccwrt

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"testing"

	"pcc-toolkit/core/internal/pcc"
	"pcc-toolkit/core/internal/pccpat"
)

func buildTestPCC(exportData [][]byte, names []string, imports []pcc.Import) ([]byte, *pcc.FileSummary) {
	exports := make([]pcc.Export, len(exportData))
	var total int
	for _, d := range exportData {
		total += len(d)
	}
	raw := make([]byte, total)
	pos := 0
	for i, d := range exportData {
		exports[i] = pcc.Export{
			Index:           i,
			ClassIndex:      3,
			ObjectNameIndex: 5 + i,
			SerialSize:      len(d),
			SerialOffset:    pos,
		}
		if 5+i < len(names) {
			exports[i].ObjectName = names[5+i]
		}
		exports[i].ClassName = "Object"
		copy(raw[pos:], d)
		pos += len(d)
	}

	return raw, &pcc.FileSummary{
		Path:        "test.pcc",
		GameProfile: pcc.ProfileME2OT,
		Compressed:  false,
		Header: pcc.Header{
			UnrealVersion:   512,
			LicenseeVersion: 130,
			Flags:           0,
			NameCount:       len(names),
			ExportCount:     len(exportData),
			ImportCount:     len(imports),
		},
		Names:   names,
		Imports: imports,
		Exports: exports,
	}
}

func TestWritePCC_ZeroByteExportData(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "zerobyte.pcc")

	rawData, summary := pccpat.BuildMinimalPCC([][]byte{{}})

	err := WritePCC(outPath, summary, rawData)
	if err != nil {
		t.Fatalf("WritePCC: %v", err)
	}

	readSum, err := pcc.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(readSum.Exports) != 1 {
		t.Fatalf("export count: want 1, got %d", len(readSum.Exports))
	}
	if readSum.Exports[0].SerialSize != 0 {
		t.Errorf("serial size: want 0, got %d", readSum.Exports[0].SerialSize)
	}
}

func TestWritePCC_ManyExports(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "many.pcc")

	names := []string{
		"None", "Class", "Core", "Object", "BioConversation",
	}
	for i := range 12 {
		names = append(names, fmt.Sprintf("export_%d", i))
	}
	imports := []pcc.Import{
		{ClassNameIndex: 1, ObjectNameIndex: 2},
	}
	exportData := make([][]byte, 12)
	for i := range exportData {
		exportData[i] = []byte(fmt.Sprintf("export data %d payload", i))
	}
	rawData, summary := buildTestPCC(exportData, names, imports)

	err := WritePCC(outPath, summary, rawData)
	if err != nil {
		t.Fatalf("WritePCC: %v", err)
	}

	readSum, err := pcc.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(readSum.Exports) != 12 {
		t.Fatalf("export count: want 12, got %d", len(readSum.Exports))
	}
	for i, exp := range readSum.Exports {
		want := len(exportData[i])
		if exp.SerialSize != want {
			t.Errorf("export %d size: want %d, got %d", i, want, exp.SerialSize)
		}
	}
}

func TestWritePCC_LargeExportData(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "large.pcc")

	payloadSize := 10 * 1024
	exportData := make([][]byte, 3)
	for i := range exportData {
		exportData[i] = make([]byte, payloadSize)
		for j := range exportData[i] {
			exportData[i][j] = byte((i*payloadSize + j) % 256)
		}
	}

	rawData, summary := pccpat.BuildMinimalPCC(exportData)

	err := WritePCC(outPath, summary, rawData)
	if err != nil {
		t.Fatalf("WritePCC: %v", err)
	}

	decRaw, readSum, err := pcc.ReadFileRaw(outPath)
	if err != nil {
		t.Fatalf("ReadFileRaw: %v", err)
	}
	if len(readSum.Exports) != 3 {
		t.Fatalf("export count: want 3, got %d", len(readSum.Exports))
	}
	for i := range readSum.Exports {
		exp := readSum.Exports[i]
		if exp.SerialSize != payloadSize {
			t.Errorf("export %d size: want %d, got %d", i, payloadSize, exp.SerialSize)
			continue
		}
		got := decRaw[exp.SerialOffset : exp.SerialOffset+exp.SerialSize]
		if !bytes.Equal(got, exportData[i]) {
			t.Errorf("export %d data mismatch", i)
		}
	}
}

func TestWritePCC_DifferentExportSizes(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "diffsize.pcc")

	sizes := []int{0, 1, 17, 255, 1024, 4097}
	names := []string{
		"None", "Class", "Core", "Object", "BioConversation",
	}
	for i := range sizes {
		names = append(names, fmt.Sprintf("exp_%d", i))
	}
	imports := []pcc.Import{
		{ClassNameIndex: 1, ObjectNameIndex: 2},
	}
	exportData := make([][]byte, len(sizes))
	for i, sz := range sizes {
		exportData[i] = make([]byte, sz)
		for j := range exportData[i] {
			exportData[i][j] = byte((i + j) % 256)
		}
	}
	rawData, summary := buildTestPCC(exportData, names, imports)

	err := WritePCC(outPath, summary, rawData)
	if err != nil {
		t.Fatalf("WritePCC: %v", err)
	}

	decRaw, readSum, err := pcc.ReadFileRaw(outPath)
	if err != nil {
		t.Fatalf("ReadFileRaw: %v", err)
	}
	if len(readSum.Exports) != len(sizes) {
		t.Fatalf("export count: want %d, got %d", len(sizes), len(readSum.Exports))
	}
	for i, exp := range readSum.Exports {
		if exp.SerialSize != sizes[i] {
			t.Errorf("export %d size: want %d, got %d", i, sizes[i], exp.SerialSize)
		}
		if sizes[i] > 0 {
			got := decRaw[exp.SerialOffset : exp.SerialOffset+exp.SerialSize]
			if !bytes.Equal(got, exportData[i]) {
				t.Errorf("export %d data mismatch", i)
			}
		}
	}
}

func TestWritePCCCompressed_LargeDataMultipleBlocks(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "comp_large.pcc")

	exportSize := 100 * 1024
	names := []string{
		"None", "Class", "Core", "Object", "BioConversation",
		"export_0", "export_1",
	}
	exportData := make([][]byte, 2)
	for i := range exportData {
		exportData[i] = make([]byte, exportSize)
		for j := range exportData[i] {
			exportData[i][j] = byte(rand.Intn(256))
		}
	}
	rawData, summary := buildTestPCC(exportData, names, []pcc.Import{
		{ClassNameIndex: 1, ObjectNameIndex: 2},
	})

	err := WritePCCCompressed(outPath, summary, rawData, true)
	if err != nil {
		t.Fatalf("WritePCCCompressed: %v", err)
	}

	fileBytes, _ := os.ReadFile(outPath)
	flags := binary.LittleEndian.Uint32(fileBytes[16:20])
	if flags&pcc.CompressedFlag == 0 {
		t.Error("expected compressed flag to be set")
	}

	decRaw, readSum, err := pcc.ReadFileRaw(outPath)
	if err != nil {
		t.Fatalf("ReadFileRaw: %v", err)
	}

	if !readSum.Compressed {
		t.Error("expected compressed=true in summary")
	}
	if len(readSum.Exports) != 2 {
		t.Fatalf("export count: want 2, got %d", len(readSum.Exports))
	}
	for i, exp := range readSum.Exports {
		if exp.SerialSize != exportSize {
			t.Errorf("export %d size: want %d, got %d", i, exportSize, exp.SerialSize)
			continue
		}
		got := decRaw[exp.SerialOffset : exp.SerialOffset+exp.SerialSize]
		if !bytes.Equal(got, exportData[i]) {
			t.Errorf("export %d data mismatch (first byte: want %d, got %d)", i, exportData[i][0], got[0])
		}
	}
}

func TestWritePCCCompressed_VerySmallData(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "comp_tiny.pcc")

	rawData, summary := pccpat.BuildMinimalPCC([][]byte{
		[]byte("tiny"),
	})

	err := WritePCCCompressed(outPath, summary, rawData, true)
	if err != nil {
		t.Fatalf("WritePCCCompressed: %v", err)
	}

	fileBytes, _ := os.ReadFile(outPath)
	flags := binary.LittleEndian.Uint32(fileBytes[16:20])
	if flags&pcc.CompressedFlag == 0 {
		t.Error("expected compressed flag")
	}

	readSum, err := pcc.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if !readSum.Compressed {
		t.Error("expected compressed=true")
	}
	if len(readSum.Exports) != 1 {
		t.Fatalf("export count: want 1, got %d", len(readSum.Exports))
	}
	if readSum.Exports[0].SerialSize != 4 {
		t.Errorf("serial size: want 4, got %d", readSum.Exports[0].SerialSize)
	}
}

func TestWritePCC_CompressedVsUncompressedIntegrity(t *testing.T) {
	dir := t.TempDir()

	exportData := [][]byte{
		[]byte("AAAA"),
		[]byte("BBBBBBBB"),
		[]byte("CCCC"),
	}

	rawData, summary := pccpat.BuildMinimalPCC(exportData)

	uncompPath := filepath.Join(dir, "uncomp.pcc")
	if err := WritePCCCompressed(uncompPath, summary, rawData, false); err != nil {
		t.Fatalf("WritePCCCompressed(false): %v", err)
	}

	compPath := filepath.Join(dir, "comp.pcc")
	if err := WritePCCCompressed(compPath, summary, rawData, true); err != nil {
		t.Fatalf("WritePCCCompressed(true): %v", err)
	}

	uncompRaw, uncompSum, err := pcc.ReadFileRaw(uncompPath)
	if err != nil {
		t.Fatalf("ReadFileRaw uncomp: %v", err)
	}
	compRaw, compSum, err := pcc.ReadFileRaw(compPath)
	if err != nil {
		t.Fatalf("ReadFileRaw comp: %v", err)
	}

	if len(uncompSum.Exports) != len(compSum.Exports) {
		t.Fatalf("export count mismatch: %d vs %d", len(uncompSum.Exports), len(compSum.Exports))
	}

	for i := range uncompSum.Exports {
		ue := uncompSum.Exports[i]
		ce := compSum.Exports[i]
		if ue.SerialSize != ce.SerialSize {
			t.Errorf("export %d size: uncomp=%d, comp=%d", i, ue.SerialSize, ce.SerialSize)
			continue
		}
		ud := uncompRaw[ue.SerialOffset : ue.SerialOffset+ue.SerialSize]
		cd := compRaw[ce.SerialOffset : ce.SerialOffset+ce.SerialSize]
		if !bytes.Equal(ud, cd) {
			t.Errorf("export %d data mismatch: uncomp and comp differ", i)
		}
	}

	if len(uncompSum.Names) != len(compSum.Names) {
		t.Errorf("name count: uncomp=%d, comp=%d", len(uncompSum.Names), len(compSum.Names))
	}
}

func TestWritePCCCompressed_RoundTripFullVerify(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "comp_rt.pcc")

	names := []string{
		"None", "Class", "Core", "Object", "BioConversation",
		"export_0", "export_1", "export_2", "export_3",
	}
	imports := []pcc.Import{
		{ClassNameIndex: 1, ObjectNameIndex: 2},
		{ClassNameIndex: 3, ObjectNameIndex: 4},
	}
	exportData := [][]byte{
		[]byte("alpha"),
		[]byte("bravo charlie"),
		[]byte("delta"),
		[]byte("echo foxtrot golf"),
	}
	rawData, summary := buildTestPCC(exportData, names, imports)

	err := WritePCCCompressed(outPath, summary, rawData, true)
	if err != nil {
		t.Fatalf("WritePCCCompressed: %v", err)
	}

	decRaw, readSum, err := pcc.ReadFileRaw(outPath)
	if err != nil {
		t.Fatalf("ReadFileRaw: %v", err)
	}

	if !readSum.Compressed {
		t.Error("expected compressed file")
	}
	if readSum.GameProfile != pcc.ProfileME2OT {
		t.Errorf("profile: want %s, got %s", pcc.ProfileME2OT, readSum.GameProfile)
	}

	if len(readSum.Names) != len(names) {
		t.Errorf("name count: want %d, got %d", len(names), len(readSum.Names))
	}
	for i, n := range names {
		if readSum.Names[i] != n {
			t.Errorf("name[%d]: want %q, got %q", i, n, readSum.Names[i])
		}
	}

	if len(readSum.Imports) != len(imports) {
		t.Errorf("import count: want %d, got %d", len(imports), len(readSum.Imports))
	}
	for i, imp := range imports {
		ri := readSum.Imports[i]
		if ri.ClassNameIndex != imp.ClassNameIndex {
			t.Errorf("import[%d] class: want %d, got %d", i, imp.ClassNameIndex, ri.ClassNameIndex)
		}
		if ri.ObjectNameIndex != imp.ObjectNameIndex {
			t.Errorf("import[%d] object: want %d, got %d", i, imp.ObjectNameIndex, ri.ObjectNameIndex)
		}
	}

	if len(readSum.Exports) != len(exportData) {
		t.Fatalf("export count: want %d, got %d", len(exportData), len(readSum.Exports))
	}
	for i := range readSum.Exports {
		exp := readSum.Exports[i]
		if exp.SerialSize != len(exportData[i]) {
			t.Errorf("export %d size: want %d, got %d", i, len(exportData[i]), exp.SerialSize)
		}
		got := decRaw[exp.SerialOffset : exp.SerialOffset+exp.SerialSize]
		if !bytes.Equal(got, exportData[i]) {
			t.Errorf("export %d data mismatch", i)
		}
	}
}

func TestWritePCCCompressed_UncompressedRoundTripFullVerify(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "uncomp_rt.pcc")

	names := []string{
		"None", "Class", "Core", "Object", "BioConversation",
		"my_export_0", "my_export_1",
	}
	exportData := [][]byte{
		[]byte("uncomp test payload A"),
		[]byte("uncomp test payload B with more data"),
	}
	rawData, summary := buildTestPCC(exportData, names, []pcc.Import{
		{ClassNameIndex: 1, ObjectNameIndex: 2},
	})

	err := WritePCCCompressed(outPath, summary, rawData, false)
	if err != nil {
		t.Fatalf("WritePCCCompressed(false): %v", err)
	}

	decRaw, readSum, err := pcc.ReadFileRaw(outPath)
	if err != nil {
		t.Fatalf("ReadFileRaw: %v", err)
	}

	if readSum.Compressed {
		t.Error("expected uncompressed file")
	}
	if readSum.GameProfile != pcc.ProfileME2OT {
		t.Errorf("profile: want %s, got %s", pcc.ProfileME2OT, readSum.GameProfile)
	}
	if len(readSum.Names) != len(names) {
		t.Errorf("name count: want %d, got %d", len(names), len(readSum.Names))
	}
	if readSum.Header.UnrealVersion != 512 {
		t.Errorf("UV: want 512, got %d", readSum.Header.UnrealVersion)
	}
	if readSum.Header.LicenseeVersion != 130 {
		t.Errorf("LV: want 130, got %d", readSum.Header.LicenseeVersion)
	}
	if len(readSum.Exports) != len(exportData) {
		t.Fatalf("export count: want %d, got %d", len(exportData), len(readSum.Exports))
	}
	for i := range readSum.Exports {
		exp := readSum.Exports[i]
		if exp.SerialSize != len(exportData[i]) {
			t.Errorf("export %d size: want %d, got %d", i, len(exportData[i]), exp.SerialSize)
		}
		got := decRaw[exp.SerialOffset : exp.SerialOffset+exp.SerialSize]
		if !bytes.Equal(got, exportData[i]) {
			t.Errorf("export %d data mismatch", i)
		}
	}
}

func TestWritePCC_EmptyNamesTable(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "nonames.pcc")

	names := []string{}
	rawData, summary := buildTestPCC(nil, names, nil)

	err := WritePCC(outPath, summary, rawData)
	if err != nil {
		t.Fatalf("WritePCC: %v", err)
	}

	readSum, err := pcc.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(readSum.Names) != 0 {
		t.Errorf("name count: want 0, got %d", len(readSum.Names))
	}
}

func TestWritePCC_ManyNames(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "manynames.pcc")

	names := []string{"None"}
	for i := range 55 {
		names = append(names, fmt.Sprintf("name_%d", i))
	}
	rawData, summary := buildTestPCC(nil, names, []pcc.Import{
		{ClassNameIndex: 1, ObjectNameIndex: 2},
	})

	err := WritePCC(outPath, summary, rawData)
	if err != nil {
		t.Fatalf("WritePCC: %v", err)
	}

	readSum, err := pcc.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(readSum.Names) != len(names) {
		t.Errorf("name count: want %d, got %d", len(names), len(readSum.Names))
	}
	for i, n := range names {
		if readSum.Names[i] != n {
			t.Errorf("name[%d]: want %q, got %q", i, n, readSum.Names[i])
		}
	}
}

func TestWritePCC_ManyImports(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "manyimports.pcc")

	names := []string{
		"None", "Class", "Core", "Object", "BioConversation",
	}
	for i := range 15 {
		names = append(names, fmt.Sprintf("package_%d", i))
	}
	imports := make([]pcc.Import, 12)
	for i := range imports {
		imports[i] = pcc.Import{
			ClassNameIndex:  1 + (i % 4),
			ObjectNameIndex: 5 + i,
		}
	}
	rawData, summary := buildTestPCC(nil, names, imports)

	err := WritePCC(outPath, summary, rawData)
	if err != nil {
		t.Fatalf("WritePCC: %v", err)
	}

	readSum, err := pcc.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(readSum.Imports) != 12 {
		t.Errorf("import count: want 12, got %d", len(readSum.Imports))
	}
	for i, imp := range imports {
		ri := readSum.Imports[i]
		if ri.ClassNameIndex != imp.ClassNameIndex {
			t.Errorf("import[%d] class: want %d, got %d", i, imp.ClassNameIndex, ri.ClassNameIndex)
		}
		if ri.ObjectNameIndex != imp.ObjectNameIndex {
			t.Errorf("import[%d] obj: want %d, got %d", i, imp.ObjectNameIndex, ri.ObjectNameIndex)
		}
	}
}

func TestWritePCC_ZeroImports(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "zeroimp.pcc")

	names := []string{
		"None", "Class", "Core", "Object", "BioConversation",
		"export_0",
	}
	exportData := [][]byte{[]byte("data")}
	rawData, summary := buildTestPCC(exportData, names, nil)

	err := WritePCC(outPath, summary, rawData)
	if err != nil {
		t.Fatalf("WritePCC: %v", err)
	}

	readSum, err := pcc.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}
	if len(readSum.Imports) != 0 {
		t.Errorf("import count: want 0, got %d", len(readSum.Imports))
	}
}

func TestWritePCC_NilExportData(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "nilexp.pcc")

	names := []string{
		"None", "Class", "Core", "Object", "BioConversation",
		"exp_0", "exp_1", "exp_2",
	}
	exportData := [][]byte{nil, {}, []byte("ok")}
	rawData, summary := buildTestPCC(exportData, names, []pcc.Import{
		{ClassNameIndex: 1, ObjectNameIndex: 2},
	})

	err := WritePCC(outPath, summary, rawData)
	if err != nil {
		t.Fatalf("WritePCC: %v", err)
	}

	decRaw, readSum, err := pcc.ReadFileRaw(outPath)
	if err != nil {
		t.Fatalf("ReadFileRaw: %v", err)
	}
	if len(readSum.Exports) != 3 {
		t.Fatalf("export count: want 3, got %d", len(readSum.Exports))
	}
	sizes := []int{0, 0, 2}
	for i, exp := range readSum.Exports {
		if exp.SerialSize != sizes[i] {
			t.Errorf("export %d size: want %d, got %d", i, sizes[i], exp.SerialSize)
		}
		if sizes[i] > 0 {
			got := decRaw[exp.SerialOffset : exp.SerialOffset+exp.SerialSize]
			if string(got) != "ok" {
				t.Errorf("export %d data: want 'ok', got %q", i, got)
			}
		}
	}
}

func TestWritePCCCompressed_ReadRawVerifyByteEquality(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "comp_raw.pcc")

	exportData := [][]byte{
		[]byte("Hello compressed world!"),
		{0x00, 0x01, 0x02, 0x03, 0xFF, 0xFE, 0xFD, 0xFC},
		bytes.Repeat([]byte{0xAB}, 256),
	}
	rawData, summary := pccpat.BuildMinimalPCC(exportData)

	err := WritePCCCompressed(outPath, summary, rawData, true)
	if err != nil {
		t.Fatalf("WritePCCCompressed: %v", err)
	}

	decRaw, readSum, err := pcc.ReadFileRaw(outPath)
	if err != nil {
		t.Fatalf("ReadFileRaw: %v", err)
	}

	if len(readSum.Exports) != len(exportData) {
		t.Fatalf("export count: want %d, got %d", len(exportData), len(readSum.Exports))
	}
	for i := range readSum.Exports {
		exp := readSum.Exports[i]
		if exp.SerialSize != len(exportData[i]) {
			t.Errorf("export %d size: want %d, got %d", i, len(exportData[i]), exp.SerialSize)
			continue
		}
		got := decRaw[exp.SerialOffset : exp.SerialOffset+exp.SerialSize]
		if !bytes.Equal(got, exportData[i]) {
			t.Errorf("export %d: data mismatch", i)
		}
	}
}

func TestWritePCCCompressed_NilInputs(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "nil.pcc")

	err := WritePCCCompressed(outPath, nil, nil, false)
	if err == nil {
		t.Fatal("expected error for nil summary")
	}

	rawData, summary := pccpat.BuildMinimalPCC([][]byte{[]byte("test")})
	err = WritePCCCompressed(outPath, summary, nil, false)
	if err == nil {
		t.Fatal("expected error for nil rawData")
	}

	err = WritePCCCompressed(outPath, nil, rawData, false)
	if err == nil {
		t.Fatal("expected error for nil summary")
	}
}

func TestWritePCCCompressed_NonME2(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "bad.pcc")

	rawData, summary := pccpat.BuildMinimalPCC([][]byte{[]byte("test")})
	summary.GameProfile = pcc.ProfileLE2

	err := WritePCCCompressed(outPath, summary, rawData, false)
	if err == nil {
		t.Fatal("expected error for non-ME2 profile")
	}

	err = WritePCCCompressed(outPath, summary, rawData, true)
	if err == nil {
		t.Fatal("expected error for non-ME2 profile (compressed)")
	}
}
