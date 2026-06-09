package editor

import (
	"encoding/binary"
	"fmt"
	"os"

	pcc "github.com/commander-spaceman/me2pcc"
	"pcc-toolkit/core/internal/dialogue"
	"pcc-toolkit/core/internal/pccpat"
)

type EditResult struct {
	Status     string                            `json:"status"`
	Output     string                            `json:"output,omitempty"`
	Validation *dialogue.ValidationReportSummary `json:"validation,omitempty"`
}

func EditConversation(
	pccPath string,
	outputPath string,
	convIndex int,
	dryRun bool,
	modifyFn func(*dialogue.Conversation) error,
) (*EditResult, error) {
	rawData, summary, err := pcc.ReadFileRaw(pccPath)
	if err != nil {
		return nil, fmt.Errorf("read PCC: %w", err)
	}

	if err := summary.RequireME2(); err != nil {
		return nil, err
	}

	result := dialogue.ParseConversations(summary, rawData, "resilient")
	if err := firstError(*result); err != nil {
		return nil, fmt.Errorf("parse conversations: %w", err)
	}

	if convIndex < 0 || convIndex >= len(result.Conversations) {
		return nil, fmt.Errorf("conversation index %d out of range [0, %d)", convIndex, len(result.Conversations))
	}

	conv := &result.Conversations[convIndex]

	exportIndex := conv.ExportIndex
	if exportIndex < 0 || exportIndex >= len(summary.Exports) {
		return nil, fmt.Errorf("export index %d out of range", exportIndex)
	}
	originalSerial := rawData[summary.Exports[exportIndex].SerialOffset:]
	if len(originalSerial) > summary.Exports[exportIndex].SerialSize {
		originalSerial = originalSerial[:summary.Exports[exportIndex].SerialSize]
	}

	if err := modifyFn(conv); err != nil {
		return nil, fmt.Errorf("modify conversation: %w", err)
	}

	var newSerialData []byte
	var addedNames []string

	if len(originalSerial) > 0 {
		newSerialData, err = SerializeConversationPreserving(*conv, originalSerial, summary.Names)
		if err != nil {
			return nil, fmt.Errorf("serialize preserving: %w", err)
		}
		addedNames = nil
	} else {
		newSerialData, addedNames, err = SerializeConversation(*conv, summary.Names)
		if err != nil {
			return nil, fmt.Errorf("serialize conversation: %w", err)
		}
	}

	if len(addedNames) > 0 {
		expandedNames := make([]string, len(summary.Names)+len(addedNames))
		copy(expandedNames, summary.Names)
		copy(expandedNames[len(summary.Names):], addedNames)
		summary.Names = expandedNames

		rawData, err = expandNameTable(rawData, summary, addedNames)
		if err != nil {
			return nil, fmt.Errorf("expand name table: %w", err)
		}

		newSerialData, _, err = SerializeConversation(*conv, summary.Names)
		if err != nil {
			return nil, fmt.Errorf("re-serialize with expanded names: %w", err)
		}
	}

	patchedData, newSummary, err := pccpat.PatchExport(rawData, summary, exportIndex, newSerialData)
	if err != nil {
		return nil, fmt.Errorf("patch export: %w", err)
	}

	validateResult := dialogue.ParseConversations(newSummary, patchedData, "resilient")
	validationReport := dialogue.BuildValidationReport(validateResult)

	editResult := &EditResult{
		Status:     "ok",
		Output:     outputPath,
		Validation: &validationReport.Summary,
	}

	if dryRun {
		editResult.Status = "dry_run"
		editResult.Output = ""
		return editResult, nil
	}

	if validationReport.Summary.Invalid > 0 {
		editResult.Status = fmt.Sprintf("written_with_%d_invalid", validationReport.Summary.Invalid)
	} else if validationReport.Summary.Warning > 0 {
		editResult.Status = fmt.Sprintf("written_with_%d_warnings", validationReport.Summary.Warning)
	}

	// Clear compressed flag in header since patchedData is uncompressed.
	// The flags offset depends on the folder name length in the PCC header.
	_ = binary.LittleEndian.Uint32(patchedData[0:4]) // magic
	cursor := 8
	cursor += 4
	folderLen := int(int32(binary.LittleEndian.Uint32(patchedData[cursor : cursor+4])))
	cursor += 4
	if folderLen > 0 {
		cursor += folderLen
	} else if folderLen < 0 {
		cursor += (-folderLen) * 2
	}
	flags := binary.LittleEndian.Uint32(patchedData[cursor : cursor+4])
	flags &^= uint32(pcc.CompressedFlag)
	binary.LittleEndian.PutUint32(patchedData[cursor:cursor+4], flags)

	if err := os.WriteFile(outputPath, patchedData, 0644); err != nil {
		return nil, fmt.Errorf("write PCC: %w", err)
	}

	return editResult, nil
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
