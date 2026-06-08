package evidence

import (
	"runtime"
	"testing"

	me2resolver "github.com/commander-spaceman/me2tlk/resolver"
	"pcc-toolkit/core/internal/scan"
)

const me2BioGameRoot = `C:\Program Files\EA Games\Mass Effect 2\BioGame`
const me2BaseTlk = `C:\Program Files\EA Games\Mass Effect 2\BioGame\CookedPC\BIOGame_INT.tlk`

func TestScanEvidenceFullPipeline(t *testing.T) {
	files, err := scan.CollectPccFiles(me2BioGameRoot)
	if err != nil {
		t.Fatalf("CollectPccFiles: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("no PCC files found")
	}

	const sampleSize = 50
	if len(files) > sampleSize {
		files = files[:sampleSize]
	}
	t.Logf("testing with %d PCC files (sampled from %d collected)", len(files), sampleSize)

	resolver, err := me2resolver.BuildResolver(me2BaseTlk, "", "INT", false)
	if err != nil {
		t.Fatalf("BuildResolver: %v", err)
	}

	candidateResults := resolver.Search("Shepard")
	if len(candidateResults) == 0 {
		t.Fatal("no TLK candidates for 'Shepard'")
	}

	candidates := make([]int32, 0, len(candidateResults))
	seen := make(map[int32]bool)
	for _, r := range candidateResults {
		if !seen[r.StringID] {
			seen[r.StringID] = true
			candidates = append(candidates, r.StringID)
		}
	}
	t.Logf("TLK candidates: %d", len(candidates))

	workers := runtime.NumCPU()
	t.Logf("using %d workers", workers)

	scanReport := scan.Run(files, candidates, workers)
	t.Logf("scanned: %d files, %d with hits, %d total hits",
		scanReport.FilesScanned, scanReport.FilesWithHits, scanReport.TotalHits)

	if scanReport.FilesScanned == 0 {
		t.Fatal("no files scanned")
	}

	for _, e := range scanReport.Errors {
		t.Logf("scan error: %s", e)
	}

	report := BuildReport(
		"Shepard",
		me2BaseTlk,
		"",
		me2BioGameRoot,
		candidates,
		scanReport,
		resolver,
	)

	if report.TotalHits == 0 {
		t.Error("expected hits for 'Shepard' query")
	}
	if report.FilesScanned < 10 {
		t.Errorf("expected >= 10 files scanned, got %d", report.FilesScanned)
	}
	if len(report.CandidateStrRefs) == 0 {
		t.Error("expected candidate strrefs")
	}

	EnrichConversationMatchesWithAST(report)

	t.Logf("evidence: %d files scanned, %d total hits, %d evidence items",
		report.FilesScanned, report.TotalHits, len(report.Evidence))

	for _, ev := range report.Evidence[:min(3, len(report.Evidence))] {
		t.Logf("  strref %d: %d bioconv, %d sem, %d fallback, text=%q",
			ev.StrRef, len(ev.BioConversation), len(ev.SemanticContainer),
			len(ev.ContainerFallback), ev.Text)
	}
}
