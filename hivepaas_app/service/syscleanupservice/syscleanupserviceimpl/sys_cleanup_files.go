package syscleanupserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
)

func (s *service) sysCleanupFiles(
	_ context.Context,
	data *sysCleanupData,
) (err error) {
	if !data.SysCleanupSettings.FileCleanup.Enabled {
		return nil
	}

	defer func() {
		if err != nil {
			data.TaskOutput.FileCleanup.Error = err.Error()
		}
	}()

	var errs []error

	// Remove outdated temp files
	errs = append(errs, s.sysCleanupTempFiles(data))

	return errors.Join(errs...)
}

func (s *service) sysCleanupTempFiles(
	data *sysCleanupData,
) (err error) {
	if data.CleanupFilesTemp == base.CleanupFlagFalse {
		return nil
	}

	baseDirs := []string{base.BaseTempDirDefault, fileutil.AppTempDir()}
	threshold := time.Now().AddDate(0, 0, -fileutil.TempDirRetentionDays)
	if data.CleanupFilesTemp == base.CleanupFlagForce {
		threshold = time.Now()
	}

	var errs []error
	for _, baseDir := range baseDirs {
		_, err := fileutil.RemoveDatedTempDirs(baseDir, threshold)
		errs = append(errs, err)
	}
	return hperrors.Wrap(errors.Join(errs...))
}
