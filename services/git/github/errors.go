package github

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

var (
	ErrGithubAppClientRequired   = hperrors.NewErr(hperrors.ErrArgumentInvalid, "github app client required")
	ErrGithubTokenClientRequired = hperrors.NewErr(hperrors.ErrArgumentInvalid, "github token client required")
	ErrAccessProviderInvalid     = hperrors.NewErr(hperrors.ErrArgumentInvalid, "access provider invalid")
	ErrAPICallFailed             = hperrors.NewErr(hperrors.ErrActionFailed, "api call failed")
)
