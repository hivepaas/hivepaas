package settingmountservice

import (
	"cmp"
	"path"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// GatedPart reports whether a part of that name is gated in any source type.
// Import reads an entry before it knows its source's type, and the registry
// names gated parts the same way throughout, so the name is enough.
func GatedPart(name string) bool {
	for _, typ := range SourceTypes() {
		if part := PartOf(typ, name); part != nil && part.Gated {
			return true
		}
	}
	return false
}

// Grant is one gated part of one source that an entry hands to its app: what
// §7's gate is about.
type Grant struct {
	Source string `json:"source"`
	Part   string `json:"part"`
}

// Grants are the gated pairs an entry hands out, sorted, each once.
func Grants(mount *entity.AppSettingMount) []Grant {
	if mount == nil {
		return nil
	}
	var grants []Grant
	for _, f := range mount.Files {
		if f != nil && GatedPart(f.Part) {
			grants = append(grants, Grant{Source: mount.Source.ID, Part: f.Part})
		}
	}
	slices.SortFunc(grants, func(a, b Grant) int {
		return cmp.Or(strings.Compare(a.Source, b.Source), strings.Compare(a.Part, b.Part))
	})
	return slices.Compact(grants)
}

// Widens is what after hands out that before did not: the pairs that take the
// Reveal Secrets permission.
func Widens(before, after []Grant) []Grant {
	var added []Grant
	for _, grant := range after {
		if !slices.Contains(before, grant) {
			added = append(added, grant)
		}
	}
	return added
}

// CheckEntry refuses an entry wrong in itself, or for its source's type.
func CheckEntry(key string, mount *entity.AppSettingMount, sourceType base.SettingType) error {
	if !ValidEntryKey(key) {
		return hperrors.Wrap(hperrors.ErrSettingMountKeyInvalid).WithParam("Name", key)
	}
	if mount == nil || len(mount.Files) == 0 {
		return hperrors.Wrap(hperrors.ErrSettingMountNoFiles)
	}
	if !IsSourceType(sourceType) {
		return hperrors.Wrap(hperrors.ErrSettingMountSourceUnsupported).WithParam("Type", string(sourceType))
	}
	parts, paths := map[string]bool{}, map[string]bool{}
	for _, f := range mount.Files {
		if f == nil || PartOf(sourceType, f.Part) == nil || parts[f.Part] {
			part := ""
			if f != nil {
				part = f.Part
			}
			return hperrors.Wrap(hperrors.ErrSettingMountPartInvalid).
				WithParam("Part", part).WithParam("Type", string(sourceType))
		}
		if !ValidPath(f.Path) || paths[f.Path] {
			return hperrors.Wrap(hperrors.ErrSettingMountPathInvalid).WithParam("Path", f.Path)
		}
		parts[f.Part], paths[f.Path] = true, true
	}
	return nil
}

// CheckPathsFree refuses a path another file of the app already has. claimed is
// ClaimedPaths: a path, and what has it.
func CheckPathsFree(claimed map[string]string, paths ...string) error {
	for _, p := range paths {
		if owner, taken := claimed[path.Clean(p)]; taken {
			return hperrors.Wrap(hperrors.ErrSettingMountPathTaken).WithParam("Path", p).WithParam("Owner", owner)
		}
	}
	return nil
}
