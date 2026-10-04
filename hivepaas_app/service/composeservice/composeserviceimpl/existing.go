package composeserviceimpl

import (
	"maps"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// Blocks of an env's settings the converter writes, and reads of an existing
// env's.
const (
	blockSecrets     = "secrets"
	blockConfigFiles = "configFiles"
)

// existingNames are the names the apps of an existing env answer to on its
// network - their keys and aliases - each with the key of its app.
func existingNames(env *specmodel.EnvDoc) map[string]string {
	out := map[string]string{}
	if env == nil {
		return out
	}
	for _, key := range slices.Sorted(maps.Keys(env.Apps)) {
		out[key] = key
	}
	for _, key := range slices.Sorted(maps.Keys(env.Apps)) {
		app := env.Apps[key]
		if app == nil || app.Deployment == nil || app.Deployment.Networks == nil {
			continue
		}
		for _, attachment := range app.Deployment.Networks.Attachments {
			if attachment == nil {
				continue
			}
			for _, alias := range attachment.Aliases {
				if _, found := out[alias]; !found {
					out[alias] = key
				}
			}
		}
	}
	return out
}

// existingSetting says whether the existing env has a setting of a block by
// a name.
func (c *converter) existingSetting(block, name string) bool {
	if c.req.Existing == nil {
		return false
	}
	entries, _ := c.req.Existing.Settings[block].(map[string]any)
	_, found := entries[name]
	return found
}

// existingApp is the existing env's app a service is used as; nil for one
// created.
func (c *converter) existingApp(name string) *specmodel.AppDoc {
	key := c.used[name]
	if key == "" || c.req.Existing == nil {
		return nil
	}
	return c.req.Existing.Apps[key]
}
