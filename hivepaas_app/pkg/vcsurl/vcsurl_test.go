package vcsurl_test

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/vcsurl"
)

func TestParse_GitHub(t *testing.T) {
	urls := []string{
		"github.com/foo/bar",
		"http://github.com/foo/bar",
		"https://github.com/foo/bar",
		"https://github.com/foo/bar.git",
		"https://api.github.com/repos/foo/bar",
		"git@github.com:foo/bar",
		"git@github.com:foo/bar.git",
		"git+ssh://github.com/foo/bar",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			vcs, err := vcsurl.Parse(url)
			assert.NoError(t, err)
			AssertVCS_GitHub(t, vcs)
		})
	}
}

func TestParse_GitHubCommittish(t *testing.T) {
	urls := []string{
		"https://github.com/foo/bar/commit/qux",
		"https://api.github.com/repos/foo/bar/commits/qux",
		"https://api.github.com/repos/foo/bar/branches/qux",
		"https://github.com/foo/bar/tree/qux",
		"https://github.com/foo/bar/releases/tag/qux",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			vcs, err := vcsurl.Parse(url)
			assert.NoError(t, err)
			assert.Equal(t, "qux", vcs.Committish)
			AssertVCS_GitHub(t, vcs)
		})
	}
}

func TestParse_GitHubCommittishSlash(t *testing.T) {
	vcs, err := vcsurl.Parse("https://github.com/foo/bar/tree/qux/baz")
	assert.NoError(t, err)
	assert.Equal(t, "qux/baz", vcs.Committish)
	AssertVCS_GitHub(t, vcs)
}

func AssertVCS_GitHub(t *testing.T, vcs *vcsurl.VCS) {
	t.Helper()
	assert.Equal(t, vcsurl.Git, vcs.Kind)
	assert.Equal(t, vcsurl.GitHub, vcs.Host)
	assert.Equal(t, "foo", vcs.Username)
	assert.Equal(t, "bar", vcs.Name)
	assert.Equal(t, "foo/bar", vcs.FullName)
}

func TestParse_Bitbucket(t *testing.T) {
	urls := []string{
		"bitbucket.org/foo/bar",
		"https://bitbucket.org/foo/bar",
		"http://bitbucket.org/foo/bar",
		"http://bitbucket.org/foo/bar.git",
		"https://baz@bitbucket.org/foo/bar.git",
		"git@bitbucket.org:foo/bar.git",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			vcs, err := vcsurl.Parse(url)
			assert.NoError(t, err)
			AssertVCS_Bitbucket(t, vcs)
		})
	}
}

func TestParse_BitbucketCommittish(t *testing.T) {
	urls := []string{
		"https://bitbucket.org/foo/bar/src/qux/",
		"https://bitbucket.org/foo/bar/commits/qux",
		"https://bitbucket.org/foo/bar/branch/qux",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			vcs, err := vcsurl.Parse(url)
			assert.NoError(t, err)
			assert.Equal(t, "qux", vcs.Committish)
			AssertVCS_Bitbucket(t, vcs)
		})
	}
}

func AssertVCS_Bitbucket(t *testing.T, vcs *vcsurl.VCS) {
	t.Helper()
	assert.Equal(t, vcsurl.Git, vcs.Kind)
	assert.Equal(t, vcsurl.Bitbucket, vcs.Host)
	assert.Equal(t, "foo", vcs.Username)
	assert.Equal(t, "bar", vcs.Name)
	assert.Equal(t, "foo/bar", vcs.FullName)
}

func TestParse_Gitlab(t *testing.T) {
	urls := []string{
		"gitlab.com/foo/bar",
		"https://gitlab.com/foo/bar",
		"https://gitlab.com/foo/bar.git",
		"git@gitlab.com:foo/bar.git",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			vcs, err := vcsurl.Parse(url)
			assert.NoError(t, err)
			assert.Equal(t, vcsurl.Git, vcs.Kind)
			assert.Equal(t, vcsurl.GitLab, vcs.Host)
			assert.Equal(t, "foo", vcs.Username)
			assert.Equal(t, "bar", vcs.Name)
			assert.Equal(t, "foo/bar", vcs.FullName)
		})
	}
}

func TestParse_GitlabSubGroup(t *testing.T) {
	urls := []string{
		"gitlab.com/foo/bar/qux",
		"https://gitlab.com/foo/bar/qux",
		"https://gitlab.com/foo/bar/qux.git",
		"git@gitlab.com:foo/bar/qux.git",
	}

	for _, url := range urls {
		t.Run(url, func(t *testing.T) {
			vcs, err := vcsurl.Parse(url)
			assert.NoError(t, err)
			assert.Equal(t, vcsurl.Git, vcs.Kind)
			assert.Equal(t, vcsurl.GitLab, vcs.Host)
			assert.Equal(t, "foo/bar", vcs.Username)
			assert.Equal(t, "qux", vcs.Name)
			assert.Equal(t, "foo/bar/qux", vcs.FullName)
		})
	}
}

