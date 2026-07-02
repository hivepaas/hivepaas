package schedjobexecserviceimpl

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"filippo.io/age"
	"github.com/itchyny/timefmt-go"
	"github.com/klauspost/compress/zstd"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/funcutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/services/aws/s3"
)

type countingWriter struct {
	w io.Writer
	n *int64
}

func (cw *countingWriter) Write(p []byte) (int, error) {
	n, err := cw.w.Write(p)
	if cw.n != nil {
		*cw.n += int64(n)
	}
	return n, err //nolint:wrapcheck
}

type writeCloserWrapper struct {
	io.Writer
	closeFunc func() error
}

func (w *writeCloserWrapper) Close() error {
	if w.closeFunc != nil {
		return w.closeFunc()
	}
	return nil
}

func (s *service) schedJobExecGetOutputFileName(
	data *schedJobExecData,
) (string, error) {
	cmdOutput := data.SchedJob.Command.Output
	timeNow := data.TimeNow

	finalFileName := cmdOutput.SaveFileName
	if finalFileName == "" {
		finalFileName = fmt.Sprintf("job_%s_output_{timestamp}", data.SchedJobSetting.ID)
	}
	finalFileName = strings.ReplaceAll(finalFileName, "{timestamp}", timeNow.Format("20060102-150405"))
	finalFileName = strings.ReplaceAll(finalFileName, "{date}", timeNow.Format(time.DateOnly))

	// Supports popular time format syntax like `%Y-%m-%d %H:%M:%S`
	finalFileName = timefmt.Format(timeNow, finalFileName)

	switch cmdOutput.CompressionFormat {
	case base.FileCompressionFormatGzip:
		finalFileName += ".gz"
	case base.FileCompressionFormatZstd:
		finalFileName += ".zst"
	case base.FileCompressionNone: // Do nothing
	default:
		return "", apperrors.New(apperrors.ErrArchiveFormatUnsupported).
			WithParam("Format", cmdOutput.CompressionFormat)
	}

	switch cmdOutput.EncryptionFormat {
	case base.FileEncryptionFormatAge:
		finalFileName += ".age"
	case base.FileEncryptionNone: // Do nothing
	default:
		return "", apperrors.New(apperrors.ErrEncryptionFormatUnsupported).
			WithParam("Format", cmdOutput.EncryptionFormat)
	}

	return finalFileName, nil
}

func (s *service) schedJobExecInitOutputFile(
	ctx context.Context,
	data *schedJobExecData,
) (err error) {
	cmdOutput := data.SchedJob.Command.Output

	fileName, err := s.schedJobExecGetOutputFileName(data)
	if err != nil {
		return apperrors.New(err)
	}

	data.File = &entity.File{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       base.ObjectScopeApp,
		ObjectID:    data.App.ID,
		Type:        base.FileTypeSchedJobOutput,
		Kind:        string(cmdOutput.FileKind),
		Status:      base.FileStatusActive,
		Name:        fileName,
		Mimetype:    "application/octet-stream",
		StorageType: base.FileStorageLocal,
		CreatedAt:   data.TimeNow,
		UpdatedAt:   data.TimeNow,
	}

	if cmdOutput.Storage.ID != "" {
		storageSetting := data.RefObjects.RefSettings[cmdOutput.Storage.ID]
		if storageSetting == nil {
			return apperrors.NewNotFound("Storage setting")
		}
		if base.CloudStorageKind(storageSetting.Kind) != base.CloudStorageKindS3 {
			return apperrors.NewUnsupported(fmt.Sprintf("Storage kind '%s'", storageSetting.Kind))
		}
		s3Client, err := s3.NewClientFromSetting(ctx, storageSetting)
		if err != nil {
			return apperrors.New(err)
		}

		data.File.StorageType = base.FileStorageCloud
		data.File.StorageID = storageSetting.ID
		data.File.Bucket = s3Client.Config.Bucket
		data.File.Path = filepath.Join(cmdOutput.SavePath, fileName)
		data.uploadFunc = func(ctx context.Context, objectKey string, content io.Reader) error {
			return s3Client.UploadEx(ctx, data.File.Bucket, objectKey, 0, 0, content)
		}
	} else {
		data.File.Path = filepath.Join(config.Current.DataPathFiles().RelPath(), data.File.ID+"-"+fileName)
	}

	return nil
}

