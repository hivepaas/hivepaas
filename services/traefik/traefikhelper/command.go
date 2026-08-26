package traefikhelper

import "strings"

//nolint:mnd
func ParseCommandArg(arg string) (key string, val string, valid bool) {
	arg = strings.TrimSpace(arg)
	if arg == "" || arg == "traefik" {
		return arg, "", false
	}
	s := strings.TrimLeft(arg, "-")
	parts := strings.SplitN(s, "=", 2)
	key = strings.ToLower(strings.TrimSpace(parts[0]))
	if len(parts) == 2 {
		val = parts[1]
	}
	return key, val, key != ""
}
