package apptemplateserviceimpl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterepo"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
)

const (
	officialSourceID     = "official"
	rawGitHubBaseURL     = "https://raw.githubusercontent.com"
	fetchTimeout         = 15 * time.Second
	maxConcurrentFetches = 4
	cacheDirPerm         = 0o750
)

// officialSource serves the templates the signed release info pins.
//
// Nothing it returns is trusted because of where it came from. The release info
// is signed; the pin in it names the sha256 of index.json; the index names the
// sha256 of every template and icon; and every file is checked against its hash
// before anything reads it. The commit in the pin only chooses the URL - nothing
// here can check a git object id against what GitHub returns.
type officialSource struct {
	hpAppService hpappservice.Service
	baseURL      string
	client       *http.Client
	logger       logging.Logger
	// cacheDir is where verified files are kept by hash, so a restart or an
	// unreachable GitHub still serves every revision already fetched.
	cacheDir   func() string
	fetchSlots chan struct{}

	mu        sync.Mutex
	indexSHA  string
	indexMemo *templatemodel.Index
	// pinMemo holds the templates pin for a few seconds. Reading it opens and
	// verifies the signed release envelope, and one request asks for it several
	// times - the index, the revision, and every file that has to be downloaded.
	pinMemo   *base.TemplatesRef
	pinReadAt time.Time
}

// pinMemoTTL is how long the pin is reused within a request or a burst of them.
// It is short because it exists to collapse the repeats of one request, not to
// avoid reading release info - that has a cache of its own, thirty minutes long.
const pinMemoTTL = 30 * time.Second

func newOfficialSource(hpAppService hpappservice.Service, logger logging.Logger) *officialSource {
	return &officialSource{
		hpAppService: hpAppService,
		baseURL:      rawGitHubBaseURL,
		client:       &http.Client{Timeout: fetchTimeout},
		logger:       logger,
		cacheDir: func() string {
			return filepath.Join(config.Current().AppPath, "cache", "app-templates", "sha256")
		},
		fetchSlots: make(chan struct{}, maxConcurrentFetches),
	}
}

func (s *officialSource) ID() string {
	return officialSourceID
}

// pin reads the templates pin for the installation's channel from the release
// info, which is fetched, verified and cached by hpappservice.
func (s *officialSource) pin(ctx context.Context) (*base.TemplatesRef, error) {
	s.mu.Lock()
	if s.pinMemo != nil && time.Since(s.pinReadAt) < pinMemoTTL {
		pin := s.pinMemo
		s.mu.Unlock()
		return pin, nil
	}
	s.mu.Unlock()

	info, err := s.hpAppService.GetAppReleaseInfo(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	release := info.Stable
	if cfg := config.Current(); cfg != nil && cfg.IsBetaEnv() {
		release = info.Beta
	}
	if release == nil || release.Templates == nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable).
			WithExtraDetail("the release info pins no templates for this channel")
	}

	s.mu.Lock()
	s.pinMemo, s.pinReadAt = release.Templates, time.Now()
	s.mu.Unlock()
	return release.Templates, nil
}

func (s *officialSource) Revision(ctx context.Context) (string, error) {
	pin, err := s.pin(ctx)
	if err != nil {
		return "", err
	}
	return pin.Commit, nil
}

