package templatemodel

import (
	"regexp"
	"strings"
)

// portsPath is where a template says which ports its app answers at on the
// cluster itself, rather than through the reverse proxy.
var portsPath = []string{"deployment", "networks", "endpointSpec", "ports"}

// paramPlaceholderPattern matches a value that is one parameter and nothing
// else: ${{ params.vpnPort }}. A port written that way is chosen when the app is
// created, and what is recorded here is which parameter chooses it.
var paramPlaceholderPattern = regexp.MustCompile(`^\$\{\{\s*params\.([A-Za-z][A-Za-z0-9_]*)\s*\}\}$`)

// PublishedPort is one address a template claims on the cluster.
//
// Published is the port as the template wrote it, or the parameter's default
// when a parameter chooses it - PublishedParam then names that parameter, so
// that what is shown before deploying can follow what a person types rather
// than the default they are about to change.
type PublishedPort struct {
	Target         uint32
	Published      uint32
	PublishedParam string
	Protocol       string
	PublishMode    string
}

// PublishedPorts are the ports a template publishes, read from the template
// rather than from a render: the store shows them before anybody deploys.
func (t *Template) PublishedPorts() []PublishedPort {
	return t.publishedPortsIn(t.App)
}

// ComponentPublishedPorts are the ports one component's app publishes.
func (t *Template) ComponentPublishedPorts(component *Component) []PublishedPort {
	return t.publishedPortsIn(component.App)
}

// publishedPortsIn reads the ports out of an app tree, a port given by a
// parameter as that parameter's default.
func (t *Template) publishedPortsIn(app map[string]any) []PublishedPort {
	node := any(app)
	for _, key := range portsPath {
		parent, ok := node.(map[string]any)
		if !ok {
			return nil
		}
		if node, ok = parent[key]; !ok {
			return nil
		}
	}
	entries, ok := node.([]any)
	if !ok {
		return nil
	}

	defaults := map[string]any{}
	for _, param := range t.Parameters {
		if param != nil {
			defaults[param.Name] = param.Default
		}
	}

	ports := make([]PublishedPort, 0, len(entries))
	for _, entry := range entries {
		fields, isMap := entry.(map[string]any)
		if !isMap {
			continue
		}
		target, _ := portNumber(fields["target"], defaults)
		published, param := portNumber(fields["published"], defaults)
		protocol, _ := fields["protocol"].(string)
		publishMode, _ := fields["publishMode"].(string)
		ports = append(ports, PublishedPort{
			Target:         target,
			Published:      published,
			PublishedParam: param,
			Protocol:       strings.ToLower(protocol),
			PublishMode:    strings.ToLower(publishMode),
		})
	}
	return ports
}

// portNumber reads a port written as a number, or as one parameter - in which
// case it is that parameter's default, and the parameter's name.
func portNumber(value any, defaults map[string]any) (uint32, string) {
	if text, isText := value.(string); isText {
		match := paramPlaceholderPattern.FindStringSubmatch(text)
		if match == nil {
			return 0, ""
		}
		number, _ := portNumber(defaults[match[1]], nil)
		return number, match[1]
	}
	number, ok := ToInt64(value)
	if !ok || number < 0 || number > maxPortNumber {
		return 0, ""
	}
	return uint32(number), ""
}

// maxPortNumber is the largest port there is.
const maxPortNumber = 65535
