package dialogue

import (
	"os"
	"path/filepath"
	"testing"

	"pcc-toolkit/core/internal/pcc"
)

func outputPCCPath(name string) string {
	wd, err := os.Getwd()
	if err != nil {
		return ""
	}
	return filepath.Join(wd, "..", "..", "..", "output", "property-parsing", name)
}

func TestSchemaGuidedParse_Synthetic(t *testing.T) {
	names := []string{
		"Core", "Class", "BioDialogEntryNode", "BioDialogReplyNode",
		"nIndex", "nSpeakerIndex", "srText", "ReplyListNew", "IntProperty",
		"StringRefProperty", "ArrayProperty", "BioDialogReplyListDetails",
		"StructProperty", "None",
	}

	buf := make([]byte, 1024)
	off := 0

	writeI32 := func(v int) {
		putI32LE(buf, off, v)
		off += 4
	}
	writeFName := func(nameIdx int) {
		writeI32(nameIdx)
		writeI32(0)
	}

	nameIdx := func(n string) int {
		for i, name := range names {
			if name == n {
				return i
			}
		}
		return -1
	}

	payloadStart := off

	// Item 0: BioDialogEntryNode
	writeFName(nameIdx("nIndex"))
	writeFName(nameIdx("IntProperty"))
	writeI32(4)
	writeI32(0)
	writeI32(0)

	writeFName(nameIdx("nSpeakerIndex"))
	writeFName(nameIdx("IntProperty"))
	writeI32(4)
	writeI32(0)
	writeI32(3)

	writeFName(nameIdx("srText"))
	writeFName(nameIdx("StringRefProperty"))
	writeI32(4)
	writeI32(0)
	writeI32(12345)

	writeFName(nameIdx("None"))

	// Item 1
	writeFName(nameIdx("nIndex"))
	writeFName(nameIdx("IntProperty"))
	writeI32(4)
	writeI32(0)
	writeI32(1)

	writeFName(nameIdx("None"))

	payloadSize := off - payloadStart

	items := parseStructArraySchemaGuided(buf, names, "BioDialogEntryNode", payloadStart, payloadSize, 2)
	if items == nil {
		t.Fatal("parseStructArraySchemaGuided returned nil")
	}
	if len(items) != 2 {
		t.Fatalf("expected 2 items, got %d", len(items))
	}

	nidx0, ok := items[0]["nIndex"]
	if !ok || nidx0.Value.(int) != 0 {
		t.Errorf("item[0].nIndex = %v, want 0", nidx0.Value)
	}

	nidx1, ok := items[1]["nIndex"]
	if !ok || nidx1.Value.(int) != 1 {
		t.Errorf("item[1].nIndex = %v, want 1", nidx1.Value)
	}

	sr, ok := items[0]["srText"]
	if !ok {
		t.Error("item[0]: missing srText")
	} else if v, ok := sr.Value.(int); !ok || v != 12345 {
		t.Errorf("item[0].srText = %v, want 12345", sr.Value)
	}
}

func TestSchemaGuidedParse_UnknownType(t *testing.T) {
	result := parseStructArraySchemaGuided(make([]byte, 64), []string{"nIndex"}, "UnknownStruct", 0, 64, 1)
	if result != nil {
		t.Error("expected nil for unknown struct type")
	}
}

func TestParseConversation_RealFile(t *testing.T) {
	pccPath := outputPCCPath("BioD_CitHub_300Dialogue_LOC_INT.pcc")
	if _, err := os.Stat(pccPath); os.IsNotExist(err) {
		t.Skipf("output file not found: %s", pccPath)
	}

	rawData, summary, err := pcc.ReadFileRaw(pccPath)
	if err != nil {
		t.Fatalf("ReadFileRaw: %v", err)
	}
	if err := summary.RequireME2(); err != nil {
		t.Fatalf("RequireME2: %v", err)
	}

	convExport := -1
	for _, e := range summary.Exports {
		if e.ClassName == "BioConversation" {
			convExport = e.Index
			break
		}
	}
	if convExport < 0 {
		t.Fatal("no BioConversation export found")
	}

	conv, err := ParseConversation(summary, rawData, convExport)
	if err != nil {
		t.Fatalf("ParseConversation: %v", err)
	}

	if conv.ParseMode != "struct_property_semantic" {
		t.Fatalf("expected parse_mode struct_property_semantic, got %s", conv.ParseMode)
	}

	if len(conv.Entries) == 0 {
		t.Fatal("no entries parsed")
	}

	hasSpeaker := false
	hasStrRef := false
	for _, entry := range conv.Entries {
		if entry.SpeakerID != nil {
			hasSpeaker = true
		}
		if entry.LineStrRef != nil {
			hasStrRef = true
		}
	}
	if !hasSpeaker {
		t.Error("no entry has speaker_id set")
	}
	if !hasStrRef {
		t.Error("no entry has line_strref set")
	}

	if len(conv.Starts) == 0 {
		t.Error("no starts parsed")
	}
}

