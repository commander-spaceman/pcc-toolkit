package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	pcc "github.com/commander-spaceman/me2pcc"
	"pcc-toolkit/core/internal/dialogue"
	"pcc-toolkit/core/internal/dumper"
	"pcc-toolkit/core/internal/editor"
	"pcc-toolkit/core/internal/evidence"
	"pcc-toolkit/core/internal/graph"
	"pcc-toolkit/core/internal/owners"
	"pcc-toolkit/core/internal/scan"
	"pcc-toolkit/core/internal/serialize"
	"pcc-toolkit/core/internal/tlkwrt"

	"github.com/commander-spaceman/me2tlk/reader"
	me2resolver "github.com/commander-spaceman/me2tlk/resolver"
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

func cmdParsePcc(args []string) {
	fs := flag.NewFlagSet("parse-pcc", flag.ExitOnError)
	file := fs.String("file", "", "Path to PCC file")
	exportsOnly := fs.Bool("exports-only", false, "Only show export table")
	exportIndex := fs.Int("export-index", -1, "Show detail for a single export")
	propertyTags := fs.Bool("property-tags", false, "Include property tags for each export")
	semanticProps := fs.Bool("semantic-props", false, "Include parsed properties for each export")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	fs.Parse(args)

	if *file == "" {
		writeError("--file is required", 2)
	}

	needsProperties := *propertyTags || *semanticProps

	if *exportIndex >= 0 {
		cmdExportDetail(*file, *exportIndex, *pretty, *propertyTags, *semanticProps)
		return
	}

	if needsProperties {
		cmdParsePccWithProperties(*file, *exportsOnly, *propertyTags, *semanticProps, *pretty)
		return
	}

	summary, err := pcc.ReadFile(*file)
	if err != nil {
		writeError(fmt.Sprintf("error: %v", err), 1)
	}

	if err := summary.RequireME2(); err != nil {
		writeError(fmt.Sprintf("error: %v", err), 1)
	}

	if *exportsOnly {
		summary.Names = nil
		summary.Imports = nil
	}

	var enc *json.Encoder
	if *pretty {
		enc = json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
	} else {
		enc = json.NewEncoder(os.Stdout)
	}
	if err := enc.Encode(summary); err != nil {
		writeError(fmt.Sprintf("failed to encode output: %v", err), 1)
	}
}

func cmdParsePccWithProperties(file string, exportsOnly, includeTags, includeSemantic, pretty bool) {
	rawData, summary, err := pcc.ReadFileRaw(file)
	if err != nil {
		writeError(fmt.Sprintf("error: %v", err), 1)
	}
	if err := summary.RequireME2(); err != nil {
		writeError(fmt.Sprintf("error: %v", err), 1)
	}

	propMap := pcc.ComputeExportProperties(rawData, summary, includeTags, includeSemantic)

	type exportWithProperties struct {
		Index           int                           `json:"index"`
		ClassIndex      int                           `json:"class_index"`
		ObjectNameIndex int                           `json:"object_name_index"`
		SerialSize      int                           `json:"serial_size"`
		SerialOffset    int                           `json:"serial_offset"`
		ObjectName      string                        `json:"object_name,omitempty"`
		ClassName       string                        `json:"class_name,omitempty"`
		PropertyTags    []pcc.PropertyTag             `json:"property_tags,omitempty"`
		SemanticProps   map[string]pcc.ParsedProperty `json:"semantic_props,omitempty"`
	}

	exportsOut := make([]exportWithProperties, len(summary.Exports))
	for i, exp := range summary.Exports {
		exportsOut[i] = exportWithProperties{
			Index:           exp.Index,
			ClassIndex:      exp.ClassIndex,
			ObjectNameIndex: exp.ObjectNameIndex,
			SerialSize:      exp.SerialSize,
			SerialOffset:    exp.SerialOffset,
			ObjectName:      exp.ObjectName,
			ClassName:       exp.ClassName,
		}
		if ep, ok := propMap[exp.Index]; ok {
			exportsOut[i].PropertyTags = ep.Tags
			exportsOut[i].SemanticProps = ep.SemanticProps
		}
	}

	type outputWithProperties struct {
		File        string                 `json:"file"`
		GameProfile pcc.GameProfile        `json:"game_profile"`
		Compressed  bool                   `json:"compressed"`
		Header      pcc.Header             `json:"header"`
		Names       []string               `json:"names,omitempty"`
		Imports     []pcc.Import           `json:"imports,omitempty"`
		Exports     []exportWithProperties `json:"exports"`
	}

	out := outputWithProperties{
		File:        summary.Path,
		GameProfile: summary.GameProfile,
		Compressed:  summary.Compressed,
		Header:      summary.Header,
		Names:       summary.Names,
		Imports:     summary.Imports,
		Exports:     exportsOut,
	}

	if exportsOnly {
		out.Names = nil
		out.Imports = nil
	}

	var enc *json.Encoder
	if pretty {
		enc = json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
	} else {
		enc = json.NewEncoder(os.Stdout)
	}
	if err := enc.Encode(out); err != nil {
		writeError(fmt.Sprintf("failed to encode output: %v", err), 1)
	}
}

