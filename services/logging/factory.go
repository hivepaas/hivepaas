package logging

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging/loggingmodel"
	"github.com/hivepaas/hivepaas/services/logging/victorialogs"
	"github.com/hivepaas/hivepaas/services/logging/vlagent"
)

// BackendConfig carries the configuration for whichever backend is named.
type BackendConfig struct {
	VictoriaLogs *victorialogs.Config
}

// CollectorConfig carries the configuration for whichever collector is named.
type CollectorConfig struct {
	Vlagent *vlagent.Config
}

// NewBackend builds the backend for the given type.
func NewBackend(t BackendType, cfg *BackendConfig) (Backend, error) {
	if cfg == nil {
		return nil, hperrors.Wrap(loggingmodel.ErrBackendUnsupported).WithParam("Name", string(t))
	}
	switch t {
	case loggingmodel.BackendTypeVictoriaLogs:
		if cfg.VictoriaLogs == nil {
			return nil, hperrors.Wrap(loggingmodel.ErrBackendUnsupported).WithParam("Name", string(t))
		}
		return victorialogs.New(cfg.VictoriaLogs), nil
	default:
		return nil, hperrors.Wrap(loggingmodel.ErrBackendUnsupported).WithParam("Name", string(t))
	}
}

// NewDeployer builds a backend that HivePaaS also runs.
//
// Separate from NewBackend because only a managed backend can be deployed, and
// asking a user's backend to describe itself is a programming error rather than
// a configuration one.
func NewDeployer(t BackendType, cfg *BackendConfig) (Deployer, error) {
	b, err := NewBackend(t, cfg)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	d, ok := b.(Deployer)
	if !ok {
		return nil, hperrors.Wrap(loggingmodel.ErrBackendUnsupported).WithParam("Name", string(t))
	}
	return d, nil
}

// NewCollector builds the collector for the given type.
func NewCollector(t CollectorType, cfg *CollectorConfig) (Collector, error) {
	if cfg == nil {
		return nil, hperrors.Wrap(loggingmodel.ErrCollectorUnsupported).WithParam("Name", string(t))
	}
	switch t {
	case loggingmodel.CollectorTypeVlagent:
		if cfg.Vlagent == nil {
			return nil, hperrors.Wrap(loggingmodel.ErrCollectorUnsupported).WithParam("Name", string(t))
		}
		return vlagent.New(cfg.Vlagent), nil
	default:
		return nil, hperrors.Wrap(loggingmodel.ErrCollectorUnsupported).WithParam("Name", string(t))
	}
}
