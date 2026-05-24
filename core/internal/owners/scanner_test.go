package owners

import (
	"encoding/binary"
	"testing"

	"pcc-toolkit/core/internal/pcc"
)

func writeI32LE(buf []byte, off int, v int) {
	binary.LittleEndian.PutUint32(buf[off:off+4], uint32(v))
}

func TestResolveObjectName_Export(t *testing.T) {
	summary := &pcc.FileSummary{
		Exports: []pcc.Export{
			{Index: 0, ObjectName: "SomeObject"},
			{Index: 1, ObjectName: "TargetConv"},
		},
	}
	got := resolveObjectName(2, summary) // UIndex 2 = export[1]
	if got != "TargetConv" {
		t.Errorf("resolveObjectName = %q, want TargetConv", got)
	}
}

func TestResolveObjectName_Import(t *testing.T) {
	summary := &pcc.FileSummary{
		Names: []string{"ImportedConv"},
		Imports: []pcc.Import{
			{ObjectNameIndex: 0},
		},
	}
	got := resolveObjectName(-1, summary) // UIndex -1 = import[0]
	if got != "ImportedConv" {
		t.Errorf("resolveObjectName = %q, want ImportedConv", got)
	}
}

func TestResolveObjectName_Zero(t *testing.T) {
	summary := &pcc.FileSummary{}
	got := resolveObjectName(0, summary)
	if got != "" {
		t.Errorf("resolveObjectName = %q, want empty", got)
	}
}

func TestScanOwners_NoStartConversation(t *testing.T) {
	summary := &pcc.FileSummary{
		Exports: []pcc.Export{
			{Index: 0, ClassName: "BioConversation", ObjectName: "SomeConv", SerialOffset: 0, SerialSize: 4},
		},
	}
	rawData := []byte{0, 0, 0, 0}
	output := ScanOwners(rawData, summary, "test.pcc")
	if len(output.Owners) != 0 {
		t.Errorf("expected 0 owners, got %d", len(output.Owners))
	}
}

func TestScanOwners_InvalidData(t *testing.T) {
	summary := &pcc.FileSummary{
		Exports: []pcc.Export{
			{Index: 0, ClassName: "BioSeqAct_StartConversation", ObjectName: "StartConv", SerialOffset: 0, SerialSize: 4},
		},
	}
	rawData := []byte{0, 0, 0, 0}
	output := ScanOwners(rawData, summary, "test.pcc")
	if len(output.Owners) != 0 {
		t.Errorf("expected 0 owners for invalid data, got %d", len(output.Owners))
	}
}

func TestScanOwners_StartConversationWithoutOwner(t *testing.T) {
	names := []string{
		"Core", "Class", "BioConversation", "AnyConv",
		"Conv", "ObjectProperty", "VariableLinks", "ArrayProperty",
		"StructProperty", "LinkDesc", "StrProperty", "LinkedVariables",
		"Owner", "SeqVar_Object", "ObjValue", "Tag",
		"NameProperty", "IntProperty", "None",
	}

	nameIdx := func(n string) int {
		for i, name := range names {
			if name == n {
				return i
			}
		}
		return -1
	}

	writeFName := func(buf []byte, off int, nameIdx int) int {
		writeI32LE(buf, off, nameIdx)
		writeI32LE(buf, off+4, 0)
		return off + 8
	}

	// Build a StartConversation export with Conv property and VariableLinks (no Owner link)
	buf := make([]byte, 1024)
	off := 0

	// Conv property: ObjectProperty pointing to export at index 0 (UIndex 1)
	off = writeFName(buf, off, nameIdx("Conv"))
	off = writeFName(buf, off, nameIdx("ObjectProperty"))
	writeI32LE(buf, off, 4) // propSize
	writeI32LE(buf, off+4, 0)
	off += 8
	writeI32LE(buf, off, 1) // ObjectProperty value: UIndex 1 = export[0]
	off += 4

	// VariableLinks: ArrayProperty<StructProperty> with one item but LinkDesc != "Owner"
	off = writeFName(buf, off, nameIdx("VariableLinks"))
	off = writeFName(buf, off, nameIdx("ArrayProperty"))
	writeI32LE(buf, off, 4+8+4) // propSize: count(4) + StructProperty FName(8) + struct payload(4)
	writeI32LE(buf, off+4, 0)
	off += 8
	// Array metadata: StructProperty FName
	off = writeFName(buf, off, nameIdx("StructProperty"))
	// Array count
	writeI32LE(buf, off, 1) // count = 1
	off += 4
	// Struct item: LinkDesc = "NotOwner"
	structOff := off
	off = writeFName(buf, off, nameIdx("LinkDesc"))
	off = writeFName(buf, off, nameIdx("StrProperty"))
	writeI32LE(buf, off, 4+8) // propSize: count(4) + "NotOwner"(8)
	writeI32LE(buf, off+4, 0)
	off += 8
	writeI32LE(buf, off, 8) // string length
	off += 4
	copy(buf[off:], "NotOwner")
	off += 8
	off = writeFName(buf, off, nameIdx("None"))
	propSize := off - structOff
	// Fix the VariableLinks size
	writeI32LE(buf, 40, 4+8+propSize)

	exp := pcc.Export{
		Index:        0,
		ClassName:    "BioSeqAct_StartConversation",
		ObjectName:   "StartConv_NoOwner",
		SerialOffset: 0,
		SerialSize:   off,
	}
	summary := &pcc.FileSummary{
		Names: names,
		Exports: []pcc.Export{
			{Index: 0, ObjectName: "AnyConv", ClassName: "BioConversation"},
			exp,
		},
	}

	output := ScanOwners(buf, summary, "test.pcc")
	if output.File != "test.pcc" {
		t.Errorf("file = %q, want test.pcc", output.File)
	}
	// Should find the conversation name but owner is "Not found"
	if len(output.Owners) != 1 {
		t.Errorf("expected 1 owner entry, got %d", len(output.Owners))
	} else {
		if output.Owners[0].ConversationName != "AnyConv" {
			t.Errorf("conversation = %q, want AnyConv", output.Owners[0].ConversationName)
		}
		if output.Owners[0].OwnerTag != "Not found" {
			t.Errorf("owner = %q, want Not found", output.Owners[0].OwnerTag)
		}
	}
}
