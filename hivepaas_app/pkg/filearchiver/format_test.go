package filearchiver

import (
	"testing"
)

func TestDetectArchiveFormat(t *testing.T) {
	tests := []struct {
		name     string
		filename string
		want     ArchiveFormat
	}{
		// Tar Gz cases
		{"tar.gz extension", "archive.tar.gz", ArchiveFormatTarGz},
		{"tgz extension", "archive.tgz", ArchiveFormatTarGz},
		{"tar.gz uppercase", "archive.TAR.GZ", ArchiveFormatTarGz},
		{"tgz uppercase", "archive.TGZ", ArchiveFormatTarGz},
		{"tar.gz mixed case", "archive.tAr.Gz", ArchiveFormatTarGz},

		// Tar Lz4 cases
		{"tar.lz4 extension", "archive.tar.lz4", ArchiveFormatTarLz4},
		{"tar.lz4 uppercase", "archive.TAR.LZ4", ArchiveFormatTarLz4},
		{"tar.lz4 mixed case", "archive.TaR.Lz4", ArchiveFormatTarLz4},

		// Tar Zstd cases
		{"tar.zst extension", "archive.tar.zst", ArchiveFormatTarZstd},
		{"tzst extension", "archive.tzst", ArchiveFormatTarZstd},
		{"tar.zst uppercase", "archive.TAR.ZST", ArchiveFormatTarZstd},
		{"tzst uppercase", "archive.TZST", ArchiveFormatTarZstd},
		{"tar.zst mixed case", "archive.tAr.ZsT", ArchiveFormatTarZstd},

		// Raw compression formats (gz, lz4, zst)
		{"gz extension", "archive.gz", ArchiveFormatGz},
		{"gz uppercase", "archive.GZ", ArchiveFormatGz},
		{"lz4 extension", "archive.lz4", ArchiveFormatLz4},
		{"lz4 uppercase", "archive.LZ4", ArchiveFormatLz4},
		{"zst extension", "archive.zst", ArchiveFormatZstd},
		{"zst uppercase", "archive.ZST", ArchiveFormatZstd},

		// Negative / Unsupported cases
		{"empty filename", "", ""},
		{"no extension", "archive", ""},
		{"unsupported zip extension", "archive.zip", ""},
		{"unsupported tar extension", "archive.tar", ""},
		{"extension in the middle", "archive.tar.gz.txt", ""},
		{"extension prefix only", "tar.gz.archive", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DetectArchiveFormat(tt.filename)
			if got != tt.want {
				t.Errorf("DetectArchiveFormat(%q) = %q, want %q", tt.filename, got, tt.want)
			}
		})
	}
}
