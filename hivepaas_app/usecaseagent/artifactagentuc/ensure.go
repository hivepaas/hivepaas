package artifactagentuc

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/artifactagentuc/artifactagentdto"
)

// maxSize is templatemodel.MaxArtifactSize. The agent checks it again: what a
// request says is not what a template said until it is.
const maxSize = 256 << 20

var sha256Pattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

// Ensure has the file on this node and says where the host has it. A file
// already there is checked against its hash once per agent start and kept; one
// that is not there, or not right, is downloaded - from the mirror first when
// there is one, then from where it is published - into a temp file beside its
// place, hashed while it streams, and renamed into place only when it is whole
// and right. Nothing of a download that failed is left behind.
func (uc *UC) Ensure(ctx context.Context, req *artifactagentdto.EnsureReq) (*artifactagentdto.EnsureResp, error) {
	if err := validateEnsureReq(req); err != nil {
		return nil, err
	}
	// The host's path is what the app's bind mount names: without
	// HP_STORAGE_HOST_DIR it would be relative, which the backend cannot tell for
	// an artifact's and the task cannot start with.
	if !filepath.IsAbs(uc.hostCacheDir) {
		return nil, hperrors.Wrap(hperrors.ErrArtifactHostDirUnset)
	}
	lock := uc.lockFor(req.SHA256)
	lock.Lock()
	defer lock.Unlock()

	resp := &artifactagentdto.EnsureResp{HostPath: filepath.Join(uc.hostCacheDir, req.SHA256)}
	target := filepath.Join(uc.cacheDir, req.SHA256)
	present, err := uc.isPresent(target, req)
	if err != nil {
		return nil, err
	}
	if present {
		return resp, nil
	}

	uc.logger.Infof("Artifact %s: downloading %s", req.SHA256, req.URL)
	// The source's error is the one returned: a mirror that failed is only logged,
	// since the source was tried after it.
	for _, source := range []string{req.MirrorURL, req.URL} {
		if source == "" {
			continue
		}
		if err = uc.download(ctx, source, target, req); err == nil {
			uc.verified.Store(req.SHA256, true)
			resp.Downloaded = true
			return resp, nil
		}
		uc.logger.Warnf("Artifact %s: %s: %v", req.SHA256, source, err)
	}
	return nil, err
}

func validateEnsureReq(req *artifactagentdto.EnsureReq) error {
	reason := ""
	switch {
	case req == nil:
		reason = "nothing was asked for"
	case !isHTTPS(req.URL):
		reason = "its URL is not https"
	case req.MirrorURL != "" && !isHTTPS(req.MirrorURL):
		reason = "its mirror URL is not https"
	case !sha256Pattern.MatchString(req.SHA256):
		reason = "its sha256 is not 64 lowercase hex characters"
	case req.Size <= 0 || req.Size > maxSize:
		reason = fmt.Sprintf("its size is not from 1 byte to %d MiB", maxSize>>20) //nolint:mnd // MiB
	}
	if reason == "" {
		return nil
	}
	return hperrors.Wrap(hperrors.ErrArtifactInvalid).WithParam("Reason", reason)
}

func isHTTPS(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Scheme == "https" && parsed.Host != ""
}

// isPresent reports whether target is the file asked for. One of the right size
// is hashed the first time this agent is asked for it; one that is wrong is
// removed, to be downloaded again.
func (uc *UC) isPresent(target string, req *artifactagentdto.EnsureReq) (bool, error) {
	info, err := os.Stat(target)
	if os.IsNotExist(err) {
		return false, nil
	}
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	if _, checked := uc.verified.Load(req.SHA256); checked && info.Size() == req.Size {
		return true, nil
	}
	if info.Size() == req.Size {
		sum, hashErr := hashFile(target)
		if hashErr != nil {
			return false, hashErr
		}
		if sum == req.SHA256 {
			uc.verified.Store(req.SHA256, true)
			return true, nil
		}
	}
	uc.logger.Warnf("Artifact %s: the file kept is not what its hash says, downloading it again", req.SHA256)
	uc.verified.Delete(req.SHA256)
	if err = os.Remove(target); err != nil && !os.IsNotExist(err) {
		return false, hperrors.Wrap(err)
	}
	return false, nil
}

func hashFile(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", hperrors.Wrap(err)
	}
	defer func() { _ = file.Close() }()
	hash := sha256.New()
	if _, err = io.Copy(hash, file); err != nil {
		return "", hperrors.Wrap(err)
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}

// download writes source to target when it is the file asked for, and leaves
// nothing behind otherwise.
func (uc *UC) download(ctx context.Context, source, target string, req *artifactagentdto.EnsureReq) (err error) {
	name := filepath.Base(source)
	failed := func(format string, args ...any) error {
		return hperrors.Wrap(hperrors.ErrArtifactDownloadFailed).WithParam("Name", name).
			WithExtraDetail(format, args...)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodGet, source, nil)
	if err != nil {
		return failed("%s", err.Error())
	}
	httpResp, err := uc.client.Do(httpReq)
	if err != nil {
		return failed("%s", err.Error())
	}
	defer func() { _ = httpResp.Body.Close() }()
	if httpResp.StatusCode != http.StatusOK {
		return failed("the source answered %s", httpResp.Status)
	}
	if httpResp.ContentLength > req.Size {
		return failed("the source sends %d bytes, and the file is %d", httpResp.ContentLength, req.Size)
	}

	if err = os.MkdirAll(uc.cacheDir, 0o755); err != nil { //nolint:gosec,mnd // read by every container
		return hperrors.Wrap(err)
	}
	temp, err := os.CreateTemp(uc.cacheDir, ".download-"+req.SHA256+"-*")
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer func() {
		_ = temp.Close()
		if err != nil {
			_ = os.Remove(temp.Name())
		}
	}()

	hash := sha256.New()
	// One byte more than the size, to see a source that sends more than it should.
	written, err := io.Copy(io.MultiWriter(temp, hash), io.LimitReader(httpResp.Body, req.Size+1))
	switch {
	case err != nil:
		return failed("%s", err.Error())
	case written != req.Size:
		return failed("the source sent %d bytes, and the file is %d", written, req.Size)
	}
	if sum := hex.EncodeToString(hash.Sum(nil)); sum != req.SHA256 {
		return hperrors.Wrap(hperrors.ErrArtifactHashMismatch).WithParam("Name", name).
			WithExtraDetail("sha256 %s, expected %s", sum, req.SHA256)
	}
	if err = temp.Sync(); err != nil {
		return hperrors.Wrap(err)
	}
	if err = temp.Chmod(0o755); err != nil { //nolint:mnd // a binary every container may run
		return hperrors.Wrap(err)
	}
	if err = temp.Close(); err != nil {
		return hperrors.Wrap(err)
	}
	if err = os.Rename(temp.Name(), target); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
