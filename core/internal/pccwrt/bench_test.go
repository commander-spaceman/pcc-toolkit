package pccwrt

import (
	"fmt"
	"testing"

	pcc "github.com/commander-spaceman/me2pcc"

	"github.com/commander-spaceman/me2lzo/compress"
	"github.com/commander-spaceman/me2lzo/decompress"
)

func BenchmarkLZOCompress_1KB(b *testing.B) {
	data := make([]byte, 1024)
	for i := range data {
		data[i] = byte(i % 256)
	}
	b.ResetTimer()
	for b.Loop() {
		compress.Compress(data)
	}
}

func BenchmarkLZOCompress_128KB(b *testing.B) {
	data := make([]byte, pcc.MaxBlockSizeOT)
	for i := range data {
		data[i] = byte(i % 256)
	}
	b.ResetTimer()
	for b.Loop() {
		compress.Compress(data)
	}
}

func BenchmarkLZORoundTrip_128KB(b *testing.B) {
	data := make([]byte, pcc.MaxBlockSizeOT)
	for i := range data {
		data[i] = byte(i % 256)
	}
	var compressed []byte
	b.ResetTimer()
	for b.Loop() {
		compressed, _ = compress.Compress(data)
		decompress.Decompress(compressed, make([]byte, len(data)))
	}
}

func BenchmarkLZORoundTrip_1MB(b *testing.B) {
	size := 1024 * 1024
	data := make([]byte, size)
	for i := range data {
		data[i] = byte(i % 256)
	}
	b.ResetTimer()
	for b.Loop() {
		compressed, _ := compress.Compress(data)
		decompress.Decompress(compressed, make([]byte, len(data)))
	}
}

func BenchmarkWritePCC_Uncompressed_1MB(b *testing.B) {
	rawData, summary := buildBenchPCC(1024 * 1024)
	b.ResetTimer()
	for b.Loop() {
		buildUncompressedPCC(summary, rawData)
	}
}

func BenchmarkWritePCC_Compressed_1MB(b *testing.B) {
	rawData, summary := buildBenchPCC(1024 * 1024)
	b.ResetTimer()
	for b.Loop() {
		buildCompressedPCC(summary, rawData)
	}
}

func BenchmarkWritePCC_Compressed_5MB(b *testing.B) {
	rawData, summary := buildBenchPCC(5 * 1024 * 1024)
	b.ResetTimer()
	for b.Loop() {
		buildCompressedPCC(summary, rawData)
	}
}

func buildBenchPCC(exportSize int) ([]byte, *pcc.FileSummary) {
	names := make([]string, 50)
	names[0] = "None"
	for i := 1; i < 50; i++ {
		names[i] = fmt.Sprintf("Name_%d", i)
	}

	exports := make([]pcc.Export, 1)
	exports[0] = pcc.Export{
		Index:           0,
		ClassIndex:      0,
		ObjectNameIndex: 1,
		SerialSize:      exportSize,
		SerialOffset:    2000,
		ObjectName:      "BenchExport",
		ClassName:       "Object",
	}

	rawData := make([]byte, 2000+exportSize)
	for i := range rawData {
		rawData[i] = byte(i % 256)
	}

	imports := []pcc.Import{
		{ClassNameIndex: 1, ObjectNameIndex: 2},
	}

	return rawData, &pcc.FileSummary{
		Path:        "bench.pcc",
		GameProfile: pcc.ProfileME2OT,
		Compressed:  false,
		Header: pcc.Header{
			UnrealVersion:   512,
			LicenseeVersion: 130,
			Flags:           0,
			NameCount:       len(names),
			ExportCount:     1,
			ImportCount:     1,
		},
		Names:   names,
		Imports: imports,
		Exports: exports,
	}
}
