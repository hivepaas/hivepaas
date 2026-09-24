package specserviceimpl

import (
	"errors"
	"maps"
	"slices"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// refKind says what a reference in a bundle names, and so how it is resolved.
type refKind int

const (
	// refPath is a path into the bundle, which export wrote for a setting it
	// held.
	refPath refKind = iota
	// refExternal is an external block, which export wrote for a setting it did
	// not hold.
	refExternal
	// refRawID is an id export could not resolve, because the setting was gone.
	refRawID
	// refAppID is an app's id: export writes app references as they are stored.
	refAppID
	// refSourceApp is the key of an app in the same env, whose directory a mount
	// reaches.
	refSourceApp
)

// What holds a reference, as an issue's detail names it.
const (
	refInSetting = "setting"
	refInMount   = "mount"
)

// bundleRef is one reference a bundle object makes.
type bundleRef struct {
	kind     refKind
	value    string
	external *specmodel.ExternalRef
	// in and holder name what holds the reference, for an issue's detail:
	// "setting" or "mount", and which one.
	in, holder string
}

// externalRefPlaceholder stands in for an external block while a body is read
// as its type: the block is where a string id was, and a string has to be there
// for the type to parse.
const externalRefPlaceholder = "hivepaas-spec-external:"

// settingRefs reads the references one setting body makes. The body is parsed
// as its type and asked, so only the strings the type calls references are
// followed - a config file's content that happens to look like a path is not
// one. A body from a newer HivePaaS has none here: SETTING_VERSION_NEWER already
// blocks the import.
func settingRefs(typ base.SettingType, at, holder string, body any) ([]bundleRef, error) {
	var externals []*specmodel.ExternalRef
	data, err := readBundleSetting(typ, at, holder, withExternalPlaceholders(body, &externals))
	if err != nil || data == nil {
		return nil, err
	}
	ids := data.GetRefObjectIDs()
	if ids == nil {
		return nil, nil
	}

	var refs []bundleRef
	seen := map[string]bool{}
	add := func(ref bundleRef) {
		id := strconv.Itoa(int(ref.kind)) + ":" + ref.value
		if seen[id] {
			return
		}
		seen[id] = true
		ref.in, ref.holder = refInSetting, holder
		refs = append(refs, ref)
	}
	for _, id := range ids.RefSettingIDs {
		switch {
		case id == "":
		case strings.HasPrefix(id, externalRefPlaceholder):
			if n, ok := placeholderIndex(id, len(externals)); ok {
				add(bundleRef{kind: refExternal, value: id, external: externals[n]})
			}
		default:
			if _, ok := parseRefPath(id); ok {
				add(bundleRef{kind: refPath, value: id})
			} else {
				add(bundleRef{kind: refRawID, value: id})
			}
		}
	}
	for _, id := range ids.RefAppIDs {
		if id != "" {
			add(bundleRef{kind: refAppID, value: id})
		}
	}
	return refs, nil
}

// readBundleSetting reads one setting body of the bundle as its type. It reads
// nothing from a newer HivePaaS - SETTING_VERSION_NEWER already blocks the
// import - and refuses a body that is not its type. External references have to
// be out of the way first: withExternalPlaceholders.
func readBundleSetting(typ base.SettingType, at, holder string, body any) (entity.SettingData, error) {
	_, data, err := decodeImportedSetting(specmodel.Block(holder), typ, holder, body)
	if errors.Is(err, hperrors.ErrDataVerNewerThanSystemVer) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(hperrors.ErrSpecBundleInvalid).WithCause(err).
			WithExtraDetail("%s: %s does not read as a %s setting", at, holder, typ)
	}
	return data, nil
}

// placeholderIndex is the external block a placeholder stands for, by its
// position among count collected.
func placeholderIndex(value string, count int) (int, bool) {
	number, found := strings.CutPrefix(value, externalRefPlaceholder)
	if !found {
		return 0, false
	}
	n, err := strconv.Atoi(number)
	return n, err == nil && n >= 0 && n < count
}

