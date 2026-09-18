package templaterepo

import (
	"fmt"
	"maps"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
)

// lintKeyPrefix makes the stand-in key a dependency is bound to while linting.
const lintKeyPrefix = "lint_"

// lintDependencies checks what a template's dependencies need from the rest of
// the repository. A problem here makes a bound render meaningless, so Lint skips
// the renders of a template that has one.
func lintDependencies(repo *Repo, file *TemplateFile) []Problem {
	var problems []Problem
	report := func(format string, args ...any) {
		problems = append(problems, Problem{Path: file.Path, Message: fmt.Sprintf(format, args...)})
	}
	for _, dep := range file.Template.Dependencies {
		prefix := "dependency " + dep.Name
		target := repo.FindTemplate(dep.Template)
		if target == nil {
			report("%s: template %q is not in this repository", prefix, dep.Template)
			continue
		}
		depTmpl := target.Template
		if len(depTmpl.Dependencies) > 0 {
			report("%s: template %q has dependencies of its own, and dependencies go one level deep",
				prefix, dep.Template)
		}
		if dep.Version != "" {
			version := depTmpl.FindVersion(dep.Version)
			switch {
			case version == nil:
				report("%s: template %q has no version %q", prefix, dep.Template, dep.Version)
			case version.Deprecated:
				report("%s: version %q of %q is deprecated", prefix, dep.Version, dep.Template)
			}
		}
		if dep.Variant != "" && depTmpl.FindVariant(dep.Variant) == nil {
			report("%s: template %q has no variant %q", prefix, dep.Template, dep.Variant)
		}
		for _, name := range slices.Sorted(maps.Keys(dep.Params)) {
			declared := slices.ContainsFunc(depTmpl.Parameters, func(param *templatemodel.Parameter) bool {
				return param != nil && param.Name == name
			})
			if !declared {
				report("%s: template %q declares no parameter %q", prefix, dep.Template, name)
			}
		}
	}
	return problems
}

// lintBindings renders each dependency the way creating an app would - its fixed
// parameters, a stand-in for each one a person is asked - and binds it under a
// stand-in key to what it shares.
func lintBindings(
	repo *Repo,
	file *TemplateFile,
	owner map[string]*templaterender.Value,
) (map[string]*templaterender.DepBinding, []Problem) {
	tmpl := file.Template
	bindings := map[string]*templaterender.DepBinding{}
	var problems []Problem
	for _, dep := range tmpl.Dependencies {
		depTmpl := repo.FindTemplate(dep.Template).Template
		asked := map[string]any{}
		for _, param := range dep.AskedParams(depTmpl) {
			asked[param.Name] = lintStandIn(param)
		}
		input, err := templaterender.DependencyParams(tmpl.Metadata.Name, dep, depTmpl, owner, asked)
		if err == nil {
			var result *templaterender.Result
			result, err = templaterender.Render(&templaterender.Request{
				Template: depTmpl, Version: dep.Version, Variant: dep.Variant, Params: input,
			})
			if err == nil {
				bindings[dep.Name] = &templaterender.DepBinding{
					AppKey: lintKeyPrefix + dep.Name, SharedVars: templaterender.SharedVarsOf(result.Doc),
				}
				continue
			}
		}
		problems = append(problems, Problem{Path: file.Path,
			Message: "dependency " + dep.Name + ": " + ErrorText(err)})
	}
	return bindings, problems
}

// lintStandIn is a value a person could have given for a parameter they are asked.
func lintStandIn(param *templatemodel.Parameter) any {
	switch param.Type {
	case templatemodel.ParamTypeVolume:
		return lintVolumeID
	case templatemodel.ParamTypeSecret:
		return lintSecret
	case templatemodel.ParamTypeDomain:
		return lintDomain
	case templatemodel.ParamTypeInt, templatemodel.ParamTypeSize:
		if param.Min != nil {
			return param.Min
		}
		if param.Type == templatemodel.ParamTypeSize {
			return "1MB"
		}
		return 1
	case templatemodel.ParamTypeBool:
		return false
	case templatemodel.ParamTypeSelect:
		if len(param.Options) > 0 {
			return param.Options[0].Value
		}
		return ""
	case templatemodel.ParamTypeString:
		return "lint"
	}
	return "lint"
}
