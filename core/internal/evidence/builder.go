package evidence

import (
	me2resolver "github.com/commander-spaceman/me2tlk/resolver"
	"pcc-toolkit/core/internal/dialogue"
	"pcc-toolkit/core/internal/owners"
	"pcc-toolkit/core/internal/pcc"
	"pcc-toolkit/core/internal/scan"
)

func BuildReport(
	query string,
	tlkPath string,
	dlcDir string,
	bioGameRoot string,
	candidates []int32,
	scanReport *scan.ScanReport,
	resolver *me2resolver.Resolver,
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

	report.NarrativeProfiles = computeNarrativeProfiles(report)

	return report
}

func findSourceTLK(resolver *me2resolver.Resolver, strref int32) string {
	result := resolver.ResolveWithSource(strref)
	if result != nil {
		return result.SourceTLK
	}
	return ""
}

type nodeInfo struct {
	nodeType       string
	nodeID         int
	speakerTag     string
	listenerTag    string
	conversationID string
}

type nodeKey struct {
	exportIndex int
	strRef      int
}

func EnrichConversationMatchesWithAST(report *EvidenceReport) {
	filesToParse := make(map[string]bool)
	for i := range report.Evidence {
		ev := &report.Evidence[i]
		for j := range ev.BioConversation {
			filesToParse[ev.BioConversation[j].FilePath] = true
		}
	}

	if len(filesToParse) == 0 {
		return
	}

	fileNodeMaps := make(map[string]map[nodeKey][]nodeInfo)
	fileOwnerMaps := make(map[string]map[string]string)

	for filePath := range filesToParse {
		rawData, summary, err := pcc.ReadFileRaw(filePath)
		if err != nil {
			continue
		}
		result := dialogue.ParseConversations(summary, rawData, "resilient")
		strRefMap := make(map[nodeKey][]nodeInfo)
		for _, conv := range result.Conversations {
			for _, entry := range conv.Entries {
				if entry.LineStrRef != nil && *entry.LineStrRef > 0 {
					key := nodeKey{exportIndex: conv.ExportIndex, strRef: *entry.LineStrRef}
					strRefMap[key] = append(strRefMap[key], nodeInfo{
						nodeType:       "entry",
						nodeID:         entry.ID,
						speakerTag:     entry.SpeakerTag,
						listenerTag:    entry.ListenerTag,
						conversationID: conv.ID,
					})
				}
			}
			for _, reply := range conv.Replies {
				if reply.LineStrRef != nil && *reply.LineStrRef > 0 {
					key := nodeKey{exportIndex: conv.ExportIndex, strRef: *reply.LineStrRef}
					strRefMap[key] = append(strRefMap[key], nodeInfo{
						nodeType:       "reply",
						nodeID:         reply.ID,
						speakerTag:     "player",
						listenerTag:    "",
						conversationID: conv.ID,
					})
				}
			}
		}
		fileNodeMaps[filePath] = strRefMap

		ownerOutput := owners.ScanOwners(rawData, summary, filePath)
		convOwnerMap := make(map[string]string)
		for _, oe := range ownerOutput.Owners {
			convOwnerMap[oe.ConversationName] = oe.OwnerTag
		}
		fileOwnerMaps[filePath] = convOwnerMap
	}

	for i := range report.Evidence {
		ev := &report.Evidence[i]
		for j := range ev.BioConversation {
			cm := &ev.BioConversation[j]
			if nodeMap, ok := fileNodeMaps[cm.FilePath]; ok {
				if infos, ok := nodeMap[nodeKey{exportIndex: cm.ExportIndex, strRef: cm.StrRef}]; ok && len(infos) == 1 {
					info := infos[0]
					cm.NodeType = info.nodeType
					cm.NodeID = info.nodeID
					cm.SpeakerTag = info.speakerTag
					cm.ListenerTag = info.listenerTag
					if info.conversationID != "" {
						cm.ConversationID = info.conversationID
					}
				}
			}
			if ownerMap, ok := fileOwnerMaps[cm.FilePath]; ok {
				if cm.ConversationID != "" {
					if ownerTag, ok := ownerMap[cm.ConversationID]; ok {
						cm.OwnerTag = ownerTag
					}
				}
			}
		}
	}
}

func computeNarrativeProfiles(report *EvidenceReport) []ProfileMatch {
	var combined string
	for _, ev := range report.Evidence {
		if ev.Text != "" {
			if combined != "" {
				combined += " "
			}
			combined += ev.Text
		}
	}
	if combined == "" {
		combined = report.Query
	}
	if combined == "" {
		return nil
	}
	return MatchProfile(combined)
}
