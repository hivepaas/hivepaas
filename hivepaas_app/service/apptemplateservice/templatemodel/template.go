// Package templatemodel is the app template format: what a template file, the
// category and tag vocabularies and the generated index contain.
//
// It is the durable half of the feature. A template published today is read by
// versions of HivePaaS that do not exist yet, so the shape here changes only with
// APIVersion, and it shares no types with the dashboard DTOs.
package templatemodel

import (
	"bytes"

	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// APIVersion is the format version of every file in a templates repository. It
// is the spec format's version, because a template's app block is a spec AppDoc.
const APIVersion = "hivepaas.com/v1"

const (
	KindAppTemplate        = "AppTemplate"
	KindTemplateIndex      = "TemplateIndex"
	KindTemplateCategories = "TemplateCategories"
	KindTemplateTags       = "TemplateTags"
)

// Template is one file under templates/.
type Template struct {
	APIVersion string       `yaml:"apiVersion"`
	Kind       string       `yaml:"kind"`
	Metadata   Metadata     `yaml:"metadata"`
	Parameters []*Parameter `yaml:"parameters,omitempty"`
	Variants   []*Variant   `yaml:"variants,omitempty"`
	Versions   []*Version   `yaml:"versions"`

	// App is a specmodel.AppDoc with placeholders in it. It stays an untyped tree
	// until it is rendered: version overrides and placeholders apply to the tree,
	// and only the result has to decode as an AppDoc.
	App map[string]any `yaml:"app"`
}

type Metadata struct {
	Name        string   `yaml:"name"`
	Title       string   `yaml:"title"`
	Tagline     string   `yaml:"tagline"`
	Description string   `yaml:"description"`
	Categories  []string `yaml:"categories"`
	Tags        []string `yaml:"tags,omitempty"`
	Aliases     []string `yaml:"aliases,omitempty"`
	Icon        string   `yaml:"icon"`
	Links       *Links   `yaml:"links,omitempty"`
	License     string   `yaml:"license,omitempty"`
	Requires    Requires `yaml:"requires"`
}

type Links struct {
	Website       string `yaml:"website,omitempty" json:"website,omitempty"`
	Documentation string `yaml:"documentation,omitempty" json:"documentation,omitempty"`
	Source        string `yaml:"source,omitempty" json:"source,omitempty"`
}

// Requires names the oldest HivePaaS able to provision a template, by version
// code (base.CurrentVersion). Codes are fixed-width, so they compare as strings.
type Requires struct {
	VersionCode string `yaml:"versionCode" json:"versionCode"`
}

// IsCompatible reports whether a HivePaaS at currentVersionCode can provision a
// template with these requirements.
func IsCompatible(requires Requires, currentVersionCode string) bool {
	return requires.VersionCode <= currentVersionCode
}

type ParamType string

const (
	ParamTypeString ParamType = "string"
	ParamTypeSecret ParamType = "secret"
	ParamTypeInt    ParamType = "int"
	ParamTypeSize   ParamType = "size"
	ParamTypeBool   ParamType = "bool"
	ParamTypeSelect ParamType = "select"
	ParamTypeVolume ParamType = "volume"
)

// AllParamTypes is closed on purpose: the dashboard generates a form field per
// parameter, and a type it does not know is a field it cannot draw.
//
// TODO: app templates phase 3 - more parameter types (ssl-cert, registry-auth,
// domain). See docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
var AllParamTypes = []ParamType{ParamTypeString, ParamTypeSecret, ParamTypeInt, ParamTypeSize,
	ParamTypeBool, ParamTypeSelect, ParamTypeVolume}

type Parameter struct {
	Name        string    `yaml:"name"`
	Title       string    `yaml:"title"`
	Description string    `yaml:"description,omitempty"`
	Type        ParamType `yaml:"type"`
	Default     any       `yaml:"default,omitempty"`
	Optional    bool      `yaml:"optional,omitempty"`

	Pattern   string          `yaml:"pattern,omitempty"`
	MinLength *int            `yaml:"minLength,omitempty"`
	MaxLength *int            `yaml:"maxLength,omitempty"`
	Min       any             `yaml:"min,omitempty"`
	Max       any             `yaml:"max,omitempty"`
	Generate  *Generate       `yaml:"generate,omitempty"`
	Options   []*SelectOption `yaml:"options,omitempty"`
}

const (
	CharsetAlnum = "alnum"
	CharsetHex   = "hex"
)

type Generate struct {
	Length  int    `yaml:"length"`
	Charset string `yaml:"charset,omitempty"`
}

type SelectOption struct {
	Value string `yaml:"value"`
	Title string `yaml:"title"`
}

type Variant struct {
	Name        string `yaml:"name"`
	Title       string `yaml:"title"`
	Description string `yaml:"description,omitempty"`
	Default     bool   `yaml:"default,omitempty"`
}

// Version is a major line, pinned to an exact image per variant.
type Version struct {
	Name       string            `yaml:"name"`
	Release    string            `yaml:"release"`
	Default    bool              `yaml:"default,omitempty"`
	Deprecated bool              `yaml:"deprecated,omitempty"`
	Image      string            `yaml:"image,omitempty"`
	Images     map[string]string `yaml:"images,omitempty"`
	Vars       map[string]string `yaml:"vars,omitempty"`
	Override   *Override         `yaml:"override,omitempty"`
}

// Override is what a version changes in the template. It is a struct with one
// field rather than the app tree itself so that a version can later override
// something other than app.
//
// TODO: app templates later - per-version parameter defaults. See
// docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
type Override struct {
	App map[string]any `yaml:"app,omitempty"`
}

// ImageFor returns this version's image for a variant; variant is empty for a
// template without variants.
func (v *Version) ImageFor(variant string) string {
	if variant == "" {
		return v.Image
	}
	return v.Images[variant]
}

func (t *Template) DefaultVersion() *Version {
	for _, version := range t.Versions {
		if version != nil && version.Default {
			return version
		}
	}
	return nil
}

func (t *Template) FindVersion(name string) *Version {
	for _, version := range t.Versions {
		if version != nil && version.Name == name {
			return version
		}
	}
	return nil
}

func (t *Template) DefaultVariant() *Variant {
	for _, variant := range t.Variants {
		if variant != nil && variant.Default {
			return variant
		}
	}
	return nil
}

func (t *Template) FindVariant(name string) *Variant {
	for _, variant := range t.Variants {
		if variant != nil && variant.Name == name {
			return variant
		}
	}
	return nil
}

// DecodeTemplate parses a template file strictly. An unknown field is an error:
// an older HivePaaS that skipped a field it did not understand would provision
// something other than what the template describes, and nothing would say so.
func DecodeTemplate(data []byte) (*Template, error) {
	tmpl := &Template{}
	if err := decodeYAMLStrict(data, tmpl); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("%s", err.Error())
	}
	return tmpl, nil
}

func decodeYAMLStrict(data []byte, out any) error {
	decoder := yaml.NewDecoder(bytes.NewReader(data))
	decoder.KnownFields(true)
	if err := decoder.Decode(out); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
