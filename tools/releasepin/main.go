// Command releasepin pins every image release.json names to the digest its tag
// points to now: `redis:8.6-alpine` becomes `redis:8.6-alpine@sha256:…`. That is
// each `…Image` field, and each image of the `functionRuntimes` map.
//
// A signed release is only a claim about what runs if what it names cannot
// change: a tag can be re-pushed, a digest cannot. The digest is the image
// index's, so one pin serves every architecture.
//
//	go run ./tools/releasepin              # pin release.json in place
//	go run ./tools/releasepin -check       # only report what is not pinned to the current digest
//	go run ./tools/releasepin -check -deps # the same, leaving out the app and the agent
//
// The app's and the agent's images are built after the release commit, from the
// tag on it: until then their tags are not in the registry, and they are left
// as they are, to be pinned by a run after the build. Every other image is
// pinned before the commit - the binary is built with release.json as it is.
//
// It also checks each channel's templates pin: the index.json of the pinned
// commit hashes to its indexSha256, as a server will check it. A pin that does
// not stops it, -check or not; the pin itself is made with
// `go run ./tools/apptemplate pin`.
//
// It needs docker with buildx, and asks the registries and GitHub; run it before
// `make release-sign`, never after.
package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"slices"
	"strings"
)

// imageField is a `"somethingImage": "reference"` line of release.json.
var imageField = regexp.MustCompile(`("[A-Za-z]+Image"\s*:\s*")([^"]+)(")`)

// builtWithTheRelease are the fields naming the images a release builds itself.
var builtWithTheRelease = regexp.MustCompile(`^"(appImage|agentImage)"`)

// functionRuntimes is the `"functionRuntimes": {…}` map of a release, and
// runtimeField one `"runtime": "reference"` line of it.
var (
	functionRuntimes = regexp.MustCompile(`("functionRuntimes"\s*:\s*\{)([^}]*)(\})`)
	runtimeField     = regexp.MustCompile(`("[a-z0-9-]+"\s*:\s*")([^"]+)(")`)
)

// patchTag is a tag with three numbers, 8.6.2-alpine or v1.52.0: one release.
// A tag with fewer may be following a line - redis 8.6 is whatever 8.6.x is
// newest - and names less than its digest pins. Postgres, which numbers its
// releases with two (18.3), is flagged too: this cannot tell the two apart.
var patchTag = regexp.MustCompile(`^v?\d+\.\d+\.\d+`)

type report struct {
	// Changed are the references whose pin was added or moved, as `old -> new`.
	Changed []string
	// Floating are the tags that name less than a patch.
	Floating []string
	// NotBuilt are the app's and the agent's images whose tags the registry
	// does not have yet: the release is not built.
	NotBuilt []string
}

// pin pins every image of release to its tag's current digest, keeping every
// other byte of the file as it is.
func pin(release []byte, resolve func(ref string) (string, error)) ([]byte, report, error) {
	return pinWith(release, resolve, false)
}

// pinWith is pin, leaving the app's and the agent's images as they are when
// depsOnly: before a release is built they are not its to pin - the run that
// builds them makes their digests, a rebuild under the same tag new ones.
func pinWith(release []byte, resolve func(ref string) (string, error), depsOnly bool) ([]byte, report, error) {
	p := &pinner{resolve: resolve, depsOnly: depsOnly}
	out := p.pinMatches(release, imageField)
	out = functionRuntimes.ReplaceAllFunc(out, func(block []byte) []byte {
		parts := functionRuntimes.FindSubmatch(block)
		return slices.Concat(parts[1], p.pinMatches(parts[2], runtimeField), parts[3])
	})
	if p.failed != nil {
		return nil, p.report, p.failed
	}
	return out, p.report, nil
}

// pinner pins references, and remembers what it changed and the first
// reference it could not resolve.
type pinner struct {
	resolve  func(ref string) (string, error)
	depsOnly bool
	report   report
	failed   error
}