//nolint:gocognit
func (s *service) schedJobExecInitWriter(
	ctx context.Context,
	data *schedJobExecData,
) (writer io.WriteCloser, err error) {
	cmdOutput := data.SchedJob.Command.Output
	if cmdOutput == nil || !cmdOutput.Enabled {
		return nil, nil
	}

	err = s.schedJobExecInitOutputFile(ctx, data)
	if err != nil {
		return nil, apperrors.New(err)
	}

	var baseWriter io.WriteCloser

	if data.File.StorageID != "" {
		pr, pw := io.Pipe()
		data.uploadErrChan = make(chan error, 1)
		go func() {
			defer funcutil.EnsureNoPanic(nil)
			err := data.uploadFunc(ctx, data.File.Path, pr)
			if err != nil {
				data.uploadErrChan <- err
				_ = pr.CloseWithError(err)
			} else {
				data.uploadErrChan <- nil
			}
		}()
		baseWriter = &writeCloserWrapper{
			Writer:    &countingWriter{w: pw, n: &data.File.Size},
			closeFunc: func() error { return pw.Close() },
		}
	} else {
		destFilePath := filepath.Join(config.Current.AppPath, data.File.Path)
		f, err := os.Create(destFilePath)
		if err != nil {
			return nil, apperrors.New(err)
		}
		baseWriter = &writeCloserWrapper{
			Writer: &countingWriter{w: f, n: &data.File.Size},
		}
	}

	var (
		encW  io.WriteCloser
		compW io.WriteCloser
	)

	writer = baseWriter

	// 1. Encryption
	if cmdOutput.EncryptionFormat == base.FileEncryptionFormatAge {
		encSecret, err := cmdOutput.EncryptionSecret.GetPlain()
		if err != nil {
			_ = baseWriter.Close()
			return nil, apperrors.New(err)
		}
		if encSecret == "" {
			_ = baseWriter.Close()
			return nil, apperrors.NewMissing("Encryption secret")
		}
		recipient, err := age.NewScryptRecipient(encSecret)
		if err != nil {
			_ = baseWriter.Close()
			return nil, apperrors.New(err)
		}
		encW, err = age.Encrypt(writer, recipient)
		if err != nil {
			_ = baseWriter.Close()
			return nil, apperrors.New(err)
		}
		writer = encW
	}

	// 2. Compression
	switch cmdOutput.CompressionFormat {
	case base.FileCompressionNone:
		// Do nothing
	case base.FileCompressionFormatGzip:
		compW = gzip.NewWriter(writer)
		writer = compW
	case base.FileCompressionFormatZstd:
		zstdW, err := zstd.NewWriter(writer)
		if err != nil {
			_ = baseWriter.Close()
			return nil, apperrors.New(err)
		}
		compW = zstdW
		writer = compW
	}

	data.closeStack = func() error {
		var errs []error
		if compW != nil {
			if err := compW.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		if encW != nil {
			if err := encW.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		if baseWriter != nil {
			if err := baseWriter.Close(); err != nil {
				errs = append(errs, err)
			}
		}
		return errors.Join(errs...)
	}

	return writer, nil
}

func (s *service) schedJobExecFinalize(
	ctx context.Context,
	db database.IDB,
	err error,
	data *schedJobExecData,
) error {
	if data.closeStack != nil {
		if closeErr := data.closeStack(); closeErr != nil {
			err = errors.Join(err, closeErr)
		}
		if data.uploadErrChan != nil {
			if uploadErr := <-data.uploadErrChan; uploadErr != nil {
				err = errors.Join(err, uploadErr)
			}
		}
	}
	if err != nil {
		return apperrors.New(err)
	}

	if data.File != nil {
		if err = s.fileRepo.Insert(ctx, db, data.File); err != nil {
			return apperrors.New(err)
		}
	}

	return nil
}

func (s *service) schedJobExecCleanup(
	execErr error,
	data *schedJobExecData,
) {
	if execErr != nil && data.File != nil && data.File.StorageType == base.FileStorageLocal {
		_ = os.RemoveAll(filepath.Join(config.Current.AppPath, data.File.Path))
	}
}
