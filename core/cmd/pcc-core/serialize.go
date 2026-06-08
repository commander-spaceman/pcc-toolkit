package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"pcc-toolkit/core/internal/serialize"
)

func cmdSerialize(args []string) {
	fs := flag.NewFlagSet("serialize", flag.ExitOnError)
	file := fs.String("file", "", "Path to PCC file")
	game := fs.String("game", "", "Game profile (me2_ot)")
	resolveTlk := fs.String("resolve-tlk", "", "Path to TLK for text resolution")
	dlcDir := fs.String("dlc-dir", "", "DLC directory for TLK overrides")
	language := fs.String("language", "INT", "TLK language code")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	fs.Parse(args)

	if *file == "" {
		writeError("--file is required", 2)
	}

	_ = game

	output, err := serialize.Run(*file, *resolveTlk, *dlcDir, *language, "resilient")
	if err != nil {
		writeError(fmt.Sprintf("error: %v", err), 1)
	}

	var enc *json.Encoder
	if *pretty {
		enc = json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
	} else {
		enc = json.NewEncoder(os.Stdout)
	}
	if err := enc.Encode(output); err != nil {
		writeError(fmt.Sprintf("failed to encode output: %v", err), 1)
	}
}

func cmdBatchExtract(args []string) {
	fs := flag.NewFlagSet("batch-extract", flag.ExitOnError)
	dir := fs.String("dir", "", "Directory to scan for PCC files")
	globFlag := fs.String("glob", "*.pcc", "Glob pattern for PCC files")
	outputDir := fs.String("output-dir", "", "Output directory for extracted JSON files")
	resolveTlk := fs.String("resolve-tlk", "", "Path to TLK for text resolution")
	dlcDir := fs.String("dlc-dir", "", "DLC directory for TLK overrides")
	language := fs.String("language", "INT", "TLK language code")
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

	if *outputDir != "" {
		if err := os.MkdirAll(*outputDir, 0755); err != nil {
			writeError(fmt.Sprintf("create output dir: %v", err), 1)
		}
	}

	type BatchExtractResult struct {
		File          string `json:"file"`
		Conversations int    `json:"conversations"`
		OutputPath    string `json:"output_path,omitempty"`
		Error         string `json:"error,omitempty"`
	}

	type BatchExtractReport struct {
		Dir        string               `json:"dir"`
		Pattern    string               `json:"pattern"`
		FilesFound int                  `json:"files_found"`
		FilesOK    int                  `json:"files_ok"`
		FilesError int                  `json:"files_error"`
		Results    []BatchExtractResult `json:"results"`
	}

	report := BatchExtractReport{
		Dir:     *dir,
		Pattern: *globFlag,
	}

	for _, m := range matches {
		report.FilesFound++
		output, err := serialize.Run(m, *resolveTlk, *dlcDir, *language, "resilient")
		if err != nil {
			report.FilesError++
			report.Results = append(report.Results, BatchExtractResult{
				File:  m,
				Error: err.Error(),
			})
			continue
		}

		report.FilesOK++
		br := BatchExtractResult{
			File:          m,
			Conversations: len(output.Conversations),
		}

		if *outputDir != "" {
			base := filepath.Base(m)
			ext := filepath.Ext(base)
			outName := base[:len(base)-len(ext)] + ".json"
			outPath := filepath.Join(*outputDir, outName)
			data, _ := json.MarshalIndent(output, "", "  ")
			if err := os.WriteFile(outPath, data, 0644); err != nil {
				br.Error = err.Error()
			}
			br.OutputPath = outPath
		}

		report.Results = append(report.Results, br)
	}

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
