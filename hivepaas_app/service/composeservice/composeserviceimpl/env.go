package composeserviceimpl

import (
	"bytes"
	"maps"
	"regexp"
	"slices"
	"strings"

	"github.com/compose-spec/compose-go/v2/dotenv"
	"github.com/compose-spec/compose-go/v2/types"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// envVars are a service's environment as the app's variables: its env_files
// first and its environment over them, as compose reads them. A secret
// variable in a value is a reference to the env secret holding it; a variable
// whose own name reads as a secret's and whose value the file writes out is
// kept as a secret of the app, which the variable refers to - secrets
// collects them.
func (c *converter) envVars(
	appPath, name string, svc types.ServiceConfig, secrets map[string]any,
) (map[string]any, []string) {
	values := map[string]string{}
	for _, file := range svc.EnvFiles {
		c.readEnvFile(appPath, name, file, values)
	}
	for key, value := range svc.Environment {
		if value != nil {
			values[key] = *value
			continue
		}
		// `- KEY` alone takes the variable of the same name, when there is one.
		if from, ok := c.lookup(key); ok {
			values[key] = from
		}
	}

	data := make([]any, 0, len(values))
	var kept []string
	for _, key := range slices.Sorted(maps.Keys(values)) {
		value, plainNames := c.r.markers.replacePlain(values[key], c.r.values)
		value, names := c.r.markers.replace(value, func(variable string) string {
			c.secretVariables[variable] = true
			return "${secrets." + c.secretOfVariable(variable) + "}"
		})
		entry := map[string]any{"k": key, "v": value}
		switch {
		case len(names) > 0:
		case strings.Contains(value, "${"):
			// Not a reference of HivePaaS's - written `$${` in compose, or a
			// file's own: it reaches the container as it is.
			entry["literal"] = true
		case len(plainNames) > 0:
			// A variable the review says is not secret fills it.
		case value != "" && secretByName(key) && secretNamePattern.MatchString(key):
			secrets[key] = map[string]any{"key": key, "value": value,
				specmodel.SettingMetaKey: map[string]any{detailName: key}}
			entry["v"] = "${secrets." + key + "}"
			kept = append(kept, key)
		}
		data = append(data, entry)
	}
	if len(kept) > 0 {
		c.add(appPath, "", composeservice.CodeSecretEnv, map[string]any{"variables": kept},
			"kept as secrets of the app, which its variables refer to")
	}
	if len(data) == 0 {
		return nil, kept
	}
	return map[string]any{"data": data}, kept
}

// secretNamePattern is a name a variable can refer to a secret by.
var secretNamePattern = regexp.MustCompile(`^[a-zA-Z_][a-zA-Z0-9_]*$`)

// readEnvFile reads one env_file into values, from the request's files.
func (c *converter) readEnvFile(appPath, name string, file types.EnvFile, values map[string]string) {
	rel, ok := cleanPath(file.Path)
	if !ok {
		c.add(appPath, specmodel.SeverityFixable, composeservice.CodeFileMissing,
			map[string]any{detailPath: file.Path}, "not read: it leaves the compose file's directory")
		return
	}
	content, given := c.r.files[rel]
	c.need(rel, composeservice.NeedEnvFile, name, given)
	if !given {
		if bool(file.Required) {
			c.add(appPath, specmodel.SeverityFixable, composeservice.CodeFileMissing,
				map[string]any{detailPath: rel}, "its variables are left out until it is given")
		}
		return
	}
	read := map[string]string{}
	if err := dotenv.ParseWithFormat(bytes.NewReader(content), rel, read, c.lookup, file.Format); err != nil {
		c.add(appPath, specmodel.SeverityFixable, composeservice.CodeFileMissing,
			map[string]any{detailPath: rel, "error": err.Error()}, "its variables are left out: it cannot be read")
		return
	}
	maps.Copy(values, read)
}

// lookup is a variable as the file is read with it: a secret one as its
// marker.
func (c *converter) lookup(name string) (string, bool) {
	value, ok := c.r.values[name]
	if marked, isMarked := c.r.marked(name, value); isMarked {
		return marked, true
	}
	return value, ok
}

// secretOfVariable is the env secret a secret variable is kept in: named after
// it, unless one of the file's secrets is.
func (c *converter) secretOfVariable(variable string) string {
	name := variable
	for {
		body, taken := c.envSecrets[name]
		if !taken || isVariableSecret(body, variable) {
			return name
		}
		name += "_VARIABLE"
	}
}

// variableKey marks which variable an env secret holds, while the bundle is
// made; it is removed before it is written.
const variableKey = "variable"

func isVariableSecret(body any, variable string) bool {
	fields, _ := body.(map[string]any)
	return fields[variableKey] == variable
}

// secretVariableSettings are the env secrets the secret variables an
// environment refers to are kept in.
func (c *converter) secretVariableSettings() {
	for _, variable := range slices.Sorted(maps.Keys(c.secretVariables)) {
		name := c.secretOfVariable(variable)
		body := c.fileSetting(name, partValue, []byte(c.r.values[variable]), true, "")
		body[variableKey] = variable
		c.envSecrets[name] = body
		detail := map[string]any{"variable": variable, "secret": name}
		if c.existingSetting(blockSecrets, name) {
			c.add(c.envPath(), specmodel.SeverityWarning, composeservice.CodeSecretExists, detail,
				"the env's secret is used as it is: the value given here is not written")
			continue
		}
		c.add(c.envPath(), "", composeservice.CodeSecretVariable, detail,
			"kept as an env secret, which the apps' variables refer to")
	}
}