func TestExportProperties_RealFile(t *testing.T) {
	pccPath := outputPCCPath("BioD_CitHub_300Dialogue_LOC_INT.pcc")
	if _, err := os.Stat(pccPath); os.IsNotExist(err) {
		t.Skipf("output file not found: %s", pccPath)
	}

	rawData, summary, err := pcc.ReadFileRaw(pccPath)
	if err != nil {
		t.Fatalf("ReadFileRaw: %v", err)
	}
	if err := summary.RequireME2(); err != nil {
		t.Fatalf("RequireME2: %v", err)
	}

	propMap := pcc.ComputeExportProperties(rawData, summary, true, true)

	convExport := -1
	for _, e := range summary.Exports {
		if e.ClassName == "BioConversation" {
			convExport = e.Index
			break
		}
	}
	if convExport < 0 {
		t.Fatal("no BioConversation export found")
	}

	ep, ok := propMap[convExport]
	if !ok {
		t.Fatal("no properties computed for BioConversation export")
	}

	if len(ep.Tags) == 0 {
		t.Error("expected property tags for BioConversation")
	}
	if ep.SemanticProps == nil {
		t.Error("expected semantic props for BioConversation")
	}
}

func TestParseConversation_RealFile_ReplyChoices(t *testing.T) {
	pccPath := outputPCCPath("BioD_CitHub_300Dialogue_LOC_INT.pcc")
	if _, err := os.Stat(pccPath); os.IsNotExist(err) {
		t.Skipf("output file not found: %s", pccPath)
	}

	rawData, summary, err := pcc.ReadFileRaw(pccPath)
	if err != nil {
		t.Fatalf("ReadFileRaw: %v", err)
	}
	if err := summary.RequireME2(); err != nil {
		t.Fatalf("RequireME2: %v", err)
	}

	convExport := -1
	for _, e := range summary.Exports {
		if e.ClassName == "BioConversation" {
			convExport = e.Index
			break
		}
	}
	if convExport < 0 {
		t.Fatal("no BioConversation export found")
	}

	conv, err := ParseConversation(summary, rawData, convExport)
	if err != nil {
		t.Fatalf("ParseConversation: %v", err)
	}

	hasReplyChoices := false
	hasOrder := false
	for _, entry := range conv.Entries {
		if len(entry.ReplyChoices) > 0 {
			hasReplyChoices = true
			for _, rc := range entry.ReplyChoices {
				if rc.Order > 0 {
					hasOrder = true
				}
				if rc.FromEntryID != entry.ID {
					t.Errorf("ReplyChoice.FromEntryID = %d, want %d", rc.FromEntryID, entry.ID)
				}
				if rc.ToReplyID == 0 && len(entry.ReplyChoices) == 1 {
					continue
				}
			}
		}
	}

	if !hasReplyChoices {
		t.Error("no entries have reply_choices with ReplyListNew data")
	}
	if !hasOrder {
		t.Log("no reply_choices with Order > 0 found (may be valid for this conversation)")
	}
}

func TestParseConversation_RealFile_ReplyChoices_Category(t *testing.T) {
	pccPath := outputPCCPath("BioD_CitHub_300Dialogue_LOC_INT.pcc")
	if _, err := os.Stat(pccPath); os.IsNotExist(err) {
		t.Skipf("output file not found: %s", pccPath)
	}

	rawData, summary, err := pcc.ReadFileRaw(pccPath)
	if err != nil {
		t.Fatalf("ReadFileRaw: %v", err)
	}
	if err := summary.RequireME2(); err != nil {
		t.Fatalf("RequireME2: %v", err)
	}

	parseResult := ParseConversations(summary, rawData, "resilient")
	if len(parseResult.Errors) > 0 {
		t.Logf("ParseConversations warnings: %v", parseResult.Errors)
	}

	totalChoices := 0
	withCategory := 0
	for _, conv := range parseResult.Conversations {
		for _, entry := range conv.Entries {
			for _, rc := range entry.ReplyChoices {
				totalChoices++
				if rc.Category != "" {
					withCategory++
				}
			}
		}
	}

	if totalChoices == 0 {
		t.Skip("no reply choices found in any conversation")
	}

	t.Logf("reply choices: %d total, %d with category", totalChoices, withCategory)
}

func putI32LE(buf []byte, offset int, v int) {
	u := uint32(v)
	buf[offset] = byte(u)
	buf[offset+1] = byte(u >> 8)
	buf[offset+2] = byte(u >> 16)
	buf[offset+3] = byte(u >> 24)
}