func cmdExportDetail(path string, index int, pretty, includeTags, includeSemantic bool) {
	rawData, summary, err := pcc.ReadFileRaw(path)
	if err != nil {
		writeError(fmt.Sprintf("error: %v", err), 1)
	}

	if err := summary.RequireME2(); err != nil {
		writeError(fmt.Sprintf("error: %v", err), 1)
	}

	if index < 0 || index >= len(summary.Exports) {
		writeError(fmt.Sprintf("export index %d out of range [0, %d)", index, len(summary.Exports)), 2)
	}

	exp := summary.Exports[index]

	type detail struct {
		pcc.Export
		SerialData    string                        `json:"serial_data,omitempty"`
		PropertyTags  []pcc.PropertyTag             `json:"property_tags,omitempty"`
		SemanticProps map[string]pcc.ParsedProperty `json:"semantic_props,omitempty"`
	}

	detailOut := detail{Export: exp}

	start := exp.SerialOffset
	end := exp.SerialOffset + exp.SerialSize
	if start >= 0 && end <= len(rawData) && end > start {
		detailOut.SerialData = base64.StdEncoding.EncodeToString(rawData[start:end])

		if includeTags {
			tags, err := pcc.ParsePropertyTags(rawData, summary.Names, start, exp.SerialSize, false)
			if err == nil {
				detailOut.PropertyTags = tags
			}
		}
		if includeSemantic {
			props, _ := pcc.ParsePropertyCollection(rawData, summary.Names, start, exp.SerialSize)
			if props != nil {
				detailOut.SemanticProps = props
			}
		}
	}

	var enc *json.Encoder
	if pretty {
		enc = json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
	} else {
		enc = json.NewEncoder(os.Stdout)
	}
	if err := enc.Encode(detailOut); err != nil {
		writeError(fmt.Sprintf("failed to encode output: %v", err), 1)
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
		writeError("--file is required", 2)
	}

	tlkFile, err := reader.ReadFile(*file)
	if err != nil {
		writeError(fmt.Sprintf("error: %v", err), 1)
	}

	type parseTlkOutput struct {
		File         string         `json:"file"`
		Header       reader.Header  `json:"header"`
		Entries      []reader.Entry `json:"entries,omitempty"`
		Results      []reader.Entry `json:"results,omitempty"`
		TotalEntries int            `json:"total_entries"`
	}

	out := parseTlkOutput{
		File:   *file,
		Header: tlkFile.Header,
	}

	switch {
	case *strref >= 0:
		text, ok := reader.ResolveString(tlkFile, int32(*strref), true)
		if ok {
			out.Entries = []reader.Entry{{StringID: int32(*strref), Text: text}}
		}
	case *search != "":
		out.Results = tlkFile.Search(*search)
	case *dumpAll:
		tlkFile.IterEntriesWithSource()(func(id int32, text string, source string) bool {
			out.Entries = append(out.Entries, reader.Entry{StringID: id, Text: text, Source: source})
			return true
		})
	default:
		out.Entries = nil
	}

	out.TotalEntries = len(out.Entries)
	if out.TotalEntries == 0 && out.Results != nil {
		out.TotalEntries = len(out.Results)
	}

	var enc *json.Encoder
	if *pretty {
		enc = json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
	} else {
		enc = json.NewEncoder(os.Stdout)
	}
	if err := enc.Encode(out); err != nil {
		writeError(fmt.Sprintf("failed to encode output: %v", err), 1)
	}
}

