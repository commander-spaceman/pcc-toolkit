package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/commander-spaceman/me2tlk/reader"
	me2resolver "github.com/commander-spaceman/me2tlk/resolver"
)

type multiFlag []string

func (m *multiFlag) String() string { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error {
	*m = append(*m, v)
	return nil
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
