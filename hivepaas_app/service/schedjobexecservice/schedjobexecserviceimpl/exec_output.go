package schedjobexecserviceimpl

import (
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
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

func (s *service) initOutputWriter(
	ctx context.Context,
	data *execData,
) (writer io.WriteCloser, err error) {
	cmdOutput := data.SchedJob.CommandOutput
	if cmdOutput == nil || !cmdOutput.Enabled {
		return nil, nil
	}

	if cmdOutput.SaveToFile != nil {
		return s.initOutputWriterToFile(ctx, data)
	}
	if cmdOutput.PipeToApp != nil {
		return s.initOutputWriterToApp(ctx, data)
	}
	return nil, apperrors.Wrap(apperrors.ErrSettingMissing).
		WithParam("Name", "command output")
}

func (s *service) finalize(
	ctx context.Context,
	db database.IDB,
	err error,
	data *execData,
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
		return apperrors.Wrap(err)
	}

	if data.File != nil {
		if err = s.fileRepo.Insert(ctx, db, data.File); err != nil {
			return apperrors.Wrap(err)
		}
	}

	return nil
}

func (s *service) cleanup(
	execErr error,
	data *execData,
) {
	if execErr != nil && data.File != nil && data.File.StorageType == base.FileStorageLocal {
		_ = os.RemoveAll(filepath.Join(config.Current.AppPath, data.File.Path))
	}
}
