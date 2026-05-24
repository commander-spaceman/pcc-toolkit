package tlk

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"testing"
)

func buildMinimalTLK() []byte {
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
	writeI32(2)
	writeI32(4)

	writeI32(1)
	writeI32(0)

	writeI32(-66)
	writeI32(1)

	writeI32(-67)
	writeI32(-1)

	buf = append(buf, 0b00011010, 0, 0, 0)

	return buf
}

func TestParseTLKHeader(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}
	if tlkFile.Header.Magic != TLKMagic {
		t.Errorf("Magic = 0x%08X, want 0x%08X", tlkFile.Header.Magic, TLKMagic)
	}
	if tlkFile.Header.Version != 3 {
		t.Errorf("Version = %d, want 3", tlkFile.Header.Version)
	}
	if tlkFile.Header.MaleEntryCount != 1 {
		t.Errorf("MaleEntryCount = %d, want 1", tlkFile.Header.MaleEntryCount)
	}
	if tlkFile.Header.FemaleEntryCount != 0 {
		t.Errorf("FemaleEntryCount = %d, want 0", tlkFile.Header.FemaleEntryCount)
	}
	if tlkFile.Header.TreeNodeCount != 2 {
		t.Errorf("TreeNodeCount = %d, want 2", tlkFile.Header.TreeNodeCount)
	}
}

func TestParseTLKInvalidMagic(t *testing.T) {
	data := buildMinimalTLK()
	binary.LittleEndian.PutUint32(data[0:4], 0xDEADBEEF)
	_, err := Parse(data, "test.tlk")
	if err == nil {
		t.Fatal("expected error for invalid magic")
	}
}

func TestDecodeString(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	text, ok := ResolveString(tlkFile, 1, true)
	if !ok {
		t.Fatal("expected string for id 1")
	}
	if text != "AB" {
		t.Errorf("text = %q, want AB", text)
	}
}

func TestDecodeStringNotFound(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	_, ok := ResolveString(tlkFile, 999, true)
	if ok {
		t.Fatal("expected false for unknown id")
	}
}

func TestGetBit(t *testing.T) {
	data := []byte{0b10101010, 0b01010101}
	if getBit(data, 0) != false {
		t.Error("bit 0 should be 0")
	}
	if getBit(data, 1) != true {
		t.Error("bit 1 should be 1")
	}
	if getBit(data, 7) != true {
		t.Error("bit 7 should be 1")
	}
	if getBit(data, 8) != true {
		t.Error("bit 8 should be 1")
	}
	if getBit(data, 9) != false {
		t.Error("bit 9 should be 0")
	}
}

func TestStringIDs(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	ids := tlkFile.StringIDs()
	if len(ids) != 1 {
		t.Fatalf("got %d ids, want 1", len(ids))
	}
	if ids[0] != 1 {
		t.Errorf("ids[0] = %d, want 1", ids[0])
	}
}

func TestIterEntries(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	count := 0
	tlkFile.IterEntries()(func(id int32, text string) bool {
		count++
		return true
	})
	if count != 1 {
		t.Errorf("got %d entries, want 1", count)
	}
}

func TestSearch(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	results := tlkFile.Search("AB")
	if len(results) != 1 {
		t.Fatalf("got %d search results, want 1", len(results))
	}
	if results[0].StringID != 1 {
		t.Errorf("StringID = %d, want 1", results[0].StringID)
	}

	results = tlkFile.Search("ZZZ")
	if len(results) != 0 {
		t.Errorf("expected 0 results for 'ZZZ', got %d", len(results))
	}
}

func TestResolverResolve(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	resolver := &Resolver{Files: []*File{tlkFile}}
	text, ok := resolver.Resolve(1)
	if !ok {
		t.Fatal("expected resolution for id 1")
	}
	if text != "AB" {
		t.Errorf("text = %q, want AB", text)
	}
}

