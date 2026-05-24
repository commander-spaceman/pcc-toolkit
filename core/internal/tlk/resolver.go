package tlk

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

type Resolver struct {
	Files []*File
}

func (r *Resolver) Resolve(stringID int32) (string, bool) {
	for _, tlk := range r.Files {
		text, ok := ResolveString(tlk, stringID, true)
		if ok {
			return text, true
		}
	}
	return "", false
}

func (r *Resolver) ResolveWithSource(stringID int32) *ResolveResult {
	for _, tlk := range r.Files {
		text, ok := ResolveString(tlk, stringID, true)
		if ok {
			return &ResolveResult{
				StringID:  stringID,
				Text:      text,
				SourceTLK: tlk.Path,
				Found:     true,
			}
		}
	}
	return &ResolveResult{StringID: stringID, Text: "", Found: false}
}

func (r *Resolver) IterAllEntries() func(func(int32, string, string) bool) {
	return func(yield func(int32, string, string) bool) {
		seen := make(map[int32]bool)
		for _, tlk := range r.Files {
			tlk.IterEntries()(func(id int32, text string) bool {
				if seen[id] {
					return true
				}
				seen[id] = true
				return yield(id, text, tlk.Path)
			})
		}
	}
}

func (r *Resolver) TotalUniqueEntries() int {
	count := 0
	seen := make(map[int32]bool)
	for _, tlk := range r.Files {
		for id := range tlk.MaleEntries {
			if !seen[id] {
				seen[id] = true
				count++
			}
		}
	}
	return count
}

func (r *Resolver) Search(query string) []ResolveResult {
	var results []ResolveResult
	r.IterAllEntries()(func(id int32, text string, source string) bool {
		if containsFold(text, query) || containsFold(fmt.Sprintf("%d", id), query) {
			results = append(results, ResolveResult{
				StringID:  id,
				Text:      text,
				SourceTLK: source,
			})
		}
		return true
	})
	return results
}

var mountPriorityRe = regexp.MustCompile(`(?i)MountPriority\s*=\s*(\d+)`)

func ReadMountPriority(dlcRoot string) int {
	data, err := os.ReadFile(filepath.Join(dlcRoot, "Mount.dlc"))
	if err != nil {
		return 0
	}
	match := mountPriorityRe.FindSubmatch(data)
	if len(match) < 2 {
		return 0
	}
	pri, err := strconv.Atoi(string(match[1]))
	if err != nil {
		return 0
	}
	return pri
}

type tlkCandidate struct {
	Path     string
	Priority int
}

func FindDlcTlkFiles(dlcDir string, language string, includeTestTlks bool) ([]string, error) {
	if language == "" {
		language = "INT"
	}

	entries, err := os.ReadDir(dlcDir)
	if err != nil {
		return nil, fmt.Errorf("read dlc dir: %w", err)
	}

	hasDLCFolders := false
	var candidates []tlkCandidate

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if !strings.HasPrefix(strings.ToUpper(name), "DLC_") {
			continue
		}
		hasDLCFolders = true

		dlcRoot := filepath.Join(dlcDir, name)
		priority := ReadMountPriority(dlcRoot)

		cookedMatches, _ := filepath.Glob(filepath.Join(dlcRoot, "CookedPC*", fmt.Sprintf("*_%s.TLK", language)))
		for _, match := range cookedMatches {
			if !includeTestTlks && strings.Contains(strings.ToLower(filepath.Base(match)), "_test_") {
				continue
			}
			candidates = append(candidates, tlkCandidate{Path: match, Priority: priority})
		}
	}

	if !hasDLCFolders {
		matches, _ := filepath.Glob(filepath.Join(dlcDir, "**", "*.tlk"))
		for _, match := range matches {
			candidates = append(candidates, tlkCandidate{Path: match, Priority: 0})
		}
	}

	sort.Slice(candidates, func(i, j int) bool {
		if candidates[i].Priority != candidates[j].Priority {
			return candidates[i].Priority < candidates[j].Priority
		}
		return strings.ToLower(candidates[i].Path) < strings.ToLower(candidates[j].Path)
	})

	result := make([]string, len(candidates))
	for i, c := range candidates {
		result[i] = c.Path
	}
	return result, nil
}

func BuildResolver(baseTlkPath string, dlcDir string, language string, includeTestTlks bool) (*Resolver, error) {
	var files []*File

	baseTLK, err := ReadFile(baseTlkPath)
	if err == nil {
		files = append(files, baseTLK)
	}

	if dlcDir != "" {
		dlcPaths, dlcErr := FindDlcTlkFiles(dlcDir, language, includeTestTlks)
		if dlcErr == nil {
			for _, path := range dlcPaths {
				tlkFile, readErr := ReadFile(path)
				if readErr != nil {
					continue
				}
				files = append(files, tlkFile)
			}
		}
	}

	if len(files) == 0 {
		if err != nil {
			return nil, fmt.Errorf("read base tlk: %w", err)
		}
		return nil, fmt.Errorf("no TLK files found")
	}

	return &Resolver{Files: files}, nil
}

func containsFold(s, substr string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(substr))
}
