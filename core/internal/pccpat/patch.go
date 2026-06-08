package pccpat

import (
	"encoding/binary"
	"errors"
	"fmt"

	pcc "github.com/commander-spaceman/me2pcc"
)

type exportEntryMeta struct {
	entryOffset      int
	entrySize        int
	serialSizeOffset int
	serialOffsetOff  int
}

func buildExportEntryMap(data []byte, header pcc.Header) ([]exportEntryMeta, error) {
	if header.ExportCount <= 0 {
		return nil, nil
	}
	cursor := header.ExportOffset
	if cursor < 0 || cursor >= len(data) {
		return nil, errors.New("export table offset out of range")
	}

	entries := make([]exportEntryMeta, header.ExportCount)
	for i := 0; i < header.ExportCount; i++ {
		if cursor+40 > len(data) {
			return nil, fmt.Errorf("truncated export table at index %d", i)
		}

		entryOffset := cursor
		headerLen := 40

		if header.UnrealVersion <= 512 {
			if cursor+44 > len(data) {
				return nil, fmt.Errorf("truncated component map at export %d", i)
			}
			componentCount := int(int32(binary.LittleEndian.Uint32(data[cursor+40:])))
			if componentCount < 0 {
				return nil, fmt.Errorf("negative component map count at export %d", i)
			}
			headerLen += 4 + (componentCount * 12)
		}

		genPos := cursor + headerLen
		if genPos+8 > len(data) {
			return nil, fmt.Errorf("truncated generation header at export %d", i)
		}
		generationCount := int(int32(binary.LittleEndian.Uint32(data[genPos+4:])))
		if generationCount < 0 {
			return nil, fmt.Errorf("negative generation count at export %d", i)
		}
		headerLen += 8 + (generationCount * 4) + 16
		if !(header.UnrealVersion == 491 && header.LicenseeVersion <= 110) {
			headerLen += 4
		}

		if cursor+headerLen > len(data) {
			return nil, fmt.Errorf("truncated export entry %d", i)
		}

		entries[i] = exportEntryMeta{
			entryOffset:      entryOffset,
			entrySize:        headerLen,
			serialSizeOffset: entryOffset + 32,
			serialOffsetOff:  entryOffset + 36,
		}
		cursor += headerLen
	}
	return entries, nil
}

func PatchExport(rawData []byte, summary *pcc.FileSummary, exportIndex int, newData []byte) ([]byte, *pcc.FileSummary, error) {
	if rawData == nil || summary == nil {
		return nil, nil, errors.New("rawData and summary are required")
	}
	if exportIndex < 0 || exportIndex >= len(summary.Exports) {
		return nil, nil, fmt.Errorf("export index %d out of range [0, %d)", exportIndex, len(summary.Exports))
	}

	entries, err := buildExportEntryMap(rawData, summary.Header)
	if err != nil {
		return nil, nil, fmt.Errorf("export table: %w", err)
	}
	if len(entries) != len(summary.Exports) {
		return nil, nil, fmt.Errorf("export count mismatch: table has %d, summary has %d", len(entries), len(summary.Exports))
	}

	updatedSummary := cloneSummary(summary)
	exp := &updatedSummary.Exports[exportIndex]
	meta := entries[exportIndex]

	oldOffset := exp.SerialOffset
	oldSize := exp.SerialSize

	if oldOffset < 0 || oldOffset+oldSize > len(rawData) {
		return nil, nil, fmt.Errorf("export %d data out of range: offset=%d size=%d len=%d", exportIndex, oldOffset, oldSize, len(rawData))
	}

	newSize := len(newData)
	delta := newSize - oldSize

	var result []byte
	if delta != 0 {
		result = shiftData(rawData, oldOffset+oldSize, delta)
	} else {
		result = make([]byte, len(rawData))
		copy(result, rawData)
	}

	copy(result[oldOffset:oldOffset+newSize], newData)

	writeI32(result, meta.serialSizeOffset, newSize)
	exp.SerialSize = newSize

	for j := exportIndex + 1; j < len(updatedSummary.Exports); j++ {
		nextMeta := entries[j]
		nextExp := &updatedSummary.Exports[j]
		newOffset := nextExp.SerialOffset + delta
		writeI32(result, nextMeta.serialOffsetOff, newOffset)
		nextExp.SerialOffset = newOffset
	}

	return result, updatedSummary, nil
}

func shiftData(data []byte, pivot int, delta int) []byte {
	newLen := len(data) + delta
	result := make([]byte, newLen)

	headLen := pivot
	if headLen > newLen {
		headLen = newLen
	}
	copy(result[:headLen], data[:headLen])

	tailStart := pivot
	if tailStart < len(data) {
		copy(result[pivot+delta:], data[tailStart:])
	}

	return result
}

func writeI32(buf []byte, offset int, v int) {
	binary.LittleEndian.PutUint32(buf[offset:], uint32(int32(v)))
}

func cloneSummary(s *pcc.FileSummary) *pcc.FileSummary {
	exports := make([]pcc.Export, len(s.Exports))
	copy(exports, s.Exports)
	names := make([]string, len(s.Names))
	copy(names, s.Names)
	imports := make([]pcc.Import, len(s.Imports))
	copy(imports, s.Imports)
	return &pcc.FileSummary{
		Path:        s.Path,
		GameProfile: s.GameProfile,
		Compressed:  s.Compressed,
		Header:      s.Header,
		Names:       names,
		Imports:     imports,
		Exports:     exports,
	}
}
