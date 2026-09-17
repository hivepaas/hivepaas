package templatemodel

import (
	"fmt"
	"maps"
	"regexp"
	"slices"
)

// MaxDependencies caps what one template creates besides its own app. One
// database is the common case, a database and a cache the next; the cap is what
// lets a creation dialog show everything a request is about to create.
const MaxDependencies = 3

// dependencyNamePattern is narrower than a parameter name. The name becomes the
// suffix of an app name, and that app's key is what environment references use,
// so it has to come through slugifying unchanged.
var dependencyNamePattern = regexp.MustCompile(`^[a-z][a-z0-9]{0,15}$`)

// Dependency is another template of the same repository whose app is created
// alongside this template's own: a web app's database.
type Dependency struct {
	// Name is the role - db, cache - and the suffix of the created app's name.
	Name  string `yaml:"name"`
	Title string `yaml:"title"`
	// Template names a template in the same repository.
	Template string `yaml:"template"`
	// Version and Variant are empty for the dependency template's defaults.
	Version string `yaml:"version,omitempty"`
	Variant string `yaml:"variant,omitempty"`
	// Params are the values this template fixes for the dependency. A string may
	// use this template's own ${{ params.x }} placeholders.
	Params map[string]any `yaml:"params,omitempty"`
}

// DependencyAppName is the name of the app created for a dependency.
func DependencyAppName(appName, dependencyName string) string {
	return appName + "-" + dependencyName
}

func (t *Template) FindDependency(name string) *Dependency {
	for _, dep := range t.Dependencies {
		if dep != nil && dep.Name == name {
			return dep
		}
	}
	return nil
}

// AskedParams are the parameters of the dependency's template that the person
// creating the app fills in: those this dependency does not fix, with no default,
// not optional, and not secrets HivePaaS generates. For a database that is its
// data volume. An empty default counts as none, as it does when parameters are
// resolved.
func (d *Dependency) AskedParams(depTemplate *Template) []*Parameter {
	var asked []*Parameter
	for _, param := range depTemplate.Parameters {
		if param == nil {
			continue
		}
		if _, fixed := d.Params[param.Name]; fixed {
			continue
		}
		defaulted := param.Default != nil && param.Default != ""
		generated := param.Type == ParamTypeSecret && param.Generate != nil
		if defaulted || param.Optional || generated {
			continue
		}
		asked = append(asked, param)
	}
	return asked
}

// validateDependencies checks what a template says about its dependencies. What
// needs the templates they name is templaterepo's.
func validateDependencies(t *Template, p *problems) {
	if len(t.Dependencies) > MaxDependencies {
		p.add("dependencies: at most %d", MaxDependencies)
	}
	seen := map[string]bool{}
	for i, dep := range t.Dependencies {
		if dep == nil {
			p.add("dependencies[%d] is empty", i)
			continue
		}
		prefix := fmt.Sprintf("dependencies[%s]", dep.Name)
		if !dependencyNamePattern.MatchString(dep.Name) {
			p.add("dependencies[%d].name %q must be lowercase letters and digits, starting with a letter", i, dep.Name)
		}
		if seen[dep.Name] {
			p.add("%s is declared twice", prefix)
		}
		seen[dep.Name] = true
		if dep.Title == "" {
			p.add("%s.title is required", prefix)
		}
		switch {
		case !templateNamePattern.MatchString(dep.Template):
			p.add("%s.template %q is not a template name", prefix, dep.Template)
		case dep.Template == t.Metadata.Name:
			p.add("%s: a template cannot depend on itself", prefix)
		}
		if dep.Version != "" && !choiceNamePattern.MatchString(dep.Version) {
			p.add("%s.version %q is invalid", prefix, dep.Version)
		}
		if dep.Variant != "" && !choiceNamePattern.MatchString(dep.Variant) {
			p.add("%s.variant %q is invalid", prefix, dep.Variant)
		}
		for _, name := range slices.Sorted(maps.Keys(dep.Params)) {
			if !paramNamePattern.MatchString(name) {
				p.add("%s.params: %q is not a parameter name", prefix, name)
			}
		}
	}
}
