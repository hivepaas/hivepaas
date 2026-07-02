package filearchiver

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// Decompress extracts the archive file based on the specified format.
// If the format is empty or ArchiveFormatAuto, it will be automatically detected from the filename.
func Decompress(filename, dirPath string, format ArchiveFormat) (errStr string, err error) {
	if format == "" {
		format = DetectArchiveFormat(filename)
		if format == "" {
			err = apperrors.New(apperrors.ErrUnrecognized).WithParam("Name", "Archive format")
			return err.Error(), err
		}
	}

	if err := os.MkdirAll(dirPath, base.DirModeDefault); err != nil {
		err = fmt.Errorf("failed to create destination directory: %w", err)
		return err.Error(), err
	}

	switch format {
	case ArchiveFormatTarZstd:
		return DecompressTarZstd(filename, dirPath)
	case ArchiveFormatTarLz4:
		return DecompressTarLz4(filename, dirPath)
	case ArchiveFormatTarGz:
		return DecompressTarGz(filename, dirPath)

	case ArchiveFormatZstd:
		return DecompressZstd(filename, dirPath)
	case ArchiveFormatLz4:
		return DecompressLz4(filename, dirPath)
	case ArchiveFormatGz:
		return DecompressGz(filename, dirPath)

	case ArchiveFormatAuto:
		fallthrough

	default:
		err = apperrors.New(apperrors.ErrArchiveFormatUnsupported).WithParam("Format", format)
		return err.Error(), err
	}
}

// DecompressTarGz decompresses a .tar.gz (or .tgz) archive into dirPath.
func DecompressTarGz(filename, dirPath string) (cmdErr string, err error) {
	var stderr bytes.Buffer
	cmd := exec.Command("tar", "-xzf", filename, "-C", dirPath)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stderr.String(), fmt.Errorf("tar gzip extraction failed: %w, stderr: %s", err, stderr.String())
	}
	return "", nil
}

// DecompressTarLz4 decompresses a .tar.lz4 archive into dirPath using lz4 and tar in a pipe.
func DecompressTarLz4(filename, dirPath string) (cmdErr string, err error) {
	lz4Cmd := exec.Command("lz4", "-dc", filename)
	tarCmd := exec.Command("tar", "-xf", "-", "-C", dirPath)

	pipe, err := lz4Cmd.StdoutPipe()
	if err != nil {
		err = fmt.Errorf("failed to create pipe for lz4 to tar: %w", err)
		return err.Error(), err
	}
	tarCmd.Stdin = pipe

	var lz4Stderr bytes.Buffer
	lz4Cmd.Stderr = &lz4Stderr

	var tarStderr bytes.Buffer
	tarCmd.Stderr = &tarStderr

	if err := lz4Cmd.Start(); err != nil {
		err = fmt.Errorf("failed to start lz4: %w", err)
		return err.Error(), err
	}

	if err := tarCmd.Start(); err != nil {
		_ = lz4Cmd.Process.Kill()
		err = fmt.Errorf("failed to start tar: %w", err)
		return err.Error(), err
	}

	if err := tarCmd.Wait(); err != nil {
		_ = lz4Cmd.Process.Kill()
		return tarStderr.String(), fmt.Errorf("tar lz4 extraction failed: %w, stderr: %s", err, tarStderr.String())
	}

	if err := lz4Cmd.Wait(); err != nil {
		return lz4Stderr.String(), fmt.Errorf("lz4 decompression failed: %w, stderr: %s", err, lz4Stderr.String())
	}

	return "", nil
}

// DecompressTarZstd decompresses a .tar.zst (or .tzst) archive into dirPath using zstd and tar in a pipe.
func DecompressTarZstd(filename, dirPath string) (cmdErr string, err error) {
	zstdCmd := exec.Command("zstd", "-dc", filename)
	tarCmd := exec.Command("tar", "-xf", "-", "-C", dirPath)

	pipe, err := zstdCmd.StdoutPipe()
	if err != nil {
		err = fmt.Errorf("failed to create pipe for zstd to tar: %w", err)
		return err.Error(), err
	}
	tarCmd.Stdin = pipe

	var zstdStderr bytes.Buffer
	zstdCmd.Stderr = &zstdStderr

	var tarStderr bytes.Buffer
	tarCmd.Stderr = &tarStderr

	if err := zstdCmd.Start(); err != nil {
		err = fmt.Errorf("failed to start zstd: %w", err)
		return err.Error(), err
	}

	if err := tarCmd.Start(); err != nil {
		_ = zstdCmd.Process.Kill()
		err = fmt.Errorf("failed to start tar: %w", err)
		return err.Error(), err
	}

	if err := tarCmd.Wait(); err != nil {
		_ = zstdCmd.Process.Kill()
		return tarStderr.String(), fmt.Errorf("tar zstd extraction failed: %w, stderr: %s", err, tarStderr.String())
	}

	if err := zstdCmd.Wait(); err != nil {
		return zstdStderr.String(), fmt.Errorf("zstd decompression failed: %w, stderr: %s", err, zstdStderr.String())
	}

	return "", nil
}

// DecompressGz decompresses a raw .gz file into dirPath.
func DecompressGz(filename, dirPath string) (cmdErr string, err error) {
	outPath := filepath.Join(dirPath, getDecompressedFilename(filename, ".gz"))
	outFile, err := os.Create(outPath)
	if err != nil {
		err = fmt.Errorf("failed to create output file %s: %w", outPath, err)
		return err.Error(), err
	}
	defer outFile.Close()

	var stderr bytes.Buffer
	cmd := exec.Command("gzip", "-d", "-c", filename)
	cmd.Stdout = outFile
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stderr.String(), fmt.Errorf("gzip decompression failed: %w, stderr: %s", err, stderr.String())
	}
	return "", nil
}

// DecompressLz4 decompresses a raw .lz4 file into dirPath.
func DecompressLz4(filename, dirPath string) (cmdErr string, err error) {
	outPath := filepath.Join(dirPath, getDecompressedFilename(filename, ".lz4"))
	outFile, err := os.Create(outPath)
	if err != nil {
		err = fmt.Errorf("failed to create output file %s: %w", outPath, err)
		return err.Error(), err
	}
	defer outFile.Close()

	var stderr bytes.Buffer
	cmd := exec.Command("lz4", "-d", "-c", filename)
	cmd.Stdout = outFile
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stderr.String(), fmt.Errorf("lz4 decompression failed: %w, stderr: %s", err, stderr.String())
	}
	return "", nil
}

// DecompressZstd decompresses a raw .zst file into dirPath.
func DecompressZstd(filename, dirPath string) (cmdErr string, err error) {
	outPath := filepath.Join(dirPath, getDecompressedFilename(filename, ".zst"))
	outFile, err := os.Create(outPath)
	if err != nil {
		err = fmt.Errorf("failed to create output file %s: %w", outPath, err)
		return err.Error(), err
	}
	defer outFile.Close()

	var stderr bytes.Buffer
	cmd := exec.Command("zstd", "-d", "-c", filename)
	cmd.Stdout = outFile
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return stderr.String(), fmt.Errorf("zstd decompression failed: %w, stderr: %s", err, stderr.String())
	}
	return "", nil
}

// getDecompressedFilename strips the specified extension from the filename in a case-insensitive manner.
func getDecompressedFilename(filename, ext string) string {
	baseName := filepath.Base(filename)
	if len(baseName) > len(ext) && strings.EqualFold(baseName[len(baseName)-len(ext):], ext) {
		return baseName[:len(baseName)-len(ext)]
	}
	return baseName
}
