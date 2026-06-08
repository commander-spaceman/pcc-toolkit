package tlk

import (
	"encoding/binary"
	"fmt"
	"os"
	"sort"
)

func BuildCodeTable(nodes []Node) map[uint16]string {
	table := make(map[uint16]string)
	buildCodeTableRecursive(nodes, 0, "", table)
	return table
}

func buildCodeTableRecursive(nodes []Node, nodeIdx int32, prefix string, table map[uint16]string) {
	if nodeIdx < 0 || nodeIdx >= int32(len(nodes)) {
		return
	}
	node := nodes[nodeIdx]

	if node.LeftNodeID >= 0 {
		buildCodeTableRecursive(nodes, node.LeftNodeID, prefix+"0", table)
	} else {
		charCode := uint16((0xFFFF - int(node.LeftNodeID)) & 0xFFFF)
		table[charCode] = prefix + "0"
	}

	if node.RightNodeID >= 0 {
		buildCodeTableRecursive(nodes, node.RightNodeID, prefix+"1", table)
	} else {
		charCode := uint16((0xFFFF - int(node.RightNodeID)) & 0xFFFF)
		table[charCode] = prefix + "1"
	}
}

type bitWriter struct {
	bytes []byte
	pos   int
}

func newBitWriter() *bitWriter {
	return &bitWriter{bytes: make([]byte, 0, 1024)}
}

func (bw *bitWriter) writeBit(bit bool) {
	byteIdx := bw.pos >> 3
	bitIdx := bw.pos & 7
	for byteIdx >= len(bw.bytes) {
		bw.bytes = append(bw.bytes, 0)
	}
	if bit {
		bw.bytes[byteIdx] |= 1 << bitIdx
	}
	bw.pos++
}

func (bw *bitWriter) writeBits(code string) {
	for _, c := range code {
		bw.writeBit(c == '1')
	}
}

func (bw *bitWriter) bytesCopy() []byte {
	result := make([]byte, len(bw.bytes))
	copy(result, bw.bytes)
	return result
}

func EncodeString(text string, codeTable map[uint16]string) ([]byte, error) {
	bw := newBitWriter()

	for _, r := range text {
		code, ok := codeTable[uint16(r)]
		if !ok {
			return nil, fmt.Errorf("character %q (U+%04X) not found in code table", r, r)
		}
		bw.writeBits(code)
	}

	nullCode, ok := codeTable[0]
	if !ok {
		return nil, fmt.Errorf("null terminator not found in code table")
	}
	bw.writeBits(nullCode)

	for bw.pos&7 != 0 {
		bw.writeBit(false)
	}

	return bw.bytesCopy(), nil
}

type StringEntry struct {
	StringID int32
	Text     string
	Male     bool
}

func WriteFile(path string, tlkFile *File, newEntries []StringEntry) error {
	if err := AddEntries(tlkFile, newEntries); err != nil {
		return err
	}
	buf, err := BuildBytes(tlkFile, nil)
	if err != nil {
		return err
	}
	return os.WriteFile(path, buf, 0644)
}

func WriteFileBytes(tlkFile *File) ([]byte, error) {
	return BuildBytes(tlkFile, nil)
}

