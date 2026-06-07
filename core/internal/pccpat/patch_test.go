package pccpat

import (
	"testing"
)

func TestPatchExport_SameSize(t *testing.T) {
	original := []byte("hello world")
	data, summary := BuildMinimalPCC([][]byte{original})

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
	data, summary := BuildMinimalPCC([][]byte{original})

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
	data, summary := BuildMinimalPCC([][]byte{original})

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
	data, summary := BuildMinimalPCC([][]byte{original1, original2})

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
	data, summary := BuildMinimalPCC([][]byte{[]byte("test")})

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
	data, summary := BuildMinimalPCC([][]byte{original})

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
	_, summary := BuildMinimalPCC([][]byte{[]byte("a"), []byte("b"), []byte("c")})

	buf, _ := BuildMinimalPCC([][]byte{[]byte("a"), []byte("b"), []byte("c")})
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
