package registryserviceimpl

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/registryservice"
)

// ProbeDomain asks the registry's own address what answers there.
//
// It is advisory and says so: HivePaaS cannot see somebody's DNS settings, and a
// load balancer in front of the cluster is not a proxy in the sense that matters
// here. What it can do is name what it saw, early, instead of leaving an operator
// to discover it from a build that dies on a 100 MB layer.
func (s *service) ProbeDomain(ctx context.Context, domain string) (*registryservice.DomainProbe, error) {
	if strings.TrimSpace(domain) == "" {
		return nil, hperrors.Wrap(hperrors.ErrRegistrySettingsInvalid).
			WithExtraDetail("A domain is required.")
	}

	url := fmt.Sprintf("https://%s/v2/", domain)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		// Not reachable is not the same as not proxied, and saying so is the
		// honest answer while the registry is still starting.
		return &registryservice.DomainProbe{Reached: false}, nil
	}
	defer func() { _ = resp.Body.Close() }()

	evidence := proxyEvidence(resp.Header)
	return &registryservice.DomainProbe{
		Reached:  true,
		Proxied:  len(evidence) > 0,
		Evidence: evidence,
	}, nil
}

// proxyEvidence names the headers that say something answered instead of the
// registry. It returns sentences rather than header names, because they are shown
// to a person deciding what to change in their DNS.
func proxyEvidence(header http.Header) []string {
	var evidence []string
	if ray := header.Get("cf-ray"); ray != "" {
		evidence = append(evidence, fmt.Sprintf("a cf-ray header (%s), which only Cloudflare sends", ray))
	}
	if server := header.Get("server"); strings.EqualFold(server, "cloudflare") {
		evidence = append(evidence, "a server header saying cloudflare")
	}
	if via := header.Get("via"); via != "" {
		evidence = append(evidence, fmt.Sprintf("a via header (%s), so something is relaying this", via))
	}
	return evidence
}
