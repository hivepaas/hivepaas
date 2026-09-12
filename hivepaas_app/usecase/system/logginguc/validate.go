package logginguc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/services/logging"
)

// validateSettings refuses a configuration that would deploy something broken.
//
// Nothing is checked while logging is disabled: that is the default state, and
// requiring a complete form to turn the feature off would be perverse.
func validateSettings(cfg *entity.Logging) error {
	if cfg == nil || !cfg.Enabled {
		return nil
	}

	// Only container logs are collected, so only these two say anything.
	if !cfg.Sources.Apps && !cfg.Sources.HivePaaS {
		return hperrors.Wrap(logging.ErrNoSources)
	}

	if cfg.Backend.Managed {
		vl := cfg.Backend.VictoriaLogs
		if vl == nil {
			return hperrors.Wrap(loggingservice.ErrNotConfigured)
		}
		if vl.NodeID == "" {
			return hperrors.Wrap(loggingservice.ErrBackendNodeMissing)
		}
		if vl.VolumeID == "" {
			return hperrors.Wrap(loggingservice.ErrVolumeMissing)
		}
	} else if cfg.Backend.Ingest == nil || cfg.Backend.Ingest.URL == "" {
		return hperrors.Wrap(logging.ErrIngestEndpointRequired)
	}

	names := map[string]bool{}
	for i := range cfg.Forwards {
		f := &cfg.Forwards[i]
		if f.Name == "" {
			return hperrors.NewArgumentInvalid("Forwards").
				WithExtraDetail("a forward needs a name to be reported on or removed")
		}
		if names[f.Name] {
			return hperrors.NewArgumentInvalid("Forwards").
				WithExtraDetail("forward name '%s' is used twice", f.Name)
		}
		names[f.Name] = true
		if f.Endpoint.URL == "" {
			return hperrors.Wrap(logging.ErrIngestEndpointRequired)
		}
	}
	return nil
}