// withExternalPlaceholders copies a body with every external block swapped for
// a placeholder, and collects the blocks in the order the placeholders number
// them.
func withExternalPlaceholders(node any, externals *[]*specmodel.ExternalRef) any {
	switch typed := node.(type) {
	case map[string]any:
		if ref := asExternalRef(typed); ref != nil {
			*externals = append(*externals, ref)
			return externalRefPlaceholder + strconv.Itoa(len(*externals)-1)
		}
		out := make(map[string]any, len(typed))
		for key, value := range typed {
			out[key] = withExternalPlaceholders(value, externals)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, value := range typed {
			out[i] = withExternalPlaceholders(value, externals)
		}
		return out
	}
	return node
}

// asExternalRef reads an external block: a mapping holding only `external`,
// which names a type.
func asExternalRef(fields map[string]any) *specmodel.ExternalRef {
	if len(fields) != 1 {
		return nil
	}
	inner, ok := fields[externalRefKey].(map[string]any)
	if !ok {
		return nil
	}
	text := func(key string) string {
		value, _ := inner[key].(string)
		return value
	}
	if text("type") == "" {
		return nil
	}
	return &specmodel.ExternalRef{Type: text("type"), Name: text("name"), Kind: text("kind"), ID: text("id")}
}

// mountRefs reads the references an app's managed mounts make: the volume each
// reaches, and the app whose directory it is when it is not this one's.
func mountRefs(storage *specmodel.Storage) []bundleRef {
	if storage == nil {
		return nil
	}
	var refs []bundleRef
	for _, target := range slices.Sorted(maps.Keys(storage.Mounts)) {
		m := storage.Mounts[target]
		switch {
		case m.External != nil:
			external := *m.External
			refs = append(refs, bundleRef{kind: refExternal, value: external.Name, external: &external})
		case m.Source != "":
			kind := refRawID
			if _, ok := parseRefPath(m.Source); ok {
				kind = refPath
			}
			refs = append(refs, bundleRef{kind: kind, value: m.Source})
		}
		if m.SourceApp != nil && m.SourceApp.App != "" {
			refs = append(refs, bundleRef{kind: refSourceApp, value: m.SourceApp.App})
		}
		for i := range refs {
			if refs[i].in == "" {
				refs[i].in, refs[i].holder = refInMount, target
			}
		}
	}
	return refs
}

// refTarget is where a path in a bundle leads: the node that holds the setting,
// and the setting's block and key among that node's settings.
type refTarget struct {
	node                string
	project, env, app   string
	block, key          string
	settingType         base.SettingType
	isCollectionSetting bool
}

// parseRefPath reads a path export wrote for a setting it held: the scope's path
// - global, a project, an env or an app - then the block, then for a
// collection the entry's key, which may itself hold a slash.
func parseRefPath(path string) (refTarget, bool) {
	segments := strings.Split(path, "/")
	var t refTarget
	var rest []string
	switch {
	case len(segments) >= 2 && segments[0] == globalFilenameStem:
		t.node, rest = globalFilenameStem, segments[1:]
	case len(segments) >= 3 && segments[0] == projectsSegment:
		t.project, rest = segments[1], segments[2:]
		scope := projectsSegment + "/" + t.project
		t.node = scope + "/settings"
		if len(rest) >= 3 && rest[0] == envsSegment {
			t.env, rest = rest[1], rest[2:]
			scope += "/envs/" + t.env
			t.node = scope + "/settings"
			if len(rest) >= 3 && rest[0] == "apps" {
				t.app, rest = rest[1], rest[2:]
				t.node = scope + "/apps/" + t.app
			}
		}
	default:
		return refTarget{}, false
	}

	t.block = rest[0]
	if typ, ok := specmodel.SingletonTypeOf(t.block); ok {
		t.settingType = typ
		return t, len(rest) == 1
	}
	typ, ok := specmodel.CollectionTypeOf(t.block)
	if !ok || len(rest) < 2 {
		return refTarget{}, false
	}
	t.settingType, t.isCollectionSetting = typ, true
	t.key = strings.Join(rest[1:], "/")
	return t, t.key != ""
}
