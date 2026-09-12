package loggingmodel

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// NOTE: these live here rather than in package `logging` because `logging`
// imports the implementations (victorialogs, vlagent), so those packages cannot
// import it back. `logging` re-exports them.
//
// They are also deliberately not in hperrors/constants.go. Their translations
// live in errors.logging.en.toml, which names this file.
var (
	// Configuration
	ErrBackendUnsupported     = hperrors.NewErr(hperrors.ErrUnsupported, "ERR_LOGGING_BACKEND_UNSUPPORTED")
	ErrCollectorUnsupported   = hperrors.NewErr(hperrors.ErrUnsupported, "ERR_LOGGING_COLLECTOR_UNSUPPORTED")
	ErrIngestEndpointRequired = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_LOGGING_INGEST_ENDPOINT_REQUIRED")
	ErrNoSources              = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_LOGGING_NO_SOURCES")
	ErrForwardFormatInvalid   = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_LOGGING_FORWARD_FORMAT_INVALID")

	// Talking to a backend
	ErrQueryInvalid       = hperrors.NewErr(hperrors.ErrBadRequest, "ERR_LOGGING_QUERY_INVALID")
	ErrBackendUnreachable = hperrors.NewErr(hperrors.ErrActionFailed, "ERR_LOGGING_BACKEND_UNREACHABLE")
)
