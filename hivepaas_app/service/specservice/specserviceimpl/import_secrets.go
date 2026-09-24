package specserviceimpl

import (
	"maps"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// generatedSecretTypes hold values HivePaaS owns: an app's database, cache or
// storage credentials, and a project's webhook secret. Nothing depends on them
// before the object exists, so a created one missing its value is given one.
var generatedSecretTypes = []base.SettingType{base.SettingTypeAppKind, base.SettingTypeRepoWebhook}

// appKindCredentials are where an app-kind setting keeps its credentials:
// TestAppKindCredentialsAreItsSecrets holds this to the type's secrets.
var appKindCredentials = [][]string{
	{"database", "password"}, {"database", "rootPassword"}, {"cache", "password"}, {"storage", "secret"},
}

// createdSettings names the settings import creates in a node, the way changes
// name them: all of a node being created, the ones the target lacks of one
// being updated, the ones closure pulled into one it selected.
func (p *planner) createdSettings(node *specmodel.PlanNode) []string {
	switch {
	case node.Action == specmodel.ActionCreate:
		return settingNames(p.settingsOf[node.Path])
	case node.Action != specmodel.ActionUpdate:
		return nil
	case node.SelectedBy == selectedByDependency:
		return p.pulled[node.Path]
	}
	prefix := p.settingsPrefix(node)
	var names []string
	for _, missing := range p.missing[node.Path] {
		if name, ok := strings.CutPrefix(missing, prefix); ok {
			names = append(names, name)
		}
	}
	return names
}

// checkSecrets says what becomes of the secrets of what import creates. A
// bundle that omitted its secrets leaves each created one empty - or, for a
// value HivePaaS owns, generates it. Whatever exists keeps the target's secret
// in every mode, so it is not looked at.
func (p *planner) checkSecrets() error {
	if p.bundle.Manifest.SecretsMode != specmodel.SecretsModeOmit {
		p.noteKeptCredentials()
		return nil
	}
	for _, node := range p.nodes {
		if !node.Selected {
			continue
		}
		prefix := p.settingsPrefix(node)
		settings := p.settingsOf[node.Path]
		for _, name := range p.createdSettings(node) {
			block, key, _ := strings.Cut(name, "/")
			body, typ, found := settingBody(settings, block, key)
			if !found {
				continue
			}
			var externals []*specmodel.ExternalRef
			data, err := readBundleSetting(typ, node.Path, prefix+name, withExternalPlaceholders(body, &externals))
			if err != nil {
				return err
			}
			empty := entity.CountEmptySecrets(data)
			switch {
			case empty == 0:
			case slices.Contains(generatedSecretTypes, typ):
				node.Notes = append(node.Notes, specmodel.Issue{
					Code: specmodel.CodeSecretGenerated, Path: node.Path,
					Detail: map[string]any{refInSetting: prefix + name},
					Action: "the bundle omitted its secrets, so HivePaaS generates them",
				})
			default:
				node.Issues = append(node.Issues, specmodel.Issue{
					Severity: specmodel.SeverityFixable, Code: specmodel.CodeSecretOmitted, Path: node.Path,
					Detail: map[string]any{refInSetting: prefix + name, "secrets": empty},
					Action: "created with its secrets empty, and pending until they are given",
				})
			}
		}
	}
	return nil
}

// noteKeptCredentials notes each app that keeps the target's credential: an
// app matched by key is another app with the same key - recreated since the
// export, or on another installation - and its data was initialized with the
// target's credential, not the bundle's.
func (p *planner) noteKeptCredentials() {
	kindBlock := specmodel.SingletonBlockName(base.SettingTypeAppKind)
	for _, node := range p.writingApps() {
		if node.MatchedBy != specmodel.MatchedByKey {
			continue
		}
		current := p.currentApp(node)
		if current == nil || !hasCredential(current.Settings[kindBlock]) {
			continue
		}
		node.Notes = append(node.Notes, specmodel.Issue{
			Code: specmodel.CodeCredentialKept, Path: node.Path,
			Detail: map[string]any{refInSetting: "settings." + kindBlock},
			Action: "the app keeps this installation's credential, which its data was initialized with",
		})
	}
}

// withTargetCredential is an app's settings as import would write them, for the
// diff: an app matched by key keeps the target's credential, so the bundle's is
// no difference.
func (p *planner) withTargetCredential(
	bundle, current map[string]any, matchedBy specmodel.MatchedBy,
) map[string]any {
	kindBlock := specmodel.SingletonBlockName(base.SettingTypeAppKind)
	currentKind := current[kindBlock]
	if matchedBy != specmodel.MatchedByKey || !p.bundle.Manifest.SecretsMode.RevealsSecrets() ||
		!hasCredential(currentKind) {
		return bundle
	}
	kind, ok := bundle[kindBlock].(map[string]any)
	if !ok {
		return bundle
	}
	out := maps.Clone(bundle)
	kind = maps.Clone(kind)
	for _, path := range appKindCredentials {
		value, found := valueAt(currentKind, path)
		if !found {
			continue
		}
		parent, _ := kind[path[0]].(map[string]any)
		parent = maps.Clone(parent)
		if parent == nil {
			parent = map[string]any{}
		}
		parent[path[1]] = value
		kind[path[0]] = parent
	}
	out[kindBlock] = kind
	return out
}

func hasCredential(kind any) bool {
	for _, path := range appKindCredentials {
		if value, found := valueAt(kind, path); found && value != "" {
			return true
		}
	}
	return false
}

func valueAt(node any, path []string) (any, bool) {
	for _, key := range path {
		fields, ok := node.(map[string]any)
		if !ok {
			return nil, false
		}
		if node, ok = fields[key]; !ok {
			return nil, false
		}
	}
	return node, true
}
