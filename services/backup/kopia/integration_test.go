package kopia

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

// newTestRepo builds a client pointed at a fresh filesystem repository under a temp dir. It skips
// the test when the kopia binary is not installed, so the suite stays runnable anywhere.
func newTestRepo(t *testing.T, repoName string) (*Client, string) {
	t.Helper()
	if _, err := exec.LookPath("kopia"); err != nil {
		t.Skip("kopia binary not installed, skipping integration test")
	}

	baseDir := t.TempDir()
	repoPath := filepath.Join(baseDir, repoName)
	mustNoError(t, exec.Command("mkdir", "-p", repoPath).Run())

	return NewClient(&backupmodel.Storage{
		RepositoryPassword: "integration-test-password",
		StorageLocal:       &backupmodel.StorageLocal{Path: repoPath},
		ConfigFile:         filepath.Join(baseDir, repoName+".config"),
	}, backupmodel.DefaultCommandExecutor), baseDir
}

func TestIntegration_InitRepoThenConnectAndList(t *testing.T) {
	client, baseDir := newTestRepo(t, "repo")
	ctx := context.Background()

	err := client.InitRepo(ctx, &backupmodel.InitRepoOptions{
		Description: "integration repo",
		PackSizeMB:  32,
		Compression: "zstd-fastest",
	})
	mustNoError(t, err)

	// A brand new repository holds nothing.
	listResp, err := client.ListSnapshots(ctx, nil)
	mustNoError(t, err)
	assert.Empty(t, listResp.Items)

	// Put something in it so the import flow has data to find.
	dataDir := filepath.Join(baseDir, "data")
	mustNoError(t, exec.Command("mkdir", "-p", dataDir).Run())
	mustNoError(t, exec.Command("sh", "-c", "echo hello > "+filepath.Join(dataDir, "f1.txt")).Run())

	backupResp, err := client.BackupDirectory(ctx, dataDir, &backupmodel.BackupOptions{Tags: []string{"app:demo"}})
	mustNoError(t, err)
	mustNotNil(t, backupResp.Item)
	assert.NotEmpty(t, backupResp.Item.ID)

	// This is what importing an existing repository does: connect with a config file that has
	// never seen this repository, then read back what is already stored.
	importClient := NewClient(&backupmodel.Storage{
		RepositoryPassword: "integration-test-password",
		StorageLocal:       &backupmodel.StorageLocal{Path: filepath.Join(baseDir, "repo")},
		ConfigFile:         filepath.Join(baseDir, "imported.config"),
	}, backupmodel.DefaultCommandExecutor)

	mustNoError(t, importClient.ConnectRepo(ctx))

	importedList, err := importClient.ListSnapshots(ctx, nil)
	mustNoError(t, err)
	mustLen(t, importedList.Items, 1)

	snapshot := importedList.Items[0]
	assert.Equal(t, backupResp.Item.ID, snapshot.ID)
	assert.Len(t, snapshot.ShortID, 8)
	assert.Contains(t, snapshot.Paths, dataDir)
	assert.NotZero(t, snapshot.SizeBytes)
	assert.False(t, snapshot.Time.IsZero())
	// Kopia's own "tag:" namespace must be stripped, otherwise the stored tag would be
	// "tag:app:demo" and would never match what a user searches for.
	assert.Contains(t, snapshot.Tags, "app:demo")
}

func TestIntegration_InitRepo_FailsOnExistingRepo(t *testing.T) {
	client, _ := newTestRepo(t, "repo")
	ctx := context.Background()

	mustNoError(t, client.InitRepo(ctx, nil))

	// Creating over an initialized location must fail rather than silently adopt it: that is what
	// makes `importExisting` a deliberate choice.
	mustError(t, client.InitRepo(ctx, nil))
}

func TestIntegration_ConnectRepo_FailsWhenMissing(t *testing.T) {
	client, baseDir := newTestRepo(t, "repo")
	ctx := context.Background()

	missing := NewClient(&backupmodel.Storage{
		RepositoryPassword: "integration-test-password",
		StorageLocal:       &backupmodel.StorageLocal{Path: filepath.Join(baseDir, "does-not-exist")},
		ConfigFile:         filepath.Join(baseDir, "missing.config"),
	}, backupmodel.DefaultCommandExecutor)

	err := missing.ConnectRepo(ctx)
	mustError(t, err)

	// The engine's own diagnostics have to reach the caller; without them a failure is just
	// "exit status 1" and there is nothing to act on.
	assert.Contains(t, err.Error(), "kopia repository connect failed")
	assert.Contains(t, err.Error(), "cannot access storage path")

	_ = client
}

