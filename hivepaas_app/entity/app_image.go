package entity

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

const (
	// repoSegmentMaxLen is what one of the two parts of a repository name may
	// hold. A project and an app may each be named with 100 characters, and
	// <address>/<username>/ still goes in front of the result.
	repoSegmentMaxLen = 50
	// envSegmentMaxLen leaves room in a 128-character tag for a commit, or for a
	// custom tag of the length base.ImageCustomTagMaxLen allows.
	envSegmentMaxLen = 20
	// truncatedSuffixLen is how much of the global key's digest is appended when
	// a segment was cut, so that two long names that share a prefix stay apart.
	truncatedSuffixLen = 6
	commitHashLen      = 7
)

// ImageRepoName is the repository a build is pushed to: one per app, shared by
// the app's environments, unique across the installation.
//
// It needs Project loaded. Both build paths load it - the deployment helper and
// the agent's own loader - and an app without it is a programming error rather
// than a name to guess at, so it is an error and not a shorter name.
func (app *App) ImageRepoName() (string, error) {
	if app.Project == nil || app.Project.Key == "" {
		return "", hperrors.Wrap(hperrors.ErrMissing).
			WithMsgLog("app %v has no project loaded, so its image has no name", app.ID)
	}

	project, projectCut := cutImageSegment(app.Project.Key, repoSegmentMaxLen)
	appPart, appCut := cutImageSegment(app.Key, repoSegmentMaxLen)

	name := normalizeImagePart(project + "-" + appPart)
	if projectCut || appCut {
		name += "-" + imageDigestOf(app.GlobalKey)
	}
	return name, nil
}

// ImageTagPrefix is the environment part every tag of this app carries, the
// commit's and any the deployment added.
func (app *App) ImageTagPrefix() (string, error) {
	if app.ProjectEnv == nil || app.ProjectEnv.Key == "" {
		return "", hperrors.Wrap(hperrors.ErrMissing).
			WithMsgLog("app %v has no environment loaded, so its image has no tag", app.ID)
	}

	env, _ := cutImageSegment(app.ProjectEnv.Key, envSegmentMaxLen)
	return normalizeImagePart(env), nil
}

// ImageTag is what one build is tagged with.
func (app *App) ImageTag(commitHash string) (string, error) {
	if len(commitHash) < commitHashLen {
		return "", hperrors.Wrap(hperrors.ErrValueInvalid).
			WithMsgLog("commit hash %q is too short to tag an image with", commitHash)
	}

	prefix, err := app.ImageTagPrefix()
	if err != nil {
		return "", err
	}
	return prefix + "-" + commitHash[:commitHashLen], nil
}

// ImageReference is what a build pushes: the repository under the account of the
// registry it is being pushed to, and one tag.
func ImageReference(address, username, repoName, tag string) string {
	return fmt.Sprintf("%s/%s/%s:%s", address, username, repoName, tag)
}

// cutImageSegment shortens one part of a name and says whether it had to.
func cutImageSegment(value string, limit int) (string, bool) {
	if len(value) <= limit {
		return value, false
	}
	return value[:limit], true
}

// normalizeImagePart makes a joined name or tag legal: the OCI grammar takes a
// separator between alphanumerics and nothing else, so a doubled or dangling one
// has to go. Keys arrive slugged, and this is what handles the ones that were
// slugged under older rules.
func normalizeImagePart(value string) string {
	replacer := strings.NewReplacer("__", "_", "--", "-")
	for strings.Contains(value, "__") || strings.Contains(value, "--") {
		value = replacer.Replace(value)
	}
	return strings.Trim(value, "-_.")
}

func imageDigestOf(globalKey string) string {
	sum := sha256.Sum256([]byte(globalKey))
	return hex.EncodeToString(sum[:])[:truncatedSuffixLen]
}
