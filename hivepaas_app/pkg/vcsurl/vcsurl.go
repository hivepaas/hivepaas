package vcsurl

import (
	"errors"
	"fmt"
	"net/url"
	"path/filepath"
	"regexp"
	"strings"
)

// Errors returned by Parse function.
var (
	ErrUnknownURL      = errors.New("unknown URL format")
	ErrUnableParse     = errors.New("unable to determine name or full name")
	ErrEmptyURL        = errors.New("empty URL")
	ErrEmptyPath       = errors.New("empty path in URL")
	ErrUnknownProtocol = errors.New("remote protocol should be SSH or HTTPS")
)

// Host VCS provider.
type Host string

// Supported VCS host provider.
const (
	GitHub    Host = "github.com"
	Bitbucket Host = "bitbucket.org"
	GitLab    Host = "gitlab.com"

	gitHubAPI Host = "api.github.com"
)

// Kind of VCS
type Kind string

// Supported VCS kinds.
const (
	Git Kind = "git"
)

// Protocol of remote
type Protocol string

// Supported VCS protocols.
const (
	SSH   Protocol = "ssh"
	HTTPS Protocol = "https"
)

var kindByHost = map[Host]Kind{
	GitHub:    Git,
	gitHubAPI: Git,
	GitLab:    Git,
	Bitbucket: Git,
}

// VCS describes a VCS repository.
type VCS struct {
	// ID unique repository identification (e.g. github.com/owner/repo or gitea.com/owner/repo).
	ID string
	// Kind of VCS.
	Kind Kind
	// Host is the public web host of the repository.
	Host Host
	// Username of repo owner or namespace on repo hosting site.
	Username string
	// Name base name of repo on repo hosting site.
	Name string
	// FullName full name of repo on repo hosting site (e.g. owner/repo or group/subgroup/repo).
	FullName string
	// Committish is a reference to an object that can be recursively
	// dereferenced to a commit object. They can be commits, tags or branches.
	Committish string
	// Raw is the original parsed URL.
	Raw string
}

var (
	removeDotGit    = regexp.MustCompile(`\.git$`)
	gitPreprocessRE = regexp.MustCompile(`^git@([a-zA-Z0-9-_\.]+)\:(.*)$`)
)

// Parse parses a string that resembles a VCS repository URL.
//
//nolint:mnd
func Parse(raw string) (*VCS, error) {
	if len(raw) == 0 {
		return nil, fmt.Errorf("parse %q: %w", raw, ErrEmptyURL)
	}

	spec := raw
	if parts := gitPreprocessRE.FindStringSubmatch(spec); len(parts) == 3 {
		spec = fmt.Sprintf("git://%s/%s", parts[1], parts[2])
	}

	parsedURL, err := url.Parse(spec)
	if err != nil {
		return nil, fmt.Errorf("url parse: %w", err)
	}

	if parsedURL.Scheme == "" {
		spec = "https://" + spec
		if parsedURL, err = url.Parse(spec); err != nil {
			return nil, fmt.Errorf("url parse: %w", err)
		}
	}

	vcs := &VCS{}
	vcs.Raw = raw
	vcs.Host = Host(parsedURL.Host)
	if kind, ok := kindByHost[vcs.Host]; ok {
		vcs.Kind = kind
	} else {
		vcs.Kind = Git
	}

	switch vcs.Host {
	case GitHub, gitHubAPI:
		err = vcs.parseGitHub(parsedURL)
	case Bitbucket:
		err = vcs.parseBitbucket(parsedURL)
	case GitLab:
		err = vcs.parseGitlab(parsedURL)
	default:
		err = vcs.parseDefault(parsedURL)
	}

	if err != nil {
		return nil, fmt.Errorf("parse %q: %w", raw, err)
	}

	if vcs.ID == "" {
		vcs.ID = fmt.Sprintf("%s/%s", string(vcs.Host), vcs.FullName)
	}

	return vcs, nil
}

