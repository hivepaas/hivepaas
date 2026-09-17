package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterepo"
	"github.com/hivepaas/hivepaas/services/registry"
)

// maxBumpTags is what a single repository contributes. It has to reach the end of
// the list: a registry answers in lexical order, so the newest release sorts after
// every older major.
const maxBumpTags = 5000

var errAmbiguousBump = errors.New("the image to replace appears more than once in the file")

// tagLister is what bump needs from a registry client, as an interface so the
// tests answer without a network.
type tagLister interface {
	ListTags(ctx context.Context, ref registry.Reference, maxTags int) (*registry.ListTagsResult, error)
}

var newTagLister = func() tagLister { return registry.New() }

// bumpPlan is one version line moving to a newer release of the same major.
type bumpPlan struct {
	Template       string
	Version        string
	CurrentRelease string
	Release        string
	// Images and CurrentImages are keyed by variant name, or by "" for a template
	// without variants.
	CurrentImages map[string]string
	Images        map[string]string
}

func (p *bumpPlan) String() string {
	return fmt.Sprintf("%s %s: %s -> %s", p.Template, p.Version, p.CurrentRelease, p.Release)
}

// runBump moves every version line it can to the newest release its major line
// publishes, then rewrites index.json so the repository stays consistent.
func runBump(args []string, out io.Writer) error {
	flags := flag.NewFlagSet("bump", flag.ContinueOnError)
	dryRun := flags.Bool("dry-run", false, "print what would change without writing")
	if err := flags.Parse(args); err != nil {
		return err
	}
	dir, err := singleDir(flags)
	if err != nil {
		return err
	}

	repo, _, err := templaterepo.Load(os.DirFS(dir))
	if err != nil {
		return errors.New(templaterepo.ErrorText(err))
	}

	lister := newTagLister()
	tagsByRepo := map[string][]string{}
	changed := 0
	for _, file := range repo.Templates {
		plans, planErr := planTemplateBumps(context.Background(), lister, file, tagsByRepo)
		if planErr != nil {
			return planErr
		}
		if len(plans) == 0 {
			continue
		}
		content := file.Content
		for _, plan := range plans {
			fmt.Fprintln(out, plan)
			if content, err = applyBump(content, plan); err != nil {
				return fmt.Errorf("%s: %w", file.Path, err)
			}
		}
		changed++
		if *dryRun {
			continue
		}
		if err = os.WriteFile(filepath.Join(dir, file.Path), content, 0o644); err != nil { //nolint:gosec
			return err
		}
	}

	if changed == 0 {
		fmt.Fprintln(out, "every template is current")
		return nil
	}
	if *dryRun {
		fmt.Fprintln(out, "dry run: nothing written")
		return nil
	}
	return runIndex([]string{dir}, out)
}

// planTemplateBumps plans every version line of one template, reading each
// repository once into tagsByRepo.
func planTemplateBumps(
	ctx context.Context,
	lister tagLister,
	file *templaterepo.TemplateFile,
	tagsByRepo map[string][]string,
) ([]*bumpPlan, error) {
	tmpl := file.Template
	variants := []string{""}
	if len(tmpl.Variants) > 0 {
		variants = variants[:0]
		for _, variant := range tmpl.Variants {
			variants = append(variants, variant.Name)
		}
	}

	plans := make([]*bumpPlan, 0, len(tmpl.Versions))
	for _, version := range tmpl.Versions {
		if version.Deprecated {
			continue
		}
		current := map[string]string{}
		for _, variant := range variants {
			if image := version.ImageFor(variant); image != "" {
				current[variant] = image
			}
		}
		for _, image := range current {
			ref := registry.ParseRepository(image)
			if _, read := tagsByRepo[ref.String()]; read {
				continue
			}
			result, err := lister.ListTags(ctx, ref, maxBumpTags)
			if err != nil {
				return nil, fmt.Errorf("%s: %w", file.Path, err)
			}
			tagsByRepo[ref.String()] = result.Tags
		}

		plan, ok := planVersionBump(version.Name, current, tagsByRepo)
		if !ok {
			continue
		}
		plan.Template, plan.Version, plan.CurrentRelease = tmpl.Metadata.Name, version.Name, version.Release
		plans = append(plans, plan)
	}
	return plans, nil
}

