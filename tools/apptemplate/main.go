// Command apptemplate works on a checkout of the app-templates repository: it
// lints templates, writes index.json, renders a template the way HivePaaS will,
// and prints the pin release.json carries.
//
// It is built on the packages HivePaaS renders templates with, so a template this
// tool accepts is one HivePaaS accepts.
//
// Usage:
//
//	go run ./tools/apptemplate lint   <dir>
//	go run ./tools/apptemplate index  [-check] <dir>
//	go run ./tools/apptemplate render [-version 18] [-variant alpine] [-param name=value]... <dir> <template>
//	go run ./tools/apptemplate pin    <dir>
package main

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterepo"
)

var (
	errProblems   = errors.New("the repository has problems")
	errIndexStale = errors.New("index.json is out of date: run `apptemplate index`")

	githubRemotePattern = regexp.MustCompile(
		`^(?:https://github\.com/|git@github\.com:|ssh://git@github\.com/)([A-Za-z0-9][A-Za-z0-9-]*/[A-Za-z0-9._-]+?)(?:\.git)?/?$`)
)

func main() {
	if len(os.Args) < 2 { //nolint:mnd
		usage()
	}
	args := os.Args[2:]
	var err error
	switch os.Args[1] {
	case "lint":
		err = runLint(args, os.Stdout)
	case "index":
		err = runIndex(args, os.Stdout)
	case "render":
		err = runRender(args, os.Stdout)
	case "pin":
		err = runPin(args, os.Stdout)
	default:
		usage()
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, "apptemplate:", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: apptemplate lint|index|render|pin [flags] <dir> ...")
	os.Exit(2) //nolint:mnd
}

func runLint(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("lint", flag.ContinueOnError)
	if err := flags.Parse(args); err != nil {
		return err
	}
	dir, err := singleDir(flags)
	if err != nil {
		return err
	}

	repo, problems, err := templaterepo.Load(os.DirFS(dir))
	if err != nil {
		return errors.New(templaterepo.ErrorText(err))
	}
	problems = append(problems, templaterepo.Lint(repo)...)
	for _, problem := range problems {
		fmt.Fprintln(out, problem)
	}
	if len(problems) > 0 {
		return fmt.Errorf("%w: %d", errProblems, len(problems))
	}
	fmt.Fprintf(out, "OK: %d template(s)\n", len(repo.Templates))
	return nil
}

func runIndex(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("index", flag.ContinueOnError)
	check := flags.Bool("check", false, "fail if index.json differs from what would be written")
	if err := flags.Parse(args); err != nil {
		return err
	}
	dir, err := singleDir(flags)
	if err != nil {
		return err
	}

	data, err := buildIndex(dir)
	if err != nil {
		return err
	}
	path := filepath.Join(dir, templaterepo.IndexFile)
	if *check {
		current, readErr := os.ReadFile(path)
		if readErr != nil || !bytes.Equal(current, data) {
			return errIndexStale
		}
		fmt.Fprintln(out, "index.json is up to date")
		return nil
	}
	if err = os.WriteFile(path, data, 0o644); err != nil { //nolint:gosec
		return err
	}
	fmt.Fprintf(out, "written: %s\n", path)
	return nil
}

// buildIndex refuses a repository with problems: an index is published, and an
// index of broken templates is a store of forms that cannot be submitted.
func buildIndex(dir string) ([]byte, error) {
	repo, problems, err := templaterepo.Load(os.DirFS(dir))
	if err != nil {
		return nil, errors.New(templaterepo.ErrorText(err))
	}
	if problems = append(problems, templaterepo.Lint(repo)...); len(problems) > 0 {
		return nil, fmt.Errorf("%w: run `apptemplate lint` to see them", errProblems)
	}
	index, err := templaterepo.BuildIndex(repo)
	if err != nil {
		return nil, errors.New(templaterepo.ErrorText(err))
	}
	return templaterepo.MarshalIndex(index)
}

type paramFlags map[string]any

func (p paramFlags) String() string {
	pairs := make([]string, 0, len(p))
	for name, value := range p {
		pairs = append(pairs, fmt.Sprintf("%s=%v", name, value))
	}
	sort.Strings(pairs)
	return strings.Join(pairs, ",")
}

func (p paramFlags) Set(value string) error {
	name, v, found := strings.Cut(value, "=")
	if !found || name == "" {
		return fmt.Errorf("-param %q: want name=value", value)
	}
	p[name] = v
	return nil
}

func runRender(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("render", flag.ContinueOnError)
	version := flags.String("version", "", "version to render; empty for the default")
	variant := flags.String("variant", "", "variant to render; empty for the default")
	params := paramFlags{}
	flags.Var(params, "param", "a parameter as name=value; repeatable")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 2 { //nolint:mnd
		return errors.New("usage: apptemplate render [flags] <dir> <template>")
	}

	repo, _, err := templaterepo.Load(os.DirFS(flags.Arg(0)))
	if err != nil {
		return errors.New(templaterepo.ErrorText(err))
	}
	file := repo.FindTemplate(flags.Arg(1))
	if file == nil {
		return fmt.Errorf("no template %q in %s (a template that does not load is left out: run lint)",
			flags.Arg(1), flags.Arg(0))
	}

	result, err := templaterender.Render(&templaterender.Request{
		Template:        file.Template,
		Version:         *version,
		Variant:         *variant,
		Params:          params,
		AllowDeprecated: true,
	})
	if err != nil {
		return errors.New(templaterepo.ErrorText(err))
	}

	fmt.Fprintf(out, "# %s %s (%s), base sha256 %s\n", file.Template.Metadata.Name, result.Version.Name,
		result.Image, result.BaseSHA256)
	encoder := yaml.NewEncoder(out)
	encoder.SetIndent(2) //nolint:mnd
	if err = encoder.Encode(result.Doc); err != nil {
		return err
	}
	return encoder.Close()
}

type pin struct {
	Repo        string `json:"repo"`
	Commit      string `json:"commit"`
	IndexSHA256 string `json:"indexSha256"`
}

// runPin prints what release.json's templates field carries for the checkout's
// HEAD. The hash is of index.json as committed, because that is the file the
// official source fetches by commit.
func runPin(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("pin", flag.ContinueOnError)
	if err := flags.Parse(args); err != nil {
		return err
	}
	dir, err := singleDir(flags)
	if err != nil {
		return err
	}

	status, err := gitOutput(dir, "status", "--porcelain", "--", templaterepo.IndexFile)
	if err != nil {
		return err
	}
	if len(bytes.TrimSpace(status)) > 0 {
		return errors.New("index.json has uncommitted changes: the pin names the committed file")
	}
	commit, err := gitOutput(dir, "rev-parse", "HEAD")
	if err != nil {
		return err
	}
	index, err := gitOutput(dir, "show", "HEAD:"+templaterepo.IndexFile)
	if err != nil {
		return err
	}
	remote, err := gitOutput(dir, "remote", "get-url", "origin")
	if err != nil {
		return err
	}
	repo, err := repoFromRemote(string(remote))
	if err != nil {
		return err
	}

	sum := sha256.Sum256(index)
	encoder := json.NewEncoder(out)
	encoder.SetIndent("", "  ")
	return encoder.Encode(pin{
		Repo:        repo,
		Commit:      strings.TrimSpace(string(commit)),
		IndexSHA256: hex.EncodeToString(sum[:]),
	})
}

// repoFromRemote reads owner/name from a GitHub remote. Only GitHub: the
// official source builds raw.githubusercontent.com URLs from it.
func repoFromRemote(remote string) (string, error) {
	match := githubRemotePattern.FindStringSubmatch(strings.TrimSpace(remote))
	if match == nil {
		return "", fmt.Errorf("remote %q is not a GitHub repository", strings.TrimSpace(remote))
	}
	return match[1], nil
}

func gitOutput(dir string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", append([]string{"-C", dir}, args...)...)
	var stderr bytes.Buffer
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("git %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(stderr.String()))
	}
	return out, nil
}

func singleDir(flags *flag.FlagSet) (string, error) {
	if flags.NArg() != 1 {
		return "", fmt.Errorf("usage: apptemplate %s <dir>", flags.Name())
	}
	return flags.Arg(0), nil
}
