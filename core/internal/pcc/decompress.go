package pcc

import (
	"errors"
	"fmt"

	lzo "github.com/commander-spaceman/me2lzo/decompress"
)

func decompressME2OT(data []byte) ([]byte, error) {
	cursor, err := locateCompressionInfoOffsetME2OT(data)
	if err != nil {
		return nil, err
	}
	if cursor+8 > len(data) {
		return nil, errors.New("compression header out of range")
	}
	compressionType := readI32(data, cursor)
	numChunks := readI32(data, cursor+4)
	cursor += 8
	if compressionType != CompressionLZO {
		return nil, fmt.Errorf("unsupported compression type %d (expected LZO=%d)", compressionType, CompressionLZO)
	}
	if numChunks <= 0 {
		return nil, errors.New("invalid chunk count")
	}

	type chunkInfo struct {
		uncompressedOffset int
		uncompressedSize   int
		compressedOffset   int
		compressedSize     int
	}
	chunks := make([]chunkInfo, 0, numChunks)
	for i := 0; i < numChunks; i++ {
		if cursor+16 > len(data) {
			return nil, errors.New("chunk table out of range")
		}
		chunks = append(chunks, chunkInfo{
			uncompressedOffset: readI32(data, cursor),
			uncompressedSize:   readI32(data, cursor+4),
			compressedOffset:   readI32(data, cursor+8),
			compressedSize:     readI32(data, cursor+12),
		})
		cursor += 16
	}

	firstChunkOffset := chunks[0].uncompressedOffset
	maxEnd := 0
	for _, c := range chunks {
		if c.uncompressedOffset < firstChunkOffset {
			firstChunkOffset = c.uncompressedOffset
		}
		end := c.uncompressedOffset + c.uncompressedSize
		if end > maxEnd {
			maxEnd = end
		}
	}
	if firstChunkOffset < 0 || maxEnd <= 0 || firstChunkOffset > len(data) {
		return nil, errors.New("invalid chunk offsets")
	}

	type blockHeader struct {
		compressedSize   int
		uncompressedSize int
	}
	type chunkBlocks struct {
		uncompressedOffset int
		chunkBlob          []byte
		blockSize          int
		blockDataOffset    int
		blocks             []blockHeader
	}
	var parsedChunks []chunkBlocks
	maxUncompressedBlockSize := 0

	for _, c := range chunks {
		if c.compressedOffset < 0 || c.compressedOffset+c.compressedSize > len(data) {
			return nil, errors.New("compressed chunk out of range")
		}
		chunkBlob := data[c.compressedOffset : c.compressedOffset+c.compressedSize]
		if len(chunkBlob) < ChunkHeaderSize {
			return nil, errors.New("truncated chunk header")
		}
		magic := readU32(chunkBlob, 0)
		blockSize := readI32(chunkBlob, 4)
		compressedSizeHeader := readI32(chunkBlob, 8)
		uncompressedSizeHeader := readI32(chunkBlob, 12)
		if magic != ChunkHeaderMagic {
			return nil, errors.New("invalid chunk magic")
		}
		if uncompressedSizeHeader != c.uncompressedSize {
			return nil, errors.New("chunk size mismatch")
		}
		if compressedSizeHeader+ChunkHeaderSize > c.compressedSize {
			return nil, errors.New("truncated chunk payload")
		}
		if blockSize <= 0 {
			return nil, errors.New("invalid block size")
		}
		if blockSize > MaxBlockSizeOT {
			return nil, fmt.Errorf("block size %d exceeds max %d for ME2 OT", blockSize, MaxBlockSizeOT)
		}

		blockCount := uncompressedSizeHeader / blockSize
		if uncompressedSizeHeader%blockSize != 0 {
			blockCount++
		}
		blockTableOffset := ChunkHeaderSize
		blockDataOffset := blockTableOffset + (blockCount * ChunkBlockHeaderSize)
		if blockDataOffset > len(chunkBlob) {
			return nil, errors.New("invalid block table")
		}

		blocks := make([]blockHeader, blockCount)
		for i := 0; i < blockCount; i++ {
			blockHeaderOffset := blockTableOffset + (i * ChunkBlockHeaderSize)
			if blockHeaderOffset+8 > len(chunkBlob) {
				return nil, errors.New("block header out of range")
			}
			compSz := readI32(chunkBlob, blockHeaderOffset)
			uncompSz := readI32(chunkBlob, blockHeaderOffset+4)
			if compSz < 0 || uncompSz < 0 {
				return nil, errors.New("invalid block sizes")
			}
			if uncompSz > blockSize {
				return nil, errors.New("block uncompressed size exceeds block size")
			}
			blocks[i] = blockHeader{compressedSize: compSz, uncompressedSize: uncompSz}
			if uncompSz > maxUncompressedBlockSize {
				maxUncompressedBlockSize = uncompSz
			}
		}

		parsedChunks = append(parsedChunks, chunkBlocks{
			uncompressedOffset: c.uncompressedOffset,
			chunkBlob:          chunkBlob,
			blockSize:          blockSize,
			blockDataOffset:    blockDataOffset,
			blocks:             blocks,
		})
	}

	output := make([]byte, maxEnd)
	copy(output[:firstChunkOffset], data[:firstChunkOffset])

	reusableBuf := make([]byte, maxUncompressedBlockSize)

	for _, pc := range parsedChunks {
		writeOffset := pc.uncompressedOffset
		dataCursor := pc.blockDataOffset
		for _, bh := range pc.blocks {
			if dataCursor+bh.compressedSize > len(pc.chunkBlob) {
				return nil, errors.New("compressed block out of range")
			}
			compressedBlock := pc.chunkBlob[dataCursor : dataCursor+bh.compressedSize]
			dataCursor += bh.compressedSize

			written, decErr := lzo.Decompress(compressedBlock, reusableBuf)
			if decErr != nil {
				return nil, decErr
			}
			if written != bh.uncompressedSize {
				return nil, errors.New("decompressed block size mismatch")
			}
			endOffset := writeOffset + bh.uncompressedSize
			if endOffset > len(output) {
				return nil, errors.New("decompressed output out of range")
			}
			copy(output[writeOffset:endOffset], reusableBuf[:bh.uncompressedSize])
			writeOffset = endOffset
		}
	}

	return output, nil
}

func locateCompressionInfoOffsetME2OT(data []byte) (int, error) {
	cursor := 8
	if cursor+4 > len(data) {
		return 0, errors.New("truncated header")
	}
	cursor += 4
	folderLen := readI32(data, cursor)
	cursor += 4
	if folderLen > 0 {
		cursor += folderLen
	} else if folderLen < 0 {
		cursor += (-folderLen) * 2
	}
	if cursor+4 > len(data) {
		return 0, errors.New("truncated header folder")
	}
	cursor += 4  // flags (uint32)
	cursor += 24 // nameCount, nameOffset, exportCount, exportOffset, importCount, importOffset (6 x int32)
	cursor += 4  // dependsCount
	cursor += 16 // dependsOffset + padding? + guid start
	if cursor+4 > len(data) {
		return 0, errors.New("truncated generations")
	}
	generations := int(readU32(data, cursor))
	cursor += 4
	if generations > 0 {
		cursor += 12
		cursor += (generations - 1) * 12
	}
	cursor += 8  // engineVersion (uint32) + cookerVersion (uint32)
	cursor += 16 // compression types (uint32 x4: flags, fullHeaderSize, packageTag, packageGuid part?)
	cursor += 8  // packageSource + additionalPackagesToCook
	if cursor < 0 || cursor > len(data) {
		return 0, errors.New("compression info out of range")
	}
	return cursor, nil
}
