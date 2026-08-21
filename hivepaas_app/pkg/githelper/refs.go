package githelper

import (
	"strings"
)

const (
	refHeadsPrefix         = "refs/heads/"
	refTagsPrefix          = "refs/tags/"
	refPullPrefix          = "refs/pull/"
	refMergeRequestsPrefix = "refs/merge-requests/"
)

type ReferenceName string

func (r ReferenceName) String() string {
	return string(r)
}

func (r ReferenceName) Short() string {
	s := string(r)
	if after, ok := strings.CutPrefix(s, refHeadsPrefix); ok {
		return after
	}
	if after, ok := strings.CutPrefix(s, refTagsPrefix); ok {
		return after
	}
	return s
}

func NewBranchReferenceName(name string) ReferenceName {
	return ReferenceName(refHeadsPrefix + name)
}

func NewTagReferenceName(name string) ReferenceName {
	return ReferenceName(refTagsPrefix + name)
}

type RefType string

const (
	RefBranch RefType = "branch"
	RefTag    RefType = "tag"
	RefPull   RefType = "pull"
)

func (rt RefType) IsBranch() bool {
	return rt == RefBranch
}

func (rt RefType) IsTag() bool {
	return rt == RefTag
}

func (rt RefType) IsPull() bool {
	return rt == RefPull
}

func (rt RefType) CanCheckout() bool {
	return rt == RefBranch || rt == RefTag || rt == RefPull
}

func NormalizeRepoRef(ref string) ReferenceName {
	if ref == "" || ref == "HEAD" { //nolint:goconst
		return "HEAD"
	}
	ref, _ = strings.CutPrefix(ref, "refs/")

	// Heads ref (branch)
	if after, ok := strings.CutPrefix(ref, "heads/"); ok {
		return NewBranchReferenceName(after)
	}

	// Tags ref
	if after, ok := strings.CutPrefix(ref, "tags/"); ok {
		return NewTagReferenceName(after)
	}

	// Pull ref (github, gitea)
	if after, ok := strings.CutPrefix(ref, "pull/"); ok {
		ref = after
		ref, _ = strings.CutSuffix(ref, "/head")
		return ReferenceName(refPullPrefix + ref + "/head")
	}

	// Merge request ref (gitlab)
	if after, ok := strings.CutPrefix(ref, "merge-requests/"); ok {
		ref = after
		ref, _ = strings.CutSuffix(ref, "/head")
		return ReferenceName(refMergeRequestsPrefix + ref + "/head")
	}

	// Branch
	return NewBranchReferenceName(ref)
}

func GetRefType(ref string) RefType {
	if strings.HasPrefix(ref, refHeadsPrefix) {
		return RefBranch
	}
	if strings.HasPrefix(ref, refTagsPrefix) {
		return RefTag
	}
	if strings.HasPrefix(ref, refPullPrefix) || strings.HasPrefix(ref, refMergeRequestsPrefix) {
		return RefPull
	}
	return ""
}

func GetRefShort(ref string) (RefType, string) {
	refType := GetRefType(ref)
	if refType == RefBranch || refType == RefTag {
		return refType, ReferenceName(ref).Short()
	}
	return refType, ref
}
