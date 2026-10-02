package traefikservice

import (
	"regexp"
	"strconv"
	"strings"
)

// AppRouteName is what an app's routers, services and middlewares are named by
// in Traefik: its id, lower-cased. Not its key: a key is unique in its env
// only, and two services naming the same router make Traefik drop it from
// both - an app named "app" in any project took the HivePaaS dashboard's
// router-app-0 down that way. An id is unique, and survives a rename.
func AppRouteName(appID string) string {
	return strings.ToLower(appID)
}

// AppHTTPServiceName is the Traefik service of one of an app's domains, as its
// access log names it without the provider: svc-<id>-<domain index>.
func AppHTTPServiceName(appID string, domainIndex int) string {
	return "svc-" + AppRouteName(appID) + "-" + strconv.Itoa(domainIndex)
}

// AppHTTPServicePattern matches, exactly, the Traefik services of an app's
// domains as the access log names them, provider included: svc-<id>-<n>@swarm.
// Anchored, so that an app's id never matches as the start of another's.
func AppHTTPServicePattern(appID string) string {
	return "^svc-" + regexp.QuoteMeta(AppRouteName(appID)) + "-[0-9]+@swarm$"
}

// traefikNameLabel is a router, service or middleware label: its kind and its
// name.
var traefikNameLabel = regexp.MustCompile(`^traefik\.(?:http|tcp|udp)\.(?:routers|services|middlewares)\.([^.]+)\.`)

// HasStaleRouteNames reports whether a service's labels still name routers,
// services or middlewares after something other than the app's id - written
// before AppRouteName - so that its routing has to be applied again. An
// operator's x-custom- labels are theirs and are not looked at.
func HasStaleRouteNames(labels map[string]string, appID string) bool {
	name := AppRouteName(appID)
	for key := range labels {
		m := traefikNameLabel.FindStringSubmatch(key)
		if m == nil || strings.HasPrefix(m[1], "x-custom-") {
			continue
		}
		if !strings.Contains(m[1], "-"+name+"-") && !strings.HasSuffix(m[1], "-"+name) {
			return true
		}
	}
	return false
}
