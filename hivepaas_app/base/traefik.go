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
