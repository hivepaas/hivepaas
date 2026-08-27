package containerfileserviceimpl

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/klauspost/compress/zstd"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerfileservice"
)

const (
	filePermStandard       = 0644
	maxDecompressEntrySize = 10 * 1024 * 1024 * 1024 // 10GB per file entry limit to prevent decompression bomb
)

func (s *service) PrepareUploadTarStream(
	ctx context.Context,
	req *containerfileservice.PrepareUploadTarStreamReq,
) (_ *containerfileservice.PrepareUploadTarStreamResp, err error) {
	if req.Content == nil {
		return nil, hperrors.NewArgumentInvalid("content").WithMsgLog("upload content is required")
	}

	cleanPath := filepath.Clean(req.Path)
	if cleanPath == "." || cleanPath == "" {
		cleanPath = "/"
	}

	// Use Case 1: Upload single file (extract == false)
	if !req.Extract {
		var (
			dstDir    string
			entryName string
		)

		if strings.HasSuffix(req.Path, "/") || cleanPath == "/" {
			dstDir = cleanPath
			entryName = req.FileName
			if entryName == "" {
				entryName = "uploaded_file"
			}
		} else {
			// Path specifies full file path e.g. /app/config.json
			dstDir = filepath.Dir(cleanPath)
			entryName = filepath.Base(cleanPath)
		}

		tarStream, err := createSingleFileTarStream(req.Content, entryName, req.FileSize)
		if err != nil {
			_ = req.Content.Close()
			return nil, hperrors.Wrap(err)
		}

		return &containerfileservice.PrepareUploadTarStreamResp{
			DestPath:  dstDir,
			TarStream: tarStream,
		}, nil
	}

	// Use Case 2: Extract archive into target directory
	dstDir := cleanPath
	format := detectCompressionFormat(req.FileName, req.CompressionFormat)
	var tarStream io.ReadCloser

	switch format {
	case base.FileCompressionFormatZip:
		tarStream, err = convertZipToTarStream(req.Content, req.FileSize)
		if err != nil {
			_ = req.Content.Close()
			return nil, hperrors.Wrap(err)
		}

	case base.FileCompressionFormatGzip:
		gzReader, err := gzip.NewReader(req.Content)
		if err != nil {
			_ = req.Content.Close()
			return nil, hperrors.NewArgumentInvalid("file").WithMsgLog("failed to decompress gzip: %v", err)
		}
		tarStream = sanitizeTarStream(gzReader, req.Content)

	case base.FileCompressionFormatZstd:
		zstdReader, err := zstd.NewReader(req.Content)
		if err != nil {
			_ = req.Content.Close()
			return nil, hperrors.NewArgumentInvalid("file").WithMsgLog("failed to decompress zstd: %v", err)
		}
		tarStream = sanitizeTarStream(zstdReader.IOReadCloser(), req.Content)

	case base.FileCompressionFormatTar, base.FileCompressionNone:
		tarStream = sanitizeTarStream(req.Content, req.Content)

	default:
		tarStream = sanitizeTarStream(req.Content, req.Content)
	}

	return &containerfileservice.PrepareUploadTarStreamResp{
		DestPath:  dstDir,
		TarStream: tarStream,
	}, nil
}

func detectCompressionFormat(
	fileName string,
	explicitFormat base.FileCompressionFormat,
) base.FileCompressionFormat {
	if explicitFormat != "" {
		return explicitFormat
	}

	lowerName := strings.ToLower(fileName)
	switch {
	case strings.HasSuffix(lowerName, ".zip"):
		return base.FileCompressionFormatZip
	case strings.HasSuffix(lowerName, ".tar.gz"), strings.HasSuffix(lowerName, ".tgz"),
		strings.HasSuffix(lowerName, ".gz"):
		return base.FileCompressionFormatGzip
	case strings.HasSuffix(lowerName, ".tar.zst"), strings.HasSuffix(lowerName, ".zst"):
		return base.FileCompressionFormatZstd
	case strings.HasSuffix(lowerName, ".tar"):
		return base.FileCompressionFormatTar
	default:
		return base.FileCompressionNone
	}
}

func createSingleFileTarStream(
	content io.ReadCloser,
	entryName string,
	fileSize int64,
) (io.ReadCloser, error) {
	// Sanitize entry name
	entryName = sanitizeTarPath(entryName)
	if entryName == "" {
		entryName = "uploaded_file"
	}

	// If size is unknown (<= 0), read into temp buffer first so tar header has accurate size
	if fileSize <= 0 {
		buf := &bytes.Buffer{}
		n, err := io.Copy(buf, content)
		_ = content.Close()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		fileSize = n
		content = io.NopCloser(buf)
	}

	pr, pw := io.Pipe()
	go func() {
		defer content.Close()

		tarWriter := tar.NewWriter(pw)
		header := &tar.Header{
			Name:    entryName,
			Mode:    filePermStandard,
			Size:    fileSize,
			ModTime: time.Now(),
		}

		if err := tarWriter.WriteHeader(header); err != nil {
			_ = pw.CloseWithError(err)
			return
		}

		if _, err := io.Copy(tarWriter, content); err != nil {
			_ = pw.CloseWithError(err)
			return
		}

		if err := tarWriter.Close(); err != nil {
			_ = pw.CloseWithError(err)
			return
		}

		_ = pw.Close()
	}()

	return pr, nil
}

