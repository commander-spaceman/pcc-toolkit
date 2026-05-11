package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"pcc-toolkit/core/internal/dialogue"
	"pcc-toolkit/core/internal/pcc"
	"pcc-toolkit/core/internal/tlk"
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
	case "parse-tlk":
		cmdParseTlk(args[1:])
	case "resolve-tlk":
		cmdResolveTlk(args[1:])
	case "parse-conversations":
		cmdParseConversations(args[1:])
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

func cmdParseTlk(args []string) {
	fs := flag.NewFlagSet("parse-tlk", flag.ExitOnError)
	file := fs.String("file", "", "Path to TLK file")
	search := fs.String("search", "", "Search query for text")
	strref := fs.Int("strref", -1, "Resolve a single StringRef")
	dumpAll := fs.Bool("dump-all", false, "Dump all entries")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	fs.Parse(args)

	if *file == "" {
		fmt.Fprintln(os.Stderr, "--file is required")
		os.Exit(2)
	}

	tlkFile, err := tlk.ReadFile(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	type parseTlkOutput struct {
		File    string        `json:"file"`
		Header  tlk.Header    `json:"header"`
		Entries []tlk.Entry   `json:"entries,omitempty"`
		Results []tlk.Entry   `json:"results,omitempty"`
	}

	out := parseTlkOutput{
		File:   *file,
		Header: tlkFile.Header,
	}

	switch {
	case *strref >= 0:
		text, ok := tlk.ResolveString(tlkFile, int32(*strref), true)
		if ok {
			out.Entries = []tlk.Entry{{StringID: int32(*strref), Text: text}}
		}
	case *search != "":
		out.Results = tlkFile.Search(*search)
	case *dumpAll:
		tlkFile.IterEntries()(func(id int32, text string) bool {
			out.Entries = append(out.Entries, tlk.Entry{StringID: id, Text: text})
			return true
		})
	default:
		out.Entries = nil
	}

	var enc *json.Encoder
	if *pretty {
		enc = json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
	} else {
		enc = json.NewEncoder(os.Stdout)
	}
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode output: %v\n", err)
		os.Exit(1)
	}
}

func cmdResolveTlk(args []string) {
	fs := flag.NewFlagSet("resolve-tlk", flag.ExitOnError)
	base := fs.String("base", "", "Path to base TLK file")
	dlcDir := fs.String("dlc-dir", "", "Path to DLC directory")
	strrefFlags := &multiFlag{}
	fs.Var(strrefFlags, "strref", "StringRef to resolve (repeatable)")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	fs.Parse(args)

	if *base == "" {
		fmt.Fprintln(os.Stderr, "--base is required")
		os.Exit(2)
	}
	if len(*strrefFlags) == 0 {
		fmt.Fprintln(os.Stderr, "at least one --strref is required")
		os.Exit(2)
	}

	resolver, err := tlk.BuildResolver(*base, *dlcDir, "INT", false)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	type resolveTlkOutput struct {
		Base    string               `json:"base"`
		DlcDir  string               `json:"dlc_dir,omitempty"`
		Results []tlk.ResolveResult  `json:"results"`
	}

	out := resolveTlkOutput{
		Base:   *base,
		DlcDir: *dlcDir,
	}

	for _, raw := range *strrefFlags {
		var id int
		if _, err := fmt.Sscanf(raw, "%d", &id); err != nil {
			out.Results = append(out.Results, tlk.ResolveResult{StringID: 0, Text: ""})
			continue
		}
		result := resolver.ResolveWithSource(int32(id))
		if result == nil {
			out.Results = append(out.Results, tlk.ResolveResult{StringID: int32(id), Text: ""})
		} else {
			out.Results = append(out.Results, *result)
		}
	}

	var enc *json.Encoder
	if *pretty {
		enc = json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
	} else {
		enc = json.NewEncoder(os.Stdout)
	}
	if err := enc.Encode(out); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode output: %v\n", err)
		os.Exit(1)
	}
}

func cmdParseConversations(args []string) {
	fs := flag.NewFlagSet("parse-conversations", flag.ExitOnError)
	file := fs.String("file", "", "Path to PCC file")
	convIndex := fs.Int("conv-index", -1, "Parse a single conversation by export index")
	resolveTlk := fs.String("resolve-tlk", "", "Path to TLK file for text resolution")
	dlcDir := fs.String("dlc-dir", "", "DLC directory for TLK overrides")
	mode := fs.String("mode", "resilient", "Parse mode: resilient or strict")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	fs.Parse(args)

	if *file == "" {
		fmt.Fprintln(os.Stderr, "--file is required")
		os.Exit(2)
	}

	rawData, summary, err := pcc.ReadFileRaw(*file)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	var resolver *tlk.Resolver
	if *resolveTlk != "" {
		resolver, err = tlk.BuildResolver(*resolveTlk, *dlcDir, "INT", false)
		if err != nil {
			fmt.Fprintf(os.Stderr, "tlk resolver error: %v\n", err)
			os.Exit(1)
		}
	}

	var result *dialogue.ParseResult
	if *convIndex >= 0 {
		conv, err := dialogue.ParseConversation(summary, rawData, *convIndex)
		if err != nil {
			fmt.Fprintf(os.Stderr, "error: %v\n", err)
			os.Exit(1)
		}
		if resolver != nil {
			resolveConversationTLK(conv, resolver)
		}
		result = &dialogue.ParseResult{
			File:          *file,
			GameProfile:   string(summary.GameProfile),
			Conversations: []dialogue.Conversation{*conv},
		}
	} else {
		result = dialogue.ParseConversations(summary, rawData, *mode)
		if resolver != nil {
			for i := range result.Conversations {
				resolveConversationTLK(&result.Conversations[i], resolver)
			}
		}
	}

	var enc *json.Encoder
	if *pretty {
		enc = json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
	} else {
		enc = json.NewEncoder(os.Stdout)
	}
	if err := enc.Encode(result); err != nil {
		fmt.Fprintf(os.Stderr, "failed to encode output: %v\n", err)
		os.Exit(1)
	}
}

func resolveConversationTLK(conv *dialogue.Conversation, resolver *tlk.Resolver) {
	for i := range conv.Entries {
		if conv.Entries[i].LineStrRef != nil {
			text, ok := resolver.Resolve(int32(*conv.Entries[i].LineStrRef))
			if ok {
				conv.Entries[i].LineText = text
			}
		}
	}
	for i := range conv.Replies {
		if conv.Replies[i].LineStrRef != nil {
			text, ok := resolver.Resolve(int32(*conv.Replies[i].LineStrRef))
			if ok {
				conv.Replies[i].LineText = text
			}
		}
	}
	for i := range conv.Speakers {
		if conv.Speakers[i].DisplayName != "" && len(conv.Speakers[i].DisplayName) > 7 &&
			conv.Speakers[i].DisplayName[:7] == "strref:" {
			var strref int
			fmt.Sscanf(conv.Speakers[i].DisplayName, "strref:%d", &strref)
			text, ok := resolver.Resolve(int32(strref))
			if ok {
				conv.Speakers[i].DisplayName = text
			}
		}
	}
}

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}
