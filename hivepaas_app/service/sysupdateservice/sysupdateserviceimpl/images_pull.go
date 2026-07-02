package sysupdateserviceimpl

import (
	"context"
	"errors"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/batchrecvchan"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/services/docker"
)

//nolint:unused
func (s *service) pullAllImages(
	ctx context.Context,
	data *sysUpdateData,
) (err error) {
	args := gofn.Must(data.Task.ArgsAsSystemUpdate())

	errMap := gofn.ExecTasksEx(ctx, 0, true,
		func(ctx context.Context) error {
			return s.pullImage(ctx, args.TargetVersion.AppImage, data)
		},
		func(ctx context.Context) error {
			return s.pullImage(ctx, args.TargetVersion.RedisImage, data)
		},
		func(ctx context.Context) error {
			return s.pullImage(ctx, args.TargetVersion.DbImage, data)
		},
		func(ctx context.Context) error {
			return s.pullImage(ctx, args.TargetVersion.TraefikImage, data)
		},
	)
	for _, err := range errMap {
		return err
	}
	return nil
}

//nolint:unused
func (s *service) pullImage(
	ctx context.Context,
	image string,
	data *sysUpdateData,
) (err error) {
	if image == "" {
		return nil
	}

	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Pulling image "+image, tasklog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Pulling image "+image+" finished in "+duration.String()+
				" with error: "+err.Error(), tasklog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Pulling image "+image+" finished in "+duration.String(),
				tasklog.TsNow))
		}
	}()

	logsReader, err := s.dockerManager.ImagePull(ctx, image)
	if err != nil {
		return apperrors.New(err)
	}

	logsChan, _ := docker.StartScanningJSONMsg(ctx, logsReader, batchrecvchan.Options{})
	for msgs := range logsChan {
		for _, msg := range msgs {
			frameCreator := tasklog.NewOutFrame
			if msg.Error != nil {
				err = errors.Join(err, msg.Error)
				frameCreator = tasklog.NewErrFrame
			}
			if msg.String() != "" {
				_ = data.LogStore.Add(ctx, frameCreator(msg.String(), tasklog.TsNow))
			}
		}
	}
	if err != nil {
		return apperrors.New(err)
	}

	return nil
}
