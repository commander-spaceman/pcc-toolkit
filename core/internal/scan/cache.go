package scan

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

const cacheVersion = 1

type FileCacheEntry struct {
	Size               int64 `json:"size"`
	ModTime            int64 `json:"modtime"`
	StrRefs            []int `json:"strrefs"`
	HasBioConversation bool  `json:"has_bioconversation"`
}

type FileCache struct {
	Version int                       `json:"version"`
	Files   map[string]FileCacheEntry `json:"files"`
}

func LoadFileCache(path string) (*FileCache, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return &FileCache{Version: cacheVersion, Files: make(map[string]FileCacheEntry)}, nil
	}
	var fc FileCache
	if err := json.Unmarshal(data, &fc); err != nil {
		return &FileCache{Version: cacheVersion, Files: make(map[string]FileCacheEntry)}, nil
	}
	if fc.Files == nil {
		fc.Files = make(map[string]FileCacheEntry)
	}
	fc.Version = cacheVersion
	return &fc, nil
}

func (fc *FileCache) Save(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
		return err
	}
	fc.Version = cacheVersion
	data, err := json.MarshalIndent(fc, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}

func (fc *FileCache) GetCachedStrRefs(filename string, size int64, modtime int64) ([]int, bool) {
	key := strings.ToLower(filename)
	entry, ok := fc.Files[key]
	if !ok {
		return nil, false
	}
	if entry.Size != size || entry.ModTime != modtime {
		return nil, false
	}
	return entry.StrRefs, true
}

func (fc *FileCache) SetEntry(filename string, size int64, modtime int64, strrefs []int, hasBioC bool) {
	key := strings.ToLower(filename)
	sorted := make([]int, len(strrefs))
	copy(sorted, strrefs)
	sort.Ints(sorted)
	fc.Files[key] = FileCacheEntry{
		Size:               size,
		ModTime:            modtime,
		StrRefs:            sorted,
		HasBioConversation: hasBioC,
	}
}

func (fc *FileCache) HasBioConversation(filename string, size int64, modtime int64) bool {
	key := strings.ToLower(filename)
	entry, ok := fc.Files[key]
	if !ok {
		return false
	}
	return entry.Size == size && entry.ModTime == modtime && entry.HasBioConversation
}

func (fc *FileCache) FindFilesWithStrRef(strref int) map[string]bool {
	result := make(map[string]bool)
	for key, entry := range fc.Files {
		for _, sr := range entry.StrRefs {
			if sr == strref {
				result[key] = true
				break
			}
		}
	}
	return result
}