// pinMatches pins the reference of every match of field in text: the match's
// second group.
func (p *pinner) pinMatches(text []byte, field *regexp.Regexp) []byte {
	return field.ReplaceAllFunc(text, func(match []byte) []byte {
		if p.failed != nil {
			return match
		}
		parts := field.FindSubmatch(match)
		if p.depsOnly && builtWithTheRelease.Match(parts[1]) {
			return match
		}
		current := string(parts[2])
		name, _, _ := strings.Cut(current, "@")

		digest, err := p.resolve(name)
		if err != nil && builtWithTheRelease.Match(parts[1]) {
			p.report.NotBuilt = append(p.report.NotBuilt, name)
			return match
		}
		if err != nil {
			p.failed = fmt.Errorf("%s: %w", name, err)
			return match
		}
		pinned := name + "@" + digest
		if pinned != current {
			p.report.Changed = append(p.report.Changed, current+" -> "+pinned)
		}
		if !namesAPatch(name) {
			p.report.Floating = append(p.report.Floating, name)
		}
		return []byte(string(parts[1]) + pinned + string(parts[3]))
	})
}

// namesAPatch reports whether a reference's tag has three version numbers.
func namesAPatch(ref string) bool {
	slash := strings.LastIndex(ref, "/")
	colon := strings.LastIndex(ref, ":")
	if colon < slash {
		return false // no tag: latest
	}
	return patchTag.MatchString(ref[colon+1:])
}

// resolveWithDocker asks the registry, through docker buildx, for the digest of
// the image index a tag points to.
func resolveWithDocker(ref string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.Command("docker", "buildx", "imagetools", "inspect", "--format", "{{json .Manifest}}", ref)
	cmd.Stdout, cmd.Stderr = &stdout, &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker buildx imagetools inspect: %w: %s", err, strings.TrimSpace(stderr.String()))
	}
	var manifest struct {
		Digest string `json:"digest"`
	}
	if err := json.Unmarshal(stdout.Bytes(), &manifest); err != nil {
		return "", fmt.Errorf("reading the manifest of %s: %w", ref, err)
	}
	if !strings.HasPrefix(manifest.Digest, "sha256:") {
		return "", errors.New("no digest for " + ref)
	}
	return manifest.Digest, nil
}

func main() {
	file := flag.String("file", "release.json", "the release file to pin")
	check := flag.Bool("check", false, "report what is not pinned to the current digest, and change nothing")
	deps := flag.Bool("deps", false, "leave out the app's and the agent's images, which the release builds")
	flag.Parse()

	release, err := os.ReadFile(*file)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	checked, err := checkTemplates(release, fetchHTTP)
	if err != nil {
		fmt.Fprintln(os.Stderr, "templates pin:", err)
		os.Exit(1)
	}
	for _, pin := range checked {
		fmt.Println("templates pin matches its index:", pin)
	}
	out, rep, err := pinWith(release, resolveWithDocker, *deps)
	if err != nil {
		fmt.Fprintln(os.Stderr, "not pinned:", err)
		os.Exit(1)
	}
	for _, change := range rep.Changed {
		fmt.Println(change)
	}
	for _, ref := range rep.Floating {
		fmt.Printf("note: %s may name a line rather than one release; the digest pins it, the name does not say which\n", ref)
	}
	for _, ref := range rep.NotBuilt {
		fmt.Printf("note: %s is not built yet; pin it again once the release is\n", ref)
	}
	switch {
	case *check && len(rep.Changed) > 0:
		fmt.Fprintf(os.Stderr, "%d image(s) are not pinned to their current digest\n", len(rep.Changed))
		os.Exit(1)
	case *check:
		fmt.Println("every image is pinned to its current digest")
	case len(rep.Changed) == 0:
		fmt.Println("nothing to change")
	default:
		if err := os.WriteFile(*file, out, 0o644); err != nil { //nolint:gosec // a file of the repository
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		fmt.Printf("pinned %d image(s) in %s\n", len(rep.Changed), *file)
	}
}