func (s *officialSource) Index(ctx context.Context) (*templatemodel.Index, error) {
	pin, err := s.pin(ctx)
	if err != nil {
		return nil, err
	}

	s.mu.Lock()
	if s.indexMemo != nil && s.indexSHA == pin.IndexSHA256 {
		index := s.indexMemo
		s.mu.Unlock()
		return index, nil
	}
	s.mu.Unlock()

	data, err := s.fetchVerified(ctx, pin, templaterepo.IndexFile, pin.IndexSHA256, templaterepo.MaxIndexSize)
	if err != nil {
		return nil, err
	}
	index, err := templatemodel.DecodeIndex(data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	s.mu.Lock()
	s.indexMemo, s.indexSHA = index, pin.IndexSHA256
	s.mu.Unlock()

	// Reaching here means this index is not the one held a moment ago - a new pin,
	// or the first read of this process - so whatever the previous revision left
	// in the cache is no longer named by anything.
	s.pruneCache(index, pin.IndexSHA256)
	return index, nil
}

// pruneCache removes the cached files the index does not name.
//
// The cache is addressed by content, so this needs no retention and no clock:
// the index names the sha256 of every template and icon of the revision being
// served, and a file whose name is not one of those belongs to a revision that
// has been replaced. It runs when the index changes, which is the only moment
// the answer can change.
//
// Nothing here is allowed to fail a request. A directory that cannot be read or
// a file that cannot be removed costs disk space until the next pin, and that is
// the whole consequence.
func (s *officialSource) pruneCache(index *templatemodel.Index, indexSHA string) {
	keep := make(map[string]bool, 2*len(index.Templates)+1)
	keep[indexSHA] = true
	for _, entry := range index.Templates {
		keep[entry.File.SHA256] = true
		keep[entry.Icon.SHA256] = true
	}

	dir := s.cacheDir()
	entries, err := os.ReadDir(dir)
	if err != nil {
		if !os.IsNotExist(err) {
			s.logger.Warnf("app templates: cannot read the cache directory to prune it: %s", err.Error())
		}
		return
	}

	removed, failed := 0, 0
	for _, entry := range entries {
		name := entry.Name()
		// A .tmp file is a write in progress, and the writer removes its own.
		if entry.IsDir() || keep[name] || strings.HasSuffix(name, ".tmp") {
			continue
		}
		if err = os.Remove(filepath.Join(dir, name)); err != nil {
			failed++
			continue
		}
		removed++
	}
	if removed > 0 || failed > 0 {
		s.logger.Infof("app templates: pruned %d cached file(s) of older revisions, %d could not be removed",
			removed, failed)
	}
}

func (s *officialSource) TemplateFile(ctx context.Context, entry *templatemodel.IndexEntry) ([]byte, error) {
	pin, err := s.pin(ctx)
	if err != nil {
		return nil, err
	}
	return s.fetchVerified(ctx, pin, entry.File.Path, entry.File.SHA256, templaterepo.MaxTemplateSize)
}

func (s *officialSource) Icon(ctx context.Context, sha256Hex string) ([]byte, error) {
	index, err := s.Index(ctx)
	if err != nil {
		return nil, err
	}
	entry := index.FindIcon(sha256Hex)
	if entry == nil {
		return nil, hperrors.NewNotFound("Icon")
	}
	pin, err := s.pin(ctx)
	if err != nil {
		return nil, err
	}
	return s.fetchVerified(ctx, pin, entry.Icon.Path, entry.Icon.SHA256, templaterepo.MaxIconSize)
}

// fetchVerified returns a file of the pinned revision whose sha256 must be want,
// from the cache when it has it and from GitHub otherwise.
func (s *officialSource) fetchVerified(
	ctx context.Context,
	pin *base.TemplatesRef,
	path, want string,
	maxSize int64,
) ([]byte, error) {
	if data, ok := s.readCache(want); ok {
		return data, nil
	}

	url := fmt.Sprintf("%s/%s/%s/%s", s.baseURL, pin.Repo, pin.Commit, path)
	data, err := s.download(ctx, url, path, maxSize)
	if err != nil {
		return nil, err
	}
	if hashHex(data) != want {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateVerificationFailed).
			WithExtraDetail("%s does not match its sha256", path)
	}
	s.writeCache(want, data)
	return data, nil
}

func (s *officialSource) download(ctx context.Context, url, path string, maxSize int64) ([]byte, error) {
	select {
	case s.fetchSlots <- struct{}{}:
	case <-ctx.Done():
		return nil, hperrors.Wrap(ctx.Err())
	}
	defer func() { <-s.fetchSlots }()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	res, err := s.client.Do(req)
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable).WithExtraDetail("%s: %s", path, err.Error())
	}
	defer func() { _ = res.Body.Close() }()

	if res.StatusCode != http.StatusOK {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable).WithExtraDetail("%s: %s", path, res.Status)
	}
	data, err := io.ReadAll(io.LimitReader(res.Body, maxSize+1))
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplatesUnavailable).WithExtraDetail("%s: %s", path, err.Error())
	}
	if int64(len(data)) > maxSize {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateFileTooLarge).
			WithExtraDetail("%s is larger than %d bytes", path, maxSize)
	}
	return data, nil
}

// readCache returns a cached file only if it still matches its hash. A file that
// does not is removed, and fetched again by the caller.
func (s *officialSource) readCache(sha256Hex string) ([]byte, bool) {
	path := filepath.Join(s.cacheDir(), sha256Hex)
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, false
	}
	if hashHex(data) != sha256Hex {
		_ = os.Remove(path)
		return nil, false
	}
	return data, true
}

// writeCache stores a verified file by its hash. A write that fails costs a
// later refetch, not the request.
func (s *officialSource) writeCache(sha256Hex string, data []byte) {
	dir := s.cacheDir()
	if err := os.MkdirAll(dir, cacheDirPerm); err != nil {
		return
	}
	tmp, err := os.CreateTemp(dir, sha256Hex+".*.tmp")
	if err != nil {
		return
	}
	_, writeErr := tmp.Write(data)
	closeErr := tmp.Close()
	if writeErr != nil || closeErr != nil {
		_ = os.Remove(tmp.Name())
		return
	}
	if err = os.Rename(tmp.Name(), filepath.Join(dir, sha256Hex)); err != nil {
		_ = os.Remove(tmp.Name())
	}
}

func hashHex(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}
