package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"runtime"

	me2resolver "github.com/commander-spaceman/me2tlk/resolver"
	"pcc-toolkit/core/internal/evidence"
	"pcc-toolkit/core/internal/scan"
)

func cmdScanEvidence(args []string) {
	fs := flag.NewFlagSet("scan-evidence", flag.ExitOnError)
	query := fs.String("query", "", "Search query text")
	tlkPath := fs.String("tlk", "", "Path to base TLK file")
	dlcDir := fs.String("dlc-dir", "", "DLC directory for TLK overrides")
	language := fs.String("language", "INT", "TLK language code (INT, DEU, FRA, etc.)")
	bioGameRoot := fs.String("biogame-root", "", "BioGame root directory for PCC scanning")
	cachePath := fs.String("cache", "", "Path to file cache JSON (default: none)")
	workers := fs.Int("workers", 0, "Number of concurrent workers (default: CPU count)")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	fs.Parse(args)

	if *query == "" {
		writeError("--query is required", 2)
	}
	if *tlkPath == "" {
		writeError("--tlk is required", 2)
	}

	resolver, err := me2resolver.BuildResolver(*tlkPath, *dlcDir, *language, false)
	if err != nil {
		writeError(fmt.Sprintf("tlk resolver error: %v", err), 1)
	}

	candidateResults := resolver.Search(*query)
	if len(candidateResults) == 0 {
		out := evidence.EvidenceReport{
			Query:            *query,
			TlkPath:          *tlkPath,
			DlcDir:           *dlcDir,
			BioGameRoot:      *bioGameRoot,
			CandidateStrRefs: []int{},
			Evidence:         []evidence.StrRefEvidence{},
		}
		var enc *json.Encoder
		if *pretty {
			enc = json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
		} else {
			enc = json.NewEncoder(os.Stdout)
		}
		enc.Encode(out)
		return
	}

	candidates := make([]int32, 0, len(candidateResults))
	seen := make(map[int32]bool)
	for _, r := range candidateResults {
		if !seen[r.StringID] {
			seen[r.StringID] = true
			candidates = append(candidates, r.StringID)
		}
	}

	var scanReport *scan.ScanReport
	if *bioGameRoot != "" {
		files, err := scan.CollectPccFiles(*bioGameRoot)
		if err != nil {
			writeError(fmt.Sprintf("file collection error: %v", err), 1)
		}
		if *workers <= 0 {
			*workers = runtime.NumCPU()
		}

		if *cachePath != "" {
			fileCache, cacheErr := scan.LoadFileCache(*cachePath)
			if cacheErr == nil {
				scanReport = scan.RunWithCache(files, candidates, *workers, fileCache)
				_ = fileCache.Save(*cachePath)
			} else {
				scanReport = scan.Run(files, candidates, *workers)
			}
		} else {
			scanReport = scan.Run(files, candidates, *workers)
		}
	} else {
		scanReport = &scan.ScanReport{}
	}

	report := evidence.BuildReport(
		*query,
		*tlkPath,
		*dlcDir,
		*bioGameRoot,
		candidates,
		scanReport,
		resolver,
	)

	evidence.EnrichConversationMatchesWithAST(report)

	var enc *json.Encoder
	if *pretty {
		enc = json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
	} else {
		enc = json.NewEncoder(os.Stdout)
	}
	if err := enc.Encode(report); err != nil {
		writeError(fmt.Sprintf("failed to encode output: %v", err), 1)
	}
}
