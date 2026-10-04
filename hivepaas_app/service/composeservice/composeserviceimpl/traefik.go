package composeserviceimpl

import (
	"maps"
	"regexp"
	"slices"
	"strconv"
	"strings"

	"github.com/compose-spec/compose-go/v2/types"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// The Traefik labels a service is routed by: its routers' rules and services,
// and the ports those services balance to.
const (
	traefikEnable     = "traefik.enable"
	traefikRouters    = "traefik.http.routers."
	traefikServices   = "traefik.http.services."
	traefikRule       = ".rule"
	traefikService    = ".service"
	traefikServerPort = ".loadbalancer.server.port"
)

var (
	// hostMatcher is a rule's Host(...) matcher - not HostRegexp's, nor
	// HostSNI's, which are not HTTP hosts.
	hostMatcher = regexp.MustCompile(`\bHost\(([^)]*)\)`)
	// hostArg is a quoted argument of one.
	hostArg = regexp.MustCompile("[`\"']([^`\"']+)[`\"']")
	// pathMatcher is a rule's matcher of a path, which a domain does not keep.
	pathMatcher = regexp.MustCompile(`\bPath(Prefix|Regexp)?\(`)
	// hostName is a host a domain can be: DNS labels, dotted.
	hostName = regexp.MustCompile(`^([a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?\.)+[a-z0-9]([a-z0-9-]{0,61}[a-z0-9])?$`)
)

// traefikRoute is what a service's Traefik labels route to one of its
// container's ports: the hosts, in the order the labels give them, and
// whether a rule matched a path besides.
type traefikRoute struct {
	hosts []string
	paths bool
}

// traefikRoutes are the HTTP routes a service's Traefik labels - its own and
// its deploy's - give it, by the container port they reach: a domain of the
// app's each, as HivePaaS routes. A router whose port cannot be told is said
// so, and left out.
func (c *converter) traefikRoutes(appPath string, svc types.ServiceConfig) map[uint32]*traefikRoute {
	labels := map[string]string{}
	maps.Copy(labels, svc.Labels)
	if svc.Deploy != nil {
		maps.Copy(labels, svc.Deploy.Labels)
	}
	if strings.EqualFold(strings.TrimSpace(labels[traefikEnable]), "false") {
		return nil
	}
	servicePorts := map[string]uint32{}
	for key, value := range labels {
		name, found := strings.CutPrefix(key, traefikServices)
		if !found || !strings.HasSuffix(name, traefikServerPort) {
			continue
		}
		port, err := strconv.ParseUint(strings.TrimSpace(c.plain(appPath, "labels", value)), 10, 16)
		if err == nil && port > 0 {
			servicePorts[strings.TrimSuffix(name, traefikServerPort)] = uint32(port)
		}
	}

	routes := map[uint32]*traefikRoute{}
	for _, key := range slices.Sorted(maps.Keys(labels)) {
		router, found := strings.CutPrefix(key, traefikRouters)
		if !found || !strings.HasSuffix(router, traefikRule) {
			continue
		}
		router = strings.TrimSuffix(router, traefikRule)
		rule := c.plain(appPath, "labels", labels[key])
		hosts := ruleHosts(rule)
		if len(hosts) == 0 {
			continue
		}
		port := c.routerPort(svc, servicePorts, strings.TrimSpace(labels[traefikRouters+router+traefikService]))
		if port == 0 {
			c.add(appPath, specmodel.SeverityFixable, composeservice.CodeDomainMissing,
				map[string]any{"router": router, "hosts": hosts},
				"not routed: the Traefik labels do not say which port of the container they reach")
			continue
		}
		route := routes[port]
		if route == nil {
			route = &traefikRoute{}
			routes[port] = route
		}
		for _, host := range hosts {
			if !slices.Contains(route.hosts, host) {
				route.hosts = append(route.hosts, host)
			}
		}
		route.paths = route.paths || pathMatcher.MatchString(rule)
	}
	return routes
}

// routerPort is the container port a Traefik router reaches: its service's,
// the one service's the labels give, or the one port the service has.
func (c *converter) routerPort(svc types.ServiceConfig, servicePorts map[string]uint32, service string) uint32 {
	if port, found := servicePorts[service]; found {
		return port
	}
	if len(servicePorts) == 1 {
		for _, port := range servicePorts {
			return port
		}
	}
	var candidates []uint32
	for _, p := range svc.Ports {
		if !slices.Contains(candidates, p.Target) {
			candidates = append(candidates, p.Target)
		}
	}
	for _, expose := range svc.Expose {
		port, _, _ := strings.Cut(expose, "/")
		if parsed, err := strconv.ParseUint(port, 10, 16); err == nil && !slices.Contains(candidates, uint32(parsed)) {
			candidates = append(candidates, uint32(parsed))
		}
	}
	if len(candidates) == 1 {
		return candidates[0]
	}
	return 0
}

// ruleHosts are the hosts of a rule's Host matchers that a domain can be -
// lowercased, each once.
func ruleHosts(rule string) []string {
	var hosts []string
	for _, matcher := range hostMatcher.FindAllStringSubmatch(rule, -1) {
		for _, arg := range hostArg.FindAllStringSubmatch(matcher[1], -1) {
			host := strings.ToLower(strings.TrimSpace(arg[1]))
			if hostName.MatchString(host) && !slices.Contains(hosts, host) {
				hosts = append(hosts, host)
			}
		}
	}
	return hosts
}
