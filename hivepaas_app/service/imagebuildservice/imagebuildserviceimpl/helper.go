package imagebuildserviceimpl

import (
	"context"
	"fmt"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/envvarservice"
)

func (s *service) calcBuildImageTags(
	imageTags []string,
	data *imageBuildData,
) ([]string, error) {
	// If `pushToRegistry` is set in the settings, need to prepend the registry domain and
	// username to the tags.
	// E.g. `app_name:latest` will likely become `docker.io/username/app_name:latest`
	var regAuth *entity.RegistryAuth
	if data.PushToRegistry.ID != "" {
		regAuthSetting := data.RefObjects.RefSettings[data.PushToRegistry.ID]
		if regAuthSetting == nil {
			return nil, apperrors.NewMissing("Registry auth to push image")
		}
		regAuth = regAuthSetting.MustAsRegistryAuth()
	}

	if len(imageTags) > 0 {
		if regAuth == nil {
			return imageTags, nil
		}
		for i, imageTag := range imageTags {
			_, _, found := strings.Cut(imageTag, "/")
			if found {
				continue
			}
			imageTags[i] = regAuth.Address + "/" + regAuth.Username + "/" + imageTag
		}
		return imageTags, nil
	}

	imageName := data.ImageName
	if imageName == "" || imageName == "auto" {
		imageName = data.App.GetAutoImageName()
	}

	commitHashPortion := data.CommitHash[:7]
	tagCurrent := fmt.Sprintf("%s:%s", imageName, commitHashPortion)

	// If `pushToRegistry` is set in the settings, need to prepend the registry domain and
	// username to the tags.
	// E.g. `app_name:latest` will likely become `docker.io/username/app_name:latest`
	if regAuth != nil {
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
	envResp, err := s.envVarService.BuildEnvVarsInApp(ctx, db, &envvarservice.BuildEnvVarsInAppReq{
		App: data.App,
		LoadOptions: envvarservice.EnvLoadOptions{
			BuildPhase: true,
		},
		BuildOptions: envvarservice.EnvBuildOptions{
			BuildPhaseOnly: true,
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if data.LogStore != nil && len(envResp.EnvVars) > 0 {
		secrets := make(map[string]struct{}, 10) //nolint:mnd
		for _, env := range envResp.EnvVars {
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

	result := make(map[string]*string, len(envResp.EnvVars))
	for _, envVar := range envResp.EnvVars {
		result[envVar.Key] = &envVar.Value
	}

	return result, nil
}