func TestResolverNotFound(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	resolver := &Resolver{Files: []*File{tlkFile}}
	_, ok := resolver.Resolve(999)
	if ok {
		t.Fatal("expected not found for id 999")
	}
}

func TestResolverPriority(t *testing.T) {
	base, _ := Parse(buildMinimalTLK(), "base.tlk")
	override, _ := Parse(buildMinimalTLK(), "override.tlk")

	resolver := &Resolver{Files: []*File{override, base}}
	text, ok := resolver.Resolve(1)
	if !ok {
		t.Fatal("expected resolution from override")
	}
	if text != "AB" {
		t.Errorf("text = %q, want AB", text)
	}
}

func TestResolveWithSource(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	resolver := &Resolver{Files: []*File{tlkFile}}
	result := resolver.ResolveWithSource(1)
	if result == nil {
		t.Fatal("expected result for id 1")
	}
	if !result.Found {
		t.Fatal("expected Found=true for id 1")
	}
	if result.SourceTLK != "test.tlk" {
		t.Errorf("SourceTLK = %q, want test.tlk", result.SourceTLK)
	}
	if result.Text != "AB" {
		t.Errorf("Text = %q, want AB", result.Text)
	}
}

func TestResolveWithSourceNotFound(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	resolver := &Resolver{Files: []*File{tlkFile}}
	result := resolver.ResolveWithSource(999)
	if result == nil {
		t.Fatal("expected non-nil result for id 999")
	}
	if result.Found {
		t.Fatal("expected Found=false for unknown id 999")
	}
	if result.StringID != 999 {
		t.Errorf("StringID = %d, want 999", result.StringID)
	}
}

func TestResolverSearch(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	resolver := &Resolver{Files: []*File{tlkFile}}
	results := resolver.Search("ab")
	if len(results) != 1 {
		t.Fatalf("got %d search results, want 1", len(results))
	}
	if results[0].SourceTLK != "test.tlk" {
		t.Errorf("SourceTLK = %q, want test.tlk", results[0].SourceTLK)
	}
}

func TestResolverTotalUnique(t *testing.T) {
	data := buildMinimalTLK()
	tlkFile, err := Parse(data, "test.tlk")
	if err != nil {
		t.Fatalf("Parse: %v", err)
	}

	resolver := &Resolver{Files: []*File{tlkFile}}
	n := resolver.TotalUniqueEntries()
	if n != 1 {
		t.Errorf("TotalUniqueEntries = %d, want 1", n)
	}
}

func TestParseBioEngineModules(t *testing.T) {
	t.Run("valid modules section", func(t *testing.T) {
		iniContent := `[Engine.DLCModules]
DLC_HEN_MT=3
DLC_CER_02=2
`
		tmpDir := t.TempDir()
		iniPath := filepath.Join(tmpDir, "BIOEngine.ini")
		if err := os.WriteFile(iniPath, []byte(iniContent), 0644); err != nil {
			t.Fatalf("write ini: %v", err)
		}

		modules, err := ParseBioEngineModules(iniPath)
		if err != nil {
			t.Fatalf("ParseBioEngineModules: %v", err)
		}
		if len(modules) != 2 {
			t.Errorf("got %d modules, want 2", len(modules))
		}
		if v, ok := modules["DLC_HEN_MT"]; !ok || v != 3 {
			t.Errorf("DLC_HEN_MT = %d, want 3", v)
		}
		if v, ok := modules["DLC_CER_02"]; !ok || v != 2 {
			t.Errorf("DLC_CER_02 = %d, want 2", v)
		}
	})

	t.Run("file not found", func(t *testing.T) {
		_, err := ParseBioEngineModules("nonexistent.ini")
		if err == nil {
			t.Fatal("expected error for missing file")
		}
	})

	t.Run("no dlcmodules section", func(t *testing.T) {
		iniContent := `[Core.System]
!CookPaths=CLEAR
`
		tmpDir := t.TempDir()
		iniPath := filepath.Join(tmpDir, "BIOEngine.ini")
		os.WriteFile(iniPath, []byte(iniContent), 0644)

		modules, err := ParseBioEngineModules(iniPath)
		if err != nil {
			t.Fatalf("ParseBioEngineModules: %v", err)
		}
		if len(modules) != 0 {
			t.Errorf("got %d modules, want 0", len(modules))
		}
	})

	t.Run("ignore comments and empty lines", func(t *testing.T) {
		iniContent := `
; This is a comment
# Another comment
[Engine.DLCModules]
; Skip this
DLC_MOD=5
`
		tmpDir := t.TempDir()
		iniPath := filepath.Join(tmpDir, "BIOEngine.ini")
		os.WriteFile(iniPath, []byte(iniContent), 0644)

		modules, err := ParseBioEngineModules(iniPath)
		if err != nil {
			t.Fatalf("ParseBioEngineModules: %v", err)
		}
		if v, ok := modules["DLC_MOD"]; !ok || v != 5 {
			t.Errorf("DLC_MOD = %d, want 5", v)
		}
	})
}

