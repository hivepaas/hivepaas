// Package imageref compares container image references.
//
// The system updater uses it to answer one question: is the image a release
// wants us to run newer than the one the service is running now? Applying an
// image that is not newer is not free - every swarm service update restarts the
// task, and for the database it also re-runs the migrations - so "the release
// named an image" is not on its own a reason to act.
//
// Image tags are not semantic versions and this does not pretend otherwise.
// `8.6-alpine`, `v3.7` and `v1.52.0` all appear in this repo's own release file.
// The comparison answers only where it is confident, and says so when it is not;
// the caller decides what an unanswerable comparison means.
package imageref

import (
	"strconv"
	"strings"
)

// Ref is an image reference split into the parts that matter here.
type Ref struct {
	// Repository is everything before the tag, registry host included.
	Repository string
	// Tag is empty when the reference carried none.
	Tag string
	// Digest is empty unless the reference was pinned. Docker Desktop pins every
	// image it deploys, so a running service's image usually carries one.
	Digest string
}

// Parse splits a reference of the form repository[:tag][@digest].
//
// It does not validate: a reference this cannot make sense of comes back with
// whatever could be read, and the comparison below reports that it cannot decide
// rather than guessing.
func Parse(ref string) Ref {
	out := Ref{}

	if name, digest, found := strings.Cut(ref, "@"); found {
		out.Digest = digest
		ref = name
	}

	// The last colon is only a tag separator when it comes after the last slash.
	// Without that check, the port in `registry.example:5000/redis` reads as a tag.
	colon := strings.LastIndex(ref, ":")
	if colon > strings.LastIndex(ref, "/") {
		out.Repository = ref[:colon]
		out.Tag = ref[colon+1:]
		return out
	}

	out.Repository = ref
	return out
}

// IsUpgrade says whether target should replace current, and why.
//
// The reason is meant to be read by whoever is watching the update, so it is
// returned for both answers: a step that decides to do nothing should be able to
// say what it decided.
func IsUpgrade(current, target string) (bool, string) {
	if target == "" {
		return false, "no target image"
	}
	if current == "" {
		return true, "nothing is running yet, applying " + target
	}

	cur, tgt := Parse(current), Parse(target)

	// A different repository is a decision the release made - moving to a fork, a
	// mirror, or a different distribution of the same thing. There is no version
	// line shared between the two to compare along.
	if cur.Repository != tgt.Repository {
		return true, "image changed from " + cur.Repository + " to " + tgt.Repository
	}
	if cur.Tag == tgt.Tag {
		return false, "already at " + tgt.Repository + ":" + tgt.Tag
	}

	cmp, ok := compareTags(cur.Tag, tgt.Tag)
	if !ok {
		return true, "cannot order tag " + tgt.Tag + " against " + cur.Tag + ", applying it"
	}
	if cmp >= 0 {
		return false, "target " + tgt.Tag + " is not newer than " + cur.Tag
	}
	return true, "upgrading " + tgt.Repository + " from " + cur.Tag + " to " + tgt.Tag
}

// compareTags orders two tags, reporting whether it could.
//
// A tag is read as an optional `v`, dot-separated numbers, and an optional
// suffix: `v1.52.0`, `8.6-alpine`, `18.3`. Anything else is not ordered.
func compareTags(a, b string) (int, bool) {
	aNums, aSuffix, aOK := splitTag(a)
	bNums, bSuffix, bOK := splitTag(b)
	if !aOK || !bOK {
		return 0, false
	}

	for i := 0; i < len(aNums) || i < len(bNums); i++ {
		x, y := segment(aNums, i), segment(bNums, i)
		if x != y {
			if x < y {
				return -1, true
			}
			return 1, true
		}
	}

	if aSuffix == bSuffix {
		return 0, true
	}
	// Same numbers, different suffix. `8.6-alpine` against `8.6` is a change of
	// variant, not of version, and `1.0.0-beta1` against `1.0.0` is a release
	// leaving prerelease - opposite directions with the same shape. Neither is
	// worth guessing at.
	return 0, false
}

// splitTag reads `v1.52.0` as ([1 52 0], "") and `8.6-alpine` as ([8 6], "-alpine").
func splitTag(tag string) (nums []int, suffix string, ok bool) {
	if tag == "" {
		return nil, "", false
	}
	rest := strings.TrimPrefix(strings.TrimPrefix(tag, "v"), "V")

	if i := strings.IndexAny(rest, "-_+"); i >= 0 {
		suffix = rest[i:]
		rest = rest[:i]
	}
	if rest == "" {
		return nil, "", false
	}

	for _, part := range strings.Split(rest, ".") {
		n, err := strconv.Atoi(part)
		if err != nil || n < 0 {
			return nil, "", false
		}
		nums = append(nums, n)
	}
	return nums, suffix, true
}

// segment reads past the end as zero, so 1.2 and 1.2.0 compare equal.
func segment(nums []int, i int) int {
	if i < len(nums) {
		return nums[i]
	}
	return 0
}

// MajorVersion reads the leading number of a reference's tag.
//
// It answers one question the updater has to ask before it swaps a database
// image: is this the same major line? Postgres will not start on a data
// directory written by a different major version, so that is a migration, not an
// image swap. Reports false when the tag carries no version it can read.
func MajorVersion(ref string) (int, bool) {
	nums, _, ok := splitTag(Parse(ref).Tag)
	if !ok || len(nums) == 0 {
		return 0, false
	}
	return nums[0], true
}
