package imagebuildserviceimpl

import (
	"context"

	"github.com/moby/moby/api/types/registry"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// ResolveBuildInputs reads what a build takes from settings - its variables, the
// project's registries, the registry it pushes to - and opens the secrets in
// them. Everything that needs the data encryption key in a build is here.
func (s *service) ResolveBuildInputs(
	ctx context.Context,
	db database.IDB,
	req *imagebuildservice.ImageBuildReq,
) (*imagebuildservice.BuildInputs, error) {
	inputs := &imagebuildservice.BuildInputs{}
	secrets := make(map[string]struct{})

	if req.PushToRegistry.ID != "" {
		auth, err := s.resolvePushRegistry(ctx, db, req)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		inputs.PushRegistry = auth
		secrets[auth.Password] = struct{}{}
	}

	envVars, secretEnvVars, envSecrets, err := s.calcBuildEnvVars(ctx, db, req.App)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	inputs.EnvVars, inputs.SecretEnvVars = envVars, secretEnvVars

	auths, authSecrets, err := s.calcBuildRegistryAuths(ctx, db, req.App)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	inputs.RegistryAuths = auths

	for _, secret := range append(envSecrets, authSecrets...) {
		secrets[secret] = struct{}{}
	}
	delete(secrets, "")
	inputs.Secrets = gofn.MapKeys(secrets)
	return inputs, nil
}

func (s *service) resolvePushRegistry(
	ctx context.Context,
	db database.IDB,
	req *imagebuildservice.ImageBuildReq,
) (*registry.AuthConfig, error) {
	var refObjects *entity.RefObjects
	err := s.settingService.LoadRefObjectsByIDs(ctx, db, &refObjects, req.App.GetObjectScope(), true,
		&entity.RefObjectIDs{RefSettingIDs: []string{req.PushToRegistry.ID}})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	var setting *entity.Setting
	if refObjects != nil {
		setting = refObjects.RefSettings[req.PushToRegistry.ID]
	}
	if setting == nil {
		return nil, hperrors.NewMissing("Registry auth to push image")
	}
	regAuth, err := setting.AsRegistryAuth()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	password, err := regAuth.Password.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &registry.AuthConfig{
		Username:      regAuth.Username,
		Password:      password,
		ServerAddress: regAuth.Address,
	}, nil
}

// buildInputs is the inputs a build works from: the ones it was given, or, for
// a build in the app, the ones resolved here.
func (s *service) buildInputs(
	ctx context.Context,
	db database.IDB,
	req *imagebuildservice.ImageBuildReq,
) (*imagebuildservice.BuildInputs, error) {
	if req.Inputs != nil {
		return req.Inputs, nil
	}
	return s.ResolveBuildInputs(ctx, db, req)
}

// pushRegistry is the registry the image's references are named after, nil when
// the image is not pushed.
func pushRegistry(inputs *imagebuildservice.BuildInputs) *entity.RegistryAuth {
	if inputs.PushRegistry == nil {
		return nil
	}
	return &entity.RegistryAuth{
		Address:  inputs.PushRegistry.ServerAddress,
		Username: inputs.PushRegistry.Username,
	}
}

// pushAuthHeader signs a push in to the registry the image goes to.
func pushAuthHeader(inputs *imagebuildservice.BuildInputs) (string, error) {
	header, err := docker.GenerateAuthHeader(inputs.PushRegistry)
	return header, hperrors.Wrap(err)
}
