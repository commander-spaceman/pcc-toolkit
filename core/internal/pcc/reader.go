package pcc

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

func readFileNormalize(path string) (string, error) {
	abs, err := filepath.Abs(path)
	if err != nil {
		return "", fmt.Errorf("resolve path: %w", err)
	}
	return filepath.Clean(abs), nil
}

func ReadFile(path string) (*FileSummary, error) {
	normPath, err := readFileNormalize(path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(normPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}
	return ReadFileFromBytes(data, normPath)
}

func ReadFileHeaderOnly(path string) (*FileSummary, error) {
	normPath, err := readFileNormalize(path)
	if err != nil {
		return nil, err
	}
	data, err := os.ReadFile(normPath)
	if err != nil {
		return nil, fmt.Errorf("read file: %w", err)
	}

	header, err := parsePccHeader(data)
	if err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}

	profile := InferGameProfile(header.UnrealVersion, header.LicenseeVersion)
	compressed := header.Flags&CompressedFlag != 0

	if compressed {
		decompressed, decErr := decompressME2OT(data)
		if decErr != nil {
			return nil, fmt.Errorf("decompress: %w", decErr)
		}
		header, err = parsePccHeader(decompressed)
		if err != nil {
			return nil, fmt.Errorf("parse header after decompress: %w", err)
		}
		profile = InferGameProfile(header.UnrealVersion, header.LicenseeVersion)
	}

	return &FileSummary{
		Path:        normPath,
		GameProfile: profile,
		Compressed:  compressed,
		Header:      header,
		Exports:     nil,
	}, nil
}

func ReadFileFromBytes(data []byte, path string) (*FileSummary, error) {
	header, err := parsePccHeader(data)
	if err != nil {
		return nil, fmt.Errorf("parse header: %w", err)
	}

	profile := InferGameProfile(header.UnrealVersion, header.LicenseeVersion)
	compressed := header.Flags&CompressedFlag != 0

	if compressed {
		decompressed, decErr := decompressME2OT(data)
		if decErr != nil {
			return nil, fmt.Errorf("decompress: %w", decErr)
		}
		data = decompressed
		header, err = parsePccHeader(data)
		if err != nil {
			return nil, fmt.Errorf("parse header after decompress: %w", err)
		}
	}

	names, err := parsePccNames(data, header)
	if err != nil {
		return nil, fmt.Errorf("parse names: %w", err)
	}

	imports, err := parsePccImports(data, header)
	if err != nil {
		return nil, fmt.Errorf("parse imports: %w", err)
	}

	exports, err := parsePccExports(data, header)
	if err != nil {
		return nil, fmt.Errorf("parse exports: %w", err)
	}

	resolveExportNames(exports, imports, names)

	return &FileSummary{
		Path:        path,
		GameProfile: profile,
		Compressed:  compressed,
		Header:      header,
		Names:       names,
		Imports:     imports,
		Exports:     exports,
	}, nil
}

func ReadFileRaw(path string) ([]byte, *FileSummary, error) {
	normPath, err := readFileNormalize(path)
	if err != nil {
		return nil, nil, err
	}
	data, err := os.ReadFile(normPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read file: %w", err)
	}
	return ReadFileRawFromBytes(data, normPath)
}

func ReadFileRawFromBytes(data []byte, path string) ([]byte, *FileSummary, error) {
	header, err := parsePccHeader(data)
	if err != nil {
		return nil, nil, fmt.Errorf("parse header: %w", err)
	}

	profile := InferGameProfile(header.UnrealVersion, header.LicenseeVersion)
	compressed := header.Flags&CompressedFlag != 0

	if compressed {
		decompressed, decErr := decompressME2OT(data)
		if decErr != nil {
			return nil, nil, fmt.Errorf("decompress: %w", decErr)
		}
		data = decompressed
		header, err = parsePccHeader(data)
		if err != nil {
			return nil, nil, fmt.Errorf("parse header after decompress: %w", err)
		}
	}

	names, err := parsePccNames(data, header)
	if err != nil {
		return nil, nil, fmt.Errorf("parse names: %w", err)
	}

	imports, err := parsePccImports(data, header)
	if err != nil {
		return nil, nil, fmt.Errorf("parse imports: %w", err)
	}

	exports, err := parsePccExports(data, header)
	if err != nil {
		return nil, nil, fmt.Errorf("parse exports: %w", err)
	}

	resolveExportNames(exports, imports, names)

	return data, &FileSummary{
		Path:        path,
		GameProfile: profile,
		Compressed:  compressed,
		Header:      header,
		Names:       names,
		Imports:     imports,
		Exports:     exports,
	}, nil
}

