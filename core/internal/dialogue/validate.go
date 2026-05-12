package dialogue

type ValidationIssue struct {
	Severity string `json:"severity"`
	NodeType string `json:"node_type,omitempty"`
	NodeID   int    `json:"node_id,omitempty"`
	Message  string `json:"message"`
	Cause    string `json:"cause,omitempty"`
}

type ValidationResult struct {
	ConversationID string             `json:"conversation_id"`
	ExportIndex    int                `json:"export_index"`
	Status         string             `json:"status"`
	Issues         []ValidationIssue  `json:"issues,omitempty"`
	Summary        ValidationSummary  `json:"summary"`
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
	result := &ValidationResult{
		ConversationID: conv.ID,
		ExportIndex:    conv.ExportIndex,
	}
	result.Summary.EntryCount = len(conv.Entries)
	result.Summary.ReplyCount = len(conv.Replies)
	result.Summary.SpeakerCount = len(conv.Speakers)
	result.Summary.StartCount = len(conv.Starts)

	entryIDs := make(map[int]bool)
	entryMaxID := -1
	for _, e := range conv.Entries {
		entryIDs[e.ID] = true
		if e.ID > entryMaxID {
			entryMaxID = e.ID
		}
	}

	replyIDs := make(map[int]bool)
	for _, r := range conv.Replies {
		replyIDs[r.ID] = true
	}

	speakerIDs := make(map[int]bool)
	for _, s := range conv.Speakers {
		speakerIDs[s.ID] = true
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
			result.addIssue(ValidationIssue{
				Severity: "warning",
				NodeType: "entry",
				NodeID:   e.ID,
				Message:  "orphaned entry: entry " + itoa(e.ID) + " is not reachable from any start node",
				Cause:    "entry exists in the data but no dialogue path leads to it — may be unused or conditionally activated content",
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

	if len(conv.Entries) == 0 && len(conv.Replies) > 0 {
		result.addIssue(ValidationIssue{
			Severity: "info",
			Message:  "reply-only conversation: " + itoa(len(conv.Replies)) + " replies, 0 entries",
			Cause:    "this conversation has no entry nodes — common in combat barks and ambient dialogue where NPCs speak without structured back-and-forth",
		})
	}

	if len(conv.Entries) == 0 && len(conv.Replies) == 0 && len(conv.Speakers) == 0 {
		result.addIssue(ValidationIssue{
			Severity: "info",
			Message:  "empty stub: no entries, replies, or speakers",
			Cause:    "placeholder BioConversation export with no dialogue data — common in level transition and ambient master files",
		})
	}

	if conv.ParseMode == "count_or_value_fallback" {
		severity := "warning"
		cause := "the parser could not determine the array layout and fell back to counting elements — entry/reply data may be incomplete"
		if len(conv.Entries) == 0 && len(conv.Replies) == 0 && len(conv.Speakers) == 0 {
			severity = "info"
			cause = "empty conversation stub — fallback mode is expected when there is no data to parse"
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
			result.addIssue(ValidationIssue{
				Severity: severity,
				Message:  w,
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
		for _, r := range conv.Replies {
			for _, tid := range r.TargetEntryIDs {
				if !entryIDs[tid] {
					continue
				}
				if !reached[tid] {
					continue
				}
				entry := findEntry(conv.Entries, tid)
				if entry == nil {
					continue
				}
				for _, rid := range entry.ReplyLinks {
					for _, r2 := range conv.Replies {
						if r2.ID == rid {
							for _, t2 := range r2.TargetEntryIDs {
								if entryIDs[t2] && !reached[t2] {
									reached[t2] = true
									changed = true
								}
							}
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

func BuildValidationReport(result *ParseResult) *ValidationReport {
	report := &ValidationReport{
		File: result.File,
	}

	for _, conv := range result.Conversations {
		vr := ValidateConversation(&conv)
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
