package loggingserviceimpl

import (
	"fmt"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/services/logging"
	"github.com/hivepaas/hivepaas/services/logging/victorialogs"
)

const (
	// ServiceNameBackend and ServiceNameCollector are the swarm service names.
	// They are also how the collector recognizes the stack's own containers in
	// order to skip them.
	ServiceNameBackend   = "hivepaas-victoria-logs"
	ServiceNameCollector = "hivepaas-vlagent"

	// dockerContainersGlob matches every container's json-file log on a node.
	//
	// It is the only glob the collector is given. The proxy's access log and
	// the host's own logs are not collected: traefik logs to stdout, which this
	// glob already covers, and nothing outside /var/lib/docker/containers is
	// mounted into the collector - a glob pointing there would match nothing
	// and report no error.
	dockerContainersGlob = "/var/lib/docker/containers/*/*-json.log"

	// NetworkLogging is the overlay the collector and the backend talk over.
	// Only the logging stack joins it.
	NetworkLogging = "hivepaas_logging_net"
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

// backendBaseURL is where a HivePaaS-run backend answers, from any service on a
// network it is attached to.
func backendBaseURL() string {
	return fmt.Sprintf("http://%s:%d", ServiceNameBackend, victorialogs.DefaultHTTPPort)
}

// buildCollectSpec turns the stored configuration into a collection job.
func (s *service) buildCollectSpec(cfg *entity.Logging, ingestURL string) (*logging.CollectSpec, error) {
	var sources []logging.Source
	if cfg.Sources.Apps || cfg.Sources.HivePaaS {
		// One source, not two. Apps and HivePaaS's own services write into the
		// same directory under container ids, so no glob can separate them -
		// and giving the same glob twice collects the file once anyway, with
		// only the first slot's extra fields applied. Which line belongs to an
		// app is decided at read time by the identity the daemon wrote into it.
		kind := logging.SourceKindApp
		if !cfg.Sources.Apps {
			kind = logging.SourceKindHivePaaS
		}
		sources = append(sources, logging.Source{
			Kind: kind,
			Glob: dockerContainersGlob,
			// No extra field: the lines are a mix of both, so labeling them
			// all as one or the other would be a lie stored on every line.
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