func parsePccHeader(data []byte) (Header, error) {
	if len(data) < 44 {
		return Header{}, errors.New("file too small")
	}
	if readU32(data, 0) != PackageMagic {
		return Header{}, errors.New("invalid magic")
	}
	versionPack := readU32(data, 4)
	unrealVersion := int(versionPack & 0xFFFF)
	licenseeVersion := int((versionPack >> 16) & 0xFFFF)
	cursor := 8
	cursor += 4
	folderLen := readI32(data, cursor)
	cursor += 4
	if folderLen > 0 {
		cursor += folderLen
	} else if folderLen < 0 {
		cursor += (-folderLen) * 2
	}
	if cursor < 0 || cursor+28 > len(data) {
		return Header{}, errors.New("truncated header")
	}
	flags := readU32(data, cursor)
	cursor += 4
	nameCount := readI32(data, cursor)
	cursor += 4
	nameOffset := readI32(data, cursor)
	cursor += 4
	exportCount := readI32(data, cursor)
	cursor += 4
	exportOffset := readI32(data, cursor)
	cursor += 4
	importCount := readI32(data, cursor)
	cursor += 4
	importOffset := readI32(data, cursor)
	if nameCount < 0 || exportCount < 0 || importCount < 0 {
		return Header{}, errors.New("negative table count")
	}
	return Header{
		UnrealVersion:   unrealVersion,
		LicenseeVersion: licenseeVersion,
		Flags:           flags,
		NameCount:       nameCount,
		NameOffset:      nameOffset,
		ExportCount:     exportCount,
		ExportOffset:    exportOffset,
		ImportCount:     importCount,
		ImportOffset:    importOffset,
	}, nil
}

func parsePccNames(data []byte, header Header) ([]string, error) {
	names := make([]string, 0, header.NameCount)
	cursor := header.NameOffset
	for i := 0; i < header.NameCount; i++ {
		text, next, err := readUnrealString(data, cursor)
		if err != nil {
			return nil, err
		}
		cursor = next
		if header.UnrealVersion <= 491 {
			cursor += 8
		} else {
			cursor += 4
		}
		if cursor > len(data) {
			return nil, errors.New("name table out of range")
		}
		names = append(names, text)
	}
	return names, nil
}

func parsePccImports(data []byte, header Header) ([]Import, error) {
	imports := make([]Import, 0, header.ImportCount)
	cursor := header.ImportOffset
	for i := 0; i < header.ImportCount; i++ {
		if cursor+28 > len(data) {
			return nil, errors.New("truncated import table")
		}
		imports = append(imports, Import{
			ClassNameIndex:  readI32(data, cursor+8),
			ObjectNameIndex: readI32(data, cursor+20),
		})
		cursor += 28
	}
	return imports, nil
}

func parsePccExports(data []byte, header Header) ([]Export, error) {
	exports := make([]Export, 0, header.ExportCount)
	cursor := header.ExportOffset
	for i := 0; i < header.ExportCount; i++ {
		if cursor+40 > len(data) {
			return nil, errors.New("truncated export table")
		}

		classIndex := readI32(data, cursor)
		objectNameIndex := readI32(data, cursor+12)
		serialSize := readI32(data, cursor+32)
		serialOffset := readI32(data, cursor+36)

		headerLen := 40
		if header.UnrealVersion <= 512 {
			if cursor+44 > len(data) {
				return nil, errors.New("truncated export component map")
			}
			componentCount := readI32(data, cursor+40)
			if componentCount < 0 {
				return nil, errors.New("negative component map count")
			}
			headerLen += 4 + (componentCount * 12)
		}

		if cursor+headerLen+8 > len(data) {
			return nil, errors.New("truncated export generation header")
		}
		generationCount := readI32(data, cursor+headerLen+4)
		if generationCount < 0 {
			return nil, errors.New("negative generation count")
		}
		headerLen += 8 + (generationCount * 4) + 16
		if !(header.UnrealVersion == 491 && header.LicenseeVersion <= 110) {
			headerLen += 4
		}
		if cursor+headerLen > len(data) {
			return nil, errors.New("truncated export footer")
		}

		exports = append(exports, Export{
			Index:           i,
			ClassIndex:      classIndex,
			ObjectNameIndex: objectNameIndex,
			SerialSize:      serialSize,
			SerialOffset:    serialOffset,
		})
		cursor += headerLen
	}
	return exports, nil
}

func resolveExportNames(exports []Export, imports []Import, names []string) {
	for i := range exports {
		exports[i].ObjectName = resolveName(exports[i].ObjectNameIndex, names)
		classIndex := exports[i].ClassIndex
		if classIndex > 0 {
			idx := classIndex - 1
			if idx >= 0 && idx < len(exports) {
				exports[i].ClassName = resolveName(exports[idx].ObjectNameIndex, names)
			}
		} else if classIndex < 0 {
			idx := (-classIndex) - 1
			if idx >= 0 && idx < len(imports) {
				exports[i].ClassName = resolveName(imports[idx].ObjectNameIndex, names)
			}
		}
	}
}

func ReadFileExports(path string, predicate func(Export) bool) (*FileSummary, map[int][]byte, error) {
	normPath, err := readFileNormalize(path)
	if err != nil {
		return nil, nil, err
	}
	data, err := os.ReadFile(normPath)
	if err != nil {
		return nil, nil, fmt.Errorf("read file: %w", err)
	}
	return ReadFileFromBytesAndExports(data, normPath, predicate)
}

func ReadFileFromBytesAndExports(data []byte, path string, predicate func(Export) bool) (*FileSummary, map[int][]byte, error) {
	rawData, summary, err := ReadFileRawFromBytes(data, path)
	if err != nil {
		return nil, nil, err
	}

	exportData := make(map[int][]byte)
	for i := range summary.Exports {
		exp := &summary.Exports[i]
		if !predicate(*exp) {
			continue
		}
		if exp.SerialSize < 0 || exp.SerialOffset < 0 ||
			exp.SerialOffset+exp.SerialSize > len(rawData) {
			continue
		}
		buf := make([]byte, exp.SerialSize)
		copy(buf, rawData[exp.SerialOffset:exp.SerialOffset+exp.SerialSize])
		exportData[exp.Index] = buf
	}

	return summary, exportData, nil
}
