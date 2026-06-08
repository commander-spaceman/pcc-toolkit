package pccwrt

import (
	"encoding/binary"
	"fmt"
	"os"

	pcc "github.com/commander-spaceman/me2pcc"
)

const (
	me2OTUV = 512
	me2OTLV = 130
)

func WritePCC(path string, summary *pcc.FileSummary, rawData []byte) error {
	return WritePCCCompressed(path, summary, rawData, false)
}

func WritePCCCompressed(path string, summary *pcc.FileSummary, rawData []byte, compress bool) error {
	if summary == nil {
		return fmt.Errorf("summary is nil")
	}
	if rawData == nil {
		return fmt.Errorf("rawData is nil")
	}

	if err := summary.RequireME2(); err != nil {
		return err
	}

	var buf []byte
	var err error

	if compress {
		buf, err = buildCompressedPCC(summary, rawData)
	} else {
		buf, err = buildUncompressedBuffer(summary, rawData)
	}
	if err != nil {
		return err
	}

	return os.WriteFile(path, buf, 0644)
}

func buildUncompressedPCC(summary *pcc.FileSummary, rawData []byte) ([]byte, error) {
	return buildUncompressedBuffer(summary, rawData)
}

func writeHeader(buf []byte, flags uint32, nameCount, nameOffset, exportCount, exportOffset, importCount, importOffset, dependsCount, dependsOffset int) {
	pos := 0
	writeU32(buf, pos, pcc.PackageMagic)
	pos += 4
	writeU32(buf, pos, uint32(me2OTUV|(me2OTLV<<16)))
	pos += 4
	writeU32(buf, pos, 0)
	pos += 4
	writeI32(buf, pos, 0)
	pos += 4

	writeU32(buf, pos, flags)
	pos += 4
	writeI32(buf, pos, nameCount)
	pos += 4
	writeI32(buf, pos, nameOffset)
	pos += 4
	writeI32(buf, pos, exportCount)
	pos += 4
	writeI32(buf, pos, exportOffset)
	pos += 4
	writeI32(buf, pos, importCount)
	pos += 4
	writeI32(buf, pos, importOffset)
	pos += 4
	writeI32(buf, pos, dependsCount)
	pos += 4

	for i := 0; i < 4; i++ {
		writeI32(buf, pos, 0)
		pos += 4
	}

	writeI32(buf, pos, 0)
	pos += 4

	writeU32(buf, pos, 0)
	pos += 4
	writeU32(buf, pos, 0)
	pos += 4

	writeU32(buf, pos, 0)
	pos += 4

	for i := 0; i < 3; i++ {
		writeI32(buf, pos, 0)
		pos += 4
	}

	writeI32(buf, pos, 0)
	pos += 4
	writeI32(buf, pos, 0)
	pos += 4
}

func buildNameTable(names []string) []byte {
	var total int
	for _, n := range names {
		total += 4 + len(n) + 1 + 4
	}
	buf := make([]byte, total)
	pos := 0
	for _, n := range names {
		strLen := len(n) + 1
		writeI32(buf, pos, strLen)
		pos += 4
		copy(buf[pos:], n)
		pos += len(n)
		buf[pos] = 0
		pos++
		writeI32(buf, pos, -14) // ME2 name flags
		pos += 4
	}
	return buf
}

func buildImportTable(imports []pcc.Import) []byte {
	entrySize := 28
	buf := make([]byte, len(imports)*entrySize)
	for i, imp := range imports {
		pos := i * entrySize
		writeI32(buf, pos, 0)
		writeI32(buf, pos+4, 0)
		writeI32(buf, pos+8, imp.ClassNameIndex)
		writeI32(buf, pos+12, 0)
		writeI32(buf, pos+16, 0)
		writeI32(buf, pos+20, imp.ObjectNameIndex)
		writeI32(buf, pos+24, 0)
	}
	return buf
}

type exportTableMeta struct {
	entryOffset      int
	entrySize        int
	serialSizeOffset int
	serialOffsetOff  int
}

func buildExportTable(summary *pcc.FileSummary, rawData []byte) ([]byte, []exportTableMeta) {
	meta := make([]exportTableMeta, len(summary.Exports))

	var total int
	for i := range summary.Exports {
		size := exportEntrySize(summary.Header.UnrealVersion, summary.Header.LicenseeVersion)
		meta[i] = exportTableMeta{
			entryOffset:      total,
			entrySize:        size,
			serialSizeOffset: total + 32,
			serialOffsetOff:  total + 36,
		}
		total += size
	}

	buf := make([]byte, total)
	_ = rawData

	for i, exp := range summary.Exports {
		m := meta[i]
		writeI32(buf, m.entryOffset, exp.ClassIndex)
		writeI32(buf, m.entryOffset+4, 0)
		writeI32(buf, m.entryOffset+8, 0)
		writeI32(buf, m.entryOffset+12, exp.ObjectNameIndex)
		for j := 16; j < 32; j += 4 {
			writeI32(buf, m.entryOffset+j, 0)
		}
		writeI32(buf, m.entryOffset+32, exp.SerialSize)
		writeI32(buf, m.entryOffset+36, 0)

		writeI32(buf, m.entryOffset+40, 0)
		writeI32(buf, m.entryOffset+44, 0)
		writeI32(buf, m.entryOffset+48, 0)
		writeI32(buf, m.entryOffset+52, 0)
		writeI32(buf, m.entryOffset+56, 0)
		writeI32(buf, m.entryOffset+60, 0)
		writeI32(buf, m.entryOffset+64, 0)
		writeI32(buf, m.entryOffset+68, 0)
	}

	return buf, meta
}

func exportEntrySize(uv, lv int) int {
	size := 40
	if uv <= 512 {
		size += 4
	}
	size += 8 + 16 + 4
	return size
}

func buildDependTable(exports []pcc.Export) []byte {
	if len(exports) == 0 {
		return []byte{0, 0, 0, 0}
	}
	buf := make([]byte, len(exports)*4)
	for i := range exports {
		writeI32(buf, i*4, i)
	}
	return buf
}

func collectExportData(exports []pcc.Export, rawData []byte) ([]byte, error) {
	var total int
	for _, exp := range exports {
		total += exp.SerialSize
	}
	if total == 0 {
		return nil, nil
	}
	buf := make([]byte, total)
	pos := 0
	for _, exp := range exports {
		if exp.SerialSize > 0 {
			start := exp.SerialOffset
			end := start + exp.SerialSize
			if start < 0 || end > len(rawData) {
				return nil, fmt.Errorf("export %s serial data out of range: offset=%d size=%d len=%d",
					exp.ObjectName, start, exp.SerialSize, len(rawData))
			}
			copy(buf[pos:], rawData[start:end])
			pos += exp.SerialSize
		}
	}
	return buf, nil
}

func patchExportSerialOffsets(buf []byte, meta []exportTableMeta, tableBase, exportDataOffset int) {
	pos := exportDataOffset
	for _, m := range meta {
		writeI32(buf, tableBase+m.serialOffsetOff, pos)
		pos += int(int32(binary.LittleEndian.Uint32(buf[tableBase+m.serialSizeOffset:])))
	}
}

func writeI32(buf []byte, offset int, v int) {
	binary.LittleEndian.PutUint32(buf[offset:], uint32(int32(v)))
}

func writeU32(buf []byte, offset int, v uint32) {
	binary.LittleEndian.PutUint32(buf[offset:], v)
}
