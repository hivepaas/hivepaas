package composeserviceimpl

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
)

// traefikCompose writes ^ for a backquote, which a Go raw string cannot hold.
var traefikCompose = strings.ReplaceAll(`
services:
  web:
    image: ghcr.io/me/web:1
    ports: ["8080:3000"]
    labels:
      traefik.enable: "true"
      traefik.http.routers.web.rule: "Host(^app.example.org^) || Host(^www.example.org^)"
      traefik.http.services.web.loadbalancer.server.port: "3000"
  api:
    image: ghcr.io/me/api:1
    expose: ["9000"]
    deploy:
      labels:
        - "traefik.http.routers.api.rule=Host(^api.example.org^) && PathPrefix(^/v1^)"
  off:
    image: ghcr.io/me/off:1
    labels: {traefik.enable: "false", traefik.http.routers.off.rule: "Host(^off.example.org^)"}
  unknown:
    image: ghcr.io/me/unknown:1
    labels: {traefik.http.routers.u.rule: "Host(^u.example.org^)"}
`, "^", "`")

func domainsOf(t *testing.T, resp *composeservice.ConvertResp, key string) map[string]any {
	t.Helper()
	routing, _ := appOf(t, resp, key).Settings["routing"].(map[string]any)
	out := map[string]any{}
	domains, _ := routing["domains"].([]any)
	for _, d := range domains {
		fields := d.(map[string]any)
		out[fields["domain"].(string)] = fields["containerPort"]
	}
	return out
}

// A service behind Traefik need publish no port: its labels' hosts are its
// domains, on the port they route to; its published port of the same one is
// then no domain by default.
func TestConvertRoutesByTraefikLabels(t *testing.T) {
	resp := convert(t, convertReq(traefikCompose))

	assert.Equal(t, map[string]any{"app.example.org": 3000, "www.example.org": 3000}, domainsOf(t, resp, "web"))
	assert.Equal(t, map[string]any{"api.example.org": 9000}, domainsOf(t, resp, "api"))
	assert.Empty(t, domainsOf(t, resp, "off"), "traefik.enable false")
	assert.Empty(t, domainsOf(t, resp, "unknown"), "no port to tell")
	assert.Contains(t, codes(resp.Issues[envPath+"/apps/unknown"]), composeservice.CodeDomainMissing)
	for _, issue := range resp.Issues[envPath+"/apps/api"] {
		if issue.Code == composeservice.CodeTraefikRoute {
			assert.Equal(t, true, issue.Detail["paths"], "a path the rule matches is not kept")
		}
	}

	for _, view := range resp.Services {
		if view.Name != "web" {
			continue
		}
		if assert.Len(t, view.Ports, 2) {
			assert.Equal(t, composeservice.PortAsNone, view.Ports[0].As, "published: theirs is the domain")
			assert.Equal(t, composeservice.PortSourceLabels, view.Ports[1].Source)
			assert.Equal(t, "app.example.org", view.Ports[1].Domain)
			assert.Equal(t, []string{"www.example.org"}, view.Ports[1].Also)
		}
	}

	req := convertReq(traefikCompose)
	req.Services = map[string]*composeservice.ServiceReq{"web": {Ports: []*composeservice.PortReq{
		{Target: 3000, Protocol: "tcp", Source: composeservice.PortSourceLabels, As: composeservice.PortAsDomain,
			Domain: "web.mine.net"},
	}}}
	assert.Equal(t, map[string]any{"web.mine.net": 3000, "www.example.org": 3000}, domainsOf(t, convert(t, req), "web"),
		"the review's domain, and the labels' other hosts")
}

// Only a host a domain can be is one: not a port's, an empty one, a variable
// left unread.
func TestRuleHostsAreDomains(t *testing.T) {
	assert.Equal(t, []string{"a.example.org", "b.example.org"},
		ruleHosts("Host(`A.example.org`, `b.example.org`) || Host(`a.example.org`)"))
	assert.Empty(t, ruleHosts("Host(`localhost`) || Host(`a.org:8080`) || Host(``) || HostRegexp(`{x:.+}.org`)"))
	assert.Empty(t, ruleHosts("Host(`${DOMAIN}`)"))
}
