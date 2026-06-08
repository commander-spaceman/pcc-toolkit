package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	pcc "github.com/commander-spaceman/me2pcc"
	"pcc-toolkit/core/internal/dialogue"
	"pcc-toolkit/core/internal/graph"
)

func cmdLayoutGraph(args []string) {
	fs := flag.NewFlagSet("layout-graph", flag.ExitOnError)
	file := fs.String("file", "", "Path to PCC file")
	convIndex := fs.Int("conv-index", -1, "Conversation export index")
	algorithm := fs.String("algorithm", "sugiyama", "Layout algorithm (sugiyama)")
	nodeWidth := fs.Float64("node-width", 240, "Node width in pixels")
	nodeHeight := fs.Float64("node-height", 64, "Node height in pixels")
	xSpacing := fs.Float64("x-spacing", 80, "Horizontal spacing")
	ySpacing := fs.Float64("y-spacing", 120, "Vertical spacing")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	fs.Parse(args)

	if *file == "" {
		writeError("--file is required", 2)
	}
	if *convIndex < 0 {
		writeError("--conv-index is required", 2)
	}

	rawData, summary, err := pcc.ReadFileRaw(*file)
	if err != nil {
		writeError(fmt.Sprintf("error: %v", err), 1)
	}

	if err := summary.RequireME2(); err != nil {
		writeError(fmt.Sprintf("error: %v", err), 1)
	}

	conv, err := dialogue.ParseConversation(summary, rawData, *convIndex)
	if err != nil {
		writeError(fmt.Sprintf("error: %v", err), 1)
	}

	switch *algorithm {
	case "sugiyama":
	default:
		writeError(fmt.Sprintf("unsupported algorithm: %q (only sugiyama is currently implemented)", *algorithm), 2)
	}
	layout := graph.LayoutConversation(conv, *nodeWidth, *nodeHeight, *xSpacing, *ySpacing)

	var enc *json.Encoder
	if *pretty {
		enc = json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
	} else {
		enc = json.NewEncoder(os.Stdout)
	}
	if err := enc.Encode(layout); err != nil {
		writeError(fmt.Sprintf("failed to encode output: %v", err), 1)
	}
}
