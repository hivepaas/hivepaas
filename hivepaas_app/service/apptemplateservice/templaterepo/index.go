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
			Requires:   tmpl.Metadata.Requires,
		}
		for _, variant := range tmpl.Variants {
			entry.Variants = append(entry.Variants,
				&templatemodel.IndexVariant{Name: variant.Name, Default: variant.Default})
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
