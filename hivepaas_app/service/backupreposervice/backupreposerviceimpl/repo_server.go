package backupreposerviceimpl

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/reposerverservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/reposerveragentuc"
	"github.com/hivepaas/hivepaas/services/backup"
)

// openedRepoServer is a server an agent runs for a session.
type openedRepoServer struct {
	port        int
	fingerprint string
	close       func() error
	// err is why the server went by itself, if it did.
	err func() error
}

// openRepoServerOnAgent asks an agent to run a repository server.
func openRepoServerOnAgent(
	ctx context.Context,
	agentAddr string,
	req *reposerveragentuc.RunReq,
) (*openedRepoServer, error) {
	session, err := reposerverservice.Open(ctx, agentAddr, req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &openedRepoServer{
		port: session.Port, fingerprint: session.Fingerprint,
		close: session.Close, err: session.Err,
	}, nil
}

func (s *service) agentAddrForNode(ctx context.Context, nodeID, nodeLabel string) (string, error) {
	if nodeID != "" {
		addr, err := s.agentService.GetAgentAddrForNode(ctx, nodeID)
		return addr, hperrors.Wrap(err)
	}
	addr, err := s.agentService.GetAgentAddrForNodeLabel(ctx, nodeLabel)
	return addr, hperrors.Wrap(err)
}

func (s *service) OpenRepoServer(
	ctx context.Context,
	db database.IDB,
	req *backupreposervice.OpenRepoServerReq,
) (*backupreposervice.RepoServerSession, error) {
	repo, err := req.RepoSetting.AsBackupRepo()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	storage, _, err := s.buildStorage(ctx, db, req.Scope, repo, req.RepoSetting.ID, req.RefObjects, "")
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	session, err := s.openRepoServerOn(ctx, storage, req.Username)
	return session, hperrors.Wrap(err)
}

// openRepoServerOn runs a server for the repository on its volume's node, on the
// address its agent is reached by, for the user user@host.
func (s *service) openRepoServerOn(
	ctx context.Context,
	storage *backup.Storage,
	username string,
) (*backupreposervice.RepoServerSession, error) {
	local := storage.StorageLocal
	if local == nil {
		return nil, hperrors.Wrap(hperrors.ErrBadRequest).
			WithExtraDetail("only a repository on a volume is served by a repository server")
	}
	user, host, ok := strings.Cut(username, "@")
	if !ok || user == "" || host == "" {
		return nil, hperrors.Wrap(hperrors.ErrBadRequest).WithExtraDetail("a server user is user@host")
	}
	agentAddr, err := s.agentAddr(ctx, local.NodeID, local.NodeLabel)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	listenHost, _, err := net.SplitHostPort(agentAddr)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	password := repoServerUserPassword(storage.RepositoryPassword, username)
	opened, err := s.openRepoServer(ctx, agentAddr, &reposerveragentuc.RunReq{
		RepoPath:     local.Path,
		RepoPassword: storage.RepositoryPassword,
		Username:     username,
		UserPassword: password,
		ListenHost:   listenHost,
	})
	if err != nil {
		return nil, hperrors.Wrap(fmt.Errorf("the repository server on %s did not start: %w", listenHost, err))
	}
	return &backupreposervice.RepoServerSession{
		URL:         "https://" + net.JoinHostPort(listenHost, strconv.Itoa(opened.port)),
		Fingerprint: opened.fingerprint,
		Username:    user,
		Hostname:    host,
		Password:    password,
		Stop: func() error {
			if err := opened.close(); err != nil {
				return hperrors.Wrap(err)
			}
			return hperrors.Wrap(opened.err())
		},
	}, nil
}

// repoServerUserPassword is the server user's password: the repository's,
// hashed with the user. The same in every session, so two sessions on one
// repository at once do not log each other out.
func repoServerUserPassword(repoPassword, username string) string {
	mac := hmac.New(sha256.New, []byte(username))
	mac.Write([]byte(repoPassword))
	return hex.EncodeToString(mac.Sum(nil))
}

// repoServerStorage is how a client of the session reaches the repository: from
// a config file of its own, so two sessions at once do not share one.
func repoServerStorage(session *backupreposervice.RepoServerSession, repoID string) *backup.Storage {
	suffix := make([]byte, 8) //nolint:mnd
	_, _ = rand.Read(suffix)
	return &backup.Storage{
		RepositoryPassword: session.Password,
		StorageServer: &backup.StorageServer{
			URL: session.URL, Fingerprint: session.Fingerprint,
			Username: session.Username, Hostname: session.Hostname,
		},
		ConfigFile: filepath.Join(engineConfigDir, repoID, "server-"+hex.EncodeToString(suffix)+".config"),
	}
}

// withRepoServer runs fn with a client storage reaching a repository server for
// the storage's repository, and closes the session whatever fn does. A backup
// that failed because the server went says so.
func (s *service) withRepoServer(
	ctx context.Context,
	storage *backup.Storage,
	username string,
	repoID string,
	fn func(client *backup.Storage) error,
) (err error) {
	session, err := s.openRepoServerOn(ctx, storage, username)
	if err != nil {
		return hperrors.Wrap(err)
	}
	fnErr := fn(repoServerStorage(session, repoID))
	closeErr := session.Close()
	switch {
	case fnErr != nil && closeErr != nil:
		return hperrors.Wrap(errors.Join(fnErr, fmt.Errorf("the repository server stopped: %w", closeErr)))
	case fnErr != nil:
		return hperrors.Wrap(fnErr)
	}
	return hperrors.Wrap(closeErr)
}
