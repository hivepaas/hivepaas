package templaterepo

import (
	"fmt"
	"maps"
	"regexp"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// paramRefPattern matches the ${{ params.name }} placeholders of one string.
var paramRefPattern = regexp.MustCompile(`\$\{\{\s*params\.([A-Za-z0-9_]+)\s*\}\}`)

// lintConfigFileSecrets reports secret parameters written into the content of a
// config file.
//
// A config file is kept as it reads: the setting holding it is not encrypted,
// and neither is the docker config object it becomes. A secret asked of the
// person creating the app has to stay encrypted, so it belongs in
// settings.secrets - which can mount it as a file of its own when the app wants
// to read it from disk rather than from the environment.
func lintConfigFileSecrets(file *TemplateFile) []Problem {
	tmpl := file.Template
	secretParams := map[string]bool{}
	for _, param := range tmpl.Parameters {
		if param != nil && param.Type == templatemodel.ParamTypeSecret {
			secretParams[param.Name] = true
		}
	}
	if len(secretParams) == 0 {
		return nil
	}

	var problems []Problem
	report := func(where, name, param string) {
		problems = append(problems, Problem{Path: file.Path, Message: fmt.Sprintf(
			"%sconfig file %q writes the secret parameter %q into its content:"+
				" declare it under settings.%s instead", where, name, param,
			specmodel.CollectionBlockName(base.SettingTypeSecret))})
	}
	for name, param := range configFileSecretRefs(tmpl.App, secretParams) {
		report("", name, param)
	}
	for _, version := range tmpl.Versions {
		if version == nil || version.Override == nil {
			continue
		}
		for name, param := range configFileSecretRefs(version.Override.App, secretParams) {
			report("version "+version.Name+": ", name, param)
		}
	}
	return problems
}

// configFileSecretRefs maps the name of each config file of an app tree that
// refers to a secret parameter to the first such parameter, in a fixed order.
func configFileSecretRefs(app map[string]any, secretParams map[string]bool) map[string]string {
	settings, _ := app["settings"].(map[string]any)
	entries, _ := settings[specmodel.CollectionBlockName(base.SettingTypeConfigFile)].(map[string]any)
	refs := map[string]string{}
	for _, name := range slices.Sorted(maps.Keys(entries)) {
		entry, _ := entries[name].(map[string]any)
		content, _ := entry["content"].(string)
		for _, match := range paramRefPattern.FindAllStringSubmatch(content, -1) {
			if secretParams[match[1]] {
				refs[name] = match[1]
				break
			}
		}
	}
	return refs
}
