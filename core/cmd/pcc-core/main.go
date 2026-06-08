package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
)

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
