package templatemodel

import (
	"bytes"
	"encoding/json"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// Index is index.json: everything the store lists, filters and searches by,
// without the templates themselves.
type Index struct {
	APIVersion string        `json:"apiVersion"`
	Kind       string        `json:"kind"`
	Categories []*Category   `json:"categories"`
	Tags       []*Tag        `json:"tags"`
	Templates  []*IndexEntry `json:"templates"`
}

type FileRef struct {
	Path   string `json:"path"`
	SHA256 string `json:"sha256"`
}

type IndexEntry struct {
	Name       string   `json:"name"`
	File       FileRef  `json:"file"`
	Icon       FileRef  `json:"icon"`
	Title      string   `json:"title"`
	Tagline    string   `json:"tagline"`
	Categories []string `json:"categories"`
	Tags       []string `json:"tags,omitempty"`
	Aliases    []string `json:"aliases,omitempty"`
	// License is an SPDX identifier where one fits, and the license's own name
	// where it does not - Timescale-License, BUSL-1.1. It is in the index because
	// the store lists it: what a template costs to run is part of choosing it, and
	// reading every template file to find out would defeat the index.
	License  string          `json:"license,omitempty"`
	Variants []*IndexVariant `json:"variants,omitempty"`
	Versions []*IndexVersion `json:"versions"`
	Requires Requires        `json:"requires"`
}

type IndexVariant struct {
	Name    string `json:"name"`
	Default bool   `json:"default,omitempty"`
}

type IndexVersion struct {
	Name       string   `json:"name"`
	Release    string   `json:"release"`
	Default    bool     `json:"default,omitempty"`
	Deprecated bool     `json:"deprecated,omitempty"`
	Variants   []string `json:"variants,omitempty"`
}

type Category struct {
	ID       string      `yaml:"id" json:"id"`
	Title    string      `yaml:"title" json:"title"`
	Children []*Category `yaml:"children,omitempty" json:"children,omitempty"`
}

type Tag struct {
	ID    string `yaml:"id" json:"id"`
	Title string `yaml:"title" json:"title"`
}

// Categories is categories.yaml.
type Categories struct {
	APIVersion string      `yaml:"apiVersion"`
	Kind       string      `yaml:"kind"`
	Categories []*Category `yaml:"categories"`
}

// Tags is tags.yaml.
type Tags struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       string `yaml:"kind"`
	Tags       []*Tag `yaml:"tags"`
}

func (i *Index) FindTemplate(name string) *IndexEntry {
	for _, entry := range i.Templates {
		if entry.Name == name {
			return entry
		}
	}
	return nil
}

// FindIcon finds the template whose icon has this hash. Icons are served by
// hash, and only a hash listed here is served.
func (i *Index) FindIcon(sha256 string) *IndexEntry {
	for _, entry := range i.Templates {
		if entry.Icon.SHA256 == sha256 {
			return entry
		}
	}
	return nil
}

func DecodeIndex(data []byte) (*Index, error) {
	index := &Index{}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(index); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("index.json: %s", err.Error())
	}
	if err := checkHeader("index.json", index.APIVersion, index.Kind, KindTemplateIndex); err != nil {
		return nil, err
	}
	return index, nil
}

func DecodeCategories(data []byte) (*Categories, error) {
	categories := &Categories{}
	if err := decodeYAMLStrict(data, categories); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("categories.yaml: %s", err.Error())
	}
	if err := checkHeader("categories.yaml", categories.APIVersion, categories.Kind,
		KindTemplateCategories); err != nil {
		return nil, err
	}
	return categories, nil
}

func DecodeTags(data []byte) (*Tags, error) {
	tags := &Tags{}
	if err := decodeYAMLStrict(data, tags); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("tags.yaml: %s", err.Error())
	}
	if err := checkHeader("tags.yaml", tags.APIVersion, tags.Kind, KindTemplateTags); err != nil {
		return nil, err
	}
	return tags, nil
}

func checkHeader(file, apiVersion, kind, wantKind string) error {
	if apiVersion != APIVersion || kind != wantKind {
		return hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: expected apiVersion %q and kind %q", file, APIVersion, wantKind)
	}
	return nil
}
