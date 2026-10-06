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

// TraefikAccessLogFields are the access log's fields the Access Log option
// keeps; Traefik drops the others. Every line is stored and read back on each
// query, and a line with all of them - addresses and ports twice, the start
// time twice, sizes HivePaaS does not count by - is nearly twice as long: an
// app's HTTP numbers took a fifth less time without them, its request load
// half, on Traefik 3.7.13 and VictoriaLogs 1.53.0.
var TraefikAccessLogFields = []string{
	// What an app's HTTP numbers and its request load are counted from.
	"ServiceName", "ServiceURL", "RequestMethod", "RequestPath",
	"DownstreamStatus", "OriginStatus", "Duration", "OriginDuration",
	// What a person reading a line asks: who, to which host, by which route,
	// how much came back, whether it was retried.
	"ClientHost", "ClientUsername", "RequestHost", "RouterName", "RequestProtocol",
	"DownstreamContentSize", "RetryAttempts", "TLSVersion", "entryPointName",
}

// TraefikAccessLogArgs are what the Access Log option turns on: the access
// log, as JSON - an app's HTTP numbers are counted from it - with the query
// string dropped, which can carry a token or an email, and with only the
// fields TraefikAccessLogFields names.
var TraefikAccessLogArgs = func() []string {
	args := []string{
		"--accesslog=true",
		"--accesslog.format=json",
		"--accesslog.fields.queryparameters.defaultmode=drop",
		"--accesslog.fields.defaultmode=drop",
	}
	for _, field := range TraefikAccessLogFields {
		args = append(args, "--accesslog.fields.names."+field+"=keep")
	}
	return args
}()

// traefikAccessLogFieldKeys are the keys of TraefikAccessLogFields' arguments,
// lower-cased as a key is compared.
var traefikAccessLogFieldKeys = func() map[string]struct{} {
	keys := make(map[string]struct{}, len(TraefikAccessLogFields))
	for _, field := range TraefikAccessLogFields {
		keys["accesslog.fields.names."+strings.ToLower(field)] = struct{}{}
	}
	return keys
}()

// IsTraefikAccessLogArg reports whether an argument's key is one the Access
// Log option owns: written by it, not by the operator's own arguments. A field
// it keeps is its own, so that no argument drops one; another field an
// operator keeps, --accesslog.fields.names.StartUTC=keep, is theirs.
func IsTraefikAccessLogArg(key string) bool {
	key = strings.ToLower(key)
	switch key {
	case "accesslog", "accesslog.format", "accesslog.fields.queryparameters.defaultmode",
		"accesslog.fields.defaultmode":
		return true
	}
	_, ok := traefikAccessLogFieldKeys[key]
	return ok
}

func IsTraefikCmdArgSettable(key string) bool {
	key = strings.ToLower(key)
	if strings.HasPrefix(key, "entrypoints.tcp-svc-") || strings.HasPrefix(key, "entrypoints.udp-svc-") {
		return false
	}
	_, exists := mapTraefikUnsettableCmdArgs[key]
	return !exists
}
