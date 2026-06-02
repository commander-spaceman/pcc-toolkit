package evidence

import (
	"testing"

	"pcc-toolkit/core/internal/pcc"
	"pcc-toolkit/core/internal/scan"
	"pcc-toolkit/core/internal/tlk"
)

func TestClassifyTier(t *testing.T) {
	tests := []struct {
		className string
		hasBioC   bool
		wantTier  EvidenceTier
	}{
		{"BioConversation", false, TierBioConversation},
		{"BioConversation", true, TierBioConversation},
		{"BioPawn", true, TierSemanticContainer},
		{"BioPawn", false, TierContainerFallback},
		{"SeqAct_Interp", true, TierSemanticContainer},
		{"SeqAct_Interp", false, TierContainerFallback},
		{"", true, TierSemanticContainer},
		{"", false, TierContainerFallback},
	}

	for _, tt := range tests {
		t.Run(string(tt.wantTier)+"_"+tt.className, func(t *testing.T) {
			got := ClassifyTier(tt.className, tt.hasBioC)
			if got != tt.wantTier {
				t.Errorf("ClassifyTier(%q, %v) = %s, want %s",
					tt.className, tt.hasBioC, got, tt.wantTier)
			}
		})
	}
}

func TestBuildFileHasBioCMap(t *testing.T) {
	report := &scan.ScanReport{
		Results: []scan.ScanResult{
			{FilePath: "file_a.pcc", HasBioConversation: true},
			{FilePath: "file_b.pcc", HasBioConversation: false},
		},
	}

	m := BuildFileHasBioCMap(report)
	if !m["file_a.pcc"] {
		t.Error("file_a.pcc should have BioConversation")
	}
	if m["file_b.pcc"] {
		t.Error("file_b.pcc should NOT have BioConversation")
	}
}

func TestBuildFileHasBioCMap_FromBioConversationFiles(t *testing.T) {
	report := &scan.ScanReport{
		BioConversationFiles: []string{"file_c.pcc", "file_d.pcc"},
	}

	m := BuildFileHasBioCMap(report)
	if !m["file_c.pcc"] {
		t.Error("file_c.pcc should have BioConversation from BioConversationFiles")
	}
	if !m["file_d.pcc"] {
		t.Error("file_d.pcc should have BioConversation from BioConversationFiles")
	}
}

func TestBuildFileHasBioCMap_BothSources(t *testing.T) {
	report := &scan.ScanReport{
		Results: []scan.ScanResult{
			{FilePath: "file_a.pcc", HasBioConversation: true},
		},
		BioConversationFiles: []string{"file_b.pcc"},
	}

	m := BuildFileHasBioCMap(report)
	if !m["file_a.pcc"] {
		t.Error("file_a.pcc should have BioConversation from Results")
	}
	if !m["file_b.pcc"] {
		t.Error("file_b.pcc should have BioConversation from BioConversationFiles")
	}
}

func TestBuildReport(t *testing.T) {
	scanReport := &scan.ScanReport{
		FilesScanned:  2,
		FilesWithHits: 1,
		TotalHits:     2,
		Results: []scan.ScanResult{
			{
				FilePath:           "test.pcc",
				HasBioConversation: true,
				Hits: []scan.ContainerHit{
					{ContainerHit: pcc.ContainerHit{StrRef: 42, ExportIndex: 0, ClassName: "BioConversation", ExportName: "BioD_Test_Conv1"}},
					{ContainerHit: pcc.ContainerHit{StrRef: 42, ExportIndex: 1, ClassName: "BioPawn", ExportName: "BioPawn_Test"}},
				},
			},
		},
	}

	resolver := &tlk.Resolver{}

	report := BuildReport(
		"test query",
		"test.tlk",
		"",
		"C:\\BioGame",
		[]int32{42},
		scanReport,
		resolver,
	)

	if report.Query != "test query" {
		t.Errorf("query = %q, want %q", report.Query, "test query")
	}
	if report.FilesScanned != 2 {
		t.Errorf("files_scanned = %d, want 2", report.FilesScanned)
	}
	if len(report.CandidateStrRefs) != 1 {
		t.Errorf("candidate_strrefs = %d, want 1", len(report.CandidateStrRefs))
	}
	if len(report.Evidence) != 1 {
		t.Fatalf("evidence = %d, want 1", len(report.Evidence))
	}

	ev := report.Evidence[0]
	if ev.StrRef != 42 {
		t.Errorf("evidence[0].strref = %d, want 42", ev.StrRef)
	}
	if len(ev.BioConversation) != 1 {
		t.Errorf("bioconversation hits = %d, want 1", len(ev.BioConversation))
	}
	if len(ev.SemanticContainer) != 1 {
		t.Errorf("semantic_container hits = %d, want 1", len(ev.SemanticContainer))
	}
	if len(ev.ContainerFallback) != 0 {
		t.Errorf("container_fallback hits = %d, want 0", len(ev.ContainerFallback))
	}

	if ev.BioConversation[0].ExportIndex != 0 {
		t.Errorf("tier1 export_index = %d, want 0", ev.BioConversation[0].ExportIndex)
	}
}

func TestMatchProfile(t *testing.T) {
	matches := MatchProfile("The quarian pilgrimage to the Migrant Fleet")
	if len(matches) == 0 {
		t.Error("expected at least one profile match")
	}
	found := false
	for _, m := range matches {
		if m.Profile == "Quarian Migrant Fleet" {
			found = true
			break
		}
	}
	if !found {
		t.Error("expected 'Quarian Migrant Fleet' profile match")
	}
}

func TestMatchProfileNoMatch(t *testing.T) {
	matches := MatchProfile("xyzzy plugh nothing relevant")
	if len(matches) != 0 {
		t.Errorf("expected no matches, got %d", len(matches))
	}
}

func TestBuildReport_WithNarrativeProfiles(t *testing.T) {
	resolver := &tlk.Resolver{}
	scanReport := &scan.ScanReport{
		FilesScanned:  0,
		FilesWithHits: 0,
		TotalHits:     0,
	}

	report := BuildReport(
		"quarian pilgrimage",
		"test.tlk",
		"",
		"",
		[]int32{42},
		scanReport,
		resolver,
	)

	if len(report.NarrativeProfiles) == 0 {
		t.Error("expected narrative profiles for query with keyword 'quarian'")
	}
}

func TestBuildReport_NoNarrativeMatch(t *testing.T) {
	resolver := &tlk.Resolver{}
	scanReport := &scan.ScanReport{}

	report := BuildReport(
		"xyzzy",
		"test.tlk",
		"",
		"",
		[]int32{99},
		scanReport,
		resolver,
	)

	if len(report.NarrativeProfiles) != 0 {
		t.Errorf("expected no narrative profiles, got %d", len(report.NarrativeProfiles))
	}
}
