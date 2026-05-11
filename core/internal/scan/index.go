package scan

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"os"
)

type FileIndex struct {
	Files map[string]FileIndexEntry `json:"files"`
}

type FileIndexEntry struct {
	Path       string `json:"path"`
	Hash       string `json:"hash"`
	Size       int64  `json:"size"`
	ModTime    int64  `json:"mod_time"`
}

func LoadIndex(path string) (*FileIndex, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &FileIndex{Files: make(map[string]FileIndexEntry)}, nil
	}
	var idx FileIndex
	if err := json.Unmarshal(data, &idx); err != nil {
		return &FileIndex{Files: make(map[string]FileIndexEntry)}, nil
	}
	if idx.Files == nil {
		idx.Files = make(map[string]FileIndexEntry)
	}
	return &idx, nil
}

func (idx *FileIndex) Save(path string) error {
	data, err := json.MarshalIndent(idx, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func SplitChangedFiles(files []FileEntry, idx *FileIndex) (changed, unchanged []FileEntry) {
	for _, f := range files {
		entry, exists := idx.Files[f.Path]
		if exists && entry.Size == f.Size && entry.ModTime == f.ModTimeUNIX {
			unchanged = append(unchanged, f)
		} else {
			changed = append(changed, f)
		}
	}
	return
}

func (idx *FileIndex) UpdateFrom(files []FileEntry) {
	for _, f := range files {
		hash := fmt.Sprintf("%x", md5.Sum([]byte(f.Path+fmt.Sprint(f.Size)+fmt.Sprint(f.ModTimeUNIX))))
		idx.Files[f.Path] = FileIndexEntry{
			Path:    f.Path,
			Hash:    hash,
			Size:    f.Size,
			ModTime: f.ModTimeUNIX,
		}
	}
}
