package templatemodel

import (
	"regexp"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
)

const latestTag = "latest"

// imageTagPattern is docker's own grammar for a tag. What it keeps out is the
// point: no slash, no colon and no at sign, so nothing a caller sends as a tag
// can name a registry, a repository or a digest.
var imageTagPattern = regexp.MustCompile(`^[A-Za-z0-9_][A-Za-z0-9._-]{0,127}$`)

// ValidImageTag reports whether a string is a tag and only a tag.
func ValidImageTag(tag string) bool {
	return imageTagPattern.MatchString(tag)
}

// ImageOverrideClass says how far an image a user chose is from the one the
// template pinned. It decides what the dashboard warns about; it does not decide
// whether the override is allowed, which ClassifyImageOverride answers with an
// error instead.
type ImageOverrideClass string

const (
	// ImageOverrideSameLine is another release of the version line the template
	// declares. The template's configuration applies; HivePaaS has not tested this
	// exact release.
	ImageOverrideSameLine ImageOverrideClass = "same-line"
	// ImageOverrideOtherMajor is a release line the template says nothing about -
	// another major, or another minor for a template that declares major.minor
	// lines - so the overrides written for the declared lines may not apply.
	// PostgreSQL 18 moved its data directory, and a template that does not know
	// about a line cannot carry the mount change it needs.
	ImageOverrideOtherMajor ImageOverrideClass = "other-major"
	// ImageOverrideMoving is a tag that changes what it points at, so what runs
	// can change on a redeploy with nothing recorded to say it did.
	ImageOverrideMoving ImageOverrideClass = "moving"
)

var AllImageOverrideClasses = []ImageOverrideClass{
	ImageOverrideSameLine, ImageOverrideOtherMajor, ImageOverrideMoving,
}

// ClassifyImageOverride checks an image a user wants instead of the one a template
// version pins. line is that version's name - 18 for postgres, 11.8 for mariadb -
// which is how the template's author declared what the version covers.
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
func ClassifyImageOverride(line, templateImage, override string) (ImageOverrideClass, error) {
	templateRef, overrideRef := imageref.Parse(templateImage), imageref.Parse(override)
	switch {
	// The repositories are compared in canonical form, because the same one is
	// written several ways: a template says postgres and a person pasting from a
	// registry says registry-1.docker.io/library/postgres.
	case override == "" || !imageref.SameRepository(override, templateImage):
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
	case !inVersionLine(line, overrideRef.Tag):
		return ImageOverrideOtherMajor, nil
	default:
		return ImageOverrideSameLine, nil
	}
}

// inVersionLine reports whether a tag is a release within a declared line.
//
// The line comes from the template rather than from a rule about segments,
// because no such rule holds across projects: PostgreSQL's 18.6 is a patch of line
// 18, MariaDB's 11.8.9 is a patch of line 11.8, and a semver image's 2.2.0 is a
// minor release within line 2. The author already wrote down which one applies.
// The dot after the line keeps 11.80 out of 11.8.
func inVersionLine(line, tag string) bool {
	if line == "" {
		return false
	}
	version := tagVersion(tag)
	return version == line || strings.HasPrefix(version, line+".")
}

// tagVersion is the version part of a tag - everything before the first
// separator, without a leading v: 18.6 of 18.6-alpine3.24.
func tagVersion(tag string) string {
	version := tag
	if i := strings.IndexAny(tag, "-_+"); i >= 0 {
		version = tag[:i]
	}
	return strings.TrimPrefix(strings.TrimPrefix(version, "v"), "V")
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
// line is the template version's name, as for ClassifyImageOverride.
//
// A registry answers with thousands of tags: every release, every base image,
// every alias. Three rules cut that down. A tag that does not name a release is
// dropped, because it moves. A tag from another base family is dropped -
// 18.7-trixie is not a newer build of an app running alpine, however much larger
// the number is. What remains is ordered newest first, so the top of the list is
// the answer to the question the user asked by opening it.
//
// maxCandidates of 0 means all of them.
func SelectTagCandidates(line, templateImage string, tags []string, maxCandidates int) []TagCandidate {
	current := imageref.Parse(templateImage)
	family := tagFamily(current.Tag)

	seen := map[string]bool{}
	candidates := make([]TagCandidate, 0, len(tags))
	for _, tag := range tags {
		if tag == "" || tag == current.Tag || seen[tag] || tagFamily(tag) != family {
			continue
		}
		seen[tag] = true

		class, err := ClassifyImageOverride(line, templateImage, current.Repository+":"+tag)
		if err != nil || class == ImageOverrideMoving {
			continue
		}
		// Newer is ordered, not decided the way IsUpgrade decides: that function
		// applies a tag it cannot order, which is right for a release the updater
		// was told to install and wrong for a list somebody reads. 18.6-alpine3.23
		// against 18.6-alpine3.24 cannot be ordered - same release, different base -
		// and calling it newer would be a lie on screen.
		order, ordered := imageref.CompareTags(current.Tag, tag)
		candidates = append(candidates, TagCandidate{Tag: tag, Class: class, Newer: ordered && order < 0})
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
//
// A pgNN part is the exception and keeps its number. The Postgres extension
// images - pgvector, TimescaleDB, ParadeDB - name the server they are built for
// that way, and 0.8.6-pg18 is not a newer build of 0.8.6-pg17: it is the same
// extension on the next major of PostgreSQL, which will not open the data
// directory the app already has.
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
		if pgMajorPattern.MatchString(part) {
			families = append(families, part)
			continue
		}
		if trimmed := strings.TrimRight(part, "0123456789."); trimmed != "" {
			families = append(families, trimmed)
		}
	}
	return strings.Join(families, "-")
}

// pgMajorPattern matches the pgNN part of a tag such as 0.8.6-pg18-bookworm.
var pgMajorPattern = regexp.MustCompile(`^pg[0-9]+$`)
