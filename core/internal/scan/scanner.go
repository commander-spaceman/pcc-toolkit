package scan

import (
	"encoding/binary"
	"os"
	"path/filepath"
	"runtime"
	"sync"

	pcc "github.com/commander-spaceman/me2pcc"
)

func ParseStrrefs(data []byte, candidates []int32) map[int][]int {
	if len(data) < 4 {
		return nil
	}

	result := make(map[int][]int)
	for _, strref := range candidates {
		target := make([]byte, 4)
		binary.LittleEndian.PutUint32(target, uint32(strref))
		offsets := findOffsets(data, target)
		if len(offsets) > 0 {
			result[int(strref)] = offsets
		}
	}
	return result
}

func findOffsets(data, target []byte) []int {
	if len(target) == 0 || len(data) < len(target) {
		return nil
	}
	var offsets []int
	limit := len(data) - len(target) + 1
	for i := 0; i < limit; i++ {
		match := true
		for j := 0; j < len(target); j++ {
			if data[i+j] != target[j] {
				match = false
				break
			}
		}
		if match {
			offsets = append(offsets, i)
		}
	}
	return offsets
}

func ScanFile(path string, candidates []int32) (*ScanResult, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	decompressed, summary, err := pcc.ReadFileRawFromBytes(data, path)
	if err != nil {
		return &ScanResult{
			FilePath: path,
			Error:    err.Error(),
		}, nil
	}

	hasBioC := false
	for _, e := range summary.Exports {
		if e.ClassName == "BioConversation" {
			hasBioC = true
			break
		}
	}

	if len(candidates) == 0 {
		return &ScanResult{
			FilePath:           path,
			HasBioConversation: hasBioC,
		}, nil
	}

	strRefOffsets := ParseStrrefs(decompressed, candidates)

	hits := pcc.MapOffsetsToContainers(summary.Exports, strRefOffsets, len(decompressed))

	result := &ScanResult{
		FilePath:           path,
		HasBioConversation: hasBioC,
	}

	for _, h := range hits {
		result.Hits = append(result.Hits, ContainerHit{
			ContainerHit: h,
			FilePath:     path,
		})
	}

	return result, nil
}

func Run(files []FileEntry, candidates []int32, workers int) *ScanReport {
	return runScan(files, candidates, workers, nil)
}

func RunWithCache(files []FileEntry, candidates []int32, workers int, cache *FileCache) *ScanReport {
	uncached := make([]FileEntry, 0)
	cachedResults := make([]ScanResult, 0)

	for _, f := range files {
		filename := filepath.Base(f.Path)
		cachedStrRefs, valid := cache.GetCachedStrRefs(filename, f.Size, f.ModTimeUNIX)
		if valid {
			hasBioC := cache.HasBioConversation(filename, f.Size, f.ModTimeUNIX)
			hits := make([]ContainerHit, 0)
			for _, sr := range cachedStrRefs {
				for _, c := range candidates {
					if int32(sr) == c {
						hits = append(hits, ContainerHit{
							ContainerHit: pcc.ContainerHit{StrRef: sr},
							FilePath:     f.Path,
						})
						break
					}
				}
			}
			cachedResults = append(cachedResults, ScanResult{
				FilePath:           f.Path,
				Hits:               hits,
				HasBioConversation: hasBioC,
			})
		} else {
			uncached = append(uncached, f)
		}
	}

	scanned := runScan(uncached, candidates, workers, cache)

	report := &ScanReport{
		FilesScanned: len(cachedResults) + scanned.FilesScanned,
	}

	for _, r := range cachedResults {
		if r.HasBioConversation {
			report.BioConversationFiles = append(report.BioConversationFiles, r.FilePath)
		}
		if len(r.Hits) > 0 {
			report.FilesWithHits++
			report.TotalHits += len(r.Hits)
			report.Results = append(report.Results, r)
		}
	}

	report.FilesScanned += scanned.FilesScanned
	report.FilesWithHits += scanned.FilesWithHits
	report.TotalHits += scanned.TotalHits
	report.Results = append(report.Results, scanned.Results...)
	report.Errors = append(report.Errors, scanned.Errors...)
	report.BioConversationFiles = append(report.BioConversationFiles, scanned.BioConversationFiles...)

	return report
}

func runScan(files []FileEntry, candidates []int32, workers int, cache *FileCache) *ScanReport {
	if workers <= 0 {
		workers = runtime.NumCPU()
	}

	report := &ScanReport{}

	jobs := make(chan FileEntry, len(files))
	results := make(chan ScanResult, len(files))

	var wg sync.WaitGroup
	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for f := range jobs {
				result, err := ScanFile(f.Path, candidates)
				if err != nil {
					results <- ScanResult{
						FilePath: f.Path,
						Error:    err.Error(),
					}
					continue
				}

				if cache != nil {
					strrefs := make([]int, 0, len(result.Hits))
					for _, h := range result.Hits {
						strrefs = append(strrefs, h.StrRef)
					}
					cache.SetEntry(filepath.Base(f.Path), f.Size, f.ModTimeUNIX, strrefs, result.HasBioConversation)
				}

				results <- *result
			}
		}()
	}

	go func() {
		for _, f := range files {
			jobs <- f
		}
		close(jobs)
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	for r := range results {
		report.FilesScanned++
		if r.Error != "" {
			report.Errors = append(report.Errors, r.FilePath+": "+r.Error)
			continue
		}
		if r.HasBioConversation {
			report.BioConversationFiles = append(report.BioConversationFiles, r.FilePath)
		}
		if len(r.Hits) > 0 {
			report.FilesWithHits++
			report.TotalHits += len(r.Hits)
			report.Results = append(report.Results, r)
		}
	}

	return report
}
