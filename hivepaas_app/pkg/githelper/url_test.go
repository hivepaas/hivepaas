package githelper

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetCommitHttpsUrl(t *testing.T) {
	assert.Equal(t, "https://github.com/octocat/Hello-World/commit/abc1234",
		GetCommitHttpsUrl("https://github.com/octocat/Hello-World", "abc1234"))
	assert.Equal(t, "https://github.com/octocat/Hello-World/commit/abc1234",
		GetCommitHttpsUrl("github.com/octocat/Hello-World", "abc1234"))
}