// Two repositories must not disturb each other, which is only true while each one keeps its
// connection state in its own config file.
func TestIntegration_ReposAreIsolated(t *testing.T) {
	clientA, baseDir := newTestRepo(t, "repo-a")
	ctx := context.Background()
	mustNoError(t, clientA.InitRepo(ctx, nil))

	repoBPath := filepath.Join(baseDir, "repo-b")
	mustNoError(t, exec.Command("mkdir", "-p", repoBPath).Run())
	clientB := NewClient(&backupmodel.Storage{
		RepositoryPassword: "another-password",
		StorageLocal:       &backupmodel.StorageLocal{Path: repoBPath},
		ConfigFile:         filepath.Join(baseDir, "repo-b.config"),
	}, backupmodel.DefaultCommandExecutor)
	mustNoError(t, clientB.InitRepo(ctx, nil))

	dataDir := filepath.Join(baseDir, "data")
	mustNoError(t, exec.Command("mkdir", "-p", dataDir).Run())
	mustNoError(t, exec.Command("sh", "-c", "echo hello > "+filepath.Join(dataDir, "f1.txt")).Run())

	_, err := clientB.BackupDirectory(ctx, dataDir, nil)
	mustNoError(t, err)

	// Connecting B last must not redirect A's operations onto B.
	listA, err := clientA.ListSnapshots(ctx, nil)
	mustNoError(t, err)
	assert.Empty(t, listA.Items)

	listB, err := clientB.ListSnapshots(ctx, nil)
	mustNoError(t, err)
	assert.Len(t, listB.Items, 1)
}

func mustNoError(t *testing.T, err error) {
	t.Helper()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
}

func mustError(t *testing.T, err error) {
	t.Helper()
	if !assert.Error(t, err) {
		t.FailNow()
	}
}

func mustNotNil(t *testing.T, v any) {
	t.Helper()
	if !assert.NotNil(t, v) {
		t.FailNow()
	}
}

func mustLen(t *testing.T, v any, n int) {
	t.Helper()
	if !assert.Len(t, v, n) {
		t.FailNow()
	}
}

// Compression and pack size live inside the repository, so applying them has to be visible to a
// client that was never told about them.
func TestIntegration_ApplyRepoOptions(t *testing.T) {
	client, baseDir := newTestRepo(t, "repo")
	ctx := context.Background()
	mustNoError(t, client.InitRepo(ctx, nil))

	mustNoError(t, client.ApplyRepoOptions(ctx, &backupmodel.RepoOptions{
		PackSizeMB:  32,
		Compression: "zstd-fastest",
	}))

	fresh := NewClient(&backupmodel.Storage{
		RepositoryPassword: "integration-test-password",
		StorageLocal:       &backupmodel.StorageLocal{Path: filepath.Join(baseDir, "repo")},
		ConfigFile:         filepath.Join(baseDir, "fresh.config"),
	}, backupmodel.DefaultCommandExecutor)
	mustNoError(t, fresh.ConnectRepo(ctx))

	var out bytes.Buffer
	_, err := fresh.execCommand(ctx, []string{"policy", "get", "--global"}, func(o *execOptions) {
		o.stdout = &out
	})
	mustNoError(t, err)
	assert.Contains(t, out.String(), "zstd-fastest")

	out.Reset()
	_, err = fresh.execCommand(ctx, []string{"repository", "status"}, func(o *execOptions) {
		o.stdout = &out
	})
	mustNoError(t, err)
	assert.Contains(t, out.String(), "33.6 MB")
}

// Clearing the compression must actually stop the repository compressing, which only happens
// because an empty value is normalized to the engine's explicit "none".
func TestIntegration_ApplyRepoOptions_ClearingCompression(t *testing.T) {
	client, baseDir := newTestRepo(t, "repo")
	ctx := context.Background()
	mustNoError(t, client.InitRepo(ctx, nil))

	opts := backupmodel.NewRepoOptions(0, "zstd-fastest")
	mustNoError(t, client.ApplyRepoOptions(ctx, &opts))

	// This is what the update path builds when the user empties the compression field.
	cleared := backupmodel.NewRepoOptions(0, "")
	assert.Equal(t, backupmodel.CompressionNone, cleared.Compression)
	mustNoError(t, client.ApplyRepoOptions(ctx, &cleared))

	dataDir := filepath.Join(baseDir, "data")
	mustNoError(t, os.MkdirAll(dataDir, 0o755))
	// Highly compressible: with compression on this would shrink by orders of magnitude.
	mustNoError(t, os.WriteFile(filepath.Join(dataDir, "big.txt"),
		bytes.Repeat([]byte("the quick brown fox jumps over the lazy dog. "), 20000), 0o600))

	_, err := client.BackupDirectory(ctx, dataDir, nil)
	mustNoError(t, err)

	size := dirSize(t, filepath.Join(baseDir, "repo"))
	assert.Greater(t, size, int64(500_000), "repository should hold the data uncompressed")
}

