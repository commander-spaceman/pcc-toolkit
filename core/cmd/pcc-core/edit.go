package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/commander-spaceman/me2tlk/reader"
	"pcc-toolkit/core/internal/dialogue"
	"pcc-toolkit/core/internal/editor"
	"pcc-toolkit/core/internal/tlkwrt"
)

// conversationPatch describes modifications to a conversation AST.
//
// ID semantics:
//   - add_entries.reply_links: relative offsets (0 = first newly added reply)
//   - modify_entries.reply_links: absolute reply IDs
//   - add_reply_choices.to_reply_id: absolute reply ID
//   - add_replies.target_entry_ids: absolute entry IDs
//   - modify_replies.target_entry_ids: absolute entry IDs
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
	SpeakerID            *int   `json:"speaker_id"`
	LineStrRef           *int   `json:"line_strref"`
	Text                 string `json:"text"`
	ReplyLinks           []int  `json:"reply_links"`
	ListenerTag          string `json:"listener_tag"`
	ConditionalFunc      *int   `json:"conditional_func"`
	ConditionalParam     *int   `json:"conditional_param"`
	StateTransition      *int   `json:"state_transition"`
	StateTransitionParam *int   `json:"state_transition_param"`
	FiresConditional     *bool  `json:"fires_conditional"`
	Skippable            *bool  `json:"skippable"`
}

type replyPatch struct {
	LineStrRef           *int   `json:"line_strref"`
	Text                 string `json:"text"`
	TargetEntryIDs       []int  `json:"target_entry_ids"`
	Category             string `json:"category"`
	ReplyType            string `json:"reply_type"`
	ConditionalFunc      *int   `json:"conditional_func"`
	ConditionalParam     *int   `json:"conditional_param"`
	StateTransition      *int   `json:"state_transition"`
	StateTransitionParam *int   `json:"state_transition_param"`
	FiresConditional     *bool  `json:"fires_conditional"`
	Unskippable          *bool  `json:"unskippable"`
}

type modifyEntryPatch struct {
	ID                   int   `json:"id"`
	SpeakerID            *int  `json:"speaker_id"`
	LineStrRef           *int  `json:"line_strref"`
	ReplyLinks           []int `json:"reply_links"`
	ConditionalFunc      *int  `json:"conditional_func"`
	ConditionalParam     *int  `json:"conditional_param"`
	StateTransition      *int  `json:"state_transition"`
	StateTransitionParam *int  `json:"state_transition_param"`
	FiresConditional     *bool `json:"fires_conditional"`
	Skippable            *bool `json:"skippable"`
}

type modifyReplyPatch struct {
	ID                   int    `json:"id"`
	LineStrRef           *int   `json:"line_strref"`
	TargetEntryIDs       []int  `json:"target_entry_ids"`
	Category             string `json:"category"`
	ReplyType            string `json:"reply_type"`
	ConditionalFunc      *int   `json:"conditional_func"`
	ConditionalParam     *int   `json:"conditional_param"`
	StateTransition      *int   `json:"state_transition"`
	StateTransitionParam *int   `json:"state_transition_param"`
	FiresConditional     *bool  `json:"fires_conditional"`
	Unskippable          *bool  `json:"unskippable"`
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

		if ep.ConditionalFunc != nil {
			v := *ep.ConditionalFunc
			entry.ConditionalFunc = &v
		}
		if ep.ConditionalParam != nil {
			v := *ep.ConditionalParam
			entry.ConditionalParam = &v
		}
		if ep.StateTransition != nil {
			v := *ep.StateTransition
			entry.StateTransition = &v
		}
		if ep.StateTransitionParam != nil {
			v := *ep.StateTransitionParam
			entry.StateTransitionParam = &v
		}
		if ep.FiresConditional != nil {
			v := *ep.FiresConditional
			entry.FiresConditional = &v
		}
		if ep.Skippable != nil {
			v := *ep.Skippable
			entry.Skippable = &v
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

		if rp.ConditionalFunc != nil {
			v := *rp.ConditionalFunc
			reply.ConditionalFunc = &v
		}
		if rp.ConditionalParam != nil {
			v := *rp.ConditionalParam
			reply.ConditionalParam = &v
		}
		if rp.StateTransition != nil {
			v := *rp.StateTransition
			reply.StateTransition = &v
		}
		if rp.StateTransitionParam != nil {
			v := *rp.StateTransitionParam
			reply.StateTransitionParam = &v
		}
		if rp.FiresConditional != nil {
			v := *rp.FiresConditional
			reply.FiresConditional = &v
		}
		if rp.Unskippable != nil {
			v := *rp.Unskippable
			reply.Unskippable = &v
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
			if mp.ConditionalFunc != nil {
				v := *mp.ConditionalFunc
				conv.Entries[i].ConditionalFunc = &v
			}
			if mp.ConditionalParam != nil {
				v := *mp.ConditionalParam
				conv.Entries[i].ConditionalParam = &v
			}
			if mp.StateTransition != nil {
				v := *mp.StateTransition
				conv.Entries[i].StateTransition = &v
			}
			if mp.StateTransitionParam != nil {
				v := *mp.StateTransitionParam
				conv.Entries[i].StateTransitionParam = &v
			}
			if mp.FiresConditional != nil {
				v := *mp.FiresConditional
				conv.Entries[i].FiresConditional = &v
			}
			if mp.Skippable != nil {
				v := *mp.Skippable
				conv.Entries[i].Skippable = &v
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
			if mp.ConditionalFunc != nil {
				v := *mp.ConditionalFunc
				conv.Replies[i].ConditionalFunc = &v
			}
			if mp.ConditionalParam != nil {
				v := *mp.ConditionalParam
				conv.Replies[i].ConditionalParam = &v
			}
			if mp.StateTransition != nil {
				v := *mp.StateTransition
				conv.Replies[i].StateTransition = &v
			}
			if mp.StateTransitionParam != nil {
				v := *mp.StateTransitionParam
				conv.Replies[i].StateTransitionParam = &v
			}
			if mp.FiresConditional != nil {
				v := *mp.FiresConditional
				conv.Replies[i].FiresConditional = &v
			}
			if mp.Unskippable != nil {
				v := *mp.Unskippable
				conv.Replies[i].Unskippable = &v
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

			// Avoid duplicating a reply link that already exists.
			alreadyLinked := false
			for _, rl := range conv.Entries[i].ReplyLinks {
				if rl == rcp.ToReplyID {
					alreadyLinked = true
					break
				}
			}
			if alreadyLinked {
				break
			}

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
	tlkPath := fs.String("tlk", "", "Path to TLK file for text resolution/additions")
	tlkOutput := fs.String("tlk-output", "", "Path for the output TLK file")
	convIndex := fs.Int("conv-index", 0, "Conversation index within each PCC (0-based)")

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
		editResult, err := editor.EditConversation(pccPath, outPath, *convIndex, *dryRun, modifyFn)
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
