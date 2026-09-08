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

func IsTraefikCmdArgSettable(key string) bool {
	key = strings.ToLower(key)
	if strings.HasPrefix(key, "entrypoints.tcp-svc-") || strings.HasPrefix(key, "entrypoints.udp-svc-") {
		return false
	}
	_, exists := mapTraefikUnsettableCmdArgs[key]
	return !exists
}
