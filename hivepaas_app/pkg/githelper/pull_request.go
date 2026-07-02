package githelper

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
)

func NormalizePullRef(ref string) (pullRef string, pullNumber uint64, err error) {
	var pullNumStr string
	ref, _ = strings.CutPrefix(ref, "refs/")

	for {
		// Pull ref (github, gitea)
		if refStr, ok := strings.CutPrefix(ref, "pull/"); ok {
			pullNumStr, _ = strings.CutSuffix(refStr, "/head")
			pullRef = refPullPrefix + pullNumStr + "/head"
			break
		}

		// Merge request ref (gitlab)
		if refStr, ok := strings.CutPrefix(ref, "merge-requests/"); ok {
			pullNumStr, _ = strings.CutSuffix(refStr, "/head")
			pullRef = refMergeRequestsPrefix + pullNumStr + "/head"
			break
		}

		pullNumStr = ref
		break //nolint
	}

	pullNumber, err = strconv.ParseUint(pullNumStr, 10, 64)
	if err != nil || pullNumber == 0 {
		return "", 0, apperrors.New(apperrors.ErrPullRequestInvalid).
			WithParam("PullRequest", ref)
	}

	if pullRef == "" {
		pullRef = fmt.Sprintf("refs/pull/%d/head", pullNumber)
	}

	return pullRef, pullNumber, nil
}

func GetPullNumberAsStr(ref string) (string, error) {
	after, ok := strings.CutPrefix(ref, refPullPrefix)
	if !ok {
		after, ok = strings.CutPrefix(ref, refMergeRequestsPrefix)
	}
	if !ok {
		return "", apperrors.New(apperrors.ErrPullRequestInvalid).WithParam("PullRequest", ref)
	}
	return strings.TrimSuffix(after, "/head"), nil
}

func GetPullNumber(ref string) (uint64, error) {
	pullNumberStr, err := GetPullNumberAsStr(ref)
	if err != nil {
		return 0, apperrors.New(apperrors.ErrPullRequestInvalid).WithParam("PullRequest", ref)
	}
	number, err := strconv.ParseUint(pullNumberStr, 10, 64)
	if err != nil {
		return 0, apperrors.New(apperrors.ErrPullRequestInvalid).WithParam("PullRequest", ref)
	}
	return number, nil
}

func GetPullNumberRef(prNumber int64, gitSource base.GitSource) (string, error) {
	switch gitSource {
	case base.GitSourceGithub:
		return "refs/pull/" + strconv.FormatInt(prNumber, 10) + "/head", nil
	case base.GitSourceGitlab:
		return "refs/merge-requests/" + strconv.FormatInt(prNumber, 10) + "/head", nil
	case base.GitSourceGitea, base.GitSourceGogs:
		return "refs/pull/" + strconv.FormatInt(prNumber, 10) + "/head", nil
	case base.GitSourceBitbucket:
		fallthrough
	default:
		return "", apperrors.New(apperrors.ErrGitTypeUnsupported).WithParam("Type", gitSource)
	}
}
