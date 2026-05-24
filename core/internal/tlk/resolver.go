package tlk

import (
	"encoding/binary"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

func ReadMountPriority(dlcRoot string) int {
	paths := []string{
		filepath.Join(dlcRoot, "CookedPC", "Mount.dlc"),
		filepath.Join(dlcRoot, "CookedPCConsole", "Mount.dlc"),
		filepath.Join(dlcRoot, "Mount.dlc"),
	}
	for _, p := range paths {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		pri := parseMountPriorityBinary(data)
		if pri > 0 {
			return pri
		}
	}
	return 0
}

func parseMountPriorityBinary(data []byte) int {
	if len(data) < 14 {
		return 0
	}
	if data[0] == 0x00 {
		return int(binary.LittleEndian.Uint16(data[12:14]))
	}
	if data[0] == 0xAC {
		return int(binary.LittleEndian.Uint16(data[12:14]))
	}
	return 0
}

type tlkCandidate struct {
	Path     string
	Priority int
}

func ParseBioEngineModules(iniPath string) (map[string]int, error) {
	data, err := os.ReadFile(iniPath)
	if err != nil {
		return nil, err
	}

	modules := make(map[string]int)
	lines := strings.Split(string(data), "\n")
	inSection := false

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, ";") || strings.HasPrefix(trimmed, "#") {
			continue
		}
		upperTrimmed := strings.ToUpper(trimmed)
		if strings.HasPrefix(upperTrimmed, "[") && strings.HasSuffix(upperTrimmed, "]") {
			inSection = upperTrimmed == "[ENGINE.DLCMODULES]"
			continue
		}
		if inSection && strings.Contains(trimmed, "=") {
			parts := strings.SplitN(trimmed, "=", 2)
			key := strings.TrimSpace(parts[0])
			valueStr := strings.TrimSpace(parts[1])
			var num int
			if _, err := fmt.Sscanf(valueStr, "%d", &num); err == nil {
				modules[key] = num
			}
		}
	}
	return modules, nil
}

func findDlcTlkByModule(dlcRoot string, moduleNum int, language string) string {
	candidate := filepath.Join(dlcRoot, fmt.Sprintf("DLC_%d_%s.tlk", moduleNum, language))
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	for _, cookedDir := range []string{"CookedPC", "CookedPCConsole"} {
		candidate = filepath.Join(dlcRoot, cookedDir, fmt.Sprintf("DLC_%d_%s.tlk", moduleNum, language))
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func findDlcTlkByGlob(dlcRoot string, language string, includeTestTlks bool) []string {
	var results []string
	for _, cookedDir := range []string{"CookedPC", "CookedPCConsole"} {
		matches, _ := filepath.Glob(filepath.Join(dlcRoot, cookedDir, fmt.Sprintf("*_%s.tlk", language)))
		for _, match := range matches {
			if !includeTestTlks && strings.Contains(strings.ToLower(filepath.Base(match)), "_test_") {
				continue
			}
			results = append(results, match)
		}
	}
	return results
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

		moduleTLKPath := ""
		for _, cookedDir := range []string{"CookedPC", "CookedPCConsole"} {
			bioEnginePath := filepath.Join(dlcRoot, cookedDir, "BIOEngine.ini")
			modules, iniErr := ParseBioEngineModules(bioEnginePath)
			if iniErr == nil {
				if modNum, ok := modules[name]; ok {
					moduleTLKPath = findDlcTlkByModule(dlcRoot, modNum, language)
					break
				}
			}
		}

		if moduleTLKPath != "" {
			candidates = append(candidates, tlkCandidate{Path: moduleTLKPath, Priority: priority})
		} else {
			globs := findDlcTlkByGlob(dlcRoot, language, includeTestTlks)
			for _, match := range globs {
				candidates = append(candidates, tlkCandidate{Path: match, Priority: priority})
			}
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
