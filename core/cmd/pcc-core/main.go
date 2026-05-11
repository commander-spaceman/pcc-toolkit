package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"pcc-toolkit/core/internal/pcc"
)

const version = "0.2.0"
const target = "me2_ot"

var capabilities = []string{
	"pcc_parse_v1",
	"pcc_property_tags_v1",
	"pcc_semantic_props_v1",
	"conversation_ast_v1",
	"graph_layout_v1",
	"tlk_parse_v1",
	"tlk_dlc_resolve_v1",
	"evidence_scan_v1",
	"validate_v1",
	"serialize_v1",
	"batch_validate_v1",
}

type versionOutput struct {
	Version      string   `json:"version"`
	Target       string   `json:"target"`
	Capabilities []string `json:"capabilities"`
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "subcommand required")
		os.Exit(2)
	}

	switch args[0] {
	case "version":
		cmdVersion()
	case "parse-pcc":
		cmdParsePcc(args[1:])
	default:
		fmt.Fprintf(os.Stderr, "unknown subcommand: %s\n", args[0])
		os.Exit(2)
	}
}

func cmdVersion() {
	out := versionOutput{
		Version:      version,
		Target:       target,
		Capabilities: capabilities,
	}
	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode version: %v\n", err)
		os.Exit(1)
	}
}

func cmdParsePcc(args []string) {
	fs := flag.NewFlagSet("parse-pcc", flag.ExitOnError)
	file := fs.String("file", "", "Path to PCC file")
	exportsOnly := fs.Bool("exports-only", false, "Only show export table")
	exportIndex := fs.Int("export-index", -1, "Show detail for a single export")
	propertyTags := fs.Bool("property-tags", false, "Include property tags (not yet implemented)")
	semanticProps := fs.Bool("semantic-props", false, "Include parsed properties (not yet implemented)")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	fs.Parse(args)

	if *file == "" {
		fmt.Fprintln(os.Stderr, "--file is required")
		os.Exit(2)
	}

	if *exportIndex >= 0 {
		cmdExportDetail(*file, *exportIndex, *pretty)
		return
	}

	summary, err := pcc.ReadFile(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if *exportsOnly {
		summary.Names = nil
		summary.Imports = nil
	}

	_ = propertyTags
	_ = semanticProps

	var enc *json.Encoder
	if *pretty {
		enc = json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
	} else {
		enc = json.NewEncoder(os.Stdout)
	}
	if err := enc.Encode(summary); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode output: %v\n", err)
		os.Exit(1)
	}
}

func cmdExportDetail(path string, index int, pretty bool) {
	rawData, summary, err := pcc.ReadFileRaw(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	if index < 0 || index >= len(summary.Exports) {
		fmt.Fprintf(os.Stderr, "export index %d out of range [0, %d)\n", index, len(summary.Exports))
		os.Exit(2)
	}

	exp := summary.Exports[index]

	type detail struct {
		pcc.Export
		SerialData string `json:"serial_data,omitempty"`
	}

	detailOut := detail{Export: exp}

	start := exp.SerialOffset
	end := exp.SerialOffset + exp.SerialSize
	if start >= 0 && end <= len(rawData) && end > start {
		detailOut.SerialData = base64.StdEncoding.EncodeToString(rawData[start:end])
	}

	var enc *json.Encoder
	if pretty {
		enc = json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
	} else {
		enc = json.NewEncoder(os.Stdout)
	}
	if err := enc.Encode(detailOut); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode output: %v\n", err)
		os.Exit(1)
	}
}
