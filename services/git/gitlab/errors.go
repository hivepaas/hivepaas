package gitlab

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

var (
	ErrAccessProviderInvalid = hperrors.NewErr(hperrors.ErrArgumentInvalid, "access provider invalid")
)
