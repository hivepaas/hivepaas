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
	if newConfig == nil && oldConfig == nil {
		return false
	}
	if newConfig == nil || oldConfig == nil {
		return true
	}

	if newConfig.Name != oldConfig.Name || newConfig.Content != oldConfig.Content ||
		newConfig.Base64 != oldConfig.Base64 {
		return true
	}
	if newConfig.SwarmRef == nil && oldConfig.SwarmRef == nil {
		return false
	}
	if newConfig.SwarmRef == nil || oldConfig.SwarmRef == nil {
		return true
	}
	if newConfig.SwarmRef.ConfigID != oldConfig.SwarmRef.ConfigID {
		return true
	}
	if newConfig.SwarmRef.File == nil && oldConfig.SwarmRef.File == nil {
		return false
	}
	if newConfig.SwarmRef.File == nil || oldConfig.SwarmRef.File == nil {
		return true
	}
	if newConfig.SwarmRef.File.Name != oldConfig.SwarmRef.File.Name ||
		newConfig.SwarmRef.File.Mode != oldConfig.SwarmRef.File.Mode ||
		newConfig.SwarmRef.File.UID != oldConfig.SwarmRef.File.UID ||
		newConfig.SwarmRef.File.GID != oldConfig.SwarmRef.File.GID {
		return true
	}
	return false
}
