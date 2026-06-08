package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	pcc "github.com/commander-spaceman/me2pcc"
	"pcc-toolkit/core/internal/dialogue"
)

func validationFailed(invalid, warning int, strict bool) bool {
	if invalid > 0 {
		return true
	}
	return strict && warning > 0
}

func cmdValidate(args []string) {
	fs := flag.NewFlagSet("validate", flag.ExitOnError)
	file := fs.String("file", "", "Path to PCC file")
	strict := fs.Bool("strict", false, "Fail on warnings, not just errors")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	fs.Parse(args)

	if *file == "" {
		writeError("--file is required", 2)
	}

	rawData, summary, err := pcc.ReadFileRaw(*file)
	if err != nil {
		writeError(fmt.Sprintf("error: %v", err), 1)
	}

	if err := summary.RequireME2(); err != nil {
		writeError(fmt.Sprintf("error: %v", err), 1)
	}

	result := dialogue.ParseConversations(summary, rawData, "resilient")
	report := dialogue.BuildValidationReportStrict(result, *strict)

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
	if validationFailed(report.Summary.Invalid, report.Summary.Warning, *strict) {
		os.Exit(1)
	}
}

func cmdBatchValidate(args []string) {
	fs := flag.NewFlagSet("batch-validate", flag.ExitOnError)
	dir := fs.String("dir", "", "Directory to scan for PCC files")
	globFlag := fs.String("glob", "*.pcc", "Glob pattern for PCC files")
	output := fs.String("output", "", "Output JSON file path")
	strict := fs.Bool("strict", false, "Fail on warnings")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	fs.Parse(args)

	if *dir == "" {
		writeError("--dir is required", 2)
	}

	pattern := filepath.Join(*dir, *globFlag)
	matches, err := filepath.Glob(pattern)
	if err != nil {
		writeError(fmt.Sprintf("glob error: %v", err), 1)
	}

	type BatchResult struct {
		File    string `json:"file"`
		Total   int    `json:"total"`
		Valid   int    `json:"valid"`
		Warning int    `json:"warning"`
		Invalid int    `json:"invalid"`
		Error   string `json:"error,omitempty"`
	}

	type BatchReport struct {
		Dir        string        `json:"dir"`
		Pattern    string        `json:"pattern"`
		FilesFound int           `json:"files_found"`
		FilesOK    int           `json:"files_ok"`
		FilesError int           `json:"files_error"`
		Total      int           `json:"total_conversations"`
		Valid      int           `json:"valid"`
		Warning    int           `json:"warning"`
		Invalid    int           `json:"invalid"`
		Results    []BatchResult `json:"results"`
	}

	report := BatchReport{
		Dir:     *dir,
		Pattern: *globFlag,
	}

	for _, m := range matches {
		report.FilesFound++
		rawData, summary, err := pcc.ReadFileRaw(m)
		if err != nil {
			report.FilesError++
			report.Results = append(report.Results, BatchResult{
				File:  m,
				Error: err.Error(),
			})
			continue
		}
		if err := summary.RequireME2(); err != nil {
			report.FilesError++
			report.Results = append(report.Results, BatchResult{
				File:  m,
				Error: err.Error(),
			})
			continue
		}

		result := dialogue.ParseConversations(summary, rawData, "resilient")
		valReport := dialogue.BuildValidationReportStrict(result, *strict)

		report.FilesOK++
		report.Total += valReport.Summary.Total
		report.Valid += valReport.Summary.Valid
		report.Warning += valReport.Summary.Warning
		report.Invalid += valReport.Summary.Invalid

		report.Results = append(report.Results, BatchResult{
			File:    m,
			Total:   valReport.Summary.Total,
			Valid:   valReport.Summary.Valid,
			Warning: valReport.Summary.Warning,
			Invalid: valReport.Summary.Invalid,
		})
	}

	var enc *json.Encoder
	var wr *os.File
	if *output != "" {
		wr, err = os.Create(*output)
		if err != nil {
			writeError(fmt.Sprintf("create output: %v", err), 1)
		}
		defer wr.Close()
		if *pretty {
			enc = json.NewEncoder(wr)
			enc.SetIndent("", "  ")
		} else {
			enc = json.NewEncoder(wr)
		}
	} else {
		if *pretty {
			enc = json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
		} else {
			enc = json.NewEncoder(os.Stdout)
		}
	}
	if err := enc.Encode(report); err != nil {
		writeError(fmt.Sprintf("failed to encode output: %v", err), 1)
	}
	if validationFailed(report.Invalid, report.Warning, *strict) {
		os.Exit(1)
	}
}
