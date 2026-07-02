package gitlab

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

var (
	ErrAccessProviderInvalid = apperrors.NewErr(apperrors.ErrArgumentInvalid, "access provider invalid")
)
