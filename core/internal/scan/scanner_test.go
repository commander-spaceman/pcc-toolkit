package scan

import (
	"encoding/binary"
	"testing"
)

func TestFindOffsets(t *testing.T) {
	tests := []struct {
		name   string
		data   []byte
		target []byte
		want   []int
	}{
		{
			name:   "single match",
			data:   []byte{0x0A, 0x00, 0x00, 0x00},
			target: []byte{0x0A, 0x00, 0x00, 0x00},
			want:   []int{0},
		},
		{
			name:   "nil target",
			data:   []byte{0x01, 0x02, 0x03, 0x04},
			target: nil,
			want:   nil,
		},
		{
			name:   "empty target",
			data:   []byte{0x01, 0x02, 0x03},
			target: []byte{},
			want:   nil,
		},
		{
			name:   "data too short",
			data:   []byte{0x01},
			target: []byte{0x01, 0x02, 0x03, 0x04},
			want:   nil,
		},
		{
			name:   "two matches",
			data:   []byte{0x01, 0x00, 0x00, 0x00, 0xFF, 0x01, 0x00, 0x00, 0x00},
			target: []byte{0x01, 0x00, 0x00, 0x00},
			want:   []int{0, 5},
		},
		{
			name:   "no match",
			data:   []byte{0x01, 0x02, 0x03, 0x04, 0x05},
			target: []byte{0xFF, 0xFF, 0xFF, 0xFF},
			want:   nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := findOffsets(tt.data, tt.target)
			if len(got) != len(tt.want) {
				t.Fatalf("findOffsets() len = %d, want %d", len(got), len(tt.want))
			}
			for i := range tt.want {
				if got[i] != tt.want[i] {
					t.Errorf("findOffsets()[%d] = %d, want %d", i, got[i], tt.want[i])
				}
			}
		})
	}
}

func TestParseStrrefs(t *testing.T) {
	strref := int32(1)
	data := make([]byte, 20)
	binary.LittleEndian.PutUint32(data[5:9], uint32(strref))
	binary.LittleEndian.PutUint32(data[12:16], uint32(strref))

	offsets := ParseStrrefs(data, []int32{strref})
	if len(offsets) != 1 {
		t.Fatalf("expected 1 strref entry, got %d", len(offsets))
	}
	if len(offsets[int(strref)]) != 2 {
		t.Errorf("expected 2 offsets for strref %d, got %d", strref, len(offsets[int(strref)]))
	}
	if offsets[int(strref)][0] != 5 || offsets[int(strref)][1] != 12 {
		t.Errorf("expected offsets [5, 12], got %v", offsets[int(strref)])
	}
}

func TestParseStrrefsMultiple(t *testing.T) {
	a := int32(100)
	b := int32(200)
	data := make([]byte, 30)
	binary.LittleEndian.PutUint32(data[0:4], uint32(a))
	binary.LittleEndian.PutUint32(data[10:14], uint32(b))
	binary.LittleEndian.PutUint32(data[20:24], uint32(a))

	offsets := ParseStrrefs(data, []int32{a, b})
	if len(offsets) != 2 {
		t.Fatalf("expected 2 strref entries, got %d", len(offsets))
	}
	if len(offsets[int(a)]) != 2 {
		t.Errorf("expected 2 offsets for a, got %d", len(offsets[int(a)]))
	}
	if len(offsets[int(b)]) != 1 {
		t.Errorf("expected 1 offset for b, got %d", len(offsets[int(b)]))
	}
}

func TestParseStrrefsEmpty(t *testing.T) {
	offsets := ParseStrrefs([]byte{0x01}, []int32{42})
	if offsets != nil {
		t.Errorf("expected nil for too-short data, got %v", offsets)
	}
}

func TestDeduplicateByFilename(t *testing.T) {
	files := []FileEntry{
		{Path: "C:\\game\\CookedPC\\Test.pcc", MountFile: "basegame", MountPri: -1},
		{Path: "C:\\game\\DLC\\DLC_EXP\\CookedPC\\Test.pcc", MountFile: "DLC_EXP", MountPri: 100},
		{Path: "C:\\game\\DLC\\DLC_EXP\\CookedPC\\Other.pcc", MountFile: "DLC_EXP", MountPri: 100},
	}

	result := deduplicateByFilename(files)
	if len(result) != 2 {
		t.Fatalf("expected 2 deduped files, got %d", len(result))
	}

	for _, f := range result {
		if f.MountFile == "basegame" {
			t.Error("basegame file should have been overridden by DLC")
		}
	}
}
