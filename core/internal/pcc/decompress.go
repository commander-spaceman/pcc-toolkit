package pcc

import (
	"errors"

	lzo "github.com/anchore/go-lzo"
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
		return nil, errors.New("unsupported compression type")
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

	output := make([]byte, maxEnd)
	copy(output[:firstChunkOffset], data[:firstChunkOffset])

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

		blockCount := uncompressedSizeHeader / blockSize
		if uncompressedSizeHeader%blockSize != 0 {
			blockCount++
		}
		blockTableOffset := ChunkHeaderSize
		blockDataOffset := blockTableOffset + (blockCount * ChunkBlockHeaderSize)
		if blockDataOffset > len(chunkBlob) {
			return nil, errors.New("invalid block table")
		}

		writeOffset := c.uncompressedOffset
		dataCursor := blockDataOffset
		for i := 0; i < blockCount; i++ {
			blockHeaderOffset := blockTableOffset + (i * ChunkBlockHeaderSize)
			if blockHeaderOffset+8 > len(chunkBlob) {
				return nil, errors.New("block header out of range")
			}
			blockCompressedSize := readI32(chunkBlob, blockHeaderOffset)
			blockUncompressedSize := readI32(chunkBlob, blockHeaderOffset+4)
			if blockCompressedSize < 0 || blockUncompressedSize < 0 {
				return nil, errors.New("invalid block sizes")
			}
			if dataCursor+blockCompressedSize > len(chunkBlob) {
				return nil, errors.New("compressed block out of range")
			}

			compressedBlock := chunkBlob[dataCursor : dataCursor+blockCompressedSize]
			dataCursor += blockCompressedSize
			decompressedBlock := make([]byte, blockUncompressedSize)
			written, decErr := lzo.Decompress(compressedBlock, decompressedBlock)
			if decErr != nil {
				return nil, decErr
			}
			if written != blockUncompressedSize {
				return nil, errors.New("decompressed block size mismatch")
			}
			endOffset := writeOffset + blockUncompressedSize
			if endOffset > len(output) {
				return nil, errors.New("decompressed output out of range")
			}
			copy(output[writeOffset:endOffset], decompressedBlock)
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
	cursor += 4
	cursor += 24
	cursor += 4
	cursor += 16
	if cursor+4 > len(data) {
		return 0, errors.New("truncated generations")
	}
	generations := int(readU32(data, cursor))
	cursor += 4
	if generations > 0 {
		cursor += 12
		cursor += (generations - 1) * 12
	}
	cursor += 8
	cursor += 16
	cursor += 8
	if cursor < 0 || cursor > len(data) {
		return 0, errors.New("compression info out of range")
	}
	return cursor, nil
}
