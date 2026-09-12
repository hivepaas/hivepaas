// Package logging is the entry point to the logging implementations.
//
// Callers use the aliases here rather than reaching into loggingmodel, the same
// arrangement services/backup uses.
package logging

import (
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
)

// Re-exported constants
const (
	BackendTypeVictoriaLogs = loggingmodel.BackendTypeVictoriaLogs
	CollectorTypeVlagent    = loggingmodel.CollectorTypeVlagent

	SourceKindApp           = loggingmodel.SourceKindApp
	SourceKindHivePaaS      = loggingmodel.SourceKindHivePaaS
	SourceKindTraefikAccess = loggingmodel.SourceKindTraefikAccess
	SourceKindNode          = loggingmodel.SourceKindNode
)

// Re-exported errors
var (
	ErrBackendUnsupported     = loggingmodel.ErrBackendUnsupported
	ErrCollectorUnsupported   = loggingmodel.ErrCollectorUnsupported
	ErrIngestEndpointRequired = loggingmodel.ErrIngestEndpointRequired
	ErrNoSources              = loggingmodel.ErrNoSources
	ErrForwardFormatInvalid   = loggingmodel.ErrForwardFormatInvalid
	ErrQueryInvalid           = loggingmodel.ErrQueryInvalid
	ErrBackendUnreachable     = loggingmodel.ErrBackendUnreachable
)

// Re-exported types
type (
	BackendType   = loggingmodel.BackendType
	CollectorType = loggingmodel.CollectorType
	SourceKind    = loggingmodel.SourceKind

	Backend   = loggingmodel.Backend
	Collector = loggingmodel.Collector
	Deployer  = loggingmodel.Deployer

	Endpoint      = loggingmodel.Endpoint
	Source        = loggingmodel.Source
	ForwardTarget = loggingmodel.ForwardTarget
	CollectSpec   = loggingmodel.CollectSpec
	RuntimeSpec   = loggingmodel.RuntimeSpec
	Mount         = loggingmodel.Mount
	Port          = loggingmodel.Port
	Resources     = loggingmodel.Resources
	QueryReq      = loggingmodel.QueryReq
	QueryResp     = loggingmodel.QueryResp
	LogEntry      = loggingmodel.LogEntry
)
