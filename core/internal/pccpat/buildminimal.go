package pccpat

import (
	"encoding/binary"

	pcc "github.com/commander-spaceman/me2pcc"
)

func BuildMinimalPCC(exportSerialData [][]byte) ([]byte, *pcc.FileSummary) {
	names := []string{
		"None", "Class", "Core", "Object", "BioConversation",
		"entry_0", "entry_1", "entry_2",
	}
	imports := []pcc.Import{
		{ClassNameIndex: 1, ObjectNameIndex: 2},
	}

	var nameBytes []byte
	for _, n := range names {
		nb := make([]byte, 4+len(n)+1)
		binary.LittleEndian.PutUint32(nb[0:4], uint32(len(n)+1))
		copy(nb[4:], n)
		nb[4+len(n)] = 0
		flags := make([]byte, 4)
		binary.LittleEndian.PutUint32(flags, 0xFFFFFFF2)
		nb = append(nb, flags...)
		nameBytes = append(nameBytes, nb...)
	}

	importOffset := 60 + len(nameBytes)
	exportOffset := importOffset + (len(imports) * 28)

	type exportBinEntry struct {
		data   []byte
		offset int
		size   int
	}
	exportEntries := make([]exportBinEntry, len(exportSerialData))

	exportTableSize := 0
	for range exportSerialData {
		entrySize := 40 + 4 + 8 + 16 + 4
		exportTableSize += entrySize
	}

	serialStart := exportOffset + exportTableSize
	cursor := serialStart
	for i, data := range exportSerialData {
		exportEntries[i] = exportBinEntry{
			data:   data,
			offset: cursor,
			size:   len(data),
		}
		cursor += len(data)
	}

	headerSize := 60
	buf := make([]byte, cursor)
	binary.LittleEndian.PutUint32(buf[0:], pcc.PackageMagic)
	binary.LittleEndian.PutUint32(buf[4:], uint32(512|(130<<16)))
	binary.LittleEndian.PutUint32(buf[8:], 0)
	binary.LittleEndian.PutUint32(buf[12:], 0)
	binary.LittleEndian.PutUint32(buf[16:], 0)
	binary.LittleEndian.PutUint32(buf[20:], uint32(len(names)))
	binary.LittleEndian.PutUint32(buf[24:], uint32(headerSize))
	binary.LittleEndian.PutUint32(buf[28:], uint32(len(exportSerialData)))
	binary.LittleEndian.PutUint32(buf[32:], uint32(exportOffset))
	binary.LittleEndian.PutUint32(buf[36:], uint32(len(imports)))
	binary.LittleEndian.PutUint32(buf[40:], uint32(importOffset))
	binary.LittleEndian.PutUint32(buf[44:], uint32(len(exportSerialData)))
	binary.LittleEndian.PutUint32(buf[48:], 0)
	binary.LittleEndian.PutUint32(buf[52:], 0)
	binary.LittleEndian.PutUint32(buf[56:], 0)

	copy(buf[headerSize:], nameBytes)

	impCursor := importOffset
	for _, imp := range imports {
		writeI32(buf, impCursor, 0)
		writeI32(buf, impCursor+8, imp.ClassNameIndex)
		writeI32(buf, impCursor+20, imp.ObjectNameIndex)
		impCursor += 28
	}

	expCursor := exportOffset
	for i, ed := range exportEntries {
		writeI32(buf, expCursor, 3)
		writeI32(buf, expCursor+12, 5+i)
		writeI32(buf, expCursor+32, ed.size)
		writeI32(buf, expCursor+36, ed.offset)
		entrySize := 40 + 4 + 8 + 16 + 4
		expCursor += entrySize
	}

	for _, ed := range exportEntries {
		copy(buf[ed.offset:], ed.data)
	}

	exports := make([]pcc.Export, len(exportEntries))
	for i, ed := range exportEntries {
		exports[i] = pcc.Export{
			Index:           i,
			ClassIndex:      3,
			ObjectNameIndex: 5 + i,
			SerialSize:      ed.size,
			SerialOffset:    ed.offset,
			ObjectName:      names[5+i],
			ClassName:       "Object",
		}
	}

	return buf, &pcc.FileSummary{
		Path:        "test.pcc",
		GameProfile: pcc.ProfileME2OT,
		Compressed:  false,
		Header: pcc.Header{
			UnrealVersion:   512,
			LicenseeVersion: 130,
			Flags:           0,
			NameCount:       len(names),
			NameOffset:      headerSize,
			ExportCount:     len(exports),
			ExportOffset:    exportOffset,
			ImportCount:     len(imports),
			ImportOffset:    importOffset,
		},
		Names:   names,
		Imports: imports,
		Exports: exports,
	}
}
