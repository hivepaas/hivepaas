package filearchiver

import (
	"os"
	"path/filepath"
	"testing"
)

func TestCompressAndDecompress(t *testing.T) {
	// Create temporary working directory for test files
	tempDir, err := os.MkdirTemp("", "compress_test")
	if err != nil {
		t.Fatalf("failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// Create test file structure
	srcDir := filepath.Join(tempDir, "src")
	if err := os.MkdirAll(srcDir, 0755); err != nil {
		t.Fatalf("failed to create src dir: %v", err)
	}

	testFileName := "hello.txt"
	testFileContent := "Hello, HivePAAS Compression and Decompression Tests!"
	srcFilePath := filepath.Join(srcDir, testFileName)
	if err := os.WriteFile(srcFilePath, []byte(testFileContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	formats := []ArchiveFormat{
		ArchiveFormatTarGz,
		ArchiveFormatTarLz4,
		ArchiveFormatTarZstd,
		ArchiveFormatGz,
		ArchiveFormatLz4,
		ArchiveFormatZstd,
	}

	levels := []CompressionLevel{
		CompressionLevelDefault,
		CompressionLevelFastest,
		CompressionLevelFast,
		CompressionLevelHigh,
		CompressionLevelHighest,
		CompressionLevelNone,
	}

	for _, format := range formats {
		for _, level := range levels {
			testName := string(format) + "_" + string(level)
			if level == "" {
				testName = string(format) + "_default"
			}

			t.Run(testName, func(t *testing.T) {
				archiveName := "archive_" + testName
				switch format { //nolint
				case ArchiveFormatTarGz:
					archiveName += ".tar.gz"
				case ArchiveFormatTarLz4:
					archiveName += ".tar.lz4"
				case ArchiveFormatTarZstd:
					archiveName += ".tar.zst"
				case ArchiveFormatGz:
					archiveName += ".gz"
				case ArchiveFormatLz4:
					archiveName += ".lz4"
				case ArchiveFormatZstd:
					archiveName += ".zst"
				}

				archivePath := filepath.Join(tempDir, archiveName)
				destDir := filepath.Join(tempDir, "dest_"+testName)

				// For tar formats, we compress the srcDir.
				// For raw formats, we compress the srcFilePath.
				var targetToCompress string
				if format == ArchiveFormatTarGz || format == ArchiveFormatTarLz4 || format == ArchiveFormatTarZstd {
					targetToCompress = srcDir
				} else {
					targetToCompress = srcFilePath
				}

				// 1. Compress
				if _, err := Compress(targetToCompress, archivePath, format, level); err != nil {
					t.Fatalf("failed to compress: %v", err)
				}

				// Verify archive exists and has non-zero size
				info, err := os.Stat(archivePath)
				if err != nil {
					t.Fatalf("failed to stat archive: %v", err)
				}
				if info.Size() == 0 {
					t.Errorf("compressed archive has 0 bytes")
				}

				// 2. Decompress
				if _, err := Decompress(archivePath, destDir, format); err != nil {
					t.Fatalf("failed to decompress: %v", err)
				}

				// 3. Verify contents
				var verifyPath string
				if format == ArchiveFormatTarGz || format == ArchiveFormatTarLz4 || format == ArchiveFormatTarZstd {
					// Since we compressed the contents of srcDir, the extraction output has the files directly inside destDir
					verifyPath = filepath.Join(destDir, testFileName)
				} else {
					// Raw formats decompress into a file with name minus extension inside destDir
					verifyPath = filepath.Join(destDir, getDecompressedFilename(archivePath, "."+string(format)))
				}

				got, err := os.ReadFile(verifyPath)
				if err != nil {
					t.Fatalf("failed to read decompressed file %s: %v", verifyPath, err)
				}
				if string(got) != testFileContent {
					t.Errorf("content mismatch: got %q, want %q", string(got), testFileContent)
				}
			})
		}
	}
}
