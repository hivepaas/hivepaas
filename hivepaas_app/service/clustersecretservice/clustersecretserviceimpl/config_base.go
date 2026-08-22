package clustersecretserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

const (
	configDefaultFileUID  = "0"
	configDefaultFileGID  = "0"
	configDefaultFileMode = 444
)

func (s *service) HasConfigChanges(
	newConfig, oldConfig *entity.ConfigFile,
) bool {
	if newConfig == nil || oldConfig == nil {
		return newConfig != oldConfig
	}

	if newConfig.Name != oldConfig.Name ||
		newConfig.Content != oldConfig.Content ||
		newConfig.Base64 != oldConfig.Base64 {
		return true
	}

	if (newConfig.SwarmRef == nil) != (oldConfig.SwarmRef == nil) {
		return true
	}

	if newConfig.SwarmRef != nil {
		if newConfig.SwarmRef.ConfigID != oldConfig.SwarmRef.ConfigID ||
			newConfig.SwarmRef.ConfigName != oldConfig.SwarmRef.ConfigName {
			return true
		}

		if (newConfig.SwarmRef.File == nil) != (oldConfig.SwarmRef.File == nil) {
			return true
		}

		if newConfig.SwarmRef.File != nil && *newConfig.SwarmRef.File != *oldConfig.SwarmRef.File {
			return true
		}
	}

	return false
}
