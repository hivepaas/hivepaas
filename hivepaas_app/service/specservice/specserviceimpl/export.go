package specserviceimpl

import (
	"context"
	"fmt"
	"os"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const (
	globalFilename    = "global.yaml"
	projectFilename   = "project.yaml"
	filenameTimestamp = "20060102T150405Z"
)

// Export builds a configuration bundle for a scope.
func (s *service) Export(
	ctx context.Context,
	db database.IDB,
	req *specservice.ExportReq,
) (*specservice.ExportResp, error) {
	if !req.SecretsMode.IsValid() {
		return nil, hperrors.Wrap(hperrors.ErrSpecSecretsModeInvalid).
			WithParam("Mode", string(req.SecretsMode))
	}
	if req.SecretsMode == specmodel.SecretsModeEncrypted && req.Passphrase == "" {
		return nil, hperrors.Wrap(hperrors.ErrSpecPassphraseRequired)
	}

	bundle, err := s.buildBundle(ctx, db, req)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	path, err := writeBundle(req.WorkDir, bundle)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	filename := fmt.Sprintf("hivepaas-spec-%s.tar.gz",
		bundle.Manifest.ExportedAt.Format(filenameTimestamp))

	if req.SecretsMode == specmodel.SecretsModeEncrypted {
		sealed := path + ".age"
		if err = encryptBundle(path, sealed, req.Passphrase); err != nil {
			return nil, hperrors.Wrap(err)
		}
		// The unencrypted archive must not outlive the encrypted one.
		if err = os.Remove(path); err != nil {
			return nil, hperrors.Wrap(err)
		}
		path, filename = sealed, filename+".age"
	}

	info, err := os.Stat(path)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &specservice.ExportResp{
		Path:     path,
		Filename: filename,
		Size:     info.Size(),
		Report:   bundle.Report,
		Summary:  bundle.Report.Summarize(len(bundle.Files), reportFilename),
	}, nil
}
