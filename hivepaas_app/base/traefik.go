package base

import "strings"

var (
	mapTraefikUnsettableCmdArgs = map[string]struct{}{
		"api.insecure":                     {},
		"api.dashboard":                    {},
		"providers.swarm":                  {},
		"providers.swarm.watch":            {},
		"providers.swarm.network":          {},
		"providers.swarm.exposedbydefault": {},
		"entrypoints.web.address":          {},
		"entrypoints.web.allowacmebypass":  {},
		"entrypoints.websecure.address":    {},
		"entrypoints.websecure.http.tls":   {},
		"providers.file.directory":         {},
		"providers.file.watch":             {},

		// The liveness endpoint swarm's healthcheck reads. Removing or moving any
		// of these leaves the check probing a port nothing answers on, and swarm
		// answers that by restarting a traefik that was working - repeatedly. See
		// the healthcheck in deployment/*/hivepaas.yaml.
		"ping":                     {},
		"ping.entrypoint":          {},
		"entrypoints.ping.address": {},
	}
)

// TraefikAccessLogArgs are what the Access Log option turns on: the access
// log, as JSON - an app's HTTP numbers are counted from it - and with the query
// string dropped, which can carry a token or an email.
var TraefikAccessLogArgs = []string{
	"--accesslog=true",
	"--accesslog.format=json",
	"--accesslog.fields.queryparameters.defaultmode=drop",
}

// IsTraefikAccessLogArg reports whether an argument's key is one the Access
// Log option owns: written by it, not by the operator's own arguments.
func IsTraefikAccessLogArg(key string) bool {
	switch strings.ToLower(key) {
	case "accesslog", "accesslog.format", "accesslog.fields.queryparameters.defaultmode":
		return true
	}
	return false
}

func IsTraefikCmdArgSettable(key string) bool {
	key = strings.ToLower(key)
	if strings.HasPrefix(key, "entrypoints.tcp-svc-") || strings.HasPrefix(key, "entrypoints.udp-svc-") {
		return false
	}
	_, exists := mapTraefikUnsettableCmdArgs[key]
	return !exists
}
