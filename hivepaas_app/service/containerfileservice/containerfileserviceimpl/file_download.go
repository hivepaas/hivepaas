package containerfileserviceimpl

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"path/filepath"

	"github.com/klauspost/compress/zstd"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerfileservice"
)

func (s *service) PrepareDownloadStream(
	ctx context.Context,
	req *containerfileservice.PrepareDownloadStreamReq,
) (*containerfileservice.PrepareDownloadStreamResp, error) {
	var (
		reader = req.Content
		isDir  = req.IsDir
	)

	if req.Stat != nil && req.Stat.Mode.IsDir() {
		isDir = true
	}

	cleanPath := filepath.Clean(req.Path)
	baseName := filepath.Base(cleanPath)
	if baseName == "/" || baseName == "." {
		baseName = "root"
	}

	var (
		resultReader      io.ReadCloser
		resultFileName    string
		resultFileSize    int64
		resultContentType string
	)

	if isDir {
		resultReader = reader
		resultFileName = baseName + ".tar"
		resultContentType = "application/x-tar"
	} else {
		tarReader := tar.NewReader(reader)
		header, err := tarReader.Next()
		if err != nil {
			_ = reader.Close()
			return nil, hperrors.Wrap(err)
		}

		if header.Typeflag == tar.TypeDir {
			resultReader = reader
			resultFileName = baseName + ".tar"
			resultContentType = "application/x-tar"
		} else {
			fileName := header.Name
			if fileName == "" {
				fileName = baseName
			} else {
				fileName = filepath.Base(fileName)
			}

			resultReader = &tarReadCloser{
				tarReader:  tarReader,
				underlying: reader,
			}
			resultFileName = fileName
			resultFileSize = header.Size
			resultContentType = "application/octet-stream"
		}
	}

	switch req.CompressionFormat {
	case base.FileCompressionNone, base.FileCompressionFormatZip, base.FileCompressionFormatTar:
		// No compression, keep original stream
	case base.FileCompressionFormatGzip:
		compressedReader, err := compressStream(resultReader, req.CompressionFormat)
		if err != nil {
			_ = resultReader.Close()
			return nil, hperrors.Wrap(err)
		}
		resultReader = compressedReader
		resultFileName += ".gz"
		resultContentType = "application/gzip"
		resultFileSize = 0
	case base.FileCompressionFormatZstd:
		compressedReader, err := compressStream(resultReader, req.CompressionFormat)
		if err != nil {
			_ = resultReader.Close()
			return nil, hperrors.Wrap(err)
		}
		resultReader = compressedReader
		resultFileName += ".zst"
		resultContentType = "application/zstd"
		resultFileSize = 0
	}

	return &containerfileservice.PrepareDownloadStreamResp{
		FileName:    resultFileName,
		FileSize:    resultFileSize,
		ContentType: resultContentType,
		Reader:      resultReader,
	}, nil
}

type tarReadCloser struct {
	tarReader  *tar.Reader
	underlying io.Closer
}

func (r *tarReadCloser) Read(p []byte) (int, error) {
	n, err := r.tarReader.Read(p)
	if err != nil {
		if err == io.EOF {
			return n, io.EOF
		}
		return n, hperrors.Wrap(err)
	}
	return n, nil
}

func (r *tarReadCloser) Close() error {
	if r.underlying != nil {
		if err := r.underlying.Close(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func compressStream(
	srcReader io.ReadCloser,
	format base.FileCompressionFormat,
) (io.ReadCloser, error) {
	if format == base.FileCompressionNone {
		return srcReader, nil
	}

	pr, pw := io.Pipe()

	var compWriter io.WriteCloser
	switch format {
	case base.FileCompressionNone, base.FileCompressionFormatZip, base.FileCompressionFormatTar:
		return srcReader, nil
	case base.FileCompressionFormatGzip:
		compWriter = gzip.NewWriter(pw)
	case base.FileCompressionFormatZstd:
		zstdW, err := zstd.NewWriter(pw)
		if err != nil {
			_ = pr.Close()
			_ = pw.Close()
			return nil, hperrors.Wrap(err)
		}
		compWriter = zstdW
	default:
		return srcReader, nil
	}

	go func() {
		defer func() {
			_ = srcReader.Close()
			_ = compWriter.Close()
		}()

		_, err := io.Copy(compWriter, srcReader)
		if err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		if err := compWriter.Close(); err != nil {
			_ = pw.CloseWithError(err)
			return
		}
		_ = pw.Close()
	}()

	return pr, nil
}
