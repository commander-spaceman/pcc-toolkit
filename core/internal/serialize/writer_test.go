package serialize

import (
	"encoding/json"
	"testing"

	"pcc-toolkit/core/internal/dialogue"
)

func TestRun_NonexistentFile(t *testing.T) {
	_, err := Run("nonexistent_file_xyz.pcc", "", "", "INT", "resilient")
	if err == nil {
		t.Fatal("expected error for nonexistent file, got nil")
	}
}

func TestSerializedOutput_JSONMarshal(t *testing.T) {
	output := &SerializedOutput{
		File:        "test.pcc",
		GameProfile: "me2_ot",
		Compressed:  true,
		ExportCount: 42,
		Conversations: []SerializedConversation{
			{
				Conversation: dialogue.Conversation{
					ID:          "test_conv",
					ExportIndex: 1,
					GameProfile: "me2_ot",
					ParseMode:   "struct_property_semantic",
				},
			},
		},
		Errors: []string{"parse warning"},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed SerializedOutput
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.File != "test.pcc" {
		t.Errorf("file mismatch: got %q", parsed.File)
	}
	if parsed.GameProfile != "me2_ot" {
		t.Errorf("game_profile mismatch: got %q", parsed.GameProfile)
	}
	if !parsed.Compressed {
		t.Error("compressed should be true")
	}
	if parsed.ExportCount != 42 {
		t.Errorf("export_count mismatch: got %d", parsed.ExportCount)
	}
	if len(parsed.Conversations) != 1 {
		t.Fatalf("expected 1 conversation, got %d", len(parsed.Conversations))
	}
	if parsed.Conversations[0].ID != "test_conv" {
		t.Errorf("conversation id mismatch: got %q", parsed.Conversations[0].ID)
	}
	if len(parsed.Errors) != 1 || parsed.Errors[0] != "parse warning" {
		t.Errorf("errors mismatch: got %v", parsed.Errors)
	}
}

func TestSerializedConversation_WithValidation(t *testing.T) {
	sc := SerializedConversation{
		Conversation: dialogue.Conversation{
			ID:          "validated_conv",
			ExportIndex: 5,
			ParseMode:   "row_payload",
		},
		Validation: &dialogue.ValidationResult{
			ConversationID: "validated_conv",
			ExportIndex:    5,
			Status:         "valid",
		},
	}

	data, err := json.Marshal(sc)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed SerializedConversation
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.Validation == nil {
		t.Fatal("expected validation to be present")
	}
	if parsed.Validation.Status != "valid" {
		t.Errorf("status mismatch: got %q", parsed.Validation.Status)
	}
}

func TestSerializedConversation_WithoutValidation(t *testing.T) {
	sc := SerializedConversation{
		Conversation: dialogue.Conversation{
			ID:          "no_validation_conv",
			ExportIndex: 3,
		},
	}

	data, err := json.Marshal(sc)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed SerializedConversation
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.Validation != nil {
		t.Error("expected validation to be omitted for nil value")
	}
}

func TestSerializedOutput_RoundTrip(t *testing.T) {
	output := &SerializedOutput{
		File:        "roundtrip.pcc",
		GameProfile: "me2_ot",
		Compressed:  false,
		ExportCount: 10,
		Validation: &dialogue.ValidationReportSummary{
			Total:   3,
			Valid:   2,
			Warning: 0,
			Invalid: 1,
		},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed SerializedOutput
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if parsed.Validation == nil {
		t.Fatal("expected validation_summary to be present")
	}
	if parsed.Validation.Total != 3 || parsed.Validation.Valid != 2 ||
		parsed.Validation.Warning != 0 || parsed.Validation.Invalid != 1 {
		t.Errorf("validation_summary mismatch: %+v", parsed.Validation)
	}
	if parsed.Errors != nil {
		t.Errorf("errors should be nil when empty, got %v", parsed.Errors)
	}
}

func TestSerializedOutput_EmptyConversations(t *testing.T) {
	output := &SerializedOutput{
		File:          "empty.pcc",
		GameProfile:   "me2_ot",
		Conversations: []SerializedConversation{},
	}

	data, err := json.Marshal(output)
	if err != nil {
		t.Fatalf("failed to marshal: %v", err)
	}

	var parsed SerializedOutput
	if err := json.Unmarshal(data, &parsed); err != nil {
		t.Fatalf("failed to unmarshal: %v", err)
	}

	if len(parsed.Conversations) != 0 {
		t.Errorf("expected 0 conversations, got %d", len(parsed.Conversations))
	}
}
