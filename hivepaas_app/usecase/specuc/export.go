package specuc

import (
	"context"
	"fmt"
	"io"
	"os"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/permission"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/specuc/specdto"
)

const (
	bundleContentType = "application/gzip"
	workDirPattern    = "hivepaas-spec-*"
)

// ExportSpec builds a configuration bundle for a scope.
//
// Both secret-bearing modes decrypt every secret the scope owns, so both pass
// the same capability gate and leave the same audit record as any other reveal.
// The difference between them is only whether the plaintext ends up on disk or
// inside an age envelope - not whether it was read. omit reads nothing and is
// therefore ungated.
func (uc *UC) ExportSpec(
	ctx context.Context,
	auth *basedto.Auth,
	req *specdto.ExportSpecReq,
) (_ *specdto.ExportSpecResp, err error) {
	// Refuse a malformed request before the reveal gate, not after it. The gate
	// writes an audit entry, and running it first recorded "secret-reveal:
	// allowed" for an encrypted export that then failed for want of a
	// passphrase - an audit trail saying secrets were released when nothing was.
	// specservice repeats these checks as a backstop for other callers.
	if !req.SecretsMode.IsValid() {
		return nil, hperrors.Wrap(hperrors.ErrSpecSecretsModeInvalid).
			WithParam("Mode", string(req.SecretsMode))
	}
	if req.SecretsMode == specmodel.SecretsModeEncrypted && req.Passphrase == "" {
		return nil, hperrors.Wrap(hperrors.ErrSpecPassphraseRequired)
	}

	if req.SecretsMode.RevealsSecrets() {
		err = uc.permissionManager.AuthorizeSecretReveal(ctx, uc.db, auth, &permission.RevealSubject{
			Scope:    req.Scope.ScopeType,
			ObjectID: req.Scope.ScopeObjectID(),
			Source:   base.AuditLogSourceAPIAction,
			ResType:  base.ResourceTypeSetting,
			ResName:  fmt.Sprintf("configuration spec export (%s)", req.SecretsMode),
		})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	workDir, err := os.MkdirTemp("", workDirPattern)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer func() {
		if err != nil {
			_ = os.RemoveAll(workDir)
		}
	}()

	resp, err := uc.specService.Export(ctx, uc.db, &specservice.ExportReq{
		Scope:       req.Scope,
		SecretsMode: req.SecretsMode,
		Passphrase:  req.Passphrase,
		WorkDir:     workDir,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	if err = uc.recordSpecExport(ctx, uc.db, auth, req.Scope, req.SecretsMode, resp); err != nil {
		return nil, hperrors.Wrap(err)
	}

	content, err := openBundle(resp.Path, workDir)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &specdto.ExportSpecResp{
		Data: &settings.BaseDownloadDataResp{
			ContentType:   bundleContentType,
			ContentLength: resp.Size,
			Content:       content,
			ExtraHeaders: map[string]string{
				"Content-Disposition": fmt.Sprintf("attachment; filename=%q", resp.Filename),
			},
		},
		Summary: resp.Summary,
	}, nil
}

// bundleFile hands the archive to the transport and removes the staging
// directory once it has been read.
//
// The exporter stages a tree and archives it, so the directory outlives the
// call that produced it. Closing the body is the only moment at which the
// response is known to be finished.
type bundleFile struct {
	*os.File
	workDir string
}

func (f *bundleFile) Close() error {
	err := f.File.Close()
	if removeErr := os.RemoveAll(f.workDir); removeErr != nil && err == nil {
		err = removeErr
	}
	return err
}

func openBundle(path, workDir string) (io.ReadCloser, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &bundleFile{File: file, workDir: workDir}, nil
}
