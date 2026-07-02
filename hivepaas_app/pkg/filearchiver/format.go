package filearchiver

import "strings"

type ArchiveFormat string

const (
	ArchiveFormatAuto ArchiveFormat = ""

	ArchiveFormatGz   ArchiveFormat = "gz"
	ArchiveFormatLz4  ArchiveFormat = "lz4"
	ArchiveFormatZstd ArchiveFormat = "zst"

	ArchiveFormatTarGz   ArchiveFormat = "tar.gz"
	ArchiveFormatTarLz4  ArchiveFormat = "tar.lz4"
	ArchiveFormatTarZstd ArchiveFormat = "tar.zst"
)

func (format ArchiveFormat) FileExtDefault() string {
	switch format {
	case ArchiveFormatTarGz:
		return ".tar.gz"
	case ArchiveFormatTarLz4:
		return ".tar.lz4"
	case ArchiveFormatTarZstd:
		return ".tar.zst"

	case ArchiveFormatGz:
		return ".gz"
	case ArchiveFormatLz4:
		return ".lz4"
	case ArchiveFormatZstd:
		return ".zst"

	case ArchiveFormatAuto:
		fallthrough

	default:
		return ""
	}
}

func DetectArchiveFormat(filename string) ArchiveFormat {
	filename = strings.ToLower(filename)

	switch {
	case strings.HasSuffix(filename, ".tar.gz"), strings.HasSuffix(filename, ".tgz"):
		return ArchiveFormatTarGz
	case strings.HasSuffix(filename, ".tar.lz4"):
		return ArchiveFormatTarLz4
	case strings.HasSuffix(filename, ".tar.zst"), strings.HasSuffix(filename, ".tzst"):
		return ArchiveFormatTarZstd
	case strings.HasSuffix(filename, ".gz"):
		return ArchiveFormatGz
	case strings.HasSuffix(filename, ".lz4"):
		return ArchiveFormatLz4
	case strings.HasSuffix(filename, ".zst"):
		return ArchiveFormatZstd
	default:
		return ""
	}
}
