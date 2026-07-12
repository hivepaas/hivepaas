package filearchiver

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

type CompressionLevel string

const (
	CompressionLevelDefault CompressionLevel = ""
	CompressionLevelFastest CompressionLevel = "fastest"
	CompressionLevelFast    CompressionLevel = "fast"
	CompressionLevelHigh    CompressionLevel = "high"
	CompressionLevelHighest CompressionLevel = "highest"
	CompressionLevelNone    CompressionLevel = "none"
)

// Compress archives and compresses a file or directory based on the specified format and level.
// If the format is empty, it will be automatically detected from destFilename.
func Compress(
	srcFilename, destFilename string,
	format ArchiveFormat,
	level CompressionLevel,
) (errStr string, err error) {
	if format == "" {
		format = DetectArchiveFormat(destFilename)
		if format == "" {
			err = apperrors.Wrap(apperrors.ErrUnrecognized).WithParam("Name", "Archive format")
			return err.Error(), err
		}
	}

	destDir := filepath.Dir(destFilename)
	if err := os.MkdirAll(destDir, base.DirModeDefault); err != nil {
		err = fmt.Errorf("failed to create destination directory: %w", err)
		return err.Error(), err
	}

	switch format {
	case ArchiveFormatTarZstd:
		return CompressTarZstd(srcFilename, destFilename, level)
	case ArchiveFormatTarLz4:
		return CompressTarLz4(srcFilename, destFilename, level)
	case ArchiveFormatTarGz:
		return CompressTarGz(srcFilename, destFilename, level)

	case ArchiveFormatZstd:
		return CompressZstd(srcFilename, destFilename, level)
	case ArchiveFormatLz4:
		return CompressLz4(srcFilename, destFilename, level)
	case ArchiveFormatGz:
		return CompressGz(srcFilename, destFilename, level)

	case ArchiveFormatAuto:
		fallthrough

	default:
		err = apperrors.Wrap(apperrors.ErrArchiveFormatUnsupported).WithParam("Format", format)
		return err.Error(), err
	}
}

// CompressTarGz archives and compresses filename into a .tar.gz at destFilename.
func CompressTarGz(
	srcFilename, destFilename string,
	level CompressionLevel,
) (cmdErr string, err error) {
	info, err := os.Stat(srcFilename)
	if err != nil {
		err = fmt.Errorf("failed to stat source: %w", err)
		return err.Error(), err
	}

	var dir, target string
	if info.IsDir() {
		dir = srcFilename
		target = "."
	} else {
		dir = filepath.Dir(srcFilename)
		target = filepath.Base(srcFilename)
	}

	tarCmd := exec.Command("tar", "-cf", "-", "-C", dir, target)
	gzipCmd := exec.Command("gzip", getGzipLevelFlag(level), "-c") //nolint:gosec

	outFile, err := os.Create(destFilename)
	if err != nil {
		err = fmt.Errorf("failed to create output file %s: %w", destFilename, err)
		return err.Error(), err
	}
	defer outFile.Close()

	pipe, err := tarCmd.StdoutPipe()
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	gzipCmd.Stdin = pipe
	gzipCmd.Stdout = outFile

	var tarStderr bytes.Buffer
	tarCmd.Stderr = &tarStderr

	var gzipStderr bytes.Buffer
	gzipCmd.Stderr = &gzipStderr

	if err := tarCmd.Start(); err != nil {
		err = fmt.Errorf("failed to start tar: %w", err)
		return err.Error(), err
	}
	if err := gzipCmd.Start(); err != nil {
		_ = tarCmd.Process.Kill()
		err = fmt.Errorf("failed to start gzip: %w", err)
		return err.Error(), err
	}

	if err := tarCmd.Wait(); err != nil {
		_ = gzipCmd.Process.Kill()
		return tarStderr.String(), fmt.Errorf("tar failed: %w, stderr: %s", err, tarStderr.String())
	}
	if err := gzipCmd.Wait(); err != nil {
		return gzipStderr.String(), fmt.Errorf("gzip failed: %w, stderr: %s", err, gzipStderr.String())
	}

	return "", nil
}

