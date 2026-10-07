package templatemodel

import (
	"cmp"
	"encoding/json"
	"slices"

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

	// Skipped are the entries DecodeIndex left out, for the source to log. They
	// are never written: the index the tooling builds has none.
	Skipped []*SkippedEntry `json:"-"`
}

// SkippedEntry is an entry of index.json the store does without, and why.
type SkippedEntry struct {
	// List is the list it was in: templates, categories or tags.
	List string
	// Name is its name or id, when that much could be read.
	Name   string
	Reason string
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
	License string `json:"license,omitempty"`
	// Internal keeps the template out of the store's listing and search. The
	// entry is still here, because a template that names it as a dependency is
	// fetched through the index like any other.
	Internal bool `json:"internal,omitempty"`
	// Dependencies let the store say what else a template creates without reading
	// its file.
	Dependencies []*IndexDependency `json:"dependencies,omitempty"`
	// Components are the apps a template creates for itself, for the same reason:
	// the store says "this creates thirteen apps" without reading the file.
	Components []*IndexComponent `json:"components,omitempty"`
	Variants   []*IndexVariant   `json:"variants,omitempty"`
	Versions   []*IndexVersion   `json:"versions"`
	Requires   Requires          `json:"requires"`
	// RequiresCapabilities says the template asks for kernel capabilities,
	// sysctls, ulimits or the GPU, which only someone who may change an app's
	// capabilities can grant. It is in the index so that the store can say so
	// before anybody opens the template, and what exactly it asks for is in the
	// template file.
	RequiresCapabilities bool `json:"requiresCapabilities,omitempty"`
	// RequiresDockerAPI says the template gives an app the Docker API, which
	// also takes Write on the Cluster module; in the index for the same reason.
	RequiresDockerAPI bool `json:"requiresDockerApi,omitempty"`
	// Stats are what the store orders templates by besides their name.
	Stats *IndexStats `json:"stats,omitempty"`
}

type IndexVariant struct {
	Name    string `json:"name"`
	Default bool   `json:"default,omitempty"`
}

type IndexDependency struct {
	Name     string `json:"name"`
	Title    string `json:"title"`
	Template string `json:"template"`
}

type IndexComponent struct {
	Name    string `json:"name"`
	Title   string `json:"title"`
	Primary bool   `json:"primary,omitempty"`
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

// rawIndex is index.json with its lists left unread, to be read an entry at a
// time.
type rawIndex struct {
	APIVersion string            `json:"apiVersion"`
	Kind       string            `json:"kind"`
	Categories []json.RawMessage `json:"categories"`
	Tags       []json.RawMessage `json:"tags"`
	Templates  []json.RawMessage `json:"templates"`
}

// DecodeIndex parses index.json. Unlike a template file, it accepts fields it
// does not know, and it does without an entry it cannot use.
//
// Every installation reads the index its channel pins, whatever HivePaaS version
// it runs, so an index gains fields older installations have never heard of.
// Refusing them would take the store down on every one of those installations
// the moment a newer index is pinned. That is safe here and would not be for a
// template: the index only lists, and nothing is provisioned from what an older
// HivePaaS skipped in it. Template files stay strict, and requires.versionCode is
// what keeps an older HivePaaS from provisioning a template it cannot read.
//
// For the same reason one entry this HivePaaS cannot read - a field whose type
// has changed, a template without a version to offer - costs that entry, which
// goes into Skipped, and not the whole store. Only an index that is not one at
// all, or one of another apiVersion, is refused.
func DecodeIndex(data []byte) (*Index, error) {
	raw := &rawIndex{}
	if err := json.Unmarshal(data, raw); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("index.json: %s", err.Error())
	}
	if err := checkHeader("index.json", raw.APIVersion, raw.Kind, KindTemplateIndex); err != nil {
		return nil, err
	}

	index := &Index{APIVersion: raw.APIVersion, Kind: raw.Kind}
	for _, item := range raw.Categories {
		category := &Category{}
		if err := json.Unmarshal(item, category); err != nil || category.ID == "" {
			index.skip("categories", item, reasonOf(err, "it has no id"))
			continue
		}
		index.Categories = append(index.Categories, keepUsableCategories(category))
	}
	for _, item := range raw.Tags {
		tag := &Tag{}
		if err := json.Unmarshal(item, tag); err != nil || tag.ID == "" {
			index.skip("tags", item, reasonOf(err, "it has no id"))
			continue
		}
		index.Tags = append(index.Tags, tag)
	}
	seen := make(map[string]bool, len(raw.Templates))
	for _, item := range raw.Templates {
		entry := &IndexEntry{}
		if err := json.Unmarshal(item, entry); err != nil {
			index.skip("templates", item, err.Error())
			continue
		}
		if reason := entry.normalize(); reason != "" {
			index.skip("templates", item, reason)
			continue
		}
		// FindTemplate answers with the first, so a second is one nobody could open.
		if seen[entry.Name] {
			index.skip("templates", item, "another entry has its name")
			continue
		}
		seen[entry.Name] = true
		index.Templates = append(index.Templates, entry)
	}
	return index, nil
}

// normalize drops the null items a list may hold, and reports why the entry
// cannot be offered at all, or "" when it can: it needs a name to be found by,
// a file to be fetched by hash, and a version to be created at.
func (e *IndexEntry) normalize() string {
	e.Variants = keepOnly(e.Variants, func(variant *IndexVariant) bool { return variant != nil })
	e.Dependencies = keepOnly(e.Dependencies, func(dep *IndexDependency) bool { return dep != nil })
	e.Components = keepOnly(e.Components, func(component *IndexComponent) bool { return component != nil })
	e.Versions = keepOnly(e.Versions, func(version *IndexVersion) bool { return version != nil && version.Name != "" })

	switch {
	case e.Name == "":
		return "it has no name"
	case e.File.Path == "" || e.File.SHA256 == "":
		return "it names no file"
	case len(e.Versions) == 0:
		return "it has no version"
	}
	return ""
}

func (i *Index) skip(list string, item json.RawMessage, reason string) {
	// Whatever identifies the entry, read on its own: the rest of it is what could
	// not be read.
	var named struct {
		Name string `json:"name"`
		ID   string `json:"id"`
	}
	_ = json.Unmarshal(item, &named)
	i.Skipped = append(i.Skipped, &SkippedEntry{List: list, Name: cmp.Or(named.Name, named.ID), Reason: reason})
}

func reasonOf(err error, otherwise string) string {
	if err != nil {
		return err.Error()
	}
	return otherwise
}

// keepUsableCategories drops the children of a category that are null or have
// no id, at any depth.
func keepUsableCategories(category *Category) *Category {
	category.Children = keepOnly(category.Children, func(child *Category) bool { return child != nil && child.ID != "" })
	for _, child := range category.Children {
		keepUsableCategories(child)
	}
	return category
}

// keepOnly returns the items keep accepts. A list it accepts whole is returned as
// it is, nil included, so an index with nothing to drop decodes as it was built.
func keepOnly[T any](items []T, keep func(T) bool) []T {
	if !slices.ContainsFunc(items, func(item T) bool { return !keep(item) }) {
		return items
	}
	return slices.DeleteFunc(slices.Clone(items), func(item T) bool { return !keep(item) })
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
