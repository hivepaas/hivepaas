package clustersecretserviceimpl

import (
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

const (
	secretDefaultFileUID  = "0"
	secretDefaultFileGID  = "0"
	secretDefaultFileMode = 444
)

const (
	itemRemovalRetryMax   = 2
	itemRemovalRetryDelay = 2 * time.Second
	itemRemovalRetryIncr  = 1 * time.Second
)

func (s *service) HasSecretChanges(
	newSecret, oldSecret *entity.Secret,
) bool {
	if newSecret == nil || oldSecret == nil {
		return newSecret != oldSecret
	}

	if newSecret.Key != oldSecret.Key ||
		newSecret.Base64 != oldSecret.Base64 {
		return true
	}

	equal, err := newSecret.Value.Equal(&oldSecret.Value)
	if err != nil || !equal {
		return true
	}

	if (newSecret.SwarmRef == nil) != (oldSecret.SwarmRef == nil) {
		return true
	}

	if newSecret.SwarmRef != nil {
		if newSecret.SwarmRef.SecretID != oldSecret.SwarmRef.SecretID ||
			newSecret.SwarmRef.SecretName != oldSecret.SwarmRef.SecretName {
			return true
		}

		if (newSecret.SwarmRef.File == nil) != (oldSecret.SwarmRef.File == nil) {
			return true
		}

		if newSecret.SwarmRef.File != nil && *newSecret.SwarmRef.File != *oldSecret.SwarmRef.File {
			return true
		}
	}

	return false
}
