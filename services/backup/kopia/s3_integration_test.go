package kopia

import (
	"context"
	"fmt"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

// These run against an S3-compatible endpoint. Point HIVEPAAS_TEST_S3_* at one (a local MinIO
// works) to enable them; without it there is nothing to talk to and the test skips.
func s3TestConfig(t *testing.T) *backupmodel.StorageS3 {
	t.Helper()
	if _, err := exec.LookPath("kopia"); err != nil {
		t.Skip("kopia binary not installed, skipping integration test")
	}

	endpoint := os.Getenv("HIVEPAAS_TEST_S3_ENDPOINT")
	if endpoint == "" {
		t.Skip("HIVEPAAS_TEST_S3_ENDPOINT not set, skipping S3 integration test")
	}

	hostPort := strings.TrimPrefix(strings.TrimPrefix(endpoint, "https://"), "http://")
	conn, err := net.DialTimeout("tcp", hostPort, 2*time.Second)
	if err != nil {
		t.Skipf("S3 endpoint %s is not reachable: %v", endpoint, err)
	}
	_ = conn.Close()

	return &backupmodel.StorageS3{
		Endpoint:  endpoint,
		Region:    "us-east-1",
		Bucket:    os.Getenv("HIVEPAAS_TEST_S3_BUCKET"),
		AccessKey: os.Getenv("HIVEPAAS_TEST_S3_ACCESS_KEY"),
		SecretKey: os.Getenv("HIVEPAAS_TEST_S3_SECRET_KEY"),
	}
}

// uniqueTestPrefix keeps every run on its own repository: creating one over an existing
// repository is refused, so a fixed prefix would make the test pass only once.
func uniqueTestPrefix(name string) string {
	return fmt.Sprintf("hivepaas-test/%s-%d", name, time.Now().UnixNano())
}

// Credentials are handed to kopia through the environment, never through argv. This proves kopia
// actually authenticates that way: the command line carries no keys at all.
func TestIntegration_S3_CredentialsViaEnvironment(t *testing.T) {
	cfg := s3TestConfig(t)
	baseDir := t.TempDir()
	ctx := context.Background()

	cfg.Prefix = uniqueTestPrefix("env-creds")
	client := NewClient(&backupmodel.Storage{
		RepositoryPassword: "s3-integration-password",
		StorageS3:          cfg,
		ConfigFile:         filepath.Join(baseDir, "repo.config"),
	}, backupmodel.DefaultCommandExecutor)

	flags := client.buildS3Flags(cfg)
	for _, flag := range flags {
		assert.NotContains(t, flag, cfg.AccessKey)
		assert.NotContains(t, flag, cfg.SecretKey)
	}
	assert.NotContains(t, strings.Join(flags, " "), "--access-key")
	assert.NotContains(t, strings.Join(flags, " "), "--secret-access-key")

	// If kopia did not read the credentials from the environment this would fail to authenticate.
	if err := client.InitRepo(ctx, &backupmodel.InitRepoOptions{Description: "env creds"}); err != nil {
		t.Fatalf("InitRepo over S3 failed: %v", err)
	}

	dataDir := filepath.Join(baseDir, "data")
	if err := os.MkdirAll(dataDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dataDir, "f1.txt"), []byte("hello s3"), 0o600); err != nil {
		t.Fatal(err)
	}

	backupResp, err := client.BackupDirectory(ctx, dataDir, &backupmodel.BackupOptions{Tags: []string{"app:s3demo"}})
	if err != nil {
		t.Fatalf("BackupDirectory over S3 failed: %v", err)
	}

	// Import the same repository from a config file that has never seen it.
	importClient := NewClient(&backupmodel.Storage{
		RepositoryPassword: "s3-integration-password",
		StorageS3:          cfg,
		ConfigFile:         filepath.Join(baseDir, "imported.config"),
	}, backupmodel.DefaultCommandExecutor)

	if err := importClient.ConnectRepo(ctx); err != nil {
		t.Fatalf("ConnectRepo over S3 failed: %v", err)
	}

	listResp, err := importClient.ListSnapshots(ctx, nil)
	if err != nil {
		t.Fatalf("ListSnapshots over S3 failed: %v", err)
	}
	if !assert.Len(t, listResp.Items, 1) {
		return
	}
	assert.Equal(t, backupResp.Item.ID, listResp.Items[0].ID)
	assert.Contains(t, listResp.Items[0].Tags, "app:s3demo")
}

// A wrong secret must fail, otherwise the test above would pass even if kopia silently ignored
// the credentials entirely.
func TestIntegration_S3_WrongCredentialsFail(t *testing.T) {
	cfg := s3TestConfig(t)
	baseDir := t.TempDir()

	cfg.Prefix = uniqueTestPrefix("bad-creds")
	cfg.SecretKey = "definitely-not-the-right-secret"

	client := NewClient(&backupmodel.Storage{
		RepositoryPassword: "s3-integration-password",
		StorageS3:          cfg,
		ConfigFile:         filepath.Join(baseDir, "repo.config"),
	}, backupmodel.DefaultCommandExecutor)

	err := client.InitRepo(context.Background(), nil)
	assert.Error(t, err)
	assert.NotContains(t, err.Error(), cfg.SecretKey)
}
