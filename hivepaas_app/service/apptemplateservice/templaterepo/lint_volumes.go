package templaterepo

import (
	"fmt"
	"maps"
	"regexp"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
)

// volumeParamSource is a mount source that is a volume parameter, alone.
var volumeParamSource = regexp.MustCompile(`^\$\{\{\s*params\.([A-Za-z0-9_]+)\s*\}\}$`)

// volumeMountsExempt are the templates published before lintVolumeMounts, whose
// layout stays as it is: apps created from them keep their data where it is, and
// an app created again on the same directory looks for it there. jellyfin's media
// and qbittorrent's downloads are mounted by the *arr templates, through
// sourceApp, at the paths they have now. Given one volume for both, these apps
// mix their files without failing; nothing new joins this list.
var volumeMountsExempt = []string{
	"calibre-web", "filebrowser", "jellyfin", "komga", "navidrome", "qbittorrent", "unmanic", //nolint:misspell // names
}

// lintVolumeMounts keeps an app's mounts of volume parameters out of each other.
//
// A volume gives an app one directory, and nothing stops two volume parameters
// from naming the same volume - most installations have one, called default. So
// every mount of a volume parameter lands in the same directory unless a subpath
// keeps it apart: a mount of the whole directory holds what the others write,
// and two mounts with one subpath share it. RomM, given one volume for its
// library and its data, took its own data folders for game platforms and would
// not start. One mount of a volume parameter, or mounts each with a subpath of
// their own, cannot meet.
func lintVolumeMounts(file *TemplateFile) []Problem {
	tmpl := file.Template
	if slices.Contains(volumeMountsExempt, tmpl.Metadata.Name) {
		return nil
	}
	volumes := map[string]bool{}
	for _, param := range tmpl.Parameters {
		if param != nil && param.Type == templatemodel.ParamTypeVolume {
			volumes[param.Name] = true
		}
	}
	if len(volumes) == 0 {
		return nil
	}

	var problems []Problem
	check := func(where string, app map[string]any) {
		for _, message := range overlappingMounts(app, volumes) {
			problems = append(problems, Problem{Path: file.Path, Message: where + message})
		}
	}
	apps := map[string]map[string]any{"": tmpl.App}
	for _, component := range tmpl.Components {
		if component != nil {
			apps["component "+component.Name+": "] = component.App
		}
	}
	for _, where := range slices.Sorted(maps.Keys(apps)) {
		check(where, apps[where])
		for _, version := range tmpl.Versions {
			if version != nil && version.Override != nil && version.Override.App != nil {
				merged, _ := templaterender.MergePatch(apps[where], version.Override.App).(map[string]any)
				check(fmt.Sprintf("%sversion %s: ", where, version.Name), merged)
			}
		}
	}
	return slices.CompactFunc(problems, func(a, b Problem) bool { return a == b })
}

// overlappingMounts describes the mounts of volume parameters in one app that
// could land on each other.
func overlappingMounts(app map[string]any, volumes map[string]bool) []string {
	mounts, _ := dig(app, "deployment", "storage", "mounts").(map[string]any)
	type mount struct{ target, subpath string }
	var ofVolumes []mount
	for _, target := range slices.Sorted(maps.Keys(mounts)) {
		spec, _ := mounts[target].(map[string]any)
		source, _ := spec["source"].(string)
		match := volumeParamSource.FindStringSubmatch(source)
		// Another app's volume, through sourceApp, is that app's directory: it is
		// shared on purpose.
		if spec["type"] != "volume" || match == nil || !volumes[match[1]] || spec["sourceApp"] != nil {
			continue
		}
		subpath, _ := dig(spec, "volumeOptions", "subpath").(string)
		ofVolumes = append(ofVolumes, mount{target: target, subpath: subpath})
	}
	if len(ofVolumes) < 2 { //nolint:mnd
		return nil
	}

	var out []string
	seen := map[string]string{}
	for _, m := range ofVolumes {
		if m.subpath == "" {
			out = append(out, fmt.Sprintf("%s mounts a volume parameter whole, and the app mounts others: "+
				"a volume parameter may name the same volume as another, so give it a subpath of its own", m.target))
			continue
		}
		if other, taken := seen[m.subpath]; taken {
			out = append(out, fmt.Sprintf("%s and %s both mount subpath %q of a volume parameter: "+
				"on one volume they would share it", other, m.target, m.subpath))
			continue
		}
		seen[m.subpath] = m.target
	}
	return out
}

func dig(tree map[string]any, keys ...string) any {
	var node any = tree
	for _, key := range keys {
		next, ok := node.(map[string]any)
		if !ok {
			return nil
		}
		node = next[key]
	}
	return node
}
