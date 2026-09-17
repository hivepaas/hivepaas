package apptemplateserviceimpl

import (
	"context"
	"os"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterepo"
)

const (
	localSourceID = "local"
	localRevision = "local"
)

// localDirSource reads a checkout of the templates repository directly: no
// signature, no hashes, re-read on every call so an edit shows up at once. It
// exists for authoring templates, and the service only uses it in development.
type localDirSource struct {
	dir string
}

func newLocalDirSource(dir string) *localDirSource {
	return &localDirSource{dir: dir}
}

func (s *localDirSource) ID() string {
	return localSourceID
}

func (s *localDirSource) Revision(context.Context) (string, error) {
	return localRevision, nil
}

// load reads the checkout. A template that does not decode is left out, as it
// would be left out of an index; tools/apptemplate lint says why.
func (s *localDirSource) load() (*templaterepo.Repo, *templatemodel.Index, error) {
	repo, _, err := templaterepo.Load(os.DirFS(s.dir))
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	index, err := templaterepo.BuildIndex(repo)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	return repo, index, nil
}

func (s *localDirSource) Index(context.Context) (*templatemodel.Index, error) {
	_, index, err := s.load()
	return index, err
}

func (s *localDirSource) TemplateFile(_ context.Context, entry *templatemodel.IndexEntry) ([]byte, error) {
	repo, _, err := s.load()
	if err != nil {
		return nil, err
	}
	file := repo.FindTemplate(entry.Name)
	if file == nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateNotFound).WithParam("Name", entry.Name)
	}
	return file.Content, nil
}

func (s *localDirSource) Icon(_ context.Context, sha256Hex string) ([]byte, error) {
	repo, index, err := s.load()
	if err != nil {
		return nil, err
	}
	entry := index.FindIcon(sha256Hex)
	if entry == nil {
		return nil, hperrors.NewNotFound("Icon")
	}
	return repo.Icons[entry.Icon.Path].Content, nil
}
