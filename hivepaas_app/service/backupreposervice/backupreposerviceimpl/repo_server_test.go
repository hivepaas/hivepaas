package backupreposerviceimpl

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/reposerveragentuc"
	"github.com/hivepaas/hivepaas/services/backup"
)

// fakeAgent answers where a node's agent is, and runs a fake server session.
type fakeAgent struct {
	addr     string
	addrNode string
	req      *reposerveragentuc.RunReq
	openErr  error
	closed   bool
	died     error
}

func (f *fakeAgent) service() *service {
	return &service{
		agentAddr: func(_ context.Context, nodeID, nodeLabel string) (string, error) {
			f.addrNode = nodeID + nodeLabel
			return f.addr, nil
		},
		openRepoServer: func(_ context.Context, addr string, req *reposerveragentuc.RunReq) (*openedRepoServer, error) {
			f.req = req
			if f.openErr != nil {
				return nil, f.openErr
			}
			return &openedRepoServer{
				port: 40123, fingerprint: "ab12cd",
				close: func() error { f.closed = true; return nil },
				err:   func() error { return f.died },
			}, nil
		},
	}
}

func volumeRepoStorage() *backup.Storage {
	return &backup.Storage{
		RepositoryPassword: "repo-password",
		StorageLocal:       &backup.StorageLocal{Path: "/host/srv/repos/r1", NodeID: "node-2"},
	}
}

// A server is opened on the agent of the repository's node, on the address the
// agent is reached by, for the user the snapshots are written as.
func TestOpenRepoServerOnTheRepositorysNode(t *testing.T) {
	agent := &fakeAgent{addr: "10.0.1.5:10001"}

	session, err := agent.service().openRepoServerOn(context.Background(), volumeRepoStorage(), "hivepaas@data-backup")

	if !assert.NoError(t, err) {
		return
	}
	assert.Equal(t, "node-2", agent.addrNode)
	assert.Equal(t, &reposerveragentuc.RunReq{
		RepoPath: "/host/srv/repos/r1", RepoPassword: "repo-password",
		Username: "hivepaas@data-backup", UserPassword: repoServerUserPassword("repo-password", "hivepaas@data-backup"),
		ListenHost: "10.0.1.5",
	}, agent.req)
	assert.Equal(t, "https://10.0.1.5:40123", session.URL)
	assert.Equal(t, "ab12cd", session.Fingerprint)
	assert.Equal(t, "hivepaas", session.Username)
	assert.Equal(t, "data-backup", session.Hostname)
	assert.Equal(t, agent.req.UserPassword, session.Password)

	assert.NoError(t, session.Close())
	assert.True(t, agent.closed)
}

// Only a repository on a volume has a server; the user is user@host.
func TestOpenRepoServerRefusesWhatItCannotServe(t *testing.T) {
	agent := &fakeAgent{addr: "10.0.1.5:10001"}
	cloud := &backup.Storage{RepositoryPassword: "p", StorageS3: &backup.StorageS3{Bucket: "b"}}

	_, err := agent.service().openRepoServerOn(context.Background(), cloud, "hivepaas@data-backup")
	assert.Error(t, err)

	_, err = agent.service().openRepoServerOn(context.Background(), volumeRepoStorage(), "hivepaas")
	assert.Error(t, err)
}

// The user's password is the repository's, hashed with the user: the same in
// every session, so two sessions at once do not log each other out.
func TestRepoServerUserPassword(t *testing.T) {
	password := repoServerUserPassword("repo-password", "hivepaas@data-backup")

	assert.Len(t, password, 64)
	assert.Equal(t, password, repoServerUserPassword("repo-password", "hivepaas@data-backup"))
	assert.NotEqual(t, password, repoServerUserPassword("other-password", "hivepaas@data-backup"))
	assert.NotEqual(t, password, repoServerUserPassword("repo-password", "hivepaas@restore"))
	assert.False(t, strings.Contains(password, "repo-password"))
}

// A client of a session reaches the server by its URL, pinned by fingerprint,
// as the session's user, from a config file of its own.
func TestRepoServerStorage(t *testing.T) {
	session := &backupreposervice.RepoServerSession{
		URL: "https://10.0.1.5:40123", Fingerprint: "ab12cd",
		Username: "hivepaas", Hostname: "data-backup", Password: "user-password",
	}

	storage := repoServerStorage(session, "r1")

	assert.Equal(t, "user-password", storage.RepositoryPassword)
	assert.Equal(t, &backup.StorageServer{
		URL: "https://10.0.1.5:40123", Fingerprint: "ab12cd", Username: "hivepaas", Hostname: "data-backup",
	}, storage.StorageServer)
	assert.True(t, strings.HasPrefix(storage.ConfigFile, "/tmp/hivepaas/backup-repos/r1/server-"), storage.ConfigFile)
	assert.NotEqual(t, storage.ConfigFile, repoServerStorage(session, "r1").ConfigFile,
		"two sessions at once do not share a config file")
}

// A backup through a server closes the session whatever happens, and says the
// server was gone when that is why it failed.
func TestWithRepoServerClosesTheSession(t *testing.T) {
	agent := &fakeAgent{addr: "10.0.1.5:10001"}

	err := agent.service().withRepoServer(context.Background(), volumeRepoStorage(), "hivepaas@data-backup", "r1",
		func(*backup.Storage) error { return nil })
	assert.NoError(t, err)
	assert.True(t, agent.closed)

	agent = &fakeAgent{addr: "10.0.1.5:10001", died: errors.New("repository gone")}
	err = agent.service().withRepoServer(context.Background(), volumeRepoStorage(), "hivepaas@data-backup", "r1",
		func(*backup.Storage) error { return errors.New("kopia: connection refused") })
	assert.ErrorContains(t, err, "repository gone")
	assert.True(t, agent.closed)

	agent = &fakeAgent{addr: "10.0.1.5:10001", openErr: errors.New("invalid repository password")}
	err = agent.service().withRepoServer(context.Background(), volumeRepoStorage(), "hivepaas@data-backup", "r1",
		func(*backup.Storage) error { t.Error("no backup without a server"); return nil })
	assert.ErrorContains(t, err, "invalid repository password")
}
