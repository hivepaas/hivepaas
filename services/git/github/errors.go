package github

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

var (
	ErrGithubAppClientRequired   = apperrors.NewErr(apperrors.ErrArgumentInvalid, "github app client required")
	ErrGithubTokenClientRequired = apperrors.NewErr(apperrors.ErrArgumentInvalid, "github token client required")
	ErrAccessProviderInvalid     = apperrors.NewErr(apperrors.ErrArgumentInvalid, "access provider invalid")
	ErrAPICallFailed             = apperrors.NewErr(apperrors.ErrActionFailed, "api call failed")
)
