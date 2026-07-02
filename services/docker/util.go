package docker

import (
	"github.com/moby/moby/api/pkg/authconfig"
	"github.com/moby/moby/api/types/registry"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func GenerateAuthHeader(auth *registry.AuthConfig) (string, error) {
	if auth.Username == "" || auth.Password == "" {
		return "", nil
	}
	h, err := authconfig.Encode(*auth)
	if err != nil {
		return "", apperrors.NewInfra(err)
	}
	return h, nil
}

func FilterAdd(f *client.Filters, key, value string) {
	if f == nil {
		return
	}
	if len(*f) == 0 {
		*f = make(client.Filters)
	}
	f.Add(key, value)
}

func TruncateCPUs(cpus, chunkSz float64) float64 {
	return float64(TruncateCPUsAsNano(cpus, chunkSz)) / UnitCPUNano
}

func TruncateCPUsAsNano(cpus, chunkSz float64) int64 {
	nanoCPUs := int64(cpus * UnitCPUNano)
	nanoChunkSz := int64(chunkSz * UnitCPUNano)
	return (nanoCPUs / nanoChunkSz) * nanoChunkSz
}
