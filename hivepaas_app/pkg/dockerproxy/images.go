package dockerproxy

import (
	"path"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
)

// matchImage reports whether ref is an image one of the patterns allows.
//
// A pattern is a repository, optionally with a tag, as a person writes it in a
// template: "autobase/automation:2.11.0", "openruntimes/*", "*". Repositories are
// compared after Docker's own normalization, so "alpine" and
// "docker.io/library/alpine" are the same image. A pattern without a tag allows
// every tag and digest of its repository, and "*" allows everything.
func matchImage(patterns []string, ref string) bool {
	if ref == "" {
		return false
	}
	image := imageref.Parse(ref)
	repository := imageref.NormalizeRepository(image.Repository)
	tag := image.Tag
	if tag == "" && image.Digest == "" {
		tag = "latest"
	}
	for _, pattern := range patterns {
		if pattern == "*" {
			return true
		}
		want := imageref.Parse(pattern)
		if ok, _ := path.Match(imageref.NormalizeRepository(want.Repository), repository); !ok {
			continue
		}
		if want.Tag == "" {
			return true
		}
		if ok, _ := path.Match(want.Tag, tag); ok && tag != "" {
			return true
		}
	}
	return false
}
