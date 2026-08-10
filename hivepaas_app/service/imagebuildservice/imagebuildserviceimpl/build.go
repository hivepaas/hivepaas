package imagebuildserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/moby/moby/api/types/registry"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
)

type imageBuildData struct {
	*imagebuildservice.ImageBuildReq
	Resp *imagebuildservice.ImageBuildResp

	ImageTags     []string
	EnvVars       map[string]*string
	RegistryAuths map[string]registry.AuthConfig
}

func (s *service) ImageBuild(
	ctx context.Context,
	db database.IDB,
	req *imagebuildservice.ImageBuildReq,
) (resp *imagebuildservice.ImageBuildResp, err error) {
	resp = &imagebuildservice.ImageBuildResp{}
	data := &imageBuildData{
		ImageBuildReq: req,
		Resp:          resp,
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.Join(err, apperrors.NewPanic(r))
		}
	}()

	err = s.imageBuild(ctx, db, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	// Check if the context was canceled
	if err := ctx.Err(); err != nil {
		return nil, apperrors.Wrap(err)
	}

	err = s.imagePush(ctx, data)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return resp, err
}

func (s *service) addStepStartLog(
	ctx context.Context,
	data *imageBuildData,
	msg string,
) {
	_ = data.LogStore.Add(ctx,
		tasklog.NewOutFrame("---------------------------------", tasklog.TsNow),
		tasklog.NewOutFrame(msg, tasklog.TsNow))
}

func (s *service) addStepEndLog(
	ctx context.Context,
	data *imageBuildData,
	start time.Time,
	err error,
) {
	duration := timeutil.NowUTC().Sub(start).Truncate(time.Millisecond)
	if err != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Task finished in "+duration.String()+
			" with error: "+err.Error(), tasklog.TsNow))
	} else {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("Task finished in "+duration.String(),
			tasklog.TsNow))
	}
}
