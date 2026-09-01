package kopia

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

func TestClient_BuildS3Flags(t *testing.T) {
	c := NewClient(&backupmodel.Storage{
		RepositoryPassword: "secret-password",
		StorageS3: &backupmodel.StorageS3{
			Endpoint:  "https://r2.cloudflarestorage.com",
			Bucket:    "hivepaas-vault",
			Prefix:    "prod/backups",
			Region:    "auto",
			AccessKey: "my-access-key",
			SecretKey: "my-secret-key",
		},
	}, nil)

	flags, err := c.buildStorageFlags()
	assert.NoError(t, err)
	assert.Contains(t, flags, "s3")
	assert.Contains(t, flags, "--bucket=hivepaas-vault")
	assert.Contains(t, flags, "--endpoint=r2.cloudflarestorage.com")
	assert.Contains(t, flags, "--prefix=prod/backups/")
	assert.Contains(t, flags, "--region=auto")

	// Credentials must never reach argv: kopia picks them up from the environment instead.
	assert.NotContains(t, flags, "--access-key=my-access-key")
	assert.NotContains(t, flags, "--secret-access-key=my-secret-key")
	for _, flag := range flags {
		assert.NotContains(t, flag, "my-access-key")
		assert.NotContains(t, flag, "my-secret-key")
	}

	env := c.buildEnv("")
	assert.Contains(t, env, "AWS_ACCESS_KEY_ID=my-access-key")
	assert.Contains(t, env, "AWS_SECRET_ACCESS_KEY=my-secret-key")
	assert.Contains(t, env, "KOPIA_PASSWORD=secret-password")
}

// A plaintext endpoint has to say so, otherwise kopia assumes HTTPS and cannot reach a
// self-hosted S3 that serves plain HTTP.
func TestClient_BuildS3Flags_PlaintextEndpoint(t *testing.T) {
	newClient := func(endpoint string) *Client {
		return NewClient(&backupmodel.Storage{
			StorageS3: &backupmodel.StorageS3{Endpoint: endpoint, Bucket: "b"},
		}, nil)
	}

	flags, err := newClient("http://minio.internal:9000").buildStorageFlags()
	assert.NoError(t, err)
	assert.Contains(t, flags, "--endpoint=minio.internal:9000")
	assert.Contains(t, flags, "--disable-tls")

	flags, err = newClient("https://s3.amazonaws.com").buildStorageFlags()
	assert.NoError(t, err)
	assert.Contains(t, flags, "--endpoint=s3.amazonaws.com")
	assert.NotContains(t, flags, "--disable-tls")

	flags, err = newClient("s3.amazonaws.com").buildStorageFlags()
	assert.NoError(t, err)
	assert.NotContains(t, flags, "--disable-tls")
}

func TestClient_BuildLocalFlags(t *testing.T) {
	c := NewClient(&backupmodel.Storage{
		RepositoryPassword: "secret-password",
		StorageLocal: &backupmodel.StorageLocal{
			Path:      "/mnt/backups",
			NodeID:    "node-1",
			NodeLabel: "backup-worker",
		},
	}, nil)

	flags, err := c.buildStorageFlags()
	assert.NoError(t, err)
	assert.Contains(t, flags, "filesystem")
	assert.Contains(t, flags, "--path=/mnt/backups")
}
