package scan

import "pcc-toolkit/core/internal/pcc"

type FileEntry struct {
	Path        string
	MountFile   string
	MountPri    int
	Hash        string
	Size        int64
	ModTimeUNIX int64
}

type ContainerHit struct {
	pcc.ContainerHit
	FilePath string `json:"file_path"`
}

type ScanResult struct {
	FilePath           string
	Hits               []ContainerHit
	HasBioConversation bool
	Error              string
}

type ScanReport struct {
	FilesScanned  int          `json:"files_scanned"`
	FilesWithHits int          `json:"files_with_hits"`
	TotalHits     int          `json:"total_hits"`
	Results       []ScanResult `json:"results,omitempty"`
	Errors        []string     `json:"errors,omitempty"`
}
