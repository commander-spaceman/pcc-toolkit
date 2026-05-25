package dialogue

type ValidationIssue struct {
	Severity string `json:"severity"`
	NodeType string `json:"node_type,omitempty"`
	NodeID   int    `json:"node_id,omitempty"`
	Message  string `json:"message"`
	Cause    string `json:"cause,omitempty"`
}

type ValidationResult struct {
	ConversationID string            `json:"conversation_id"`
	ExportIndex    int               `json:"export_index"`
	Status         string            `json:"status"`
	Issues         []ValidationIssue `json:"issues,omitempty"`
	Summary        ValidationSummary `json:"summary"`
}

type ValidationSummary struct {
	EntryCount      int `json:"entry_count"`
	ReplyCount      int `json:"reply_count"`
	SpeakerCount    int `json:"speaker_count"`
	StartCount      int `json:"start_count"`
	OrphanedEntries int `json:"orphaned_entries"`
	OrphanedReplies int `json:"orphaned_replies"`
	DanglingLinks   int `json:"dangling_links"`
	IssueCount      int `json:"issue_count"`
}

type ValidationReportSummary struct {
	Total   int `json:"total"`
	Valid   int `json:"valid"`
	Warning int `json:"warning"`
	Invalid int `json:"invalid"`
}

type ValidationReport struct {
	File    string                  `json:"file"`
	Results []ValidationResult      `json:"results"`
	Summary ValidationReportSummary `json:"report_summary"`
}

func ValidateConversation(conv *Conversation) *ValidationResult {
	return validateConversation(conv, false)
}

func ValidateConversationStrict(conv *Conversation) *ValidationResult {
	return validateConversation(conv, true)
}

