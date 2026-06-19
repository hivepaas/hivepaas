package githelper

import (
	"strings"

	"github.com/go-git/go-git/v5/plumbing"
)

func NormalizeRepoRef(ref string) plumbing.ReferenceName {
	if ref == "" || ref == "HEAD" { //nolint:goconst
		return "HEAD"
	}
	if strings.HasPrefix(ref, "refs/") {
		return plumbing.ReferenceName(ref)
	}

	// Tags ref
	if after, ok := strings.CutPrefix(ref, "tags/"); ok {
		ref = after
		return plumbing.NewTagReferenceName(ref)
	}

	// Heads ref
	if after, ok := strings.CutPrefix(ref, "heads/"); ok {
		ref = after
		return plumbing.NewBranchReferenceName(ref)
	}

	return plumbing.NewBranchReferenceName(ref)
}

func IsCommitHash(hash string) bool {
	if len(hash) != 40 && len(hash) != 64 { // SHA1: 40, SHA256: 64
		return false
	}
	for _, c := range hash {
		if !((c >= '0' && c <= '9') || (c >= 'a' && c <= 'f') || (c >= 'A' && c <= 'F')) { //nolint
			return false
		}
	}
	return true
}
