package tasksystemupdate

import (
	"context"
	"errors"

	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/pkg/applog"
	"github.com/localpaas/localpaas/localpaas_app/pkg/batchrecvchan"
	"github.com/localpaas/localpaas/localpaas_app/pkg/timeutil"
	"github.com/localpaas/localpaas/services/docker"
)

func (e *Executor) pullAllImages(
	ctx context.Context,
	data *taskData,
) (err error) {
	args := data.UpdateArgs

	errMap := gofn.ExecTasksEx(ctx, 0, true,
		func(ctx context.Context) error {
			return e.pullImage(ctx, args.TargetVersion.AppImage, data)
		},
		func(ctx context.Context) error {
			return e.pullImage(ctx, args.TargetVersion.RedisImage, data)
		},
		func(ctx context.Context) error {
			return e.pullImage(ctx, args.TargetVersion.DbImage, data)
		},
		func(ctx context.Context) error {
			return e.pullImage(ctx, args.TargetVersion.TraefikImage, data)
		},
	)
	for _, err := range errMap {
		return err
	}
	return nil
}

func (e *Executor) pullImage(
	ctx context.Context,
	image string,
	data *taskData,
) (err error) {
	if image == "" {
		return nil
	}

	start := timeutil.NowUTC()
	_ = data.LogStore.Add(ctx, applog.NewOutFrame("Pulling image "+image, applog.TsNow))
	defer func() {
		duration := timeutil.NowUTC().Sub(start)
		if err != nil {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Pulling image "+image+" finished in "+duration.String()+
				" with error: "+err.Error(), applog.TsNow))
		} else {
			_ = data.LogStore.Add(ctx, applog.NewOutFrame("Pulling image "+image+" finished in "+duration.String(),
				applog.TsNow))
		}
	}()

	logsReader, err := e.dockerManager.ImagePull(ctx, image)
	if err != nil {
		return apperrors.Wrap(err)
	}

	logsChan, _ := docker.StartScanningJSONMsg(ctx, logsReader, batchrecvchan.Options{})
	for msgs := range logsChan {
		for _, msg := range msgs {
			frameCreator := applog.NewOutFrame
			if msg.Error != nil {
				err = errors.Join(err, msg.Error)
				frameCreator = applog.NewErrFrame
			}
			if msg.String() != "" {
				_ = data.LogStore.Add(ctx, frameCreator(msg.String(), applog.TsNow))
			}
		}
	}
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