func cmdResolveTlk(args []string) {
	fs := flag.NewFlagSet("resolve-tlk", flag.ExitOnError)
	base := fs.String("base", "", "Path to base TLK file")
	dlcDir := fs.String("dlc-dir", "", "Path to DLC directory")
	language := fs.String("language", "INT", "TLK language code (INT, DEU, FRA, etc.)")
	strrefFlags := &multiFlag{}
	fs.Var(strrefFlags, "strref", "StringRef to resolve (repeatable)")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	fs.Parse(args)

	if *base == "" {
		writeError("--base is required", 2)
	}
	if len(*strrefFlags) == 0 {
		writeError("at least one --strref is required", 2)
	}

	resolver, err := me2resolver.BuildResolver(*base, *dlcDir, *language, false)
	if err != nil {
		writeError(fmt.Sprintf("error: %v", err), 1)
	}

	type resolveTlkOutput struct {
		Base    string                      `json:"base"`
		DlcDir  string                      `json:"dlc_dir,omitempty"`
		Results []me2resolver.ResolveResult `json:"results"`
	}

	out := resolveTlkOutput{
		Base:   *base,
		DlcDir: *dlcDir,
	}

	for _, raw := range *strrefFlags {
		var id int
		if _, err := fmt.Sscanf(raw, "%d", &id); err != nil {
			out.Results = append(out.Results, me2resolver.ResolveResult{StringID: 0, Text: "", Found: false})
			continue
		}
		refID := int32(id)
		result := resolver.ResolveWithSource(refID)
		out.Results = append(out.Results, *result)
	}

	var enc *json.Encoder
	if *pretty {
		enc = json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
	} else {
		enc = json.NewEncoder(os.Stdout)
	}
	if err := enc.Encode(out); err != nil {
		writeError(fmt.Sprintf("failed to encode output: %v", err), 1)
	}
}

