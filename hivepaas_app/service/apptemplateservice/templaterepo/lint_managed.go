package templaterepo

import (
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
)

// lintArtifactMounts keeps artifacts out from under the app's volume mounts. A
// volume mounted over an artifact's directory hides the file the engine bound
// there, and which of the two wins is Docker's mount order, not anything the
// template says. Every version is read with its override applied, since a
// version can add a mount.
func lintArtifactMounts(file *TemplateFile) []Problem {
	tmpl := file.Template
	if len(tmpl.Artifacts) == 0 {
		return nil
	}
	var problems []Problem
	check := func(where string, app map[string]any) {
		mounts, _ := dig(app, "deployment", "storage", "mounts").(map[string]any)
		for _, target := range slices.Sorted(maps.Keys(mounts)) {
			for _, artifact := range tmpl.Artifacts {
				if artifact == nil || !isUnder(artifact.MountPath, target) {
					continue
				}
				problems = append(problems, Problem{Path: file.Path, Message: fmt.Sprintf(
					"%sartifact %q at %s is under the mount at %s, which would hide it",
					where, artifact.Name, artifact.MountPath, target)})
			}
		}
	}
	check("", tmpl.App)
	for _, version := range tmpl.Versions {
		if version != nil && version.Override != nil && version.Override.App != nil {
			merged, _ := templaterender.MergePatch(tmpl.App, version.Override.App).(map[string]any)
			check(fmt.Sprintf("version %s: ", version.Name), merged)
		}
	}
	return slices.CompactFunc(problems, func(a, b Problem) bool { return a == b })
}

// isUnder reports whether p is target or inside it.
func isUnder(p, target string) bool {
	target = strings.TrimSuffix(target, "/")
	return p == target || strings.HasPrefix(p, target+"/")
}
