package permission

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

// Visibility answers, for one user, whether the screen behind something found
// across the whole system would open for them - for a page that gathers from
// everywhere, such as the home page, and must show each user only their part.
//
// It asks exactly what that screen asks - CheckAccess with the same check -
// rather than working the answer out from the user's grants. The rules that
// decide a check (an env's grant over its project's, the project's over the
// module's, an owner over all of them) then live in one place, and such a page
// can never show more than the screens it links to would.
//
// An answer is kept for the life of the Visibility, which is meant to be one
// request: grants that change meanwhile are seen by the next one.
type Visibility interface {
	// AllowsModule is whether the user may act on a module's screens: cluster,
	// system, settings.
	AllowsModule(ctx context.Context, module base.ResourceModule, action base.ActionType) (bool, error)
	// AllowsProjectEnv is whether the user may act on an env's screens, which
	// are also its apps': an app's screens ask about the env it is in.
	AllowsProjectEnv(ctx context.Context, projectID, env string, action base.ActionType) (bool, error)
}