// CompressTarLz4 archives and compresses filename into a .tar.lz4 at destFilename using lz4 and tar in a pipe.
func CompressTarLz4(
	srcFilename, destFilename string,
	level CompressionLevel,
) (cmdErr string, err error) {
	info, err := os.Stat(srcFilename)
	if err != nil {
		err = fmt.Errorf("failed to stat source: %w", err)
		return err.Error(), err
	}

	var dir, target string
	if info.IsDir() {
		dir = srcFilename
		target = "."
	} else {
		dir = filepath.Dir(srcFilename)
		target = filepath.Base(srcFilename)
	}

	tarCmd := exec.Command("tar", "-cf", "-", "-C", dir, target)
	lz4Cmd := exec.Command("lz4", "-z", getLz4LevelFlag(level)) //nolint:gosec

	outFile, err := os.Create(destFilename)
	if err != nil {
		err = fmt.Errorf("failed to create output file %s: %w", destFilename, err)
		return err.Error(), err
	}
	defer outFile.Close()

	pipe, err := tarCmd.StdoutPipe()
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	lz4Cmd.Stdin = pipe
	lz4Cmd.Stdout = outFile

	var tarStderr bytes.Buffer
	tarCmd.Stderr = &tarStderr

	var lz4Stderr bytes.Buffer
	lz4Cmd.Stderr = &lz4Stderr

	if err := tarCmd.Start(); err != nil {
		err = fmt.Errorf("failed to start tar: %w", err)
		return err.Error(), err
	}
	if err := lz4Cmd.Start(); err != nil {
		_ = tarCmd.Process.Kill()
		err = fmt.Errorf("failed to start lz4: %w", err)
		return err.Error(), err
	}

	if err := tarCmd.Wait(); err != nil {
		_ = lz4Cmd.Process.Kill()
		return tarStderr.String(), fmt.Errorf("tar failed: %w, stderr: %s", err, tarStderr.String())
	}
	if err := lz4Cmd.Wait(); err != nil {
		return lz4Stderr.String(), fmt.Errorf("lz4 failed: %w, stderr: %s", err, lz4Stderr.String())
	}

	return "", nil
}

// CompressTarZstd archives and compresses filename into a .tar.zst at destFilename using zstd and tar in a pipe.
func CompressTarZstd(
	srcFilename, destFilename string,
	level CompressionLevel,
) (cmdErr string, err error) {
	info, err := os.Stat(srcFilename)
	if err != nil {
		err = fmt.Errorf("failed to stat source: %w", err)
		return err.Error(), err
	}

	var dir, target string
	if info.IsDir() {
		dir = srcFilename
		target = "."
	} else {
		dir = filepath.Dir(srcFilename)
		target = filepath.Base(srcFilename)
	}

	tarCmd := exec.Command("tar", "-cf", "-", "-C", dir, target)
	zstdCmd := exec.Command("zstd", "-z", getZstdLevelFlag(level)) //nolint:gosec

	outFile, err := os.Create(destFilename)
	if err != nil {
		err = fmt.Errorf("failed to create output file %s: %w", destFilename, err)
		return err.Error(), err
	}
	defer outFile.Close()

	pipe, err := tarCmd.StdoutPipe()
	if err != nil {
		return "", apperrors.Wrap(err)
	}
	zstdCmd.Stdin = pipe
	zstdCmd.Stdout = outFile

	var tarStderr bytes.Buffer
	tarCmd.Stderr = &tarStderr

	var zstdStderr bytes.Buffer
	zstdCmd.Stderr = &zstdStderr

	if err := tarCmd.Start(); err != nil {
		err = fmt.Errorf("failed to start tar: %w", err)
		return err.Error(), err
	}
	if err := zstdCmd.Start(); err != nil {
		_ = tarCmd.Process.Kill()
		err = fmt.Errorf("failed to start zstd: %w", err)
		return err.Error(), err
	}

	if err := tarCmd.Wait(); err != nil {
		_ = zstdCmd.Process.Kill()
		return tarStderr.String(), fmt.Errorf("tar failed: %w, stderr: %s", err, tarStderr.String())
	}
	if err := zstdCmd.Wait(); err != nil {
		return zstdStderr.String(), fmt.Errorf("zstd failed: %w, stderr: %s", err, zstdStderr.String())
	}

	return "", nil
}

