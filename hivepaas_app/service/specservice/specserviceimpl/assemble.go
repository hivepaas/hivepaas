package specserviceimpl

import (
	"encoding/json"
	"time"

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
func assembleSettings(
	settings []*entity.Setting,
	index *refIndex,
	mode specmodel.SecretsMode,
) (map[string]any, error) {
	byType := map[base.SettingType][]*entity.Setting{}
	for _, setting := range settings {
		byType[setting.Type] = append(byType[setting.Type], setting)
	}

	out := map[string]any{}
	for _, group := range byType {
		typ := group[0].Type

		switch {
		case specmodel.IsSingletonType(typ):
			body, err := renderSetting(group[0], index, mode)
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			body[specmodel.SettingMetaKey] = settingMeta(group[0])
			out[specmodel.SingletonBlockName(typ)] = body

		case specmodel.CollectionBlockName(typ) != "":
			keys := specmodel.DeriveSettingKeys(group)
			entries := map[string]any{}
			for _, setting := range group {
				body, err := renderSetting(setting, index, mode)
				if err != nil {
					return nil, hperrors.Wrap(err)
				}
				body[specmodel.CollectionEntryIDKey] = setting.ID
				body[specmodel.SettingMetaKey] = settingMeta(setting)
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

// settingMeta is a setting's row as a bundle carries it. It is a map rather than
// the struct so that it sits in the body like the data beside it, and so that
// only what is set is written.
func settingMeta(setting *entity.Setting) map[string]any {
	meta := map[string]any{"status": string(setting.Status), "version": setting.Version}
	for key, value := range map[string]string{
		"name": setting.Name, "kind": setting.Kind, "refId": setting.RefID,
	} {
		if value != "" {
			meta[key] = value
		}
	}
	if setting.Inheritable {
		meta["inheritable"] = true
	}
	if setting.Default {
		meta["default"] = true
	}
	if !setting.ExpireAt.IsZero() {
		meta["expireAt"] = setting.ExpireAt.UTC().Format(time.RFC3339)
	}
	if setting.Version == 0 {
		delete(meta, "version")
	}
	return meta
}

// renderSetting turns one setting into the body a spec writes for it.
//
// References are rewritten in two passes because they change shape as well as
// value. An in-scope reference becomes a path, which is still a string, so
// RemapRefs handles it on the typed value with its own self-check. An
// out-of-scope one becomes a nested block, which a string swap cannot produce -
// so that pass runs over the generic map, after marshaling, where changing the
// structure is possible.
func renderSetting(
	setting *entity.Setting,
	index *refIndex,
	mode specmodel.SecretsMode,
) (map[string]any, error) {
	data, err := setting.Parse()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if policy := entity.SpecPolicyFor(setting.Type); policy != nil {
		policy.Strip(data)
	}

	// In omit mode the secrets are cleared here rather than skipped. An
	// EncryptedField marshals as the ciphertext it was loaded with, so leaving
	// it alone would write a value only the exporting installation can read -
	// present, secret-looking, and silently useless anywhere else.
	if mode == specmodel.SecretsModeOmit {
		if _, err = entity.OmitSecrets(data); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	if err = entity.RemapRefs(data, index.paths); err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Collected before marshaling, because marshaling is what turns a secret
	// back into ciphertext. See entity.SecretPlaintexts.
	var plaintexts map[string]string
	if mode != specmodel.SecretsModeOmit {
		if plaintexts, err = entity.SecretPlaintexts(data); err != nil {
			return nil, hperrors.Wrap(err)
		}
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
	// Both secret-bearing modes substitute. Leaving the stored ciphertext in an
	// encrypted bundle would defeat the point: unwrapping the age envelope on
	// another installation would yield a value sealed with a data key that
	// installation does not have. age protects the bundle; the values inside it
	// are plain so that they travel.
	if mode.RevealsSecrets() {
		replaceStrings(body, plaintexts)
	}
	return body, nil
}

// replaceStrings swaps exact string values throughout a decoded document. It is
// how plaintext mode substitutes secrets, which cannot be marshaled in the
// clear.
func replaceStrings(node any, replacements map[string]string) {
	if len(replacements) == 0 {
		return
	}
	switch typed := node.(type) {
	case map[string]any:
		for key, value := range typed {
			if text, ok := value.(string); ok {
				if replacement, found := replacements[text]; found {
					typed[key] = replacement
					continue
				}
			}
			replaceStrings(value, replacements)
		}
	case []any:
		for i, value := range typed {
			if text, ok := value.(string); ok {
				if replacement, found := replacements[text]; found {
					typed[i] = replacement
					continue
				}
			}
			replaceStrings(value, replacements)
		}
	}
}

// externalRefKey is the key an external reference is written under.
const externalRefKey = "external"

// replaceExternalRefs swaps any identifier the index could not resolve for the
// block that says what it was, so import can look it up or report it.
func replaceExternalRefs(node any, index *refIndex) {
	switch typed := node.(type) {
	case map[string]any:
		for key, value := range typed {
			if text, ok := value.(string); ok {
				if ref := index.external[text]; ref != nil {
					typed[key] = map[string]any{externalRefKey: ref}
					continue
				}
			}
			replaceExternalRefs(value, index)
		}
	case []any:
		for i, value := range typed {
			if text, ok := value.(string); ok {
				if ref := index.external[text]; ref != nil {
					typed[i] = map[string]any{externalRefKey: ref}
					continue
				}
			}
			replaceExternalRefs(value, index)
		}
	}
}
