// Package artifactagentuc keeps the artifacts of this node: files a managed
// template's engine mounts into an app's container, downloaded here from where
// their project publishes them, checked against their hash, and kept by it.
//
// The cache is <HP_STORAGE_HOST_DIR>/cache/artifacts/sha256/<hex> on the host,
// the layout of the app templates' cache. Keeping a file by its hash is what lets
// two apps share it, an app move to a new version without the old file going
// from under the task still running it, and a node prove a file is what it was
// asked for.
package artifactagentuc

import (
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// cacheSubdir is where artifacts are kept under HivePaaS's host directory.
const cacheSubdir = "cache/artifacts/sha256"

// downloadTimeout bounds one download, so a source that trickles cannot hold a
// deploy for ever.
const downloadTimeout = 10 * time.Minute

type UC struct {
	logger logging.Logger
	// cacheDir is the cache as the agent's container sees it, under /host, and
	// hostCacheDir the same directory as the host sees it, which is what a bind
	// mount is given.
	cacheDir     string
	hostCacheDir string
	client       *http.Client

	// locks serialize what is done for one hash, so that calls at once download
	// it once. verified holds the hashes whose file this agent has checked since
	// it started: a file is hashed once, not on every deploy.
	locks    sync.Map
	verified sync.Map
}

func New(logger logging.Logger) *UC {
	hostCacheDir := filepath.Join(config.Current().Storage.HostDir, cacheSubdir)
	return newUC(logger, volumeservice.HostPathPrefix+hostCacheDir, hostCacheDir)
}

func newUC(logger logging.Logger, cacheDir, hostCacheDir string) *UC {
	return &UC{
		logger:       logger,
		cacheDir:     cacheDir,
		hostCacheDir: hostCacheDir,
		client:       &http.Client{Timeout: downloadTimeout},
	}
}

func (uc *UC) lockFor(sha string) *sync.Mutex {
	lock, _ := uc.locks.LoadOrStore(sha, &sync.Mutex{})
	return lock.(*sync.Mutex) //nolint:forcetypeassert // only mutexes are stored
}