func validateConversation(conv *Conversation, strict bool) *ValidationResult {
	result := &ValidationResult{
		ConversationID: conv.ID,
		ExportIndex:    conv.ExportIndex,
	}
	result.Summary.EntryCount = len(conv.Entries)
	result.Summary.ReplyCount = len(conv.Replies)
	result.Summary.SpeakerCount = len(conv.Speakers)
	result.Summary.StartCount = len(conv.Starts)

	entryIDs := make(map[int]bool)
	for _, e := range conv.Entries {
		entryIDs[e.ID] = true
	}

	replyIDs := make(map[int]bool)
	for _, r := range conv.Replies {
		replyIDs[r.ID] = true
	}

	speakerIDs := make(map[int]bool)
	hasPlayer := false
	hasOwner := false
	for _, s := range conv.Speakers {
		speakerIDs[s.ID] = true
		if s.ID == -2 && s.Tag == "player" {
			hasPlayer = true
		}
		if s.ID == -1 && s.Tag == "owner" {
			hasOwner = true
		}
	}

	if !hasPlayer && len(conv.Speakers) > 0 {
		severity := "warning"
		if strict {
			severity = "error"
		}
		result.addIssue(ValidationIssue{
			Severity: severity,
			Message:  "missing player speaker: no speaker with id=-2 and tag=\"player\" (ME2 convention)",
			Cause:    "the conversation speaker list is missing the standard player entry — this is required by ME2 OT convention and may cause downstream issues",
		})
	}
	if !hasOwner && len(conv.Speakers) > 0 {
		severity := "warning"
		if strict {
			severity = "error"
		}
		result.addIssue(ValidationIssue{
			Severity: severity,
			Message:  "missing owner speaker: no speaker with id=-1 and tag=\"owner\" (ME2 convention)",
			Cause:    "the conversation speaker list is missing the standard owner entry — this is required by ME2 OT convention and may cause downstream issues",
		})
	}

	for _, e := range conv.Entries {
		for _, rid := range e.ReplyLinks {
			if !replyIDs[rid] {
				result.addIssue(ValidationIssue{
					Severity: "error",
					NodeType: "entry",
					NodeID:   e.ID,
					Message:  "dangling reply link: entry " + itoa(e.ID) + " → reply " + itoa(rid) + " not found",
					Cause:    "entry references a reply that does not exist in ReplyList — conversation data is structurally broken",
				})
				result.Summary.DanglingLinks++
			}
		}

		if e.SpeakerID != nil && *e.SpeakerID >= 0 && !speakerIDs[*e.SpeakerID] {
			result.addIssue(ValidationIssue{
				Severity: "warning",
				NodeType: "entry",
				NodeID:   e.ID,
				Message:  "unregistered speaker: speaker_id " + itoa(*e.SpeakerID) + " not declared in SpeakerList",
				Cause:    "entry references a speaker index that is not defined in this conversation's speaker list",
			})
		}

		if e.LineStrRef != nil && *e.LineStrRef <= 0 {
			result.addIssue(ValidationIssue{
				Severity: "warning",
				NodeType: "entry",
				NodeID:   e.ID,
				Message:  "missing TLK reference: line_strref is " + itoa(*e.LineStrRef),
				Cause:    "strref is zero or negative — this line has no valid TLK text reference in the game data",
			})
		}
	}

	for _, r := range conv.Replies {
		for _, tid := range r.TargetEntryIDs {
			if !entryIDs[tid] {
				if len(conv.Entries) > 0 {
					result.addIssue(ValidationIssue{
						Severity: "error",
						NodeType: "reply",
						NodeID:   r.ID,
						Message:  "broken reply target: reply " + itoa(r.ID) + " → entry " + itoa(tid) + " does not exist",
						Cause:    "reply points to an entry index that is not present in EntryList",
					})
					result.Summary.DanglingLinks++
				}
			}
		}

		if len(r.TargetEntryIDs) == 0 && len(conv.Entries) > 0 {
			result.addIssue(ValidationIssue{
				Severity: "warning",
				NodeType: "reply",
				NodeID:   r.ID,
				Message:  "unlinked reply: no target_entry_id — parser could not extract it from binary data",
				Cause:    "the target entry column in the reply struct could not be read (may be encoded in a non-standard format)",
			})
		}

		if r.LineStrRef != nil && *r.LineStrRef <= 0 {
			result.addIssue(ValidationIssue{
				Severity: "warning",
				NodeType: "reply",
				NodeID:   r.ID,
				Message:  "missing TLK reference: line_strref is " + itoa(*r.LineStrRef),
				Cause:    "strref is zero or negative — this line has no valid TLK text reference in the game data",
			})
		}
	}

	if conv.ParseMode == "struct_property_semantic" || conv.ParseMode == "row_payload" || conv.ParseMode == "row_payload_struct_matrix" || conv.ParseMode == "row_payload_struct_head" {
		validateReplyEntryLinks(conv, result, entryIDs, replyIDs, strict)
	}

	validateReplyChoices(conv, result, replyIDs, entryIDs, strict)

	for _, s := range conv.Starts {
		if len(s.TargetEntryIDs) == 0 {
			result.addIssue(ValidationIssue{
				Severity: "error",
				NodeType: "start",
				NodeID:   s.ID,
				Message:  "broken start node: start " + itoa(s.ID) + " has no target_entry_ids",
				Cause:    "starting list entry is missing its target entry — conversation has no valid entry point",
			})
		}
		for _, tid := range s.TargetEntryIDs {
			if !entryIDs[tid] {
				result.addIssue(ValidationIssue{
					Severity: "error",
					NodeType: "start",
					NodeID:   s.ID,
					Message:  "broken start node: start " + itoa(s.ID) + " → entry " + itoa(tid) + " not found",
					Cause:    "start node points to an entry index that does not exist in EntryList",
				})
			}
		}
	}

	reached := findReachableEntries(conv)
	for _, e := range conv.Entries {
		if !reached[e.ID] {
			cause := "entry exists in data but no dialogue path leads to it — may be unused content, conditional branch, or final line with no player response"
			if len(e.ReplyLinks) == 0 {
				cause = "leaf node — entry has no outgoing reply links (ReplyListNew is empty or absent); this is a terminal line with no player choice"
			}
			severity := "warning"
			if strict {
				severity = "error"
			}
			result.addIssue(ValidationIssue{
				Severity: severity,
				NodeType: "entry",
				NodeID:   e.ID,
				Message:  "orphaned entry: entry " + itoa(e.ID) + " is not reachable from any start node",
				Cause:    cause,
			})
			result.Summary.OrphanedEntries++
		}
	}

	for _, r := range conv.Replies {
		orphaned := len(r.TargetEntryIDs) == 0
		if !orphaned {
			orphaned = true
			for _, tid := range r.TargetEntryIDs {
				if entryIDs[tid] {
					orphaned = false
					break
				}
			}
		}
		if orphaned {
			result.Summary.OrphanedReplies++
		}
	}

	if len(conv.Entries) == 0 && len(conv.Replies) == 0 {
		result.addIssue(ValidationIssue{
			Severity: "info",
			Message: "empty stub: no entries or replies" + func() string {
				if len(conv.Speakers) > 0 && conv.ParseMode != "count_or_value_fallback" {
					return ", " + itoa(len(conv.Speakers)) + " speakers"
				}
				return ""
			}(),
			Cause: "placeholder BioConversation export with no dialogue data — common in level transition and ambient master files",
		})
	}

	if conv.ParseMode == "count_or_value_fallback" {
		severity := "warning"
		cause := "the parser could not determine the array layout and fell back to counting elements — entry/reply data may be incomplete"
		if len(conv.Entries) == 0 && len(conv.Replies) == 0 {
			severity = "info"
			cause = "empty conversation stub — fallback mode is expected when there is no data to parse"
		}
		if strict && severity == "warning" {
			severity = "error"
		}
		result.addIssue(ValidationIssue{
			Severity: severity,
			Message:  "low-confidence parse: " + conv.ParseMode,
			Cause:    cause,
		})
	}

	if len(conv.Warnings) > 0 {
		isEmptyStub := len(conv.Entries) == 0 && len(conv.Replies) == 0 && len(conv.Speakers) == 0
		for _, w := range conv.Warnings {
			severity := "warning"
			if isEmptyStub {
				severity = "info"
			}
			if strict && severity == "warning" {
				severity = "error"
			}
			result.addIssue(ValidationIssue{
				Severity: severity,
				Message:  w,
				Cause:    "parser warning from conversation data",
			})
		}
	}

	result.Summary.IssueCount = len(result.Issues)

	hasError := false
	for _, issue := range result.Issues {
		if issue.Severity == "error" {
			hasError = true
			break
		}
	}

	if hasError {
		result.Status = "invalid"
	} else {
		hasWarning := false
		for _, issue := range result.Issues {
			if issue.Severity == "warning" {
				hasWarning = true
				break
			}
		}
		if hasWarning {
			result.Status = "warning"
		} else {
			result.Status = "valid"
		}
	}

	return result
}

