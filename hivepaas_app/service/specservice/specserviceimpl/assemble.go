package specserviceimpl

import (
	"encoding/json"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// refIndex resolves a stored identifier to what a spec writes in its place.
//
// Anything reachable inside the exported scope becomes a path, so the scope
// travels with the reference and two settings sharing a name in different
// scopes can never be confused. Everything else becomes an external block: the
// exporter cannot resolve it, so it records what the importing installation
// needs to look it up. The scope rule produces that case constantly - exporting
// one project whose apps use a global certificate is the ordinary shape.
type refIndex struct {
	paths    map[string]string
	external map[string]*specmodel.ExternalRef
}

func newRefIndex() *refIndex {
	return &refIndex{
		paths:    map[string]string{},
		external: map[string]*specmodel.ExternalRef{},
	}
}

func (r *refIndex) addPath(settingID, path string) { r.paths[settingID] = path }

func (r *refIndex) addExternal(setting *entity.Setting) {
	r.external[setting.ID] = &specmodel.ExternalRef{
		Type: string(setting.Type),
		Name: setting.Name,
		Kind: setting.Kind,
		ID:   setting.ID,
	}
}

// assembleSettings turns a scope's settings into the map one document holds.
//
// A singleton takes its block name and nothing else; a collection becomes a map
// keyed by the key DeriveSettingKeys produced. A type in neither map is refused
// rather than guessed at, for the same reason an unregistered SpecPolicy is: a
// setting type added later should stop somebody, not appear under a name nobody
// chose.
func assembleSettings(settings []*entity.Setting, index *refIndex) (map[string]any, error) {
	byType := map[base.SettingType][]*entity.Setting{}
	for _, setting := range settings {
		byType[setting.Type] = append(byType[setting.Type], setting)
	}

	out := map[string]any{}
	for _, group := range byType {
		typ := group[0].Type

		switch {
		case specmodel.IsSingletonType(typ):
			body, err := renderSetting(group[0], index)
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			out[specmodel.SingletonBlockName(typ)] = body

		case specmodel.CollectionBlockName(typ) != "":
			keys := specmodel.DeriveSettingKeys(group)
			entries := map[string]any{}
			for _, setting := range group {
				body, err := renderSetting(setting, index)
				if err != nil {
					return nil, hperrors.Wrap(err)
				}
				entries[keys[setting.ID]] = body
			}
			out[specmodel.CollectionBlockName(typ)] = entries

		default:
			return nil, hperrors.Wrap(hperrors.ErrSpecSettingTypeUnclassified).
				WithParam("Type", string(typ))
		}
	}
	return out, nil
}

// renderSetting turns one setting into the body a spec writes for it.
//
// References are rewritten in two passes because they change shape as well as
// value. An in-scope reference becomes a path, which is still a string, so
// RemapRefs handles it on the typed value with its own self-check. An
// out-of-scope one becomes a nested block, which a string swap cannot produce -
// so that pass runs over the generic map, after marshaling, where changing the
// structure is possible.
func renderSetting(setting *entity.Setting, index *refIndex) (map[string]any, error) {
	data, err := setting.Parse()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if policy := entity.SpecPolicyFor(setting.Type); policy != nil {
		policy.Strip(data)
	}

	if err = entity.RemapRefs(data, index.paths); err != nil {
		return nil, hperrors.Wrap(err)
	}

	encoded, err := json.Marshal(data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	body := map[string]any{}
	if err = json.Unmarshal(encoded, &body); err != nil {
		return nil, hperrors.Wrap(err)
	}

	replaceExternalRefs(body, index)
	return body, nil
}

// replaceExternalRefs swaps any identifier the index could not resolve for the
// block that says what it was, so import can look it up or report it.
func replaceExternalRefs(node any, index *refIndex) {
	switch typed := node.(type) {
	case map[string]any:
		for key, value := range typed {
			if text, ok := value.(string); ok {
				if ref := index.external[text]; ref != nil {
					typed[key] = map[string]any{"external": ref}
					continue
				}
			}
			replaceExternalRefs(value, index)
		}
	case []any:
		for i, value := range typed {
			if text, ok := value.(string); ok {
				if ref := index.external[text]; ref != nil {
					typed[i] = map[string]any{"external": ref}
					continue
				}
			}
			replaceExternalRefs(value, index)
		}
	}
}