func TestParse_GiteaAndSelfHosted(t *testing.T) {
	tests := []struct {
		url      string
		host     vcsurl.Host
		username string
		name     string
		fullName string
		id       string
	}{
		{
			url:      "https://gitea.example.com/orgname/myrepo.git",
			host:     "gitea.example.com",
			username: "orgname",
			name:     "myrepo",
			fullName: "orgname/myrepo",
			id:       "gitea.example.com/orgname/myrepo",
		},
		{
			url:      "http://localhost:3000/gitea_admin/test-app",
			host:     "localhost:3000",
			username: "gitea_admin",
			name:     "test-app",
			fullName: "gitea_admin/test-app",
			id:       "localhost:3000/gitea_admin/test-app",
		},
		{
			url:      "git@gitea.example.com:gitea_admin/test-app.git",
			host:     "gitea.example.com",
			username: "gitea_admin",
			name:     "test-app",
			fullName: "gitea_admin/test-app",
			id:       "gitea.example.com/gitea_admin/test-app",
		},
		{
			url:      "https://gitea.com/gitea/tea.git",
			host:     "gitea.com",
			username: "gitea",
			name:     "tea",
			fullName: "gitea/tea",
			id:       "gitea.com/gitea/tea",
		},
		{
			url:      "https://gitlab.custom-domain.net/mygroup/mysubgroup/project.git",
			host:     "gitlab.custom-domain.net",
			username: "mygroup/mysubgroup",
			name:     "project",
			fullName: "mygroup/mysubgroup/project",
			id:       "gitlab.custom-domain.net/mygroup/mysubgroup/project",
		},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			vcs, err := vcsurl.Parse(tt.url)
			assert.NoError(t, err)
			assert.Equal(t, tt.host, vcs.Host)
			assert.Equal(t, tt.username, vcs.Username)
			assert.Equal(t, tt.name, vcs.Name)
			assert.Equal(t, tt.fullName, vcs.FullName)
			assert.Equal(t, tt.id, vcs.ID)
		})
	}
}

func TestParse_Empty(t *testing.T) {
	_, err := vcsurl.Parse("")
	assert.Error(t, err)
}

func TestParse_Invalid(t *testing.T) {
	_, err := vcsurl.Parse("foo")
	assert.Error(t, err)
}

func TestVCSRemote(t *testing.T) {
	tests := []struct {
		raw      string
		p        vcsurl.Protocol
		expected string
		err      error
	}{
		{"https://github.com/foo/bar", vcsurl.SSH, "git@github.com:foo/bar.git", nil},
		{"https://github.com/foo/bar", vcsurl.HTTPS, "https://github.com/foo/bar.git", nil},
		{"https://bitbucket.org/foo/bar", vcsurl.SSH, "git@bitbucket.org:foo/bar.git", nil},
		{"https://bitbucket.org/foo/bar", vcsurl.HTTPS, "https://bitbucket.org/foo/bar.git", nil},
		{"https://gitlab.com/foo/bar", vcsurl.SSH, "git@gitlab.com:foo/bar.git", nil},
		{"https://gitlab.com/foo/bar", vcsurl.HTTPS, "https://gitlab.com/foo/bar.git", nil},
		{"https://gitea.example.com/org/repo.git", vcsurl.SSH, "git@gitea.example.com:org/repo.git", nil},
		{"https://gitea.example.com/org/repo.git", vcsurl.HTTPS, "https://gitea.example.com/org/repo.git", nil},
	}

	for _, test := range tests {
		t.Run(test.raw, func(t *testing.T) {
			vcs, err := vcsurl.Parse(test.raw)
			assert.NoError(t, err)

			remote, err := vcs.Remote(test.p)
			assert.NoError(t, err)
			assert.Equal(t, test.expected, remote)
		})
	}
}

func TestSSHRemoteAndHTTPSRemote(t *testing.T) {
	t.Run("GitHub", func(t *testing.T) {
		v, err := vcsurl.Parse("https://github.com/octocat/Hello-World.git")
		assert.NoError(t, err)
		assert.Equal(t, "git@github.com:octocat/Hello-World.git", v.SSHRemote())
		assert.Equal(t, "https://github.com/octocat/Hello-World.git", v.HTTPSRemote())
	})

	t.Run("Gitea", func(t *testing.T) {
		v, err := vcsurl.Parse("http://localhost:3000/gitea_admin/test-app")
		assert.NoError(t, err)
		assert.Equal(t, "git@localhost:3000:gitea_admin/test-app.git", v.SSHRemote())
		assert.Equal(t, "https://localhost:3000/gitea_admin/test-app.git", v.HTTPSRemote())
	})

	t.Run("GitLab Subgroups", func(t *testing.T) {
		v, err := vcsurl.Parse("https://gitlab.com/company/team/project.git")
		assert.NoError(t, err)
		assert.Equal(t, "git@gitlab.com:company/team/project.git", v.SSHRemote())
		assert.Equal(t, "https://gitlab.com/company/team/project.git", v.HTTPSRemote())
	})

	t.Run("Nil safety", func(t *testing.T) {
		var v *vcsurl.VCS
		assert.Equal(t, "", v.SSHRemote())
		assert.Equal(t, "", v.HTTPSRemote())
	})
}