func validateReplyEntryLinks(conv *Conversation, result *ValidationResult, entryIDs, replyIDs map[int]bool, strict bool) {
	replyUsed := make(map[int]bool)
	for _, e := range conv.Entries {
		for _, rid := range e.ReplyLinks {
			if replyIDs[rid] {
				replyUsed[rid] = true
			}
		}
	}

	for _, r := range conv.Replies {
		if len(r.TargetEntryIDs) == 0 {
			continue
		}
		if !replyUsed[r.ID] && len(conv.Entries) > 0 && len(conv.Starts) > 0 {
			severity := "warning"
			if strict {
				severity = "error"
			}
			result.addIssue(ValidationIssue{
				Severity: severity,
				NodeType: "reply",
				NodeID:   r.ID,
				Message:  "unreachable reply: reply " + itoa(r.ID) + " is not referenced by any entry's ReplyLinks",
				Cause:    "this reply has valid entry targets but no entry links to it — the reply exists in ReplyList but is never offered as a player choice",
			})
		}

		for _, tid := range r.TargetEntryIDs {
			if entryIDs[tid] {
				entry := findEntry(conv.Entries, tid)
				if entry != nil {
					found := false
					for _, erid := range entry.ReplyLinks {
						if erid == r.ID {
							found = true
							break
						}
					}
					if !found && tid >= 0 {
						result.addIssue(ValidationIssue{
							Severity: "info",
							NodeType: "reply",
							NodeID:   r.ID,
							Message:  "non-reciprocal link: reply " + itoa(r.ID) + " → entry " + itoa(tid) + ", entry does not link back (may be intentional dialogue chain)",
							Cause:    "the reply targets this entry but the entry's ReplyListNew does not reference this reply — this is normal in a forward dialogue chain where entry A → reply B → entry C (different entries)",
						})
					}
				}
			}
		}
	}
}