// sanitizeTarStream reads an incoming TAR stream, strips/sanitizes paths against Tar-Slip,
// and pipes valid entries to a new TAR stream.
func sanitizeTarStream(
	tarReaderSource io.Reader,
	underlyingCloser io.Closer,
) io.ReadCloser {
	pr, pw := io.Pipe()

	go func() {
		defer func() {
			if underlyingCloser != nil {
				_ = underlyingCloser.Close()
			}
		}()

		tarReader := tar.NewReader(tarReaderSource)
		tarWriter := tar.NewWriter(pw)
		defer tarWriter.Close()

		for {
			header, err := tarReader.Next()
			if err != nil {
				if errors.Is(err, io.EOF) {
					_ = tarWriter.Close()
					_ = pw.Close()
					return
				}
				_ = pw.CloseWithError(hperrors.NewArgumentInvalid("tar_stream").
					WithMsgLog("invalid tar stream: %v", err))
				return
			}

			// Sanitize header name against Tar-Slip
			sanitizedName := sanitizeTarPath(header.Name)
			if sanitizedName == "" {
				continue
			}
			header.Name = sanitizedName

			if err := tarWriter.WriteHeader(header); err != nil {
				_ = pw.CloseWithError(hperrors.Wrap(err))
				return
			}

			if header.Size > 0 {
				limitedReader := io.LimitReader(tarReader, maxDecompressEntrySize)
				//nolint:gosec // G110: Decompression bomb mitigated via LimitReader
				if _, err := io.Copy(tarWriter, limitedReader); err != nil {
					_ = pw.CloseWithError(hperrors.Wrap(err))
					return
				}
			}
		}
	}()

	return pr
}

func sanitizeTarPath(path string) string {
	cleanPath := strings.ReplaceAll(path, "\\", "/")
	parts := strings.Split(cleanPath, "/")
	cleanParts := make([]string, 0, len(parts))
	for _, p := range parts {
		if p == "" || p == "." || p == ".." {
			continue
		}
		cleanParts = append(cleanParts, p)
	}
	return strings.Join(cleanParts, "/")
}

// convertZipToTarStream unpacks a zip archive and converts it to a clean TAR stream on the fly.
func convertZipToTarStream(content io.ReadCloser, size int64) (io.ReadCloser, error) {
	// Create a temp file to store zip content since zip.Reader requires io.ReaderAt
	tempFile, err := os.CreateTemp("", "hivepaas-upload-zip-*")
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	written, err := io.Copy(tempFile, content)
	_ = content.Close()
	if err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
		return nil, hperrors.Wrap(err)
	}

	if size <= 0 {
		size = written
	}

	zipReader, err := zip.NewReader(tempFile, size)
	if err != nil {
		_ = tempFile.Close()
		_ = os.Remove(tempFile.Name())
		return nil, hperrors.NewArgumentInvalid("zip").WithMsgLog("invalid zip file: %v", err)
	}

	pr, pw := io.Pipe()

	go func() {
		defer func() {
			_ = tempFile.Close()
			_ = os.Remove(tempFile.Name())
		}()

		tarWriter := tar.NewWriter(pw)
		defer tarWriter.Close()

		for _, f := range zipReader.File {
			sanitizedName := sanitizeTarPath(f.Name)
			if sanitizedName == "" {
				continue
			}

			fileInfo := f.FileInfo()
			header, err := tar.FileInfoHeader(fileInfo, "")
			if err != nil {
				_ = pw.CloseWithError(hperrors.Wrap(err))
				return
			}
			header.Name = sanitizedName

			if err := tarWriter.WriteHeader(header); err != nil {
				_ = pw.CloseWithError(hperrors.Wrap(err))
				return
			}

			if !fileInfo.IsDir() {
				rc, err := f.Open()
				if err != nil {
					_ = pw.CloseWithError(hperrors.Wrap(err))
					return
				}

				limitedReader := io.LimitReader(rc, maxDecompressEntrySize)
				//nolint:gosec // G110: Decompression bomb mitigated via LimitReader
				if _, err := io.Copy(tarWriter, limitedReader); err != nil {
					_ = rc.Close()
					_ = pw.CloseWithError(hperrors.Wrap(err))
					return
				}
				_ = rc.Close()
			}
		}

		if err := tarWriter.Close(); err != nil {
			_ = pw.CloseWithError(hperrors.Wrap(err))
			return
		}
		_ = pw.Close()
	}()

	return pr, nil
}
