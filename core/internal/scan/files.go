package scan

import (
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/commander-spaceman/me2tlk/resolver"
)

func CollectPccFiles(biogameRoot string) ([]FileEntry, error) {
	if biogameRoot == "" {
		return nil, nil
	}

	var allFiles []FileEntry

	baseCooked := filepath.Join(biogameRoot, "CookedPC")
	if entries, err := os.ReadDir(baseCooked); err == nil {
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			if !strings.EqualFold(filepath.Ext(e.Name()), ".pcc") {
				continue
			}
			info, err := e.Info()
			if err != nil {
				continue
			}
			allFiles = append(allFiles, FileEntry{
				Path:        filepath.Join(baseCooked, e.Name()),
				MountFile:   "basegame",
				MountPri:    -1,
				Size:        info.Size(),
				ModTimeUNIX: info.ModTime().Unix(),
			})
		}
	}

	dlcDir := filepath.Join(biogameRoot, "DLC")
	if entries, err := os.ReadDir(dlcDir); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				continue
			}
			dlcName := entry.Name()
			if !strings.HasPrefix(strings.ToUpper(dlcName), "DLC_") {
				continue
			}
			dlcRoot := filepath.Join(dlcDir, dlcName)
			pri := resolver.ReadMountPriority(dlcRoot)

			cookedDirs, _ := filepath.Glob(filepath.Join(dlcRoot, "CookedPC*"))
			for _, cookedDir := range cookedDirs {
				cookedEntries, err := os.ReadDir(cookedDir)
				if err != nil {
					continue
				}
				for _, ce := range cookedEntries {
					if ce.IsDir() {
						continue
					}
					if !strings.EqualFold(filepath.Ext(ce.Name()), ".pcc") {
						continue
					}
					info, err := ce.Info()
					if err != nil {
						continue
					}
					allFiles = append(allFiles, FileEntry{
						Path:        filepath.Join(cookedDir, ce.Name()),
						MountFile:   dlcName,
						MountPri:    pri,
						Size:        info.Size(),
						ModTimeUNIX: info.ModTime().Unix(),
					})
				}
			}
		}
	}

	deduped := deduplicateByFilename(allFiles)
	return deduped, nil
}

func deduplicateByFilename(files []FileEntry) []FileEntry {
	sort.Slice(files, func(i, j int) bool {
		if files[i].MountPri != files[j].MountPri {
			return files[i].MountPri < files[j].MountPri
		}
		return strings.ToLower(files[i].Path) < strings.ToLower(files[j].Path)
	})

	seen := make(map[string]FileEntry)
	for _, f := range files {
		key := strings.ToLower(filepath.Base(f.Path))
		seen[key] = f
	}

	result := make([]FileEntry, 0, len(seen))
	for _, f := range seen {
		result = append(result, f)
	}
	sort.Slice(result, func(i, j int) bool {
		return strings.ToLower(result[i].Path) < strings.ToLower(result[j].Path)
	})
	return result
}
