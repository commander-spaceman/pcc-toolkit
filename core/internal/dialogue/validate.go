package dialogue

type ValidationIssue struct {
	Severity string `json:"severity"`
	NodeType string `json:"node_type,omitempty"`
	NodeID   int    `json:"node_id,omitempty"`
	Message  string `json:"message"`
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
					Message:  "reply_link references non-existent reply " + itoa(rid),
				})
				result.Summary.DanglingLinks++
			}
		}

		if e.SpeakerID != nil && *e.SpeakerID >= 0 && !speakerIDs[*e.SpeakerID] {
			result.addIssue(ValidationIssue{
				Severity: "warning",
				NodeType: "entry",
				NodeID:   e.ID,
				Message:  "speaker_id " + itoa(*e.SpeakerID) + " not in speaker list",
			})
		}

		if e.LineStrRef != nil && *e.LineStrRef <= 0 {
			result.addIssue(ValidationIssue{
				Severity: "warning",
				NodeType: "entry",
				NodeID:   e.ID,
				Message:  "line_strref is zero or negative",
			})
		}
	}

	for _, r := range conv.Replies {
		if r.TargetEntryID != nil && !entryIDs[*r.TargetEntryID] {
			result.addIssue(ValidationIssue{
				Severity: "error",
				NodeType: "reply",
				NodeID:   r.ID,
				Message:  "target_entry_id " + itoa(*r.TargetEntryID) + " does not exist",
			})
			result.Summary.DanglingLinks++
		}

		if r.TargetEntryID == nil {
			result.addIssue(ValidationIssue{
				Severity: "warning",
				NodeType: "reply",
				NodeID:   r.ID,
				Message:  "reply has no target_entry_id (dead end)",
			})
		}

		if r.LineStrRef != nil && *r.LineStrRef <= 0 {
			result.addIssue(ValidationIssue{
				Severity: "warning",
				NodeType: "reply",
				NodeID:   r.ID,
				Message:  "line_strref is zero or negative",
			})
		}
	}

	for _, s := range conv.Starts {
		if s.TargetEntryID == nil {
			result.addIssue(ValidationIssue{
				Severity: "error",
				NodeType: "start",
				NodeID:   s.ID,
				Message:  "start has no target_entry_id",
			})
		} else if !entryIDs[*s.TargetEntryID] {
			result.addIssue(ValidationIssue{
				Severity: "error",
				NodeType: "start",
				NodeID:   s.ID,
				Message:  "start target_entry_id " + itoa(*s.TargetEntryID) + " does not exist",
			})
		}
	}

	reached := findReachableEntries(conv)
	for _, e := range conv.Entries {
		if !reached[e.ID] {
			result.addIssue(ValidationIssue{
				Severity: "warning",
				NodeType: "entry",
				NodeID:   e.ID,
				Message:  "entry is not reachable from any start or reply",
			})
			result.Summary.OrphanedEntries++
		}
	}

	for _, r := range conv.Replies {
		if r.TargetEntryID == nil || !entryIDs[*r.TargetEntryID] {
			result.Summary.OrphanedReplies++
		}
	}

	if conv.ParseMode == "count_or_value_fallback" {
		result.addIssue(ValidationIssue{
			Severity: "warning",
			Message:  "parsed with fallback mode (count_or_value_fallback) — results may be incomplete",
		})
	}

	if len(conv.Warnings) > 0 {
		for _, w := range conv.Warnings {
			result.addIssue(ValidationIssue{
				Severity: "warning",
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
	} else if len(result.Issues) > 0 {
		result.Status = "warning"
	} else {
		result.Status = "valid"
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
		if s.TargetEntryID != nil && entryIDs[*s.TargetEntryID] {
			reached[*s.TargetEntryID] = true
		}
	}

	for changed := true; changed; {
		changed = false
		for _, r := range conv.Replies {
			if r.TargetEntryID == nil || !entryIDs[*r.TargetEntryID] {
				continue
			}
			if !reached[*r.TargetEntryID] {
				continue
			}
			entry := findEntry(conv.Entries, *r.TargetEntryID)
			if entry == nil {
				continue
			}
			for _, rid := range entry.ReplyLinks {
				for _, r2 := range conv.Replies {
					if r2.ID == rid && r2.TargetEntryID != nil && entryIDs[*r2.TargetEntryID] {
						if !reached[*r2.TargetEntryID] {
							reached[*r2.TargetEntryID] = true
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
