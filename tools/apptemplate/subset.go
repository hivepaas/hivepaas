package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterepo"
)

var (
	errSubsetUsage = errors.New("usage: apptemplate subset -version-code v000001 <dir> <out-dir>")

	versionCodePattern = regexp.MustCompile(`^v[0-9]{6}$`)
)

// runVersionCode prints the version code this HivePaaS is, so that CI can ask a
// released HivePaaS which templates it should be able to read.
func runVersionCode(_ []string, out io.Writer) error {
	fmt.Fprintln(out, base.CurrentVersion)
	return nil
}

// runSubset writes a copy of the repository holding only the templates a
// HivePaaS of -version-code is offered: those whose requires.versionCode is at
// most that code. Everything else that release is shown as needing a newer
// HivePaaS, without reading it.
//
// It exists for CI: every installation of a channel reads the templates the
// channel pins, whatever release it runs, so each release still in use is
// linted - with its own lint, from its tag - against the subset it will be asked
// to read. A template that release cannot read while claiming it can is one whose
// requires.versionCode is too low.
//
// A template's dependencies need no older code than it does - lint sees to that -
// so the subset never names a template it leaves out.
func runSubset(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("subset", flag.ContinueOnError)
	code := flags.String("version-code", "", "keep the templates a HivePaaS of this code is offered")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 2 || !versionCodePattern.MatchString(*code) { //nolint:mnd
		return errSubsetUsage
	}
	dir, outDir := flags.Arg(0), flags.Arg(1)

	repo, problems, err := templaterepo.Load(os.DirFS(dir))
	if err != nil {
		return errors.New(templaterepo.ErrorText(err))
	}
	// Only a repository this release's own lint accepts is cut: a problem here
	// would otherwise show up as the older release's.
	if problems = append(problems, templaterepo.Lint(repo)...); len(problems) > 0 {
		return fmt.Errorf("%w: run `apptemplate lint` to see them", errProblems)
	}

	for _, name := range []string{templaterepo.CategoriesFile, templaterepo.TagsFile} {
		data, err := os.ReadFile(filepath.Join(dir, name))
		if err != nil {
			return err
		}
		if err = writeFileIn(outDir, name, data); err != nil {
			return err
		}
	}
	kept := 0
	for _, file := range repo.Templates {
		if !templatemodel.IsCompatible(file.Template.Metadata.Requires, *code) {
			continue
		}
		if err = writeFileIn(outDir, file.Path, file.Content); err != nil {
			return err
		}
		if icon := repo.Icons[file.Template.Metadata.Icon]; icon != nil {
			if err = writeFileIn(outDir, icon.Path, icon.Content); err != nil {
				return err
			}
		}
		kept++
	}
	fmt.Fprintf(out, "kept %d of %d template(s) for %s in %s\n", kept, len(repo.Templates), *code, outDir)
	return nil
}

func writeFileIn(dir, name string, data []byte) error {
	path := filepath.Join(dir, filepath.FromSlash(name))
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil { //nolint:gosec
		return err
	}
	return os.WriteFile(path, data, 0o644) //nolint:gosec
}
