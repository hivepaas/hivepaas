package hpappserviceimpl

import (
	"regexp"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

var (
	// templatesRepoPattern is a GitHub owner/name. The owner rule is GitHub's; a
	// name of "." or ".." is refused separately below, since either would change
	// the path of the URL the repository is read from.
	templatesRepoPattern   = regexp.MustCompile(`^[A-Za-z0-9](?:[A-Za-z0-9-]{0,38})/[A-Za-z0-9._-]{1,100}$`)
	templatesCommitPattern = regexp.MustCompile(`^[0-9a-f]{40}$`)
	sha256HexPattern       = regexp.MustCompile(`^[0-9a-f]{64}$`)
)

// validateTemplatesRef checks the form of a templates pin. Every field ends up in
// a URL or is compared against a hash, so each has exactly one accepted shape:
// a short sha or a branch name as Commit would stop pinning anything.
func validateTemplatesRef(ref *base.TemplatesRef) error {
	invalid := func(field string) error {
		return hperrors.Wrap(hperrors.ErrReleaseInfoInvalid).WithExtraDetail("templates.%s is malformed", field)
	}
	if !templatesRepoPattern.MatchString(ref.Repo) || isDotName(ref.Repo) {
		return invalid("repo")
	}
	if !templatesCommitPattern.MatchString(ref.Commit) {
		return invalid("commit")
	}
	if !sha256HexPattern.MatchString(ref.IndexSHA256) {
		return invalid("indexSha256")
	}
	return nil
}

func isDotName(repo string) bool {
	_, name, _ := strings.Cut(repo, "/")
	return name == "." || name == ".."
}
