package templatemodel

import (
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
)

const latestTag = "latest"

// ImageOverrideClass says how far an image a user chose is from the one the
// template pinned. It decides what the dashboard warns about; it does not decide
// whether the override is allowed, which ClassifyImageOverride answers with an
// error instead.
type ImageOverrideClass string

const (
	// ImageOverrideSameLine is another release of the major line the template
	// describes. The template's configuration applies; HivePaaS has not tested
	// this exact release.
	ImageOverrideSameLine ImageOverrideClass = "same-line"
	// ImageOverrideOtherMajor is a line the template says nothing about, so the
	// overrides written for the declared lines may not apply - PostgreSQL 18 moved
	// its data directory, and a template that does not know about a line cannot
	// carry the mount change it needs.
	ImageOverrideOtherMajor ImageOverrideClass = "other-major"
	// ImageOverrideMoving is a tag that changes what it points at, so what runs
	// can change on a redeploy with nothing recorded to say it did.
	ImageOverrideMoving ImageOverrideClass = "moving"
)

var AllImageOverrideClasses = []ImageOverrideClass{
	ImageOverrideSameLine, ImageOverrideOtherMajor, ImageOverrideMoving,
}

// ClassifyImageOverride checks an image a user wants instead of the template's.
//
// The repository has to be the same one: everything else the template says - the
// environment variables, the health check, the mount path, the app kind - is only
// correct for that software, so another repository is not an override but a
// different app wearing this template's configuration.
//
// Within that repository the rule is deliberately loose, because the user is
// accepting the risk for one app of theirs. What is refused is only what cannot be
// recorded: a reference with no tag - a digest on its own included - has no version
// to show in the app's binding, and latest names whatever is newest today.
func ClassifyImageOverride(templateImage, override string) (ImageOverrideClass, error) {
	templateRef, overrideRef := imageref.Parse(templateImage), imageref.Parse(override)
	switch {
	case override == "" || overrideRef.Repository != templateRef.Repository:
		return "", imageNotAllowed(override, templateRef.Repository,
			"the image must come from the same repository as the template's")
	case overrideRef.Tag == "":
		return "", imageNotAllowed(override, templateRef.Repository,
			"name a tag: a reference without one has no version to record")
	case overrideRef.Tag == latestTag:
		return "", imageNotAllowed(override, templateRef.Repository,
			"latest names whatever is newest at the time, which is not a version")
	case !isPinnedTag(overrideRef.Tag):
		return ImageOverrideMoving, nil
	}

	templateMajor, templateOK := imageref.MajorVersion(templateImage)
	overrideMajor, overrideOK := imageref.MajorVersion(override)
	if !templateOK || !overrideOK || templateMajor != overrideMajor {
		return ImageOverrideOtherMajor, nil
	}
	return ImageOverrideSameLine, nil
}

func imageNotAllowed(image, repository, reason string) error {
	return hperrors.Wrap(hperrors.ErrAppTemplateImageNotAllowed).
		WithParam("Image", image).WithParam("Repository", repository).
		WithExtraDetail("%s: %s", image, reason)
}

// TagCandidate is one tag a repository publishes that a user could move to.
type TagCandidate struct {
	Tag   string
	Class ImageOverrideClass
	Newer bool
}

// SelectTagCandidates picks, out of everything a repository has ever published,
// the tags a person would read as another build of the image in front of them.
//
// A registry answers with thousands of tags: every release, every base image,
// every alias. Three rules cut that down. A tag that does not name a release is
// dropped, because it moves. A tag from another base family is dropped -
// 18.7-trixie is not a newer build of an app running alpine, however much larger
// the number is. What remains is ordered newest first, so the top of the list is
// the answer to the question the user asked by opening it.
//
// maxCandidates of 0 means all of them.
func SelectTagCandidates(templateImage string, tags []string, maxCandidates int) []TagCandidate {
	current := imageref.Parse(templateImage)
	family := tagFamily(current.Tag)

	seen := map[string]bool{}
	candidates := make([]TagCandidate, 0, len(tags))
	for _, tag := range tags {
		if tag == "" || tag == current.Tag || seen[tag] || tagFamily(tag) != family {
			continue
		}
		seen[tag] = true

		candidate := current.Repository + ":" + tag
		class, err := ClassifyImageOverride(templateImage, candidate)
		if err != nil || class == ImageOverrideMoving {
			continue
		}
		newer, _ := imageref.IsUpgrade(templateImage, candidate)
		candidates = append(candidates, TagCandidate{Tag: tag, Class: class, Newer: newer})
	}

	slices.SortFunc(candidates, func(a, b TagCandidate) int {
		if order, ok := imageref.CompareTags(b.Tag, a.Tag); ok {
			return order
		}
		return strings.Compare(b.Tag, a.Tag)
	})
	if maxCandidates > 0 && len(candidates) > maxCandidates {
		candidates = candidates[:maxCandidates]
	}
	return candidates
}

// tagFamily is the non-numeric part of a tag's suffix: 18.6-alpine3.24 and
// 17.11-alpine3.22 are both "alpine", 11.8.9-noble is "noble", and 18.6 has no
// family at all. The numbers go because a base image moves on its own schedule -
// alpine3.24 succeeds alpine3.22 without the app's version changing.
func tagFamily(tag string) string {
	start := strings.IndexAny(tag, "-_+")
	if start < 0 {
		return ""
	}
	parts := strings.FieldsFunc(tag[start+1:], func(r rune) bool {
		return r == '-' || r == '_' || r == '+'
	})
	families := make([]string, 0, len(parts))
	for _, part := range parts {
		if trimmed := strings.TrimRight(part, "0123456789."); trimmed != "" {
			families = append(families, trimmed)
		}
	}
	return strings.Join(families, "-")
}
