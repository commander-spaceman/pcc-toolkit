package tlk

import (
	"encoding/binary"
	"errors"
	"fmt"
	"os"
)

func ReadFile(path string) (*File, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read tlk file: %w", err)
	}
	return Parse(data, path)
}

func Parse(data []byte, path string) (*File, error) {
	if len(data) < 28 {
		return nil, errors.New("tlk file too small for header")
	}

	header := Header{
		Magic:            int32(binary.LittleEndian.Uint32(data[0:4])),
		Version:          int32(binary.LittleEndian.Uint32(data[4:8])),
		MinVersion:       int32(binary.LittleEndian.Uint32(data[8:12])),
		MaleEntryCount:   int32(binary.LittleEndian.Uint32(data[12:16])),
		FemaleEntryCount: int32(binary.LittleEndian.Uint32(data[16:20])),
		TreeNodeCount:    int32(binary.LittleEndian.Uint32(data[20:24])),
		DataLen:          int32(binary.LittleEndian.Uint32(data[24:28])),
	}

	if header.Magic != TLKMagic {
		return nil, fmt.Errorf("invalid TLK magic: 0x%08X", header.Magic)
	}
	if header.MaleEntryCount < 0 || header.FemaleEntryCount < 0 ||
		header.TreeNodeCount < 0 || header.DataLen < 0 {
		return nil, errors.New("negative count in TLK header")
	}

	cursor := 28

	entrySize := int(header.MaleEntryCount+header.FemaleEntryCount) * 8
	if cursor+entrySize > len(data) {
		return nil, errors.New("tlk entry table out of range")
	}

	maleEntries := make(map[int32]int32, header.MaleEntryCount)
	for i := int32(0); i < header.MaleEntryCount; i++ {
		stringID := int32(binary.LittleEndian.Uint32(data[cursor : cursor+4]))
		bitOffset := int32(binary.LittleEndian.Uint32(data[cursor+4 : cursor+8]))
		if bitOffset >= 0 {
			maleEntries[stringID] = bitOffset
		}
		cursor += 8
	}

	femaleEntries := make(map[int32]int32, header.FemaleEntryCount)
	for i := int32(0); i < header.FemaleEntryCount; i++ {
		stringID := int32(binary.LittleEndian.Uint32(data[cursor : cursor+4]))
		bitOffset := int32(binary.LittleEndian.Uint32(data[cursor+4 : cursor+8]))
		if bitOffset >= 0 {
			femaleEntries[stringID] = bitOffset
		}
		cursor += 8
	}

	treeSize := int(header.TreeNodeCount) * 8
	if cursor+treeSize > len(data) {
		return nil, errors.New("tlk huffman tree out of range")
	}

	nodes := make([]Node, header.TreeNodeCount)
	for i := int32(0); i < header.TreeNodeCount; i++ {
		nodes[i] = Node{
			LeftNodeID:  int32(binary.LittleEndian.Uint32(data[cursor : cursor+4])),
			RightNodeID: int32(binary.LittleEndian.Uint32(data[cursor+4 : cursor+8])),
		}
		cursor += 8
	}

	dataEnd := cursor + int(header.DataLen)
	if dataEnd > len(data) {
		return nil, errors.New("tlk bitstream out of range")
	}
	bits := data[cursor:dataEnd]

	return &File{
		Path:          path,
		Header:        header,
		MaleEntries:   maleEntries,
		FemaleEntries: femaleEntries,
		Nodes:         nodes,
		Bits:          bits,
		TotalEntries:  len(maleEntries) + len(femaleEntries),
	}, nil
}

func getBit(data []byte, index int) bool {
	if index < 0 {
		return false
	}
	byteIndex := index >> 3
	if byteIndex >= len(data) {
		return false
	}
	bitIndex := index & 7
	return (data[byteIndex] & (1 << bitIndex)) != 0
}

func DecodeString(bits []byte, nodes []Node, bitOffset int32) (string, bool) {
	if len(nodes) == 0 || bitOffset < 0 {
		return "", false
	}
	root := nodes[0]
	current := root
	var chars []byte

	maxBits := len(bits) * 8
	for i := int(bitOffset); i < maxBits; i++ {
		var nextNodeID int32
		if getBit(bits, i) {
			nextNodeID = current.RightNodeID
		} else {
			nextNodeID = current.LeftNodeID
		}

		if nextNodeID >= 0 {
			if nextNodeID >= int32(len(nodes)) {
				return "", false
			}
			current = nodes[nextNodeID]
			continue
		}

		charCode := (0xFFFF - nextNodeID) & 0xFFFF
		c := byte(charCode)
		if c == 0 {
			return string(chars), true
		}
		chars = append(chars, c)
		current = root
	}
	return "", false
}

func ResolveString(tlk *File, stringID int32, male bool) (string, bool) {
	entries := tlk.MaleEntries
	if !male {
		entries = tlk.FemaleEntries
	}
	bitOffset, ok := entries[stringID]
	if !ok || bitOffset < 0 {
		return "", false
	}
	return DecodeString(tlk.Bits, tlk.Nodes, bitOffset)
}

func (f *File) StringIDs() []int32 {
	ids := make([]int32, 0, len(f.MaleEntries))
	for id := range f.MaleEntries {
		ids = append(ids, id)
	}
	for id := range f.FemaleEntries {
		ids = append(ids, id)
	}
	return ids
}

func (f *File) IterEntries() func(func(int32, string) bool) {
	return func(yield func(int32, string) bool) {
		for id := range f.MaleEntries {
			text, ok := ResolveString(f, id, true)
			if !ok {
				continue
			}
			if !yield(id, text) {
				return
			}
		}
		for id := range f.FemaleEntries {
			text, ok := ResolveString(f, id, false)
			if !ok {
				continue
			}
			if !yield(id, text) {
				return
			}
		}
	}
}

func (f *File) Search(query string) []Entry {
	var results []Entry
	f.IterEntries()(func(id int32, text string) bool {
		if containsFold(text, query) {
			results = append(results, Entry{StringID: id, Text: text})
		}
		return true
	})
	return results
}
