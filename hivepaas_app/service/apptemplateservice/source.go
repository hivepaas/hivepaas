// Package apptemplateservice serves app templates: the catalog, one template,
// its icon, and a template rendered for creating an app.
package apptemplateservice

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

// Source is where templates come from. Everything that reads templates - the
// store, provisioning, and phase 2's updates - reads them through a Source.
//
// TODO: app templates phase 3 - a git repository an administrator adds, and an
// uploaded file, as further sources. See
// docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
type Source interface {
	ID() string
	// Revision identifies what is being served, such as a commit.
	Revision(ctx context.Context) (string, error)
	Index(ctx context.Context) (*templatemodel.Index, error)
	// TemplateFile returns an entry's template file, verified wherever the source
	// is able to verify it.
	TemplateFile(ctx context.Context, entry *templatemodel.IndexEntry) ([]byte, error)
	// Icon returns the icon with this sha256. Only an icon the index lists is served.
	Icon(ctx context.Context, sha256 string) ([]byte, error)
}
