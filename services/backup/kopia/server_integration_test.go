package kopia

import (
	"bufio"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/services/backup/backupmodel"
)

var (
	serverFingerprintLine = regexp.MustCompile(`SERVER CERT SHA256: ([0-9a-f]+)`)
	serverAddressLine     = regexp.MustCompile(`SERVER ADDRESS: https://.*:(\d+)$`)
)

// startTestServer runs a kopia repository server on a fresh filesystem
// repository, with the user hivepaas@data-backup, and gives the storage a client
// reaches it with. The server stops with the test.
func startTestServer(t *testing.T) *backupmodel.Storage {
	t.Helper()
	owner, baseDir := newTestRepo(t, "repo")
	mustNoError(t, owner.InitRepo(context.Background(), &backupmodel.InitRepoOptions{}))
	ownerConfig := filepath.Join(baseDir, "repo.config")

	userAdd := exec.Command("kopia", "--config-file="+ownerConfig, "server", "user", "add",
		"hivepaas@data-backup", "--user-password=user-password")
	userAdd.Env = append(os.Environ(), "KOPIA_PASSWORD=integration-test-password")
	mustNoError(t, userAdd.Run())

	server := exec.Command("kopia", "--config-file="+ownerConfig, "server", "start",
		"--address=127.0.0.1:0", "--tls-generate-cert",
		"--tls-cert-file="+filepath.Join(baseDir, "cert.pem"), "--tls-key-file="+filepath.Join(baseDir, "key.pem"),
		"--server-username=control", "--server-password=control-password")
	server.Env = append(os.Environ(), "KOPIA_PASSWORD=integration-test-password", "KOPIA_CHECK_FOR_UPDATES=false")
	output, err := server.StderrPipe()
	mustNoError(t, err)
	server.Stdout = server.Stderr
	mustNoError(t, server.Start())
	t.Cleanup(func() {
		_ = server.Process.Kill()
		_ = server.Wait()
	})

	var fingerprint, port string
	lines := bufio.NewScanner(output)
	deadline := time.Now().Add(30 * time.Second)
	for (fingerprint == "" || port == "") && time.Now().Before(deadline) && lines.Scan() {
		if m := serverFingerprintLine.FindStringSubmatch(lines.Text()); m != nil {
			fingerprint = m[1]
		}
		if m := serverAddressLine.FindStringSubmatch(strings.TrimSpace(lines.Text())); m != nil {
			port = m[1]
		}
	}
	if fingerprint == "" || port == "" {
		t.Fatal("the kopia server did not start")
	}
	// The server blocks once its output is not read.
	go func() {
		for lines.Scan() {
			continue
		}
	}()

	return &backupmodel.Storage{
		RepositoryPassword: "user-password",
		StorageServer: &backupmodel.StorageServer{
			URL: "https://127.0.0.1:" + port, Fingerprint: fingerprint,
			Username: "hivepaas", Hostname: "data-backup",
		},
		ConfigFile: filepath.Join(t.TempDir(), "client.config"),
	}
}

// Through a repository server, a client takes a stream and a directory under its
// own identity, is refused another, and disconnects leaving no config behind.
func TestIntegration_Server_BackupsUnderTheUsersIdentity(t *testing.T) {
	if _, err := exec.LookPath("kopia"); err != nil {
		t.Skip("kopia binary not installed, skipping integration test")
	}
	storage := startTestServer(t)
	client := NewClient(storage, backupmodel.DefaultCommandExecutor)
	ctx := context.Background()

	mustNoError(t, client.ConnectRepo(ctx))

	stream, err := client.BackupStream(ctx, strings.NewReader("CREATE TABLE t();\n"), "db.sql",
		&backupmodel.BackupOptions{Source: "hivepaas@data-backup:/j1", Tags: []string{"hivepaas.run:t1"}})
	mustNoError(t, err)
	mustNotNil(t, stream.Item)
	assert.Equal(t, "data-backup", stream.Item.Hostname)

	dataDir := t.TempDir()
	mustNoError(t, os.WriteFile(filepath.Join(dataDir, "f1.txt"), []byte("hello"), 0o600))
	dir, err := client.BackupDirectory(ctx, dataDir, &backupmodel.BackupOptions{Source: "hivepaas@data-backup:/j2"})
	mustNoError(t, err)
	assert.Equal(t, []string{"/j2"}, dir.Item.Paths)

	_, err = client.BackupStream(ctx, strings.NewReader("x"), "x.sql",
		&backupmodel.BackupOptions{Source: "someone@elsewhere:/k"})
	assert.ErrorContains(t, err, "access denied")

	mustNoError(t, client.DisconnectRepo(ctx))
	_, err = os.Stat(storage.ConfigFile)
	assert.True(t, os.IsNotExist(err), "the client's config file is removed")
}