func BuildBytes(tlkFile *File, newEntries []StringEntry) ([]byte, error) {
	if tlkFile == nil {
		return nil, fmt.Errorf("tlk file is nil")
	}

	codeTable := BuildCodeTable(tlkFile.Nodes)

	maleCount := int(tlkFile.Header.MaleEntryCount)
	femaleCount := int(tlkFile.Header.FemaleEntryCount)

	var maleEntries, femaleEntries []Tuple
	for id, off := range tlkFile.MaleEntries {
		maleEntries = append(maleEntries, Tuple{StringID: id, BitOffset: off})
	}
	for id, off := range tlkFile.FemaleEntries {
		femaleEntries = append(femaleEntries, Tuple{StringID: id, BitOffset: off})
	}

	sort.Slice(maleEntries, func(i, j int) bool { return maleEntries[i].StringID < maleEntries[j].StringID })
	sort.Slice(femaleEntries, func(i, j int) bool { return femaleEntries[i].StringID < femaleEntries[j].StringID })

	newBitstream := make([]byte, len(tlkFile.Bits))
	copy(newBitstream, tlkFile.Bits)
	bitstreamPos := len(newBitstream) * 8

	for _, entry := range newEntries {
		id := entry.StringID
		isNew := true
		for _, e := range maleEntries {
			if e.StringID == id {
				isNew = false
				break
			}
		}
		for _, e := range femaleEntries {
			if e.StringID == id {
				isNew = false
				break
			}
		}
		if !isNew {
			continue
		}

		encoded, err := EncodeString(entry.Text, codeTable)
		if err != nil {
			return nil, fmt.Errorf("encode string ID %d: %w", id, err)
		}

		newTuple := Tuple{StringID: id, BitOffset: int32(bitstreamPos)}
		if entry.Male {
			maleEntries = append(maleEntries, newTuple)
			maleCount++
		} else {
			femaleEntries = append(femaleEntries, newTuple)
			femaleCount++
		}

		newBitstream = append(newBitstream, encoded...)
		bitstreamPos += len(encoded) * 8
	}

	sort.Slice(maleEntries, func(i, j int) bool { return maleEntries[i].StringID < maleEntries[j].StringID })
	sort.Slice(femaleEntries, func(i, j int) bool { return femaleEntries[i].StringID < femaleEntries[j].StringID })

	entryTableSize := (maleCount + femaleCount) * 8
	treeSize := len(tlkFile.Nodes) * 8
	headerSize := 28
	totalSize := headerSize + entryTableSize + treeSize + len(newBitstream)

	buf := make([]byte, totalSize)

	binary.LittleEndian.PutUint32(buf[0:4], uint32(TLKMagic))
	binary.LittleEndian.PutUint32(buf[4:8], uint32(tlkFile.Header.Version))
	binary.LittleEndian.PutUint32(buf[8:12], uint32(tlkFile.Header.MinVersion))
	binary.LittleEndian.PutUint32(buf[12:16], uint32(maleCount))
	binary.LittleEndian.PutUint32(buf[16:20], uint32(femaleCount))
	binary.LittleEndian.PutUint32(buf[20:24], uint32(len(tlkFile.Nodes)))
	binary.LittleEndian.PutUint32(buf[24:28], uint32(len(newBitstream)))

	pos := headerSize
	allEntries := append(maleEntries, femaleEntries...)
	for _, e := range allEntries {
		binary.LittleEndian.PutUint32(buf[pos:], uint32(e.StringID))
		binary.LittleEndian.PutUint32(buf[pos+4:], uint32(e.BitOffset))
		pos += 8
	}

	for _, node := range tlkFile.Nodes {
		binary.LittleEndian.PutUint32(buf[pos:], uint32(node.LeftNodeID))
		binary.LittleEndian.PutUint32(buf[pos+4:], uint32(node.RightNodeID))
		pos += 8
	}

	copy(buf[pos:], newBitstream)

	return buf, nil
}

func AddEntries(tlkFile *File, entries []StringEntry) error {
	if tlkFile == nil {
		return fmt.Errorf("tlk file is nil")
	}

	codeTable := BuildCodeTable(tlkFile.Nodes)

	newBitstream := make([]byte, len(tlkFile.Bits))
	copy(newBitstream, tlkFile.Bits)
	bitstreamPos := len(newBitstream) * 8

	addedMales := int32(0)
	addedFemales := int32(0)

	for _, entry := range entries {
		if _, ok := tlkFile.MaleEntries[entry.StringID]; ok {
			continue
		}
		if _, ok := tlkFile.FemaleEntries[entry.StringID]; ok {
			continue
		}

		encoded, err := EncodeString(entry.Text, codeTable)
		if err != nil {
			return fmt.Errorf("encode string ID %d: %w", entry.StringID, err)
		}

		bitOffset := int32(bitstreamPos)
		target := tlkFile.MaleEntries
		if !entry.Male {
			target = tlkFile.FemaleEntries
			addedFemales++
		} else {
			addedMales++
		}
		target[entry.StringID] = bitOffset

		newBitstream = append(newBitstream, encoded...)
		bitstreamPos += len(encoded) * 8
	}

	tlkFile.Bits = newBitstream
	tlkFile.Header.DataLen = int32(len(newBitstream))
	tlkFile.Header.MaleEntryCount += addedMales
	tlkFile.Header.FemaleEntryCount += addedFemales
	tlkFile.TotalEntries += int(addedMales + addedFemales)

	return nil
}