func TestFindDlcTlkByModule(t *testing.T) {
	t.Run("finds tlk in CookedPC", func(t *testing.T) {
		tmpDir := t.TempDir()
		cookedDir := filepath.Join(tmpDir, "CookedPC")
		os.MkdirAll(cookedDir, 0755)
		expected := filepath.Join(cookedDir, "DLC_3_INT.tlk")
		os.WriteFile(expected, []byte("dummy"), 0644)

		result := findDlcTlkByModule(tmpDir, 3, "INT")
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("finds tlk in CookedPCConsole", func(t *testing.T) {
		tmpDir := t.TempDir()
		cookedDir := filepath.Join(tmpDir, "CookedPCConsole")
		os.MkdirAll(cookedDir, 0755)
		expected := filepath.Join(cookedDir, "DLC_5_DEU.tlk")
		os.WriteFile(expected, []byte("dummy"), 0644)

		result := findDlcTlkByModule(tmpDir, 5, "DEU")
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("finds tlk in dlc root", func(t *testing.T) {
		tmpDir := t.TempDir()
		expected := filepath.Join(tmpDir, "DLC_7_INT.tlk")
		os.WriteFile(expected, []byte("dummy"), 0644)

		result := findDlcTlkByModule(tmpDir, 7, "INT")
		if result != expected {
			t.Errorf("got %q, want %q", result, expected)
		}
	})

	t.Run("returns empty when not found", func(t *testing.T) {
		tmpDir := t.TempDir()
		result := findDlcTlkByModule(tmpDir, 99, "INT")
		if result != "" {
			t.Errorf("expected empty, got %q", result)
		}
	})
}

func TestParseMountPriorityBinary(t *testing.T) {
	t.Run("me2 mount 0x00", func(t *testing.T) {
		data := make([]byte, 14)
		data[0] = 0x00
		binary.LittleEndian.PutUint16(data[12:14], 42)
		pri := parseMountPriorityBinary(data)
		if pri != 42 {
			t.Errorf("got %d, want 42", pri)
		}
	})

	t.Run("le2 mount 0xAC", func(t *testing.T) {
		data := make([]byte, 14)
		data[0] = 0xAC
		binary.LittleEndian.PutUint16(data[12:14], 10)
		pri := parseMountPriorityBinary(data)
		if pri != 10 {
			t.Errorf("got %d, want 10", pri)
		}
	})

	t.Run("unknown header returns 0", func(t *testing.T) {
		data := make([]byte, 14)
		data[0] = 0xFF
		pri := parseMountPriorityBinary(data)
		if pri != 0 {
			t.Errorf("got %d, want 0", pri)
		}
	})

	t.Run("too short returns 0", func(t *testing.T) {
		data := make([]byte, 10)
		pri := parseMountPriorityBinary(data)
		if pri != 0 {
			t.Errorf("got %d, want 0", pri)
		}
	})
}
