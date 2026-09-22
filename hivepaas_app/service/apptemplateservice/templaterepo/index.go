package templaterepo

import (
	"bytes"
	"encoding/json"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const indexIndent = "  "

// BuildIndex builds index.json from a loaded repository. It assumes Lint found
// nothing; the one thing it cannot build around is a missing icon, since the
// icon's hash is part of every entry.
func BuildIndex(repo *Repo) (*templatemodel.Index, error) {
	index := &templatemodel.Index{
		APIVersion: templatemodel.APIVersion,
		Kind:       templatemodel.KindTemplateIndex,
		Categories: repo.Categories.Categories,
		Tags:       repo.Tags.Tags,
		Templates:  []*templatemodel.IndexEntry{},
	}
	for _, file := range repo.Templates {
		tmpl := file.Template
		icon := repo.Icons[tmpl.Metadata.Icon]
		if icon == nil {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
				WithExtraDetail("%s: icon %q is missing", file.Path, tmpl.Metadata.Icon)
		}

		entry := &templatemodel.IndexEntry{
			Name:       tmpl.Metadata.Name,
			File:       templatemodel.FileRef{Path: file.Path, SHA256: file.SHA256},
			Icon:       templatemodel.FileRef{Path: icon.Path, SHA256: icon.SHA256},
			Title:      tmpl.Metadata.Title,
			Tagline:    tmpl.Metadata.Tagline,
			Categories: tmpl.Metadata.Categories,
			Tags:       tmpl.Metadata.Tags,
			Aliases:    tmpl.Metadata.Aliases,
			License:    tmpl.Metadata.License,
			Internal:   tmpl.Metadata.Internal,
			Requires:   tmpl.Metadata.Requires,

			RequiresCapabilities: requiresCapabilities(repo, tmpl),
		}
		for _, variant := range tmpl.Variants {
			entry.Variants = append(entry.Variants,
				&templatemodel.IndexVariant{Name: variant.Name, Default: variant.Default})
		}
		for _, dep := range tmpl.Dependencies {
			entry.Dependencies = append(entry.Dependencies,
				&templatemodel.IndexDependency{Name: dep.Name, Title: dep.Title, Template: dep.Template})
		}
		for _, component := range tmpl.Components {
			entry.Components = append(entry.Components, &templatemodel.IndexComponent{
				Name: component.Name, Title: component.Title, Primary: component.Primary})
		}
		for _, version := range tmpl.Versions {
			indexVersion := &templatemodel.IndexVersion{
				Name:       version.Name,
				Release:    version.Release,
				Default:    version.Default,
				Deprecated: version.Deprecated,
			}
			for _, variant := range tmpl.Variants {
				if version.ImageFor(variant.Name) != "" {
					indexVersion.Variants = append(indexVersion.Variants, variant.Name)
				}
			}
			entry.Versions = append(entry.Versions, indexVersion)
		}
		index.Templates = append(index.Templates, entry)
	}
	return index, nil
}

// requiresCapabilities reports whether creating this template grants
// capabilities to any app it creates - its own, or one of the dependencies
// created alongside it, since those are provisioned by the same request and
// gated on the same permission.
//
// A dependency naming a template this repository does not have is Lint's
// problem; here it simply grants nothing.
func requiresCapabilities(repo *Repo, tmpl *templatemodel.Template) bool {
	if tmpl.RequiresCapabilities() {
		return true
	}
	for _, dep := range tmpl.Dependencies {
		target := repo.FindTemplate(dep.Template)
		if target != nil && target.Template.RequiresCapabilities() {
			return true
		}
	}
	return false
}

// MarshalIndex writes index.json: indented, without HTML escaping, ending in a
// newline, and byte-identical for the same repository.
func MarshalIndex(index *templatemodel.Index) ([]byte, error) {
	var buf bytes.Buffer
	encoder := json.NewEncoder(&buf)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", indexIndent)
	if err := encoder.Encode(index); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return buf.Bytes(), nil
}
