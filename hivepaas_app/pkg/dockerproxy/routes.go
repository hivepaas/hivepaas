package dockerproxy

import (
	"regexp"
	"strings"
)

type route struct {
	methods []string
	pattern *regexp.Regexp
	group   Group
	handle  func(p *Proxy, c *call)
}

func on(methods, pattern string, group Group, handle func(*Proxy, *call)) route {
	return route{
		methods: strings.Split(methods, ","),
		pattern: regexp.MustCompile("^" + pattern + "$"),
		group:   group,
		handle:  handle,
	}
}

// idPart captures an id or a name in a path.
const idPart = `([^/]+)`

// routes are every endpoint an app may reach, after its API version is taken
// off. Anything else is refused.
var routes = []route{
	on("GET,HEAD", `/_ping`, "", (*Proxy).pass),
	on("GET", `/version`, "", (*Proxy).pass),
	on("GET", `/info`, "", (*Proxy).info),
	on("GET", `/images/json`, "", (*Proxy).pass),
	on("GET", `/images/(.+)/json`, "", (*Proxy).pass),
	on("POST", `/images/create`, "", (*Proxy).pull),
}