func findReachableEntries(conv *Conversation) map[int]bool {
	reached := make(map[int]bool)

	entryIDs := make(map[int]bool)
	for _, e := range conv.Entries {
		entryIDs[e.ID] = true
	}

	for _, s := range conv.Starts {
		for _, tid := range s.TargetEntryIDs {
			if entryIDs[tid] {
				reached[tid] = true
			}
		}
	}

	for changed := true; changed; {
		changed = false
		for _, e := range conv.Entries {
			if !reached[e.ID] {
				continue
			}
			for _, rid := range e.ReplyLinks {
				for _, r := range conv.Replies {
					if r.ID != rid {
						continue
					}
					for _, tid := range r.TargetEntryIDs {
						if entryIDs[tid] && !reached[tid] {
							reached[tid] = true
							changed = true
						}
					}
				}
			}
		}
	}

	return reached
}

func findEntry(entries []EntryNode, id int) *EntryNode {
	for i := range entries {
		if entries[i].ID == id {
			return &entries[i]
		}
	}
	return nil
}

func itoa(i int) string {
	if i < 0 {
		return "-" + uitoa(-i)
	}
	return uitoa(i)
}

func uitoa(n int) string {
	if n == 0 {
		return "0"
	}
	var buf [20]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = byte('0' + n%10)
		n /= 10
	}
	return string(buf[i:])
}

func (r *ValidationResult) addIssue(issue ValidationIssue) {
	r.Issues = append(r.Issues, issue)
}

func validateReplyChoices(conv *Conversation, result *ValidationResult, replyIDs, entryIDs map[int]bool, strict bool) {
	for _, e := range conv.Entries {
		for _, rc := range e.ReplyChoices {
			if rc.ToReplyID >= 0 && !replyIDs[rc.ToReplyID] && len(replyIDs) > 0 {
				result.addIssue(ValidationIssue{
					Severity: "error",
					NodeType: "entry",
					NodeID:   e.ID,
					Message:  "dangling reply choice: entry " + itoa(e.ID) + " ReplyChoice → reply " + itoa(rc.ToReplyID) + " not found",
					Cause:    "ReplyListNew entry references a reply index that does not exist in ReplyList — the ReplyChoice link is structurally broken",
				})
				result.Summary.DanglingLinks++
			}
			if rc.ParaphraseStrRef != nil && *rc.ParaphraseStrRef <= 0 {
				severity := "warning"
				if strict {
					severity = "error"
				}
				result.addIssue(ValidationIssue{
					Severity: severity,
					NodeType: "entry",
					NodeID:   e.ID,
					Message:  "invalid paraphrase strref: entry " + itoa(e.ID) + " ReplyChoice order=" + itoa(rc.Order) + " has srParaphrase=" + itoa(*rc.ParaphraseStrRef),
					Cause:    "the ReplyListNew paraphrase string reference is zero or negative — the paraphrase text will not resolve in TLK",
				})
			}
		}
	}
}

func BuildValidationReport(result *ParseResult) *ValidationReport {
	return BuildValidationReportStrict(result, false)
}

func BuildValidationReportStrict(result *ParseResult, strict bool) *ValidationReport {
	report := &ValidationReport{
		File: result.File,
	}

	for _, conv := range result.Conversations {
		var vr *ValidationResult
		if strict {
			vr = ValidateConversationStrict(&conv)
		} else {
			vr = ValidateConversation(&conv)
		}
		report.Results = append(report.Results, *vr)
		report.Summary.Total++
		switch vr.Status {
		case "valid":
			report.Summary.Valid++
		case "warning":
			report.Summary.Warning++
		case "invalid":
			report.Summary.Invalid++
		}
	}

	return report
}
