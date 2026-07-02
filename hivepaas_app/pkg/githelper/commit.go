package githelper

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/githelper/validation"
)

func IsCommitHash(hash string) bool {
	return validation.IsCommitHash(hash)
}
