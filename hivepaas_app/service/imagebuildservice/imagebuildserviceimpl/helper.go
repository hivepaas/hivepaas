package imagebuildserviceimpl

import (
	"context"
	"fmt"

	"github.com/moby/moby/api/types/registry"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) calcBuildImageTags(
	imageTags []string,
	data *imageBuildData,
) ([]string, error) {
	if len(imageTags) > 0 {
		return imageTags, nil
	}

	imageName := data.RepoSource.ImageName
	if imageName == "" || imageName == "auto" {
		imageName = data.App.GetAutoImageName()
	}

	commitHashPortion := data.RepoSource.CommitHash[:7]
	tagCurrent := fmt.Sprintf("%s:%s", imageName, commitHashPortion)

	// If `pushToRegistry` is set in the settings, need to prepend the registry domain and
	// username to the tags.
	// E.g. `app_name:latest` will likely become `docker.io/username/app_name:latest`
	repoSource := data.RepoSource
	if repoSource.PushToRegistry.ID != "" {
		regAuthSetting := data.RefObjects.RefSettings[repoSource.PushToRegistry.ID]
		if regAuthSetting == nil {
			return nil, apperrors.NewMissing("Registry auth to push image")
		}
		regAuth := regAuthSetting.MustAsRegistryAuth()
		tagCurrentWithReg := regAuth.Address + "/" + regAuth.Username + "/" + tagCurrent
		imageTags = append(imageTags, tagCurrentWithReg)
	}

	imageTags = append(imageTags, tagCurrent)
	return imageTags, nil
}

func (s *service) calcBuildEnvVars(
	ctx context.Context,
	db database.IDB,
	data *imageBuildData,
) (map[string]*string, error) {
	envResp, err := s.envVarService.ComputeAppEnvVars(ctx, db, &envvarservice.ComputeAppEnvVarsReq{
		App:            data.App,
		BuildPhaseOnly: true,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if data.LogStore != nil && len(envResp) > 0 {
		secrets := make(map[string]struct{}, 10) //nolint:mnd
		for _, env := range envResp {
			for secret := range env.RefSecrets {
				plainSecret, err := secret.Value.GetPlain()
				if err != nil {
					return nil, apperrors.Wrap(err)
				}
				secrets[plainSecret] = struct{}{}
			}
		}
		data.LogStore.UpdateRedactorAddSecrets(gofn.MapKeys(secrets))
	}

	result := make(map[string]*string, len(envResp))
	for _, envVar := range envResp {
		result[envVar.Key] = &envVar.Value
	}

	return result, nil
}

func (s *service) calcBuildRegistryAuths(
	ctx context.Context,
	db database.IDB,
	data *imageBuildData,
) (map[string]registry.AuthConfig, error) {
	settings, _, err := s.settingRepo.List(ctx, db, data.Project.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeRegistryAuth),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	result := make(map[string]registry.AuthConfig, len(settings))
	for _, setting := range settings {
		regAuth, err := setting.AsRegistryAuth()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		password, err := regAuth.Password.GetPlain()
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		result[regAuth.Address] = registry.AuthConfig{
			Username:      regAuth.Username,
			Password:      password,
			ServerAddress: regAuth.Address,
		}
	}

	return result, nil
}