func cmdParseConversations(args []string) {
	fs := flag.NewFlagSet("parse-conversations", flag.ExitOnError)
	file := fs.String("file", "", "Path to PCC file")
	convIndex := fs.Int("conv-index", -1, "Parse a single conversation by export index")
	resolveTlk := fs.String("resolve-tlk", "", "Path to TLK file for text resolution")
	dlcDir := fs.String("dlc-dir", "", "DLC directory for TLK overrides")
	language := fs.String("language", "INT", "TLK language code (INT, DEU, FRA, etc.)")
	mode := fs.String("mode", "resilient", "Parse mode: resilient or strict")
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

	var result *dialogue.ParseResult
	if *convIndex >= 0 {
		conv, err := dialogue.ParseConversation(summary, rawData, *convIndex)
		if err != nil {
			writeError(fmt.Sprintf("error: %v", err), 1)
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
		writeError(fmt.Sprintf("failed to encode output: %v", err), 1)
	}
}

func resolveConversationTLK(conv *dialogue.Conversation, resolver *me2resolver.Resolver) {
	for i := range conv.Entries {
		if conv.Entries[i].LineStrRef != nil {
			text, ok := resolver.Resolve(int32(*conv.Entries[i].LineStrRef))
			if ok {
				conv.Entries[i].LineText = text
				conv.Entries[i].LineStatus = "resolved"
			}
		}
	}
	for i := range conv.Replies {
		if conv.Replies[i].LineStrRef != nil {
			text, ok := resolver.Resolve(int32(*conv.Replies[i].LineStrRef))
			if ok {
				conv.Replies[i].LineText = text
				conv.Replies[i].LineStatus = "resolved"
			}
		}
	}
	for i := range conv.Entries {
		for j := range conv.Entries[i].ReplyChoices {
			if conv.Entries[i].ReplyChoices[j].ParaphraseStrRef != nil {
				text, ok := resolver.Resolve(int32(*conv.Entries[i].ReplyChoices[j].ParaphraseStrRef))
				if ok {
					conv.Entries[i].ReplyChoices[j].ParaphraseText = text
				}
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

func cmdScanEvidence(args []string) {
	fs := flag.NewFlagSet("scan-evidence", flag.ExitOnError)
	query := fs.String("query", "", "Search query text")
	tlkPath := fs.String("tlk", "", "Path to base TLK file")
	dlcDir := fs.String("dlc-dir", "", "DLC directory for TLK overrides")
	language := fs.String("language", "INT", "TLK language code (INT, DEU, FRA, etc.)")
	bioGameRoot := fs.String("biogame-root", "", "BioGame root directory for PCC scanning")
	cachePath := fs.String("cache", "", "Path to file cache JSON (default: none)")
	workers := fs.Int("workers", 0, "Number of concurrent workers (default: CPU count)")
	pretty := fs.Bool("pretty", false, "Pretty-print JSON output")

	fs.Parse(args)

	if *query == "" {
		writeError("--query is required", 2)
	}
	if *tlkPath == "" {
		writeError("--tlk is required", 2)
	}

	resolver, err := me2resolver.BuildResolver(*tlkPath, *dlcDir, *language, false)
	if err != nil {
		writeError(fmt.Sprintf("tlk resolver error: %v", err), 1)
	}

	candidateResults := resolver.Search(*query)
	if len(candidateResults) == 0 {
		out := evidence.EvidenceReport{
			Query:            *query,
			TlkPath:          *tlkPath,
			DlcDir:           *dlcDir,
			BioGameRoot:      *bioGameRoot,
			CandidateStrRefs: []int{},
			Evidence:         []evidence.StrRefEvidence{},
		}
		var enc *json.Encoder
		if *pretty {
			enc = json.NewEncoder(os.Stdout)
			enc.SetIndent("", "  ")
		} else {
			enc = json.NewEncoder(os.Stdout)
		}
		enc.Encode(out)
		return
	}

	candidates := make([]int32, 0, len(candidateResults))
	seen := make(map[int32]bool)
	for _, r := range candidateResults {
		if !seen[r.StringID] {
			seen[r.StringID] = true
			candidates = append(candidates, r.StringID)
		}
	}

	var scanReport *scan.ScanReport
	if *bioGameRoot != "" {
		files, err := scan.CollectPccFiles(*bioGameRoot)
		if err != nil {
			writeError(fmt.Sprintf("file collection error: %v", err), 1)
		}
		if *workers <= 0 {
			*workers = runtime.NumCPU()
		}

		if *cachePath != "" {
			fileCache, cacheErr := scan.LoadFileCache(*cachePath)
			if cacheErr == nil {
				scanReport = scan.RunWithCache(files, candidates, *workers, fileCache)
				_ = fileCache.Save(*cachePath)
			} else {
				scanReport = scan.Run(files, candidates, *workers)
			}
		} else {
			scanReport = scan.Run(files, candidates, *workers)
		}
	} else {
		scanReport = &scan.ScanReport{}
	}

	report := evidence.BuildReport(
		*query,
		*tlkPath,
		*dlcDir,
		*bioGameRoot,
		candidates,
		scanReport,
		resolver,
	)

	evidence.EnrichConversationMatchesWithAST(report)

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

func validationFailed(invalid, warning int, strict bool) bool {
	if invalid > 0 {
		return true
	}
	return strict && warning > 0
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

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
}

func cmdEditConversation(args []string) {
	fs := flag.NewFlagSet("edit-conversation", flag.ExitOnError)
	file := fs.String("file", "", "Path to PCC file")
	convIndex := fs.Int("conv-index", -1, "Export index of the conversation to edit")
	output := fs.String("output", "", "Path for the output PCC file")
	patchFile := fs.String("patch", "", "Path to JSON patch file")
	dryRun := fs.Bool("dry-run", false, "Validate without writing output")
	tlkPath := fs.String("tlk", "", "Path to TLK file for text resolution/additions")
	tlkOutput := fs.String("tlk-output", "", "Path for the output TLK file")
	backup := fs.Bool("backup", false, "Create a .bak backup of the original PCC before editing")

	fs.Parse(args)

	if *file == "" {
		writeError("--file is required", 2)
	}
	if *convIndex < 0 {
		writeError("--conv-index is required", 2)
	}
	if *patchFile == "" {
		writeError("--patch is required", 2)
	}
	if !*dryRun && *output == "" {
		writeError("--output is required (or use --dry-run)", 2)
	}

	patchData, err := os.ReadFile(*patchFile)
	if err != nil {
		writeError(fmt.Sprintf("read patch file: %v", err), 1)
	}

	var patch conversationPatch
	if err := json.Unmarshal(patchData, &patch); err != nil {
		writeError(fmt.Sprintf("parse patch JSON: %v", err), 1)
	}

	if *tlkPath != "" {
		tlkFile, err := reader.ReadFile(*tlkPath)
		if err != nil {
			writeError(fmt.Sprintf("read TLK: %v", err), 1)
		}
		if err := resolveTextToStrRefs(&patch, tlkFile); err != nil {
			writeError(fmt.Sprintf("resolve TLK text: %v", err), 1)
		}
		if *tlkOutput != "" {
			buf, err := tlkwrt.WriteFileBytes(tlkFile)
			if err != nil {
				writeError(fmt.Sprintf("build TLK bytes: %v", err), 1)
			}
			if err := os.WriteFile(*tlkOutput, buf, 0644); err != nil {
				writeError(fmt.Sprintf("write TLK: %v", err), 1)
			}
		}
	}

	if *backup && *file != "" {
		backupPath := *file + ".bak"
		src, err := os.ReadFile(*file)
		if err != nil {
			writeError(fmt.Sprintf("read file for backup: %v", err), 1)
		}
		if err := os.WriteFile(backupPath, src, 0644); err != nil {
			writeError(fmt.Sprintf("write backup: %v", err), 1)
		}
	}

	outPath := *output
	if *dryRun {
		outPath = ""
	}

	modifyFn := func(conv *dialogue.Conversation) error {
		return applyPatch(conv, &patch)
	}

	editResult, err := editor.EditConversation(*file, outPath, *convIndex, *dryRun, modifyFn)
	if err != nil {
		writeError(fmt.Sprintf("edit failed: %v", err), 1)
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(editResult); err != nil {
		writeError(fmt.Sprintf("failed to encode output: %v", err), 1)
	}
}

type conversationPatch struct {
	AddEntries      []entryPatch       `json:"add_entries"`
	AddReplies      []replyPatch       `json:"add_replies"`
	ModifyEntries   []modifyEntryPatch `json:"modify_entries"`
	ModifyReplies   []modifyReplyPatch `json:"modify_replies"`
	DeleteEntries   []int              `json:"delete_entries"`
	DeleteReplies   []int              `json:"delete_replies"`
	AddReplyChoices []replyChoicePatch `json:"add_reply_choices"`
	SetStarts       []startPatch       `json:"set_starts"`
}

type entryPatch struct {
	SpeakerID   *int   `json:"speaker_id"`
	LineStrRef  *int   `json:"line_strref"`
	Text        string `json:"text"`
	ReplyLinks  []int  `json:"reply_links"`
	ListenerTag string `json:"listener_tag"`
}

type replyPatch struct {
	LineStrRef     *int   `json:"line_strref"`
	Text           string `json:"text"`
	TargetEntryIDs []int  `json:"target_entry_ids"`
	Category       string `json:"category"`
	ReplyType      string `json:"reply_type"`
}

type modifyEntryPatch struct {
	ID         int   `json:"id"`
	SpeakerID  *int  `json:"speaker_id"`
	LineStrRef *int  `json:"line_strref"`
	ReplyLinks []int `json:"reply_links"`
}

type modifyReplyPatch struct {
	ID             int    `json:"id"`
	LineStrRef     *int   `json:"line_strref"`
	TargetEntryIDs []int  `json:"target_entry_ids"`
	Category       string `json:"category"`
	ReplyType      string `json:"reply_type"`
}

type replyChoicePatch struct {
	FromEntryID      int    `json:"from_entry_id"`
	ToReplyID        int    `json:"to_reply_id"`
	ParaphraseStrRef *int   `json:"paraphrase_strref"`
	Paraphrase       string `json:"paraphrase"`
	Category         string `json:"category"`
}

type startPatch struct {
	TargetEntryIDs []int  `json:"target_entry_ids"`
	Label          string `json:"label"`
}

func applyPatch(conv *dialogue.Conversation, patch *conversationPatch) error {
	if err := applyDeleteReplies(conv, patch.DeleteReplies); err != nil {
		return err
	}
	if err := applyDeleteEntries(conv, patch.DeleteEntries); err != nil {
		return err
	}
	if err := applyModifyEntries(conv, patch.ModifyEntries); err != nil {
		return err
	}
	if err := applyModifyReplies(conv, patch.ModifyReplies); err != nil {
		return err
	}
	if err := applyAddReplyChoices(conv, patch.AddReplyChoices); err != nil {
		return err
	}
	if err := applyAddEntries(conv, patch.AddEntries); err != nil {
		return err
	}
	if err := applyAddReplies(conv, patch.AddReplies); err != nil {
		return err
	}
	if err := applySetStarts(conv, patch.SetStarts); err != nil {
		return err
	}
	return nil
}

func applyAddEntries(conv *dialogue.Conversation, patches []entryPatch) error {
	entryBase := len(conv.Entries)
	for i, ep := range patches {
		speakerID := 0
		if ep.SpeakerID != nil {
			speakerID = *ep.SpeakerID
		}
		lineStrRef := -1
		if ep.LineStrRef != nil {
			lineStrRef = *ep.LineStrRef
		}

		entry := dialogue.EntryNode{
			ID:         entryBase + i,
			SpeakerID:  &speakerID,
			LineStrRef: &lineStrRef,
		}

		if len(ep.ReplyLinks) > 0 {
			replyBase := len(conv.Replies)
			entry.ReplyLinks = make([]int, len(ep.ReplyLinks))
			for j, replyIdx := range ep.ReplyLinks {
				actualIdx := replyBase + replyIdx
				entry.ReplyLinks[j] = actualIdx
				entry.ReplyChoices = append(entry.ReplyChoices, dialogue.ReplyChoice{
					FromEntryID: entry.ID,
					ToReplyID:   actualIdx,
					Order:       j,
				})
			}
		}

		conv.Entries = append(conv.Entries, entry)
	}
	return nil
}

func applyAddReplies(conv *dialogue.Conversation, patches []replyPatch) error {
	replyBase := len(conv.Replies)
	for i, rp := range patches {
		lineStrRef := -1
		if rp.LineStrRef != nil {
			lineStrRef = *rp.LineStrRef
		}
		category := rp.Category
		if category == "" {
			category = "REPLY_CATEGORY_DEFAULT"
		}
		replyType := rp.ReplyType
		if replyType == "" {
			replyType = "REPLY_TYPE_DEFAULT"
		}

		reply := dialogue.ReplyNode{
			ID:             replyBase + i,
			LineStrRef:     &lineStrRef,
			TargetEntryIDs: rp.TargetEntryIDs,
			Category:       category,
			ReplyType:      replyType,
		}

		conv.Replies = append(conv.Replies, reply)
	}
	return nil
}

func applyModifyEntries(conv *dialogue.Conversation, patches []modifyEntryPatch) error {
	for _, mp := range patches {
		found := false
		for i := range conv.Entries {
			if conv.Entries[i].ID != mp.ID {
				continue
			}
			found = true
			if mp.SpeakerID != nil {
				v := *mp.SpeakerID
				conv.Entries[i].SpeakerID = &v
			}
			if mp.LineStrRef != nil {
				v := *mp.LineStrRef
				conv.Entries[i].LineStrRef = &v
			}
			if mp.ReplyLinks != nil {
				conv.Entries[i].ReplyLinks = mp.ReplyLinks
				conv.Entries[i].ReplyChoices = nil
				for j, replyIdx := range mp.ReplyLinks {
					conv.Entries[i].ReplyChoices = append(conv.Entries[i].ReplyChoices, dialogue.ReplyChoice{
						FromEntryID: mp.ID,
						ToReplyID:   replyIdx,
						Order:       j,
					})
				}
			}
			break
		}
		if !found {
			return fmt.Errorf("entry %d not found for modify", mp.ID)
		}
	}
	return nil
}

func applyModifyReplies(conv *dialogue.Conversation, patches []modifyReplyPatch) error {
	for _, mp := range patches {
		found := false
		for i := range conv.Replies {
			if conv.Replies[i].ID != mp.ID {
				continue
			}
			found = true
			if mp.LineStrRef != nil {
				v := *mp.LineStrRef
				conv.Replies[i].LineStrRef = &v
			}
			if mp.TargetEntryIDs != nil {
				conv.Replies[i].TargetEntryIDs = mp.TargetEntryIDs
			}
			if mp.Category != "" {
				conv.Replies[i].Category = mp.Category
			}
			if mp.ReplyType != "" {
				conv.Replies[i].ReplyType = mp.ReplyType
			}
			break
		}
		if !found {
			return fmt.Errorf("reply %d not found for modify", mp.ID)
		}
	}
	return nil
}

func applyDeleteEntries(conv *dialogue.Conversation, ids []int) error {
	for _, delID := range ids {
		idx := -1
		for i := range conv.Entries {
			if conv.Entries[i].ID == delID {
				idx = i
				break
			}
		}
		if idx < 0 {
			return fmt.Errorf("entry %d not found for delete", delID)
		}
		conv.Entries = append(conv.Entries[:idx], conv.Entries[idx+1:]...)

		for i := range conv.Starts {
			filtered := conv.Starts[i].TargetEntryIDs[:0]
			for _, tid := range conv.Starts[i].TargetEntryIDs {
				if tid != delID {
					filtered = append(filtered, tid)
				}
			}
			conv.Starts[i].TargetEntryIDs = filtered
		}

		for i := range conv.Replies {
			filtered := conv.Replies[i].TargetEntryIDs[:0]
			for _, tid := range conv.Replies[i].TargetEntryIDs {
				if tid != delID {
					filtered = append(filtered, tid)
				}
			}
			conv.Replies[i].TargetEntryIDs = filtered
		}
	}
	return nil
}

func applyDeleteReplies(conv *dialogue.Conversation, ids []int) error {
	for _, delID := range ids {
		idx := -1
		for i := range conv.Replies {
			if conv.Replies[i].ID == delID {
				idx = i
				break
			}
		}
		if idx < 0 {
			return fmt.Errorf("reply %d not found for delete", delID)
		}
		conv.Replies = append(conv.Replies[:idx], conv.Replies[idx+1:]...)

		for i := range conv.Entries {
			filteredLinks := conv.Entries[i].ReplyLinks[:0]
			filteredChoices := conv.Entries[i].ReplyChoices[:0]
			for j, rl := range conv.Entries[i].ReplyLinks {
				if rl != delID {
					filteredLinks = append(filteredLinks, rl)
					if j < len(conv.Entries[i].ReplyChoices) {
						filteredChoices = append(filteredChoices, conv.Entries[i].ReplyChoices[j])
					}
				}
			}
			conv.Entries[i].ReplyLinks = filteredLinks
			conv.Entries[i].ReplyChoices = filteredChoices
		}
	}
	return nil
}

func applyAddReplyChoices(conv *dialogue.Conversation, patches []replyChoicePatch) error {
	for _, rcp := range patches {
		found := false
		for i := range conv.Entries {
			if conv.Entries[i].ID != rcp.FromEntryID {
				continue
			}
			found = true
			conv.Entries[i].ReplyLinks = append(conv.Entries[i].ReplyLinks, rcp.ToReplyID)

			order := len(conv.Entries[i].ReplyChoices)
			rc := dialogue.ReplyChoice{
				FromEntryID: rcp.FromEntryID,
				ToReplyID:   rcp.ToReplyID,
				Order:       order,
				Paraphrase:  rcp.Paraphrase,
				Category:    rcp.Category,
			}
			if rcp.ParaphraseStrRef != nil {
				rc.ParaphraseStrRef = rcp.ParaphraseStrRef
			}
			conv.Entries[i].ReplyChoices = append(conv.Entries[i].ReplyChoices, rc)
			break
		}
		if !found {
			return fmt.Errorf("entry %d not found for add_reply_choice", rcp.FromEntryID)
		}
	}
	return nil
}

func applySetStarts(conv *dialogue.Conversation, patches []startPatch) error {
	if patches == nil {
		return nil
	}
	conv.Starts = make([]dialogue.StartNode, len(patches))
	for i, sp := range patches {
		conv.Starts[i] = dialogue.StartNode{
			ID:             i,
			TargetEntryIDs: sp.TargetEntryIDs,
			Label:          sp.Label,
		}
	}
	return nil
}

func resolveTextToStrRefs(patch *conversationPatch, tlkFile *reader.File) error {
	maxID := int32(0)
	for id := range tlkFile.MaleEntries {
		if id > maxID {
			maxID = id
		}
	}
	for id := range tlkFile.FemaleEntries {
		if id > maxID {
			maxID = id
		}
	}

	var newTLKEntries []tlkwrt.StringEntry
	nextID := int(maxID) + 1

	for i := range patch.AddEntries {
		if patch.AddEntries[i].Text != "" {
			id := nextID
			nextID++
			patch.AddEntries[i].LineStrRef = &id
			newTLKEntries = append(newTLKEntries, tlkwrt.StringEntry{
				StringID: int32(id),
				Text:     patch.AddEntries[i].Text,
				Male:     true,
			})
		}
	}
	for i := range patch.AddReplies {
		if patch.AddReplies[i].Text != "" {
			id := nextID
			nextID++
			patch.AddReplies[i].LineStrRef = &id
			newTLKEntries = append(newTLKEntries, tlkwrt.StringEntry{
				StringID: int32(id),
				Text:     patch.AddReplies[i].Text,
				Male:     true,
			})
		}
	}

	if len(newTLKEntries) > 0 {
		return tlkwrt.AddEntries(tlkFile, newTLKEntries)
	}
	return nil
}

func cmdBatchEdit(args []string) {
	fs := flag.NewFlagSet("batch-edit", flag.ExitOnError)
	dir := fs.String("dir", "", "Directory with PCC files")
	globPat := fs.String("glob", "*.pcc", "Glob pattern for PCC files")
	patchFile := fs.String("patch", "", "Path to JSON patch file")
	outputDir := fs.String("output-dir", "", "Output directory for edited PCCs")
	dryRun := fs.Bool("dry-run", false, "Validate without writing output")

	fs.Parse(args)

	if *dir == "" {
		writeError("--dir is required", 2)
	}
	if *patchFile == "" {
		writeError("--patch is required", 2)
	}

	patchData, err := os.ReadFile(*patchFile)
	if err != nil {
		writeError(fmt.Sprintf("read patch file: %v", err), 1)
	}

	var patch conversationPatch
	if err := json.Unmarshal(patchData, &patch); err != nil {
		writeError(fmt.Sprintf("parse patch JSON: %v", err), 1)
	}

	matches, err := filepath.Glob(filepath.Join(*dir, *globPat))
	if err != nil {
		writeError(fmt.Sprintf("glob: %v", err), 1)
	}
	if len(matches) == 0 {
		writeError(fmt.Sprintf("no files match %s in %s", *globPat, *dir), 1)
	}

	type batchResult struct {
		File       string                            `json:"file"`
		Status     string                            `json:"status"`
		Error      string                            `json:"error,omitempty"`
		Output     string                            `json:"output,omitempty"`
		Validation *dialogue.ValidationReportSummary `json:"validation,omitempty"`
	}

	var results []batchResult
	for _, pccPath := range matches {
		base := filepath.Base(pccPath)
		outPath := ""
		if !*dryRun && *outputDir != "" {
			if err := os.MkdirAll(*outputDir, 0755); err != nil {
				writeError(fmt.Sprintf("create output dir: %v", err), 1)
			}
			outPath = filepath.Join(*outputDir, base)
		}

		modifyFn := func(conv *dialogue.Conversation) error {
			return applyPatch(conv, &patch)
		}
		editResult, err := editor.EditConversation(pccPath, outPath, 0, *dryRun, modifyFn)
		if err != nil {
			results = append(results, batchResult{
				File: base, Status: "error", Error: err.Error(),
			})
			continue
		}
		results = append(results, batchResult{
			File: base, Status: editResult.Status, Output: outPath,
			Validation: editResult.Validation,
		})
	}

	enc := json.NewEncoder(os.Stdout)
	enc.SetIndent("", "  ")
	if err := enc.Encode(results); err != nil {
		writeError(fmt.Sprintf("failed to encode output: %v", err), 1)
	}
}
