package serialize

import (
	"fmt"

	"pcc-toolkit/core/internal/dialogue"
	"pcc-toolkit/core/internal/pcc"
	"pcc-toolkit/core/internal/tlk"
)

type SerializedConversation struct {
	dialogue.Conversation
	Validation *dialogue.ValidationResult `json:"validation,omitempty"`
}

type SerializedOutput struct {
	File          string                            `json:"file"`
	GameProfile   string                            `json:"game_profile"`
	Compressed    bool                              `json:"compressed"`
	Header        pcc.Header                        `json:"header"`
	ExportCount   int                               `json:"export_count"`
	Conversations []SerializedConversation          `json:"conversations"`
	Validation    *dialogue.ValidationReportSummary `json:"validation_summary,omitempty"`
	Errors        []string                          `json:"errors,omitempty"`
}

func Run(
	path string,
	resolveTlk string,
	dlcDir string,
	mode string,
) (*SerializedOutput, error) {
	rawData, summary, err := pcc.ReadFileRaw(path)
	if err != nil {
		return nil, err
	}

	if err := summary.RequireME2(); err != nil {
		return nil, err
	}

	var resolver *tlk.Resolver
	if resolveTlk != "" {
		resolver, err = tlk.BuildResolver(resolveTlk, dlcDir, "INT", false)
		if err != nil {
			return nil, err
		}
	}

	result := dialogue.ParseConversations(summary, rawData, mode)

	if resolver != nil {
		for i := range result.Conversations {
			conv := &result.Conversations[i]
			resolveText(conv, resolver)
		}
	}

	validationReport := dialogue.BuildValidationReport(result)

	output := &SerializedOutput{
		File:        summary.Path,
		GameProfile: string(summary.GameProfile),
		Compressed:  summary.Compressed,
		Header:      summary.Header,
		ExportCount: len(summary.Exports),
		Validation: &dialogue.ValidationReportSummary{
			Total:   validationReport.Summary.Total,
			Valid:   validationReport.Summary.Valid,
			Warning: validationReport.Summary.Warning,
			Invalid: validationReport.Summary.Invalid,
		},
	}

	for _, conv := range result.Conversations {
		sc := SerializedConversation{Conversation: conv}
		for _, vr := range validationReport.Results {
			if vr.ConversationID == conv.ID {
				sc.Validation = &vr
				break
			}
		}
		output.Conversations = append(output.Conversations, sc)
	}

	for _, pe := range result.Errors {
		output.Errors = append(output.Errors, pe.Error)
	}

	return output, nil
}

func resolveText(conv *dialogue.Conversation, resolver *tlk.Resolver) {
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
