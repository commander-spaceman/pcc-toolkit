package pccwrt

import (
	"encoding/binary"
	"fmt"

	"pcc-toolkit/core/internal/pcc"

	"github.com/commander-spaceman/me2lzo/compress"
)

const maxChunkSize = 0x100000

type chunkPlan struct {
	uncompressedData   []byte
	uncompressedOffset int
	compressedBlocks   [][]byte
	blockSize          int
}

func buildCompressedPCC(summary *pcc.FileSummary, rawData []byte) ([]byte, error) {
	uncompressed, err := buildUncompressedBuffer(summary, rawData)
	if err != nil {
		return nil, err
	}

	chunks := divideIntoChunks(uncompressed)

	for i := range chunks {
		chunks[i].blockSize = pcc.MaxBlockSizeOT
		if err := compressChunk(&chunks[i]); err != nil {
			return nil, fmt.Errorf("compress chunk %d: %w", i, err)
		}
	}

	type chunkInfo struct {
		uncompressedOffset int
		uncompressedSize   int
		compressedOffset   int
		compressedSize     int
	}

	var chunkBlobs [][]byte
	chunkInfos := make([]chunkInfo, len(chunks))

	headerBaseSize := 100
	compressionInfoSize := 8 + len(chunks)*16
	chunkDataOffset := headerBaseSize + compressionInfoSize

	compOffset := chunkDataOffset
	for i := range chunks {
		blob := buildChunkBlob(&chunks[i])
		chunkBlobs = append(chunkBlobs, blob)

		chunkInfos[i] = chunkInfo{
			uncompressedOffset: chunks[i].uncompressedOffset,
			uncompressedSize:   len(chunks[i].uncompressedData),
			compressedOffset:   compOffset,
			compressedSize:     len(blob),
		}
		compOffset += len(blob)
	}

	totalSize := compOffset
	buf := make([]byte, totalSize)

	copy(buf[:headerBaseSize], uncompressed[:headerBaseSize])
	flags := binary.LittleEndian.Uint32(buf[16:20])
	flags |= pcc.CompressedFlag
	binary.LittleEndian.PutUint32(buf[16:20], flags)

	pos := headerBaseSize
	binary.LittleEndian.PutUint32(buf[pos:], pcc.CompressionLZO)
	pos += 4
	binary.LittleEndian.PutUint32(buf[pos:], uint32(len(chunks)))
	pos += 4

	for _, ci := range chunkInfos {
		binary.LittleEndian.PutUint32(buf[pos:], uint32(int32(ci.uncompressedOffset)))
		binary.LittleEndian.PutUint32(buf[pos+4:], uint32(int32(ci.uncompressedSize)))
		binary.LittleEndian.PutUint32(buf[pos+8:], uint32(int32(ci.compressedOffset)))
		binary.LittleEndian.PutUint32(buf[pos+12:], uint32(int32(ci.compressedSize)))
		pos += 16
	}

	for _, blob := range chunkBlobs {
		copy(buf[pos:], blob)
		pos += len(blob)
	}

	return buf, nil
}

func divideIntoChunks(uncompressed []byte) []chunkPlan {
	var chunks []chunkPlan

	chunks = append(chunks, chunkPlan{
		uncompressedData:   uncompressed,
		uncompressedOffset: 0,
	})
	return chunks
}

func compressChunk(ch *chunkPlan) error {
	blockSize := ch.blockSize
	if blockSize <= 0 {
		blockSize = pcc.MaxBlockSizeOT
	}

	data := ch.uncompressedData
	pos := 0

	for pos < len(data) {
		blockEnd := pos + blockSize
		if blockEnd > len(data) {
			blockEnd = len(data)
		}
		block := data[pos:blockEnd]

		compressed, err := compress.Compress(block)
		if err != nil {
			return fmt.Errorf("compress at offset %d: %w", pos, err)
		}

		pos = blockEnd
		ch.compressedBlocks = append(ch.compressedBlocks, compressed)
	}

	return nil
}

func buildChunkBlob(ch *chunkPlan) []byte {
	blockCount := len(ch.compressedBlocks)
	blockTableSize := blockCount * pcc.ChunkBlockHeaderSize

	uncompressedSize := len(ch.uncompressedData)

	var totalCompressed int
	for _, b := range ch.compressedBlocks {
		totalCompressed += len(b)
	}

	blobSize := pcc.ChunkHeaderSize + blockTableSize + totalCompressed
	blob := make([]byte, blobSize)

	binary.LittleEndian.PutUint32(blob[0:], pcc.ChunkHeaderMagic)
	binary.LittleEndian.PutUint32(blob[4:], uint32(int32(ch.blockSize)))
	binary.LittleEndian.PutUint32(blob[8:], uint32(int32(totalCompressed+blockTableSize)))
	binary.LittleEndian.PutUint32(blob[12:], uint32(int32(uncompressedSize)))

	btPos := pcc.ChunkHeaderSize
	dataPos := btPos + blockTableSize

	for i, block := range ch.compressedBlocks {
		blockUncompSize := ch.blockSize
		if i == blockCount-1 && uncompressedSize%ch.blockSize != 0 {
			blockUncompSize = uncompressedSize % ch.blockSize
		}
		if blockUncompSize == 0 {
			blockUncompSize = ch.blockSize
		}

		binary.LittleEndian.PutUint32(blob[btPos:], uint32(int32(len(block))))
		binary.LittleEndian.PutUint32(blob[btPos+4:], uint32(int32(blockUncompSize)))
		btPos += pcc.ChunkBlockHeaderSize

		copy(blob[dataPos:], block)
		dataPos += len(block)
	}

	return blob
}

func buildUncompressedBuffer(summary *pcc.FileSummary, rawData []byte) ([]byte, error) {
	nameBytes := buildNameTable(summary.Names)
	importBytes := buildImportTable(summary.Imports)
	exportTableBytes, expMeta := buildExportTable(summary, rawData)
	dependsBytes := buildDependTable(summary.Exports)

	exportData, err := collectExportData(summary.Exports, rawData)
	if err != nil {
		return nil, err
	}

	headerSize := 100

	nameOffset := headerSize
	importOffset := nameOffset + len(nameBytes)
	exportOffset := importOffset + len(importBytes)
	dependsOffset := exportOffset + len(exportTableBytes)
	exportDataOff := dependsOffset + len(dependsBytes)

	totalSize := exportDataOff + len(exportData)
	buf := make([]byte, totalSize)

	flags := summary.Header.Flags & ^uint32(pcc.CompressedFlag)
	writeHeader(buf, flags, len(summary.Names), nameOffset,
		len(summary.Exports), exportOffset,
		len(summary.Imports), importOffset,
		len(summary.Exports), dependsOffset)

	copy(buf[nameOffset:], nameBytes)
	copy(buf[importOffset:], importBytes)
	copy(buf[exportOffset:], exportTableBytes)
	copy(buf[dependsOffset:], dependsBytes)
	copy(buf[exportDataOff:], exportData)

	patchExportSerialOffsets(buf, expMeta, exportOffset, exportDataOff)

	return buf, nil
}
