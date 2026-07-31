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
	_, exists := mapTraefikUnsettableCmdArgs[strings.ToLower(key)]
	return !exists
}
