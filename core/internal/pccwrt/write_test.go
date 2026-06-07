package pccwrt

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"

	"pcc-toolkit/core/internal/pcc"
	"pcc-toolkit/core/internal/pccpat"
)

func TestWriteHeader_Bytes(t *testing.T) {
	buf := make([]byte, 100)
	writeHeader(buf, 0, 8, 100, 1, 255, 1, 227, 1, 327)

	importCnt := int32(binary.LittleEndian.Uint32(buf[36:40]))
	if importCnt != 1 {
		t.Errorf("importCount at bytes 36-39: want 1, got %d (raw: %v)", importCnt, buf[36:40])
	}

	importOff := int32(binary.LittleEndian.Uint32(buf[40:44]))
	if importOff != 227 {
		t.Errorf("importOffset at bytes 40-43: want 227, got %d", importOff)
	}
}

func TestWritePCC_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "test.pcc")

	rawData, summary := pccpat.BuildMinimalPCC([][]byte{
		[]byte("hello world export 0"),
	})

	err := WritePCC(outPath, summary, rawData)
	if err != nil {
		t.Fatalf("WritePCC: %v", err)
	}

	if _, err := os.Stat(outPath); os.IsNotExist(err) {
		t.Fatal("output file does not exist")
	}

	readSummary, err := pcc.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if readSummary.GameProfile != pcc.ProfileME2OT {
		t.Errorf("game profile: want %s, got %s", pcc.ProfileME2OT, readSummary.GameProfile)
	}
	if readSummary.Compressed {
		t.Error("expected uncompressed file")
	}
	if len(readSummary.Names) != len(summary.Names) {
		t.Errorf("name count: want %d, got %d", len(summary.Names), len(readSummary.Names))
	}
	for i, n := range summary.Names {
		if readSummary.Names[i] != n {
			t.Errorf("name[%d]: want %q, got %q", i, n, readSummary.Names[i])
		}
	}
	if len(readSummary.Imports) != len(summary.Imports) {
		t.Errorf("import count: want %d, got %d", len(summary.Imports), len(readSummary.Imports))
	}
	if len(readSummary.Exports) != len(summary.Exports) {
		t.Errorf("export count: want %d, got %d", len(summary.Exports), len(readSummary.Exports))
	}

	for i, exp := range summary.Exports {
		readExp := readSummary.Exports[i]
		if readExp.SerialSize != exp.SerialSize {
			t.Errorf("export %d SerialSize: want %d, got %d", i, exp.SerialSize, readExp.SerialSize)
		}
		if readExp.ObjectName != exp.ObjectName {
			t.Errorf("export %d ObjectName: want %q, got %q", i, exp.ObjectName, readExp.ObjectName)
		}
	}

	rawData2, summary2, err := pcc.ReadFileRaw(outPath)
	if err != nil {
		t.Fatalf("ReadFileRaw: %v", err)
	}

	if len(rawData2) < 100 {
		t.Errorf("rawData too small: %d bytes", len(rawData2))
	}
	_ = summary2
}

func TestWritePCC_MultipleExports(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "multi.pcc")

	rawData, summary := pccpat.BuildMinimalPCC([][]byte{
		[]byte("AAAA"),
		[]byte("BBBB"),
		[]byte("CCCCCC"),
	})

	err := WritePCC(outPath, summary, rawData)
	if err != nil {
		t.Fatalf("WritePCC: %v", err)
	}

	readSummary, err := pcc.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if len(readSummary.Exports) != 3 {
		t.Fatalf("export count: want 3, got %d", len(readSummary.Exports))
	}

	sizes := []int{4, 4, 6}
	for i, want := range sizes {
		if readSummary.Exports[i].SerialSize != want {
			t.Errorf("export %d SerialSize: want %d, got %d", i, want, readSummary.Exports[i].SerialSize)
		}
	}
}

func TestWritePCC_EmptyExports(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "empty.pcc")

	rawData, summary := pccpat.BuildMinimalPCC(nil)

	err := WritePCC(outPath, summary, rawData)
	if err != nil {
		t.Fatalf("WritePCC: %v", err)
	}

	readSummary, err := pcc.ReadFile(outPath)
	if err != nil {
		t.Fatalf("ReadFile: %v", err)
	}

	if len(readSummary.Exports) != 0 {
		t.Errorf("export count: want 0, got %d", len(readSummary.Exports))
	}
}

func TestWritePCC_NilInputs(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "nil.pcc")

	err := WritePCC(outPath, nil, nil)
	if err == nil {
		t.Fatal("expected error for nil summary")
	}

	rawData, summary := pccpat.BuildMinimalPCC([][]byte{[]byte("test")})
	err = WritePCC(outPath, summary, nil)
	if err == nil {
		t.Fatal("expected error for nil rawData")
	}
	_ = rawData
}

func TestWritePCC_NonME2(t *testing.T) {
	dir := t.TempDir()
	outPath := filepath.Join(dir, "nonme2.pcc")

	rawData, summary := pccpat.BuildMinimalPCC([][]byte{[]byte("test")})
	summary.GameProfile = pcc.ProfileLE2

	err := WritePCC(outPath, summary, rawData)
	if err == nil {
		t.Fatal("expected error for non-ME2 profile")
	}
}
