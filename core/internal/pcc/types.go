package pcc

import "fmt"

const PackageMagic = 0x9E2A83C1
const CompressedFlag = 0x02000000
const CompressionLZO = 0x2
const ChunkHeaderMagic = 0x9E2A83C1
const ChunkHeaderSize = 16
const ChunkBlockHeaderSize = 8
const MaxBlockSizeOT = 0x20000

type Header struct {
	UnrealVersion   int    `json:"unreal_version"`
	LicenseeVersion int    `json:"licensee_version"`
	Flags           uint32 `json:"flags"`
	NameCount       int    `json:"name_count"`
	NameOffset      int    `json:"name_offset"`
	ExportCount     int    `json:"export_count"`
	ExportOffset    int    `json:"export_offset"`
	ImportCount     int    `json:"import_count"`
	ImportOffset    int    `json:"import_offset"`
}

type Import struct {
	ClassNameIndex  int `json:"class_name_index"`
	ObjectNameIndex int `json:"object_name_index"`
}

type Export struct {
	Index           int    `json:"index"`
	ClassIndex      int    `json:"class_index"`
	ObjectNameIndex int    `json:"object_name_index"`
	SerialSize      int    `json:"serial_size"`
	SerialOffset    int    `json:"serial_offset"`
	ObjectName      string `json:"object_name,omitempty"`
	ClassName       string `json:"class_name,omitempty"`
}

type GameProfile string

const (
	ProfileME2OT   GameProfile = "me2_ot"
	ProfileLE2     GameProfile = "le2"
	ProfileME3OT   GameProfile = "me3_ot"
	ProfileLE3     GameProfile = "le3"
	ProfileUnknown GameProfile = "unknown"
)

func InferGameProfile(uv, lv int) GameProfile {
	switch {
	case uv == 512 && lv == 130:
		return ProfileME2OT
	case uv == 684 && lv == 168:
		return ProfileLE2
	case uv == 684 && lv == 194:
		return ProfileME3OT
	case uv == 685 && lv == 205:
		return ProfileLE3
	default:
		return ProfileUnknown
	}
}

func (p GameProfile) String() string {
	return string(p)
}

type ContainerHit struct {
	StrRef         int    `json:"strref"`
	AbsoluteOffset int    `json:"absolute_offset"`
	LocalOffset    int    `json:"local_offset"`
	ExportIndex    int    `json:"export_index"`
	ExportName     string `json:"export_name,omitempty"`
	ClassName      string `json:"class_name,omitempty"`
	SerialOffset   int    `json:"serial_offset"`
	SerialSize     int    `json:"serial_size"`
}

type FileSummary struct {
	Path        string      `json:"file"`
	GameProfile GameProfile `json:"game_profile"`
	Compressed  bool        `json:"compressed"`
	Header      Header      `json:"header"`
	Names       []string    `json:"names,omitempty"`
	Imports     []Import    `json:"imports,omitempty"`
	Exports     []Export    `json:"exports"`
}

type ExportDetail struct {
	Export     Export `json:"export"`
	SerialData string `json:"serial_data,omitempty"`
}

func (h Header) String() string {
	return fmt.Sprintf("UV=%d LV=%d flags=0x%08X names=%d exports=%d imports=%d",
		h.UnrealVersion, h.LicenseeVersion, h.Flags,
		h.NameCount, h.ExportCount, h.ImportCount)
}

func (fs *FileSummary) RequireME2() error {
	if fs.GameProfile != ProfileME2OT {
		return fmt.Errorf("expected %s package, got %s", ProfileME2OT, fs.GameProfile)
	}
	return nil
}
