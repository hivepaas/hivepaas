package base

type HTTPPathMode string

const (
	HTTPPathModeExact  HTTPPathMode = "exact"
	HTTPPathModePrefix HTTPPathMode = "prefix"
	HTTPPathModeRegex  HTTPPathMode = "regex"
)

var (
	AllHTTPPathModes = []HTTPPathMode{HTTPPathModeExact, HTTPPathModePrefix, HTTPPathModeRegex}
)
