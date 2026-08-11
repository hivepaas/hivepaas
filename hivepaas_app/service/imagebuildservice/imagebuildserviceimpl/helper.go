package imagebuildserviceimpl

import (
	"fmt"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
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
