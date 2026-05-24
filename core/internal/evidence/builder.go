package evidence

import (
	"pcc-toolkit/core/internal/pcc"
	"pcc-toolkit/core/internal/scan"
	"pcc-toolkit/core/internal/tlk"
)

func BuildReport(
	query string,
	tlkPath string,
	dlcDir string,
	bioGameRoot string,
	candidates []int32,
	scanReport *scan.ScanReport,
	resolver *tlk.Resolver,
) *EvidenceReport {
	report := &EvidenceReport{
		Query:         query,
		TlkPath:       tlkPath,
		DlcDir:        dlcDir,
		BioGameRoot:   bioGameRoot,
		FilesScanned:  scanReport.FilesScanned,
		FilesWithHits: scanReport.FilesWithHits,
		TotalHits:     scanReport.TotalHits,
	}

	report.Errors = append(report.Errors, scanReport.Errors...)

	for _, c := range candidates {
		report.CandidateStrRefs = append(report.CandidateStrRefs, int(c))
	}

	hasBioCMap := BuildFileHasBioCMap(scanReport)

	evidenceMap := make(map[int]*StrRefEvidence)
	for _, candidate := range candidates {
		ev := &StrRefEvidence{
			StrRef: int(candidate),
		}
		if resolver != nil {
			text, ok := resolver.Resolve(candidate)
			if ok {
				ev.Text = text
				ev.SourceTLK = findSourceTLK(resolver, candidate)
			}
		}
		evidenceMap[int(candidate)] = ev
	}

	for _, result := range scanReport.Results {
		fileHasBioC := hasBioCMap[result.FilePath]
		for _, hit := range result.Hits {
			ev, ok := evidenceMap[hit.StrRef]
			if !ok {
				continue
			}

			tier := ClassifyTier(hit.ClassName, fileHasBioC)

			switch tier {
			case TierBioConversation:
				cm := ConversationMatch{
					TieredHit: TieredHit{
						StrRef:     hit.StrRef,
						Text:       ev.Text,
						FilePath:   result.FilePath,
						ExportName: hit.ExportName,
						ClassName:  hit.ClassName,
						Tier:       string(TierBioConversation),
					},
					ConversationID: hit.ExportName,
					ExportIndex:    hit.ExportIndex,
				}
				ev.BioConversation = append(ev.BioConversation, cm)

			case TierSemanticContainer:
				ev.SemanticContainer = append(ev.SemanticContainer, TieredHit{
					StrRef:     hit.StrRef,
					Text:       ev.Text,
					FilePath:   result.FilePath,
					ExportName: hit.ExportName,
					ClassName:  hit.ClassName,
					Tier:       string(TierSemanticContainer),
				})

			case TierContainerFallback:
				ev.ContainerFallback = append(ev.ContainerFallback, TieredHit{
					StrRef:     hit.StrRef,
					Text:       ev.Text,
					FilePath:   result.FilePath,
					ExportName: hit.ExportName,
					ClassName:  hit.ClassName,
					Tier:       string(TierContainerFallback),
				})
			}
		}
	}

	for _, candidate := range candidates {
		ev := evidenceMap[int(candidate)]
		report.Evidence = append(report.Evidence, *ev)
	}

	return report
}

func findSourceTLK(resolver *tlk.Resolver, strref int32) string {
	result := resolver.ResolveWithSource(strref)
	if result != nil {
		return result.SourceTLK
	}
	return ""
}

func EnrichWithConversationData(
	report *EvidenceReport,
	fileData map[string]*pcc.FileSummary,
) {
	for i := range report.Evidence {
		ev := &report.Evidence[i]
		for j := range ev.BioConversation {
			cm := &ev.BioConversation[j]
			summary, ok := fileData[cm.FilePath]
			if !ok {
				continue
			}
			for _, exp := range summary.Exports {
				if exp.Index == cm.ExportIndex && exp.ClassName == "BioConversation" {
					cm.ConversationID = exp.ObjectName
					break
				}
			}
		}
	}
}