// CompressGz compresses a raw file filename into a .gz file at destFilename.
func CompressGz(
	srcFilename, destFilename string,
	level CompressionLevel,
) (cmdErr string, err error) {
	outFile, err := os.Create(destFilename)
	if err != nil {
		err = fmt.Errorf("failed to create output file %s: %w", destFilename, err)
		return err.Error(), err
	}
	defer outFile.Close()

	var stderr bytes.Buffer
	cmd := exec.Command("gzip", getGzipLevelFlag(level), "-c", srcFilename) //nolint:gosec
	cmd.Stdout = outFile
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stderr.String(), fmt.Errorf("gzip compression failed: %w, stderr: %s", err, stderr.String())
	}
	return "", nil
}

// CompressLz4 compresses a raw file filename into a .lz4 file at destFilename.
func CompressLz4(
	srcFilename, destFilename string,
	level CompressionLevel,
) (cmdErr string, err error) {
	outFile, err := os.Create(destFilename)
	if err != nil {
		err = fmt.Errorf("failed to create output file %s: %w", destFilename, err)
		return err.Error(), err
	}
	defer outFile.Close()

	var stderr bytes.Buffer
	cmd := exec.Command("lz4", "-z", getLz4LevelFlag(level), srcFilename, "-c") //nolint:gosec
	cmd.Stdout = outFile
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stderr.String(), fmt.Errorf("lz4 compression failed: %w, stderr: %s", err, stderr.String())
	}
	return "", nil
}

// CompressZstd compresses a raw file filename into a .zst file at destFilename.
func CompressZstd(
	srcFilename, destFilename string,
	level CompressionLevel,
) (cmdErr string, err error) {
	outFile, err := os.Create(destFilename)
	if err != nil {
		err = fmt.Errorf("failed to create output file %s: %w", destFilename, err)
		return err.Error(), err
	}
	defer outFile.Close()

	var stderr bytes.Buffer
	cmd := exec.Command("zstd", "-z", getZstdLevelFlag(level), srcFilename, "-c") //nolint:gosec
	cmd.Stdout = outFile
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		return stderr.String(), fmt.Errorf("zstd compression failed: %w, stderr: %s", err, stderr.String())
	}
	return "", nil
}

func getGzipLevelFlag(level CompressionLevel) string {
	switch level {
	case CompressionLevelFastest:
		return "-1"
	case CompressionLevelFast:
		return "-3"
	case CompressionLevelHigh:
		return "-7"
	case CompressionLevelHighest:
		return "-9"
	case CompressionLevelNone:
		return "-1" // gzip does not support store-only mode, using fastest instead
	case CompressionLevelDefault:
		fallthrough
	default:
		return "-6"
	}
}

func getLz4LevelFlag(level CompressionLevel) string {
	switch level {
	case CompressionLevelFastest:
		return "--fast=3"
	case CompressionLevelFast:
		return "-1"
	case CompressionLevelHigh:
		return "-9"
	case CompressionLevelHighest:
		return "-12"
	case CompressionLevelNone:
		return "--fast=10"
	case CompressionLevelDefault:
		fallthrough
	default:
		return "-1"
	}
}

func getZstdLevelFlag(level CompressionLevel) string {
	switch level {
	case CompressionLevelFastest:
		return "--fast=3"
	case CompressionLevelFast:
		return "-1"
	case CompressionLevelHigh:
		return "-12"
	case CompressionLevelHighest:
		return "-19"
	case CompressionLevelNone:
		return "--fast=5"
	case CompressionLevelDefault:
		fallthrough
	default:
		return "-3"
	}
}
