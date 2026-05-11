package pcc

func bestContainingExport(exports []Export, offset int, dataLen int) *Export {
	var best *Export
	for i := range exports {
		item := &exports[i]
		start := item.SerialOffset
		end := item.SerialOffset + item.SerialSize
		if start < 0 || end > dataLen || start >= end {
			continue
		}
		if offset < start || offset >= end {
			continue
		}
		if best == nil {
			best = item
			continue
		}
		itemHasClass := item.ClassName != ""
		bestHasClass := best.ClassName != ""
		if itemHasClass && !bestHasClass {
			best = item
			continue
		}
		if itemHasClass == bestHasClass && item.SerialSize < best.SerialSize {
			best = item
		}
	}
	return best
}

func MapOffsetsToContainers(exports []Export, offsets map[int][]int, dataLen int) []ContainerHit {
	hasOffsets := false
	for _, rows := range offsets {
		if len(rows) > 0 {
			hasOffsets = true
			break
		}
	}
	if !hasOffsets {
		return nil
	}

	var hits []ContainerHit
	for strref, strrefOffsets := range offsets {
		for _, absoluteOffset := range strrefOffsets {
			best := bestContainingExport(exports, absoluteOffset, dataLen)
			if best == nil {
				continue
			}
			hits = append(hits, ContainerHit{
				StrRef:         strref,
				AbsoluteOffset: absoluteOffset,
				LocalOffset:    absoluteOffset - best.SerialOffset,
				ExportIndex:    best.Index,
				ExportName:     best.ObjectName,
				ClassName:      best.ClassName,
				SerialOffset:   best.SerialOffset,
				SerialSize:     best.SerialSize,
			})
		}
	}
	return hits
}
