package editor

import (
	"encoding/binary"
	"fmt"

	"pcc-toolkit/core/internal/dialogue"
	"pcc-toolkit/core/internal/pcc"
	"pcc-toolkit/core/internal/pccpat"
	"pcc-toolkit/core/internal/pccwrt"
)

func EditConversation(
	pccPath string,
	outputPath string,
	convIndex int,
	modifyFn func(*dialogue.Conversation) error,
) error {
	rawData, summary, err := pcc.ReadFileRaw(pccPath)
	if err != nil {
		return fmt.Errorf("read PCC: %w", err)
	}

	if err := summary.RequireME2(); err != nil {
		return err
	}

	result := dialogue.ParseConversations(summary, rawData, "resilient")
	if err := firstError(*result); err != nil {
		return fmt.Errorf("parse conversations: %w", err)
	}

	if convIndex < 0 || convIndex >= len(result.Conversations) {
		return fmt.Errorf("conversation index %d out of range [0, %d)", convIndex, len(result.Conversations))
	}

	conv := &result.Conversations[convIndex]

	if err := modifyFn(conv); err != nil {
		return fmt.Errorf("modify conversation: %w", err)
	}

	exportIndex := conv.ExportIndex
	newSerialData, addedNames, err := SerializeConversation(*conv, summary.Names)
	if err != nil {
		return fmt.Errorf("serialize conversation: %w", err)
	}

	if len(addedNames) > 0 {
		expandedNames := make([]string, len(summary.Names)+len(addedNames))
		copy(expandedNames, summary.Names)
		copy(expandedNames[len(summary.Names):], addedNames)
		summary.Names = expandedNames

		rawData, err = expandNameTable(rawData, summary, addedNames)
		if err != nil {
			return fmt.Errorf("expand name table: %w", err)
		}

		newSerialData, _, err = SerializeConversation(*conv, summary.Names)
		if err != nil {
			return fmt.Errorf("re-serialize with expanded names: %w", err)
		}
	}

	patchedData, newSummary, err := pccpat.PatchExport(rawData, summary, exportIndex, newSerialData)
	if err != nil {
		return fmt.Errorf("patch export: %w", err)
	}

	if err := pccwrt.WritePCCCompressed(outputPath, newSummary, patchedData, true); err != nil {
		return fmt.Errorf("write PCC: %w", err)
	}

	return nil
}

func firstError(result dialogue.ParseResult) error {
	for _, pe := range result.Errors {
		return fmt.Errorf("%s (export %d): %s", pe.ID, pe.ExportIndex, pe.Error)
	}
	return nil
}

func expandNameTable(rawData []byte, summary *pcc.FileSummary, newNames []string) ([]byte, error) {
	if len(newNames) == 0 {
		return rawData, nil
	}

	newBytesLen := 0
	for _, n := range newNames {
		newBytesLen += 4 + len(n) + 1 + 4
	}

	nameTableEnd := summary.Header.ImportOffset
	if nameTableEnd <= summary.Header.NameOffset || nameTableEnd > len(rawData) {
		return nil, fmt.Errorf("invalid name table bounds: offset=%d end=%d len=%d",
			summary.Header.NameOffset, nameTableEnd, len(rawData))
	}

	insertPos := nameTableEnd

	newData := make([]byte, len(rawData)+newBytesLen)
	copy(newData[:insertPos], rawData[:insertPos])

	pos := insertPos
	for _, n := range newNames {
		strLen := len(n) + 1
		binary.LittleEndian.PutUint32(newData[pos:], uint32(strLen))
		pos += 4
		copy(newData[pos:], n)
		pos += len(n)
		newData[pos] = 0
		pos++
		binary.LittleEndian.PutUint32(newData[pos:], 0xFFFFFFF2)
		pos += 4
	}

	delta := pos - insertPos
	copy(newData[pos:], rawData[insertPos:])

	headerOff := 20
	writeI32At(newData, headerOff+4, summary.Header.NameCount+len(newNames))
	shiftOff := headerOff + 16
	writeI32At(newData, shiftOff, summary.Header.ImportOffset+delta)
	shiftOff += 4
	if summary.Header.ExportOffset > 0 {
		writeI32At(newData, shiftOff, summary.Header.ExportOffset+delta)
	}
	shiftOff += 4
	if summary.Header.ImportOffset > 0 {
		writeI32At(newData, shiftOff, summary.Header.ImportOffset+delta)
	}

	summary.Header.NameCount += len(newNames)
	summary.Header.ImportOffset += delta
	summary.Header.ExportOffset += delta

	for i := range summary.Exports {
		if summary.Exports[i].SerialOffset >= nameTableEnd {
			summary.Exports[i].SerialOffset += delta
		}
	}

	return newData, nil
}

func writeI32At(buf []byte, offset int, v int) {
	binary.LittleEndian.PutUint32(buf[offset:], uint32(v))
}
