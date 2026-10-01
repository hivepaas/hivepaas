package imagebuildserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/registry"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
)

type imageBuildData struct {
	*imagebuildservice.ImageBuildReq
	Resp *imagebuildservice.ImageBuildResp

	Inputs *imagebuildservice.BuildInputs

	ImageTags     []string
	EnvVars       map[string]*string
	SecretEnvVars map[string]string
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
	if data.LogStore == nil {
		data.LogStore = tasklog.NewNullStore()
	}

	defer func() {
		if r := recover(); r != nil {
			err = errors.Join(err, hperrors.NewPanic(r))
		}
	}()

	data.Inputs, err = s.buildInputs(ctx, db, req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	data.EnvVars = data.Inputs.EnvVars
	data.SecretEnvVars = data.Inputs.SecretEnvVars
	data.RegistryAuths = data.Inputs.RegistryAuths
	if len(data.Inputs.Secrets) > 0 {
		data.LogStore.UpdateRedactorAddSecrets(data.Inputs.Secrets)
	}

	err = s.imageBuild(ctx, db, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Check if the context was canceled
	if err := ctx.Err(); err != nil {
		return nil, hperrors.Wrap(err)
	}

	err = s.imagePush(ctx, data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return resp, err
}
