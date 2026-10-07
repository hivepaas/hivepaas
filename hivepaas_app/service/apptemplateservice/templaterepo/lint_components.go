package templaterepo

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
)

// componentAppRef is a reference to another component's app, made when the
// apps are created: ${${{ comp.server.key }}.HIVEPAAS_APP_URL}.
var componentAppRef = regexp.MustCompile(`\$\{\$\{\{\s*comp\.([A-Za-z0-9_-]+)\.key\s*\}\}\.`)

// lintComponentAppRefs keeps a component from reading an app created after it.
//
// The key ${{ comp.X.key }} is known before anything is created, so a component
// can name any other. A reference through it, ${<key>.VAR}, is resolved as the
// component's app is created, and the apps are created in the order the needs
// put them: one created later is not there to read, and the whole creation
// fails with "app not found". Headscale's server read its proxy's address while
// the proxy, needing the server, came after it.
func lintComponentAppRefs(file *TemplateFile) []Problem {
	tmpl := file.Template
	ordered, err := tmpl.ComponentOrder()
	if err != nil || len(ordered) == 0 {
		// A cycle is reported by validation.
		return nil
	}

	var problems []Problem
	created := map[string]bool{}
	for _, component := range ordered {
		read := map[string]bool{}
		eachString(component.App, func(value string) {
			for _, match := range componentAppRef.FindAllStringSubmatch(value, -1) {
				read[match[1]] = true
			}
		})
		for _, other := range slices.Sorted(maps.Keys(read)) {
			if other != component.Name && !created[other] && tmpl.FindComponent(other) != nil {
				problems = append(problems, Problem{Path: file.Path, Message: fmt.Sprintf(
					"component %s reads the app of component %s, which is created after it: "+
						"make %s need %s, or read it some other way", component.Name, other, component.Name, other)})
			}
		}
		created[component.Name] = true
	}
	return problems
}

// eachString calls visit with every string in a decoded YAML tree.
func eachString(node any, visit func(string)) {
	switch value := node.(type) {
	case string:
		visit(value)
	case map[string]any:
		for _, child := range value {
			eachString(child, visit)
		}
	case []any:
		for _, child := range value {
			eachString(child, visit)
		}
	}
}
