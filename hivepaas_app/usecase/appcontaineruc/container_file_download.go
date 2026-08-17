package appcontaineruc

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"io"
	"path/filepath"

	"github.com/klauspost/compress/zstd"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appcontaineruc/appcontainerdto"
)

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
		return n, apperrors.Wrap(err)
	}
	return n, nil
}

func (r *tarReadCloser) Close() error {
	if r.underlying != nil {
		if err := r.underlying.Close(); err != nil {
			return apperrors.Wrap(err)
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
	case base.FileCompressionNone:
		return srcReader, nil
	case base.FileCompressionFormatGzip:
		compWriter = gzip.NewWriter(pw)
	case base.FileCompressionFormatZstd:
		zstdW, err := zstd.NewWriter(pw)
		if err != nil {
			_ = pr.Close()
			_ = pw.Close()
			return nil, apperrors.Wrap(err)
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

func (uc *UC) DownloadFileFromContainer(
	ctx context.Context,
	auth *basedto.Auth,
	req *appcontainerdto.DownloadFileFromContainerReq,
) (*appcontainerdto.DownloadFileFromContainerResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if app.ServiceID == "" {
		return nil, apperrors.NewUnavailable("App service").
			WithMsgLog("service not exist for app")
	}

	res, err := uc.dockerManager.ContainerCopyFrom(ctx, req.ContainerID, req.Path)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	reader := res.Content
	stat := res.Stat

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

	if req.IsDir || stat.Mode.IsDir() {
		resultReader = reader
		resultFileName = baseName + ".tar"
		resultContentType = "application/x-tar"
	} else {
		tarReader := tar.NewReader(reader)
		header, err := tarReader.Next()
		if err != nil {
			_ = reader.Close()
			return nil, apperrors.Wrap(err)
		}

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

	switch req.CompressionFormat {
	case base.FileCompressionNone:
		// No compression, keep original stream
	case base.FileCompressionFormatGzip:
		compressedReader, err := compressStream(resultReader, req.CompressionFormat)
		if err != nil {
			_ = resultReader.Close()
			return nil, apperrors.Wrap(err)
		}
		resultReader = compressedReader
		resultFileName += ".gz"
		resultContentType = "application/gzip"
		resultFileSize = 0
	case base.FileCompressionFormatZstd:
		compressedReader, err := compressStream(resultReader, req.CompressionFormat)
		if err != nil {
			_ = resultReader.Close()
			return nil, apperrors.Wrap(err)
		}
		resultReader = compressedReader
		resultFileName += ".zst"
		resultContentType = "application/zstd"
		resultFileSize = 0
	}

	return &appcontainerdto.DownloadFileFromContainerResp{
		FileName:    resultFileName,
		FileSize:    resultFileSize,
		ContentType: resultContentType,
		Reader:      resultReader,
	}, nil
}
