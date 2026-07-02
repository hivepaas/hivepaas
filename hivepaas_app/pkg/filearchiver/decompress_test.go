package filearchiver

import (
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// getCommandPath resolves the path to an executable, checking common Homebrew paths on macOS if needed.
func getCommandPath(name string) string {
	if path, err := exec.LookPath(name); err == nil {
		return path
	}
	// Fallback to common macOS homebrew paths
	for _, dir := range []string{"/opt/homebrew/bin", "/usr/local/bin"} {
		path := filepath.Join(dir, name)
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	return name
}

func TestDecompress(t *testing.T) {
	// Create temporary working directory for test files
	tempDir, err := os.MkdirTemp("", "filearchiver_test")
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
	testFileContent := "Hello, HivePAAS!"
	srcFilePath := filepath.Join(srcDir, testFileName)
	if err := os.WriteFile(srcFilePath, []byte(testFileContent), 0644); err != nil {
		t.Fatalf("failed to write test file: %v", err)
	}

	// 1. Test TarGz
	t.Run("TarGz", func(t *testing.T) {
		archivePath := filepath.Join(tempDir, "archive.tar.gz")
		destDir := filepath.Join(tempDir, "dest_targz")

		// Compress
		cmd := exec.Command(getCommandPath("tar"), "-czf", archivePath, "-C", srcDir, testFileName)
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to compress tar.gz: %v", err)
		}

		// Decompress
		if _, err := Decompress(archivePath, destDir, ArchiveFormatTarGz); err != nil {
			t.Fatalf("failed to decompress tar.gz: %v", err)
		}

		// Verify
		got, err := os.ReadFile(filepath.Join(destDir, testFileName))
		if err != nil {
			t.Fatalf("failed to read decompressed file: %v", err)
		}
		if string(got) != testFileContent {
			t.Errorf("got %q, want %q", string(got), testFileContent)
		}
	})

	// 2. Test TarLz4
	t.Run("TarLz4", func(t *testing.T) {
		archivePath := filepath.Join(tempDir, "archive.tar.lz4")
		destDir := filepath.Join(tempDir, "dest_tarlz4")

		// Compress: tar -cf - -C srcDir hello.txt | lz4 -z > archivePath
		tarCmd := exec.Command(getCommandPath("tar"), "-cf", "-", "-C", srcDir, testFileName)
		lz4Cmd := exec.Command(getCommandPath("lz4"), "-z")

		outFile, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("failed to create archive: %v", err)
		}
		defer outFile.Close()

		pipe, err := tarCmd.StdoutPipe()
		if err != nil {
			t.Fatalf("failed to create pipe: %v", err)
		}
		lz4Cmd.Stdin = pipe
		lz4Cmd.Stdout = outFile

		if err := tarCmd.Start(); err != nil {
			t.Fatalf("failed to start tar: %v", err)
		}
		if err := lz4Cmd.Start(); err != nil {
			t.Fatalf("failed to start lz4: %v", err)
		}
		if err := tarCmd.Wait(); err != nil {
			t.Fatalf("tar wait failed: %v", err)
		}
		if err := lz4Cmd.Wait(); err != nil {
			t.Fatalf("lz4 wait failed: %v", err)
		}
		outFile.Close()

		// Decompress
		if _, err := Decompress(archivePath, destDir, ArchiveFormatTarLz4); err != nil {
			t.Fatalf("failed to decompress tar.lz4: %v", err)
		}

		// Verify
		got, err := os.ReadFile(filepath.Join(destDir, testFileName))
		if err != nil {
			t.Fatalf("failed to read decompressed file: %v", err)
		}
		if string(got) != testFileContent {
			t.Errorf("got %q, want %q", string(got), testFileContent)
		}
	})

	// 3. Test TarZstd
	t.Run("TarZstd", func(t *testing.T) {
		archivePath := filepath.Join(tempDir, "archive.tar.zst")
		destDir := filepath.Join(tempDir, "dest_tarzstd")

		// Compress: tar -cf - -C srcDir hello.txt | zstd -z > archivePath
		tarCmd := exec.Command(getCommandPath("tar"), "-cf", "-", "-C", srcDir, testFileName)
		zstdCmd := exec.Command(getCommandPath("zstd"), "-z")

		outFile, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("failed to create archive: %v", err)
		}
		defer outFile.Close()

		pipe, err := tarCmd.StdoutPipe()
		if err != nil {
			t.Fatalf("failed to create pipe: %v", err)
		}
		zstdCmd.Stdin = pipe
		zstdCmd.Stdout = outFile

		if err := tarCmd.Start(); err != nil {
			t.Fatalf("failed to start tar: %v", err)
		}
		if err := zstdCmd.Start(); err != nil {
			t.Fatalf("failed to start zstd: %v", err)
		}
		if err := tarCmd.Wait(); err != nil {
			t.Fatalf("tar wait failed: %v", err)
		}
		if err := zstdCmd.Wait(); err != nil {
			t.Fatalf("zstd wait failed: %v", err)
		}
		outFile.Close()

		// Decompress
		if _, err := Decompress(archivePath, destDir, ArchiveFormatTarZstd); err != nil {
			t.Fatalf("failed to decompress tar.zst: %v", err)
		}

		// Verify
		got, err := os.ReadFile(filepath.Join(destDir, testFileName))
		if err != nil {
			t.Fatalf("failed to read decompressed file: %v", err)
		}
		if string(got) != testFileContent {
			t.Errorf("got %q, want %q", string(got), testFileContent)
		}
	})

	// 4. Test Raw Gz
	t.Run("Gz", func(t *testing.T) {
		archivePath := filepath.Join(tempDir, "hello.txt.gz")
		destDir := filepath.Join(tempDir, "dest_gz")

		// Compress: gzip -c srcFilePath > archivePath
		outFile, err := os.Create(archivePath)
		if err != nil {
			t.Fatalf("failed to create archive: %v", err)
		}
		defer outFile.Close()

		cmd := exec.Command(getCommandPath("gzip"), "-c", srcFilePath)
		cmd.Stdout = outFile
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to compress gz: %v", err)
		}
		outFile.Close()

		// Decompress
		if _, err := Decompress(archivePath, destDir, ArchiveFormatGz); err != nil {
			t.Fatalf("failed to decompress gz: %v", err)
		}

		// Verify
		got, err := os.ReadFile(filepath.Join(destDir, testFileName))
		if err != nil {
			t.Fatalf("failed to read decompressed file: %v", err)
		}
		if string(got) != testFileContent {
			t.Errorf("got %q, want %q", string(got), testFileContent)
		}
	})

	// 5. Test Raw Lz4
	t.Run("Lz4", func(t *testing.T) {
		archivePath := filepath.Join(tempDir, "hello.txt.lz4")
		destDir := filepath.Join(tempDir, "dest_lz4")

		// Compress: lz4 -z srcFilePath archivePath
		cmd := exec.Command(getCommandPath("lz4"), "-z", srcFilePath, archivePath)
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to compress lz4: %v", err)
		}

		// Decompress
		if _, err := Decompress(archivePath, destDir, ArchiveFormatLz4); err != nil {
			t.Fatalf("failed to decompress lz4: %v", err)
		}

		// Verify
		got, err := os.ReadFile(filepath.Join(destDir, testFileName))
		if err != nil {
			t.Fatalf("failed to read decompressed file: %v", err)
		}
		if string(got) != testFileContent {
			t.Errorf("got %q, want %q", string(got), testFileContent)
		}
	})

	// 6. Test Raw Zstd
	t.Run("Zstd", func(t *testing.T) {
		archivePath := filepath.Join(tempDir, "hello.txt.zst")
		destDir := filepath.Join(tempDir, "dest_zstd")

		// Compress: zstd -z srcFilePath -o archivePath
		cmd := exec.Command(getCommandPath("zstd"), "-z", srcFilePath, "-o", archivePath)
		if err := cmd.Run(); err != nil {
			t.Fatalf("failed to compress zstd: %v", err)
		}

		// Decompress
		if _, err := Decompress(archivePath, destDir, ArchiveFormatZstd); err != nil {
			t.Fatalf("failed to decompress zstd: %v", err)
		}

		// Verify
		got, err := os.ReadFile(filepath.Join(destDir, testFileName))
		if err != nil {
			t.Fatalf("failed to read decompressed file: %v", err)
		}
		if string(got) != testFileContent {
			t.Errorf("got %q, want %q", string(got), testFileContent)
		}
	})

	// 7. Test Auto-detect Format
	t.Run("AutoDetect", func(t *testing.T) {
		archivePath := filepath.Join(tempDir, "hello.txt.gz")
		destDir := filepath.Join(tempDir, "dest_auto")

		// Decompress with empty format (auto-detect)
		if _, err := Decompress(archivePath, destDir, ""); err != nil {
			t.Fatalf("failed to auto-decompress gz: %v", err)
		}

		// Verify
		got, err := os.ReadFile(filepath.Join(destDir, testFileName))
		if err != nil {
			t.Fatalf("failed to read decompressed file: %v", err)
		}
		if string(got) != testFileContent {
			t.Errorf("got %q, want %q", string(got), testFileContent)
		}
	})
}
