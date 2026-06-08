package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	pcc "github.com/commander-spaceman/me2pcc"
	"pcc-toolkit/core/internal/owners"
)

func cmdScanOwners(args []string) {
	fs := flag.NewFlagSet("scan-owners", flag.ExitOnError)
	file := fs.String("file", "", "Path to PCC file")
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

	output := owners.ScanOwners(rawData, summary, *file)

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
