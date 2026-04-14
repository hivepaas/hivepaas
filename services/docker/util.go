package docker

import (
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/registry"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
)

func GenerateAuthHeader(username string, password string) (string, error) {
	if username == "" || password == "" {
		return "", nil
	}
	h, err := registry.EncodeAuthConfig(registry.AuthConfig{
		Username: username,
		Password: password,
	})
	if err != nil {
		return "", apperrors.NewInfra(err)
	}
	return h, nil
}

func FilterAdd(f *filters.Args, key, value string) {
	if f == nil {
		return
	}
	if f.Len() == 0 {
		*f = filters.NewArgs()
	}
	f.Add(key, value)
}
