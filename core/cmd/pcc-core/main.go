package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

var version = "0.3.0" // overridable via -ldflags "-X main.version=x.y.z"
var target = "me2_ot"

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
	"dump_lines_v1",
	"scan_owners_v1",
	"edit_conversation_v1",
	"batch_edit_v1",
}

type versionOutput struct {
	Version      string   `json:"version"`
	Target       string   `json:"target"`
	Capabilities []string `json:"capabilities"`
}

func writeError(msg string, exitCode int) {
	errPayload := map[string]string{"error": msg}
	enc := json.NewEncoder(os.Stderr)
	_ = enc.Encode(errPayload)
	os.Exit(exitCode)
}

func main() {
	flag.Parse()

	args := flag.Args()
	if len(args) == 0 {
		writeError("subcommand required", 2)
	}

	switch args[0] {
	case "version":
		cmdVersion()
	case "parse-pcc":
		cmdParsePcc(args[1:])
	case "parse-tlk":
		cmdParseTlk(args[1:])
	case "resolve-tlk":
		cmdResolveTlk(args[1:])
	case "parse-conversations":
		cmdParseConversations(args[1:])
	case "layout-graph":
		cmdLayoutGraph(args[1:])
	case "scan-evidence":
		cmdScanEvidence(args[1:])
	case "validate":
		cmdValidate(args[1:])
	case "serialize":
		cmdSerialize(args[1:])
	case "batch-validate":
		cmdBatchValidate(args[1:])
	case "batch-extract":
		cmdBatchExtract(args[1:])
	case "dump-lines":
		cmdDumpLines(args[1:])
	case "scan-owners":
		cmdScanOwners(args[1:])
	case "edit-conversation":
		cmdEditConversation(args[1:])
	case "batch-edit":
		cmdBatchEdit(args[1:])
	default:
		writeError(fmt.Sprintf("unknown subcommand: %s", args[0]), 2)
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
		writeError(fmt.Sprintf("failed to encode version: %v", err), 1)
	}
}
