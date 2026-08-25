package githelper

import (
	"fmt"
	"strings"
)

func GetCommitHttpsUrl(repoURL, commitHash string) string {
	if !strings.HasPrefix(repoURL, "https://") {
		repoURL = "https://" + repoURL
	}
	return fmt.Sprintf("%v/commit/%v", repoURL, commitHash)
}
