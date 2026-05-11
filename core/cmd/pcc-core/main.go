package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
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
