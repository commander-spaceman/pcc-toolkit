package main

import (
	"encoding/base64"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	pcc "github.com/commander-spaceman/me2pcc"
)

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
