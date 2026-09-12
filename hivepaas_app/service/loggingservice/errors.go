package loggingservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// NOTE: declared here rather than in hperrors/constants.go. The prefix differs
// from loggingmodel's ERR_LOGGING_ deliberately: both files feed one
// translation namespace, and a clash between them is a string collision the
// compiler cannot see. Translations are in errors.logging.en.toml.
var (
	ErrNotConfigured      = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_LOGGING_SVC_NOT_CONFIGURED")
	ErrBackendNodeMissing = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_LOGGING_SVC_BACKEND_NODE_MISSING")
	ErrVolumeMissing      = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_LOGGING_SVC_VOLUME_MISSING")
	ErrBackendNotReady    = hperrors.NewErr(hperrors.ErrActionFailed, "ERR_LOGGING_SVC_BACKEND_NOT_READY")
	ErrDeployFailed       = hperrors.NewErr(hperrors.ErrActionFailed, "ERR_LOGGING_SVC_DEPLOY_FAILED")
)
