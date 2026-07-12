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
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/funcutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/services/aws/s3"
)

//nolint:gocognit
func (s *service) initOutputWriterToFile(
	ctx context.Context,
	data *execData,
) (writer io.WriteCloser, err error) {
	saveToFile := data.SchedJob.CommandOutput.SaveToFile

	err = s.initOutputFile(ctx, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
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
			return nil, apperrors.Wrap(err)
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
	if saveToFile.EncryptionFormat == base.FileEncryptionFormatAge {
		encSecret, err := saveToFile.EncryptionSecret.GetPlain()
		if err != nil {
			_ = baseWriter.Close()
			return nil, apperrors.Wrap(err)
		}
		if encSecret == "" {
			_ = baseWriter.Close()
			return nil, apperrors.NewMissing("Encryption secret")
		}
		recipient, err := age.NewScryptRecipient(encSecret)
		if err != nil {
			_ = baseWriter.Close()
			return nil, apperrors.Wrap(err)
		}
		encW, err = age.Encrypt(writer, recipient)
		if err != nil {
			_ = baseWriter.Close()
			return nil, apperrors.Wrap(err)
		}
		writer = encW
	}

	// 2. Compression
	switch saveToFile.CompressionFormat {
	case base.FileCompressionNone:
		// Do nothing
	case base.FileCompressionFormatGzip:
		compW = gzip.NewWriter(writer)
		writer = compW
	case base.FileCompressionFormatZstd:
		zstdW, err := zstd.NewWriter(writer)
		if err != nil {
			_ = baseWriter.Close()
			return nil, apperrors.Wrap(err)
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

func (s *service) initOutputFile(
	ctx context.Context,
	data *execData,
) (err error) {
	cmdOutput := data.SchedJob.CommandOutput.SaveToFile

	fileName, err := s.getOutputFileName(data)
	if err != nil {
		return apperrors.Wrap(err)
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
			return apperrors.Wrap(err)
		}

		data.File.StorageType = base.FileStorageCloud
		data.File.StorageID = storageSetting.ID
		data.File.Bucket = s3Client.Config.Bucket
		data.File.Path = filepath.Join(cmdOutput.FilePath, fileName)
		data.uploadFunc = func(ctx context.Context, objectKey string, content io.Reader) error {
			return s3Client.UploadEx(ctx, data.File.Bucket, objectKey, 0, 0, content)
		}
	} else {
		data.File.Path = filepath.Join(config.Current.DataPathFiles().RelPath(), data.File.ID+"-"+fileName)
	}

	return nil
}

func (s *service) getOutputFileName(
	data *execData,
) (string, error) {
	cmdOutput := data.SchedJob.CommandOutput.SaveToFile

	finalFileName := cmdOutput.FileName
	if finalFileName == "" {
		finalFileName = fmt.Sprintf("job_%s_output_{timestamp}", data.SchedJobSetting.ID)
	}
	finalFileName = strings.ReplaceAll(finalFileName, "{timestamp}", data.TimeNow.Format("20060102-150405"))
	finalFileName = strings.ReplaceAll(finalFileName, "{date}", data.TimeNow.Format(time.DateOnly))

	// Supports popular time format syntax like `%Y-%m-%d %H:%M:%S`
	finalFileName = timefmt.Format(data.TimeNow, finalFileName)

	switch cmdOutput.CompressionFormat {
	case base.FileCompressionFormatGzip:
		finalFileName += ".gz"
	case base.FileCompressionFormatZstd:
		finalFileName += ".zst"
	case base.FileCompressionNone: // Do nothing
	default:
		return "", apperrors.Wrap(apperrors.ErrArchiveFormatUnsupported).
			WithParam("Format", cmdOutput.CompressionFormat)
	}

	switch cmdOutput.EncryptionFormat {
	case base.FileEncryptionFormatAge:
		finalFileName += ".age"
	case base.FileEncryptionNone: // Do nothing
	default:
		return "", apperrors.Wrap(apperrors.ErrEncryptionFormatUnsupported).
			WithParam("Format", cmdOutput.EncryptionFormat)
	}

	return finalFileName, nil
}
