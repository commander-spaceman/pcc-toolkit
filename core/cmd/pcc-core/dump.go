package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	pcc "github.com/commander-spaceman/me2pcc"
	me2resolver "github.com/commander-spaceman/me2tlk/resolver"
	"pcc-toolkit/core/internal/dialogue"
	"pcc-toolkit/core/internal/dumper"
)

func cmdDumpLines(args []string) {
	fs := flag.NewFlagSet("dump-lines", flag.ExitOnError)
	file := fs.String("file", "", "Path to PCC file")
	resolveTlk := fs.String("resolve-tlk", "", "Path to TLK for text resolution")
	dlcDir := fs.String("dlc-dir", "", "DLC directory for TLK overrides")
	language := fs.String("language", "INT", "TLK language code")
	format := fs.String("format", "json", "Output format: json or csv")
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

	var resolver *me2resolver.Resolver
	if *resolveTlk != "" {
		resolver, err = me2resolver.BuildResolver(*resolveTlk, *dlcDir, *language, false)
		if err != nil {
			writeError(fmt.Sprintf("tlk resolver error: %v", err), 1)
		}
	}

	result := dialogue.ParseConversations(summary, rawData, "resilient")

	if resolver != nil {
		for i := range result.Conversations {
			conv := &result.Conversations[i]
			for j := range conv.Entries {
				if conv.Entries[j].LineStrRef != nil {
					text, ok := resolver.Resolve(int32(*conv.Entries[j].LineStrRef))
					if ok {
						conv.Entries[j].LineText = text
						conv.Entries[j].LineStatus = "resolved"
					}
				}
			}
			for j := range conv.Replies {
				if conv.Replies[j].LineStrRef != nil {
					text, ok := resolver.Resolve(int32(*conv.Replies[j].LineStrRef))
					if ok {
						conv.Replies[j].LineText = text
						conv.Replies[j].LineStatus = "resolved"
					}
				}
			}
		}
	}

	output := dumper.BuildDumpLines(result)

	switch *format {
	case "csv":
		fmt.Println("conversation_id,export_index,node_type,node_id,speaker_tag,strref,line_text,line_status,file")
		for _, line := range output.Lines {
			fmt.Printf("%s,%d,%s,%d,%s,%d,\"%s\",%s,%s\n",
				line.ConversationID, line.ExportIndex, line.NodeType, line.NodeID,
				line.SpeakerTag, line.StrRef, line.LineText, line.LineStatus, line.File)
		}
	default:
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
}