// planVersionBump picks the newest release every variant of one version line
// publishes. line is the version's name, so a template declaring 11.8 moves to
// 11.8.10 and never to 11.9.
//
// Taking each variant's own newest would let alpine run 18.8 while debian stayed
// on 18.6 - one template describing two different versions of the software. So the
// releases are intersected, and the line moves only as far as all of them reach.
func planVersionBump(line string, current map[string]string, tagsByRepo map[string][]string) (*bumpPlan, bool) {
	var shared map[string]bool
	byVariant := map[string]map[string]string{} // variant -> release -> tag

	for variant, image := range current {
		tags := tagsByRepo[registry.ParseRepository(image).String()]
		releases := map[string]string{}
		for _, candidate := range templatemodel.SelectTagCandidates(line, image, tags, 0) {
			if candidate.Class != templatemodel.ImageOverrideSameLine || !candidate.Newer {
				continue
			}
			release := releaseOfTag(candidate.Tag)
			// A release can publish several bases - 18.7-alpine3.24 and
			// 18.7-alpine3.25. The newest of them is the one to move to.
			if existing, found := releases[release]; !found || isNewerTag(existing, candidate.Tag) {
				releases[release] = candidate.Tag
			}
		}
		byVariant[variant] = releases

		names := map[string]bool{}
		for release := range releases {
			names[release] = true
		}
		if shared == nil {
			shared = names
			continue
		}
		for release := range shared {
			if !names[release] {
				delete(shared, release)
			}
		}
	}
	if len(shared) == 0 {
		return nil, false
	}

	best := ""
	for _, release := range slices.Sorted(maps.Keys(shared)) {
		if best == "" || isNewerTag(best, release) {
			best = release
		}
	}

	plan := &bumpPlan{Release: best, CurrentImages: current, Images: map[string]string{}}
	for variant, image := range current {
		plan.Images[variant] = imageref.Parse(image).Repository + ":" + byVariant[variant][best]
	}
	return plan, true
}

// applyBump rewrites the file as text rather than re-marshalling it, so comments
// and formatting survive. Every image it replaces has to appear exactly once, or
// the tool would be guessing which occurrence belongs to this version line.
func applyBump(content []byte, plan *bumpPlan) ([]byte, error) {
	for variant, image := range plan.CurrentImages {
		if bytes.Count(content, []byte(image)) != 1 {
			return nil, fmt.Errorf("%w: %s", errAmbiguousBump, image)
		}
		content = bytes.Replace(content, []byte(image), []byte(plan.Images[variant]), 1)
	}

	currentRelease := []byte(fmt.Sprintf("release: %q", plan.CurrentRelease))
	if bytes.Count(content, currentRelease) != 1 {
		return nil, fmt.Errorf("%w: %s", errAmbiguousBump, currentRelease)
	}
	return bytes.Replace(content, currentRelease, []byte(fmt.Sprintf("release: %q", plan.Release)), 1), nil
}

// releaseOfTag is the version part of a tag: 18.7-alpine3.24 is release 18.7, and
// 11.8.10-noble is 11.8.10 - the same value a template's `release` field carries.
//
// A template whose release omits a level the tag carries - release 2.1 for tag
// 2.1.0 - is rewritten to the tag's own version, because release is documented as
// the exact version shown to users and the tag is where that version comes from.
func releaseOfTag(tag string) string {
	if i := strings.IndexAny(tag, "-_+"); i >= 0 {
		return tag[:i]
	}
	return tag
}

func isNewerTag(current, candidate string) bool {
	order, ok := imageref.CompareTags(current, candidate)
	return ok && order < 0
}
