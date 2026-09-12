package loggingserviceimpl

import (
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging"
)

const (
	// ServiceNameBackend and ServiceNameCollector are the swarm service names.
	// They are also how the collector recognises the stack's own containers in
	// order to skip them.
	ServiceNameBackend   = "hivepaas-victoria-logs"
	ServiceNameCollector = "hivepaas-vlagent"

	// dockerContainersGlob matches every container's json-file log on a node.
	dockerContainersGlob = "/var/lib/docker/containers/*/*-json.log"

	// traefikAccessGlob is where the proxy writes its access log.
	traefikAccessGlob = "/var/log/traefik/access.log"

	// nodeLogGlob is the host's own logs.
	nodeLogGlob = "/var/log/syslog"
)

// toEndpoint converts a stored endpoint into one services/logging can use.
//
// This is where credentials are decrypted, so that nothing under
// services/logging ever sees an EncryptedField - the same boundary
// buildS3Storage draws for backups.
func toEndpoint(ep *entity.LoggingEndpoint) (logging.Endpoint, error) {
	if ep == nil {
		return logging.Endpoint{}, nil
	}

	password, err := ep.Password.GetPlain()
	if err != nil {
		return logging.Endpoint{}, hperrors.Wrap(err)
	}
	token, err := ep.BearerToken.GetPlain()
	if err != nil {
		return logging.Endpoint{}, hperrors.Wrap(err)
	}

	return logging.Endpoint{
		URL:           ep.URL,
		Username:      ep.Username,
		Password:      password,
		BearerToken:   token,
		Headers:       ep.Headers,
		TLSSkipVerify: ep.TLSSkipVerify,
	}, nil
}

// buildCollectSpec turns the stored configuration into a collection job.
func (s *service) buildCollectSpec(cfg *entity.Logging, ingestURL string) (*logging.CollectSpec, error) {
	// The stack's own containers are skipped. Both write logs, and collecting
	// them feeds the collector its own output - with a backend that logs each
	// ingest, that is a loop rather than merely noise.
	stackExclude := fmt.Sprintf("/var/lib/docker/containers/*%s*/*-json.log", ServiceNameCollector)

	var sources []logging.Source
	if cfg.Sources.Apps {
		sources = append(sources, logging.Source{
			Kind:    logging.SourceKindApp,
			Glob:    dockerContainersGlob,
			Exclude: stackExclude,
		})
	}
	if cfg.Sources.HivePaaS {
		sources = append(sources, logging.Source{
			Kind:    logging.SourceKindHivePaaS,
			Glob:    dockerContainersGlob,
			Exclude: stackExclude,
			Labels:  map[string]string{"hivepaas_source": string(logging.SourceKindHivePaaS)},
		})
	}
	if cfg.Sources.TraefikAccess {
		sources = append(sources, logging.Source{
			Kind:   logging.SourceKindTraefikAccess,
			Glob:   traefikAccessGlob,
			Labels: map[string]string{"hivepaas_source": string(logging.SourceKindTraefikAccess)},
		})
	}
	if cfg.Sources.Nodes {
		sources = append(sources, logging.Source{
			Kind:   logging.SourceKindNode,
			Glob:   nodeLogGlob,
			Labels: map[string]string{"hivepaas_source": string(logging.SourceKindNode)},
		})
	}
	if len(sources) == 0 {
		return nil, hperrors.Wrap(logging.ErrNoSources)
	}

	forwards := make([]logging.ForwardTarget, 0, len(cfg.Forwards))
	for i := range cfg.Forwards {
		f := &cfg.Forwards[i]
		ep, err := toEndpoint(&f.Endpoint)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		forwards = append(forwards, logging.ForwardTarget{
			Name:     f.Name,
			Format:   f.Format,
			Endpoint: ep,
		})
	}

	return &logging.CollectSpec{
		Ingest:   logging.Endpoint{URL: ingestURL},
		Forwards: forwards,
		Sources:  sources,
	}, nil
}
