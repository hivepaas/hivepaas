package composeserviceimpl

import (
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// httpPorts are the ports a web server usually listens on in its container:
// published, they are offered as a domain.
var httpPorts = []uint32{80, 3000, 5000, 8000, 8080, 8888}

// privatePorts are a database's or a broker's: a compose file publishes them
// for a laptop, and a server should not.
var privatePorts = []uint32{5432, 3306, 1433, 27017, 6379, 11211, 9200, 5672, 9092}

// domainLabelMax is the longest a label of a domain name may be.
const domainLabelMax = 63

// ports are a service's published ports, and the ports its Traefik labels
// route to, as the review chose them - a domain, a port on the nodes, or none
// - into its routing and its endpoint.
func (c *converter) ports(
	appPath, name string, svc types.ServiceConfig, view *composeservice.ServiceView,
	networks *specmodel.Networks, settings map[string]any,
) {
	routes := c.traefikRoutes(appPath, svc)
	views := make([]*composeservice.PortView, 0, len(svc.Ports)+len(routes))
	modes := make([]string, 0, cap(views))
	for _, p := range svc.Ports {
		_, routed := routes[p.Target]
		views = append(views, c.portView(name, p, routed))
		modes = append(modes, p.Mode)
	}
	for _, target := range slices.Sorted(maps.Keys(routes)) {
		route := routes[target]
		views = append(views, c.labelsPortView(name, target, route))
		modes = append(modes, "")
		action := "routed by the app's domains, from its Traefik labels"
		if route.paths {
			action += ": the whole host - a path a rule matches is not kept"
		}
		c.add(appPath, "", composeservice.CodeTraefikRoute,
			map[string]any{detailTarget: target, "hosts": route.hosts, "paths": route.paths}, action)
	}

	var domains []any
	routingPort := 0
	for i, pv := range views {
		switch pv.As {
		case composeservice.PortAsDomain:
			if pv.Domain == "" || pv.Protocol != string(network.TCP) {
				c.add(appPath, specmodel.SeverityFixable, composeservice.CodeDomainMissing,
					map[string]any{detailTarget: pv.Target}, "not reachable from outside: it has no domain")
				pv.As = composeservice.PortAsNone
				break
			}
			hosts := []string{pv.Domain}
			if pv.Source == composeservice.PortSourceLabels {
				hosts = append(hosts, pv.Also...)
			}
			for _, host := range hosts {
				domains = appendDomain(domains, host, pv.Target)
			}
			if routingPort == 0 {
				routingPort = int(pv.Target)
			}
		case composeservice.PortAsNode:
			if pv.Published == 0 {
				pv.As = composeservice.PortAsNone
				break
			}
			if networks.EndpointSpec == nil {
				networks.EndpointSpec = &specmodel.EndpointSpec{}
			}
			mode := swarm.PortConfigPublishModeIngress
			if modes[i] == string(swarm.PortConfigPublishModeHost) {
				mode = swarm.PortConfigPublishModeHost
			}
			networks.EndpointSpec.Ports = append(networks.EndpointSpec.Ports, &specmodel.PortConfig{
				Target: pv.Target, Published: pv.Published, Protocol: network.IPProtocol(pv.Protocol), PublishMode: mode,
			})
		case composeservice.PortAsNone:
		}
		view.Ports = append(view.Ports, pv)
	}
	if len(domains) > 0 {
		settings["routing"] = map[string]any{"port": routingPort, "exposePublicly": true, "domains": domains}
	}
}

// appendDomain adds a domain of the app's on a port of its container, once.
func appendDomain(domains []any, host string, target uint32) []any {
	for _, existing := range domains {
		if fields, _ := existing.(map[string]any); fields["domain"] == host {
			return domains
		}
	}
	return append(domains, map[string]any{"enabled": true, "domain": host,
		"protocol": string(base.NetworkProtocolHTTP), "containerPort": int(target)})
}

// labelsPortView is the container port a service's Traefik labels route to:
// a domain on their first host, and the others too while it is one, unless
// the review says otherwise. It is not published: there is nothing to put on
// the nodes.
func (c *converter) labelsPortView(name string, target uint32, route *traefikRoute) *composeservice.PortView {
	pv := &composeservice.PortView{Target: target, Protocol: string(network.TCP),
		Source: composeservice.PortSourceLabels, Default: composeservice.PortAsDomain,
		Suggested: route.hosts[0], Also: route.hosts[1:]}
	pv.As, pv.Domain = pv.Default, pv.Suggested
	if choice := c.portChoice(name, pv); choice != nil && slices.Contains(composeservice.AllPortAs, choice.As) {
		pv.As, pv.Domain = choice.As, strings.ToLower(strings.TrimSpace(choice.Domain))
	}
	if pv.As == composeservice.PortAsDomain && pv.Domain == "" {
		pv.Domain = pv.Suggested
	}
	return pv
}

// portView is a published port, its default, and the review's choice. One
// the Traefik labels route to as well is not a domain by default: theirs is.
func (c *converter) portView(name string, p types.ServicePortConfig, routed bool) *composeservice.PortView {
	published, _ := strconv.ParseUint(p.Published, 10, 16)
	pv := &composeservice.PortView{Published: uint32(published), Target: p.Target,
		Protocol: strings.ToLower(p.Protocol)}
	if pv.Protocol == "" {
		pv.Protocol = string(network.TCP)
	}
	if c.req.RootDomain != "" {
		pv.Suggested = c.suggestedDomain(name)
	}
	pv.Default = c.defaultPortAs(pv, p.HostIP)
	if routed && pv.Default == composeservice.PortAsDomain {
		pv.Default = composeservice.PortAsNone
	}
	pv.As = pv.Default
	if choice := c.portChoice(name, pv); choice != nil && slices.Contains(composeservice.AllPortAs, choice.As) {
		pv.As, pv.Domain = choice.As, strings.ToLower(strings.TrimSpace(choice.Domain))
	}
	if pv.As == composeservice.PortAsDomain && pv.Domain == "" {
		pv.Domain = pv.Suggested
	}
	return pv
}

// defaultPortAs is what a published port becomes unless the review says
// otherwise: a domain for a web server's, none for a database's or one bound
// to the host itself, a port on the nodes for any other.
func (c *converter) defaultPortAs(pv *composeservice.PortView, hostIP string) composeservice.PortAs {
	switch {
	case pv.Published == 0, hostIP == "127.0.0.1", hostIP == "::1", slices.Contains(privatePorts, pv.Target):
		return composeservice.PortAsNone
	case pv.Protocol != string(network.TCP):
		return composeservice.PortAsNode
	case c.req.RootDomain != "" && slices.Contains(httpPorts, pv.Target):
		return composeservice.PortAsDomain
	}
	return composeservice.PortAsNode
}

func (c *converter) portChoice(name string, pv *composeservice.PortView) *composeservice.PortReq {
	choices := c.req.Services[name]
	if choices == nil {
		return nil
	}
	for _, choice := range choices.Ports {
		if choice != nil && choice.Published == pv.Published && choice.Target == pv.Target &&
			strings.EqualFold(choice.Protocol, pv.Protocol) && choice.Source == pv.Source {
			return choice
		}
	}
	return nil
}

// suggestedDomain is the app's key and the project's under the root domain:
// one label, as a wildcard record covers one, in what a DNS name may hold.
func (c *converter) suggestedDomain(name string) string {
	label := c.keys[name] + "-" + strings.ReplaceAll(c.req.ProjectKey, "_", "-")
	if len(label) > domainLabelMax {
		label = strings.TrimRight(label[:domainLabelMax], "-")
	}
	return label + "." + c.req.RootDomain
}
