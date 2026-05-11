package evidence

import "pcc-toolkit/core/internal/scan"

type EvidenceTier string

const (
	TierBioConversation    EvidenceTier = "bioconversation"
	TierSemanticContainer  EvidenceTier = "semantic_container"
	TierContainerFallback  EvidenceTier = "container_fallback"
)

type TieredHit struct {
	StrRef     int    `json:"strref"`
	Text       string `json:"text,omitempty"`
	FilePath   string `json:"file_path"`
	ExportName string `json:"export_name,omitempty"`
	ClassName  string `json:"class_name,omitempty"`
	Tier       string `json:"tier"`
}

type ConversationMatch struct {
	TieredHit
	ConversationID  string `json:"conversation_id,omitempty"`
	ExportIndex     int    `json:"export_index,omitempty"`
	NodeType        string `json:"node_type,omitempty"`
	SpeakerTag      string `json:"speaker_tag,omitempty"`
	ListenerTag     string `json:"listener_tag,omitempty"`
}

type StrRefEvidence struct {
	StrRef              int                  `json:"strref"`
	Text                string               `json:"text,omitempty"`
	SourceTLK           string               `json:"source_tlk,omitempty"`
	BioConversation     []ConversationMatch  `json:"bioconversation,omitempty"`
	SemanticContainer   []TieredHit          `json:"semantic_container,omitempty"`
	ContainerFallback   []TieredHit          `json:"container_fallback,omitempty"`
}

type EvidenceReport struct {
	Query            string            `json:"query"`
	TlkPath          string            `json:"tlk_path"`
	DlcDir           string            `json:"dlc_dir,omitempty"`
	BioGameRoot      string            `json:"biogame_root,omitempty"`
	CandidateStrRefs []int             `json:"candidate_strrefs"`
	FilesScanned     int               `json:"files_scanned"`
	FilesWithHits    int               `json:"files_with_hits"`
	TotalHits        int               `json:"total_hits"`
	Evidence         []StrRefEvidence  `json:"evidence"`
	Errors           []string          `json:"errors,omitempty"`
}

func ClassifyTier(className string, hasBioConversationInFile bool) EvidenceTier {
	if className == "BioConversation" {
		return TierBioConversation
	}
	if hasBioConversationInFile {
		return TierSemanticContainer
	}
	return TierContainerFallback
}

func BuildFileHasBioCMap(report *scan.ScanReport) map[string]bool {
	result := make(map[string]bool)
	for _, r := range report.Results {
		if r.HasBioConversation {
			result[r.FilePath] = true
		}
	}
	return result
}