//nolint:mnd
func (v *VCS) parseGitHub(u *url.URL) error {
	parts := strings.Split(u.Path, "/")
	if v.Host == gitHubAPI {
		v.Host = GitHub
		if len(parts) < 2 || parts[1] != "repos" {
			return ErrUnknownURL
		}
		parts = parts[1:]
	}

	if len(parts) < 3 {
		return ErrUnknownURL
	}

	v.Username = parts[1]
	v.Name = removeDotGit.ReplaceAllLiteralString(parts[2], "")
	v.FullName = v.Username + "/" + v.Name

	if len(parts) < 5 {
		return nil
	}

	if _, ok := githubCommittishParts[parts[3]]; ok {
		v.Committish = strings.Join(parts[4:], "/")
		return nil
	}

	if len(parts) >= 6 && parts[3] == "releases" {
		v.Committish = parts[5]
	}

	return nil
}

var githubCommittishParts = map[string]struct{}{
	"commits":  {},
	"commit":   {},
	"tree":     {},
	"branches": {},
}

//nolint:mnd
func (v *VCS) parseBitbucket(u *url.URL) error {
	parts := strings.Split(u.Path, "/")
	if len(parts) < 3 {
		return ErrUnknownURL
	}

	v.Username = parts[1]
	v.Name = removeDotGit.ReplaceAllLiteralString(parts[2], "")
	v.FullName = v.Username + "/" + v.Name

	if len(parts) >= 5 && (parts[3] == "src" || parts[3] == "commits" || parts[3] == "branch") {
		v.Committish = parts[4]
	}

	return nil
}

//nolint:mnd
func (v *VCS) parseGitlab(u *url.URL) error {
	parts := strings.Split(u.Path, "/")
	if len(parts) < 3 {
		return ErrUnknownURL
	}

	var last int
	for _, p := range parts {
		if p == "-" {
			break
		}
		last++
	}

	v.Username = strings.Join(parts[1:last-1], "/")
	v.Name = removeDotGit.ReplaceAllLiteralString(parts[last-1], "")
	v.FullName = v.Username + "/" + v.Name

	if len(parts) >= (last + 2) {
		object := parts[last+1]
		if object == "tags" || object == "commit" || object == "tree" {
			v.Committish = strings.Join(parts[last+2:], "/")
		}
	}

	return nil
}

func (v *VCS) parseDefault(u *url.URL) error {
	path := u.Path
	if len(path) == 0 {
		return ErrEmptyPath
	}

	path = strings.TrimPrefix(path, "/")
	path = removeDotGit.ReplaceAllLiteralString(path, "")
	v.FullName = path
	v.Name = filepath.Base(path)
	if strings.Contains(u.String(), "git") {
		v.Kind = Git
	}

	if v.Name == "" || v.FullName == "" {
		return ErrUnableParse
	}

	if v.FullName != v.Name && strings.HasSuffix(v.FullName, "/"+v.Name) {
		v.Username = strings.TrimSuffix(v.FullName, "/"+v.Name)
	}

	return nil
}

// Remote returns a remote URL in the given protocol. ErrUnsupportedProtocol
// is returned if the protocol is not supported by the VCS.
func (v *VCS) Remote(p Protocol) (string, error) {
	switch p {
	case SSH:
		return v.SSHRemote(), nil
	case HTTPS:
		return v.HTTPSRemote(), nil
	}

	return "", ErrUnknownProtocol
}

// SSHRemote returns the SSH remote URL in the format `git@host:fullName.git`.
func (v *VCS) SSHRemote() string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("git@%s:%s.git", v.Host, strings.TrimSuffix(v.FullName, ".git"))
}

// HTTPSRemote returns the HTTPS remote URL in the format `https://host/fullName.git`.
func (v *VCS) HTTPSRemote() string {
	if v == nil {
		return ""
	}
	return fmt.Sprintf("https://%s/%s.git", v.Host, strings.TrimSuffix(v.FullName, ".git"))
}
