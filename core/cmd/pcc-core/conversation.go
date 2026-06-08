package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"

	pcc "github.com/commander-spaceman/me2pcc"
	me2resolver "github.com/commander-spaceman/me2tlk/resolver"
	"pcc-toolkit/core/internal/dialogue"
)

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