func dirSize(t *testing.T, dir string) int64 {
	t.Helper()
	var total int64
	err := filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			total += info.Size()
		}
		return nil
	})
	mustNoError(t, err)
	return total
}

// Importing an existing repository must adopt its settings rather than assume the request's:
// what comes back has to be what the repository is actually running with.
func TestIntegration_ReadRepoConfig(t *testing.T) {
	client, baseDir := newTestRepo(t, "repo")
	ctx := context.Background()
	mustNoError(t, client.InitRepo(ctx, nil))

	mustNoError(t, client.ApplyRepoOptions(ctx, &backupmodel.RepoOptions{
		PackSizeMB:  48,
		Compression: "zstd-better-compression",
	}))

	// This is the import path: a client that has never seen the repository.
	importClient := NewClient(&backupmodel.Storage{
		RepositoryPassword: "integration-test-password",
		StorageLocal:       &backupmodel.StorageLocal{Path: filepath.Join(baseDir, "repo")},
		ConfigFile:         filepath.Join(baseDir, "imported.config"),
	}, backupmodel.DefaultCommandExecutor)
	mustNoError(t, importClient.ConnectRepo(ctx))

	config, err := importClient.ReadRepoConfig(ctx)
	mustNoError(t, err)

	assert.Equal(t, 48, config.PackSizeMB)
	assert.Equal(t, "zstd-better-compression", config.Compression)

	// The retention policy always has values, so an import can show what pruning will really do.
	mustNotNil(t, config.Retention)
	assert.Positive(t, config.Retention.KeepLast)
	assert.Positive(t, config.Retention.KeepDaily)
	assert.Positive(t, config.Retention.KeepWeekly)
	assert.Positive(t, config.Retention.KeepMonthly)
}

// A repository that was never configured reports the engine defaults, not empty values, so an
// import of one still stores something truthful.
func TestIntegration_ReadRepoConfig_UnconfiguredRepo(t *testing.T) {
	client, _ := newTestRepo(t, "repo")
	ctx := context.Background()
	mustNoError(t, client.InitRepo(ctx, nil))

	config, err := client.ReadRepoConfig(ctx)
	mustNoError(t, err)

	assert.Equal(t, backupmodel.CompressionNone, config.Compression)
	assert.Positive(t, config.PackSizeMB)
	mustNotNil(t, config.Retention)
}

// Creating a repository without asking for a pack size still leaves the repository with one, so
// reading the config back is what keeps the stored setting from claiming there is none.
func TestIntegration_ReadRepoConfig_AfterCreateWithoutOptions(t *testing.T) {
	client, _ := newTestRepo(t, "repo")
	ctx := context.Background()
	mustNoError(t, client.InitRepo(ctx, &backupmodel.InitRepoOptions{Description: "no options given"}))

	config, err := client.ReadRepoConfig(ctx)
	mustNoError(t, err)

	// The engine defaults to 20 MiB; storing the requested 0 would misreport the repository.
	assert.Equal(t, 20, config.PackSizeMB)
	assert.Equal(t, backupmodel.CompressionNone, config.Compression)
}

// What ReadRepoConfig returns has to survive a round trip through ApplyRepoOptions unchanged,
// otherwise every update would look like a change and re-apply settings forever.
func TestIntegration_RepoConfig_RoundTripsWithoutDrift(t *testing.T) {
	client, _ := newTestRepo(t, "repo")
	ctx := context.Background()
	mustNoError(t, client.InitRepo(ctx, nil))
	mustNoError(t, client.ApplyRepoOptions(ctx, &backupmodel.RepoOptions{
		PackSizeMB:  32,
		Compression: "zstd-fastest",
	}))

	first, err := client.ReadRepoConfig(ctx)
	mustNoError(t, err)

	// Feed exactly what was read back in, the way an unchanged update would.
	echoed := backupmodel.NewRepoOptions(first.PackSizeMB, first.Compression)
	assert.Equal(t, first.RepoOptions, echoed, "a read-back config must compare equal to itself")

	mustNoError(t, client.ApplyRepoOptions(ctx, &echoed))
	second, err := client.ReadRepoConfig(ctx)
	mustNoError(t, err)
	assert.Equal(t, first.RepoOptions, second.RepoOptions)
}
