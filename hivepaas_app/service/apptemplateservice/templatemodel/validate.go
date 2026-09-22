package templatemodel

import (
	"fmt"
	"maps"
	"net/url"
	"path"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
)

const (
	MaxTaglineLen = 80
	MaxCategories = 3
	MaxTags       = 8

	minGenerateLength = 8
	maxGenerateLength = 128
)

var (
	templateNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
	paramNamePattern    = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_]{0,63}$`)
	choiceNamePattern   = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._-]{0,31}$`)
	categoryRefPattern  = regexp.MustCompile(`^[a-z0-9-]+/[a-z0-9-]+$`)
	// engineNamePattern is what an app kind records as its engine: postgres,
	// mysql, mariadb. It is the same shape as a template name because that is
	// where most of them come from.
	engineNamePattern  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
	versionCodePattern = regexp.MustCompile(`^v[0-9]{6}$`)
	// pinnedVersionPattern is what separates 17.6-alpine from 17-alpine: a tag
	// naming one release carries at least major.minor. It is anchored and read
	// against the tag's version part alone, because the base image in the suffix
	// carries a version of its own - 18-alpine3.24 names a moving postgres tag on
	// a pinned alpine, and matching anywhere in the tag would call that pinned.
	//
	// It cannot know the patch level an image publishes, so it is a floor, not proof.
	pinnedVersionPattern = regexp.MustCompile(`^v?[0-9]+\.[0-9]+`)
	// pinnedDatePattern is the other shape a release takes: a date, the way minio's
	// RELEASE.2025-09-07T16-13-09Z does it. It reads the whole tag, since the date's
	// own dashes would cut a version part short.
	pinnedDatePattern = regexp.MustCompile(`\.[0-9]{4}-[0-9]{2}-[0-9]{2}`)
	// pinnedPrefixedVersionPattern is a third shape: a word, then the version, the
	// way fireflyiii/core publishes version-6.7.2 and nothing else. The version part
	// alone would be the word, so this reads the tag from the start instead.
	//
	// The word has to be followed immediately by major.minor, which is what keeps
	// stable-alpine3.24 and 18-alpine3.24 out: the first has no version after the
	// word, and the second's version is a suffix's rather than the image's.
	pinnedPrefixedVersionPattern = regexp.MustCompile(`^[a-zA-Z]+-v?[0-9]+\.[0-9]+`)
)

// problems collects everything wrong with one file, so an author sees all of it
// at once instead of one problem per lint run.
type problems []string

func (p *problems) add(format string, args ...any) {
	*p = append(*p, fmt.Sprintf(format, args...))
}

func (p problems) err(name string) error {
	if len(p) == 0 {
		return nil
	}
	return hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("%s: %s", name, strings.Join(p, "; "))
}

// Validate checks what a template says about itself. What needs the rest of the
// repository - categories and tags against their vocabularies, the icon file -
// is templaterepo's, and whether the template renders is templaterender's.
//
// fileName is the file's name without .yaml, or empty when the template did not
// come from a file.
func (t *Template) Validate(fileName string) error {
	var p problems
	if t.APIVersion != APIVersion {
		p.add("apiVersion must be %q", APIVersion)
	}
	if t.Kind != KindAppTemplate {
		p.add("kind must be %q", KindAppTemplate)
	}
	t.Metadata.validate(fileName, &p)
	validateParameters(t.Parameters, &p)
	variants := validateVariants(t.Variants, &p)
	validateVersions(t, t.Versions, variants, &p)
	validateDependencies(t, &p)
	validateComponents(t, &p)
	validateCapabilities(t, &p)

	name := t.Metadata.Name
	if name == "" {
		name = fileName
	}
	return p.err(name)
}

func (m *Metadata) validate(fileName string, p *problems) {
	switch {
	case !templateNamePattern.MatchString(m.Name):
		p.add("metadata.name %q must be lowercase letters, digits and dashes", m.Name)
	case fileName != "" && m.Name != fileName:
		p.add("metadata.name %q must equal the file name %q", m.Name, fileName)
	}
	if m.Title == "" {
		p.add("metadata.title is required")
	}
	if m.Tagline == "" || utf8.RuneCountInString(m.Tagline) > MaxTaglineLen {
		p.add("metadata.tagline is required and at most %d characters", MaxTaglineLen)
	}
	if strings.TrimSpace(m.Description) == "" {
		p.add("metadata.description is required")
	}
	if len(m.Categories) == 0 || len(m.Categories) > MaxCategories {
		p.add("metadata.categories must have 1 to %d entries", MaxCategories)
	}
	for _, category := range m.Categories {
		if !categoryRefPattern.MatchString(category) {
			p.add("metadata.categories: %q must be parent/child", category)
		}
	}
	if len(m.Tags) > MaxTags {
		p.add("metadata.tags must have at most %d entries", MaxTags)
	}
	if ext := path.Ext(m.Icon); ext != ".svg" && ext != ".png" {
		p.add("metadata.icon must be an .svg or .png file")
	}
	if m.Links != nil {
		for _, link := range [][2]string{
			{"website", m.Links.Website}, {"documentation", m.Links.Documentation}, {"source", m.Links.Source},
		} {
			if link[1] != "" && !isHTTPSURL(link[1]) {
				p.add("metadata.links.%s must be an https URL", link[0])
			}
		}
	}
	if !versionCodePattern.MatchString(m.Requires.VersionCode) {
		p.add("metadata.requires.versionCode must look like v000001")
	}
}

func isHTTPSURL(raw string) bool {
	parsed, err := url.Parse(raw)
	return err == nil && parsed.Scheme == "https" && parsed.Host != ""
}

func validateParameters(params []*Parameter, p *problems) {
	seen := map[string]bool{}
	for i, param := range params {
		if param == nil {
			p.add("parameters[%d] is empty", i)
			continue
		}
		prefix := fmt.Sprintf("parameters[%s]", param.Name)
		if !paramNamePattern.MatchString(param.Name) {
			p.add("parameters[%d].name %q must start with a letter and use letters, digits and _", i, param.Name)
		}
		if seen[param.Name] {
			p.add("%s is declared twice", prefix)
		}
		seen[param.Name] = true
		if param.Title == "" {
			p.add("%s.title is required", prefix)
		}
		validateParameterType(prefix, param, p)
	}
}

func validateParameterType(prefix string, param *Parameter, p *problems) {
	switch param.Type {
	case ParamTypeString:
		validateLengths(prefix, param, p)
		if param.Pattern != "" {
			if _, err := regexp.Compile(param.Pattern); err != nil {
				p.add("%s.pattern does not compile: %v", prefix, err)
			}
		}
		refuseAttributes(prefix, param, p, "min", "max", "generate", "options", "engine")
	case ParamTypeSecret:
		validateLengths(prefix, param, p)
		if param.Default != nil {
			p.add("%s: a secret has no default; use generate", prefix)
		}
		if gen := param.Generate; gen != nil {
			if gen.Length < minGenerateLength || gen.Length > maxGenerateLength {
				p.add("%s.generate.length must be %d to %d", prefix, minGenerateLength, maxGenerateLength)
			}
			if gen.Charset != "" && gen.Charset != CharsetAlnum && gen.Charset != CharsetHex {
				p.add("%s.generate.charset must be %s or %s", prefix, CharsetAlnum, CharsetHex)
			}
		}
		refuseAttributes(prefix, param, p, "pattern", "min", "max", "options", "engine")
	case ParamTypeInt:
		validateBounds(prefix, param, p, "integers", ToInt64)
		refuseAttributes(prefix, param, p, "pattern", "minLength", "maxLength", "generate", "options", "engine")
	case ParamTypeSize:
		validateBounds(prefix, param, p, "sizes", sizeBytes)
		refuseAttributes(prefix, param, p, "pattern", "minLength", "maxLength", "generate", "options", "engine")
	case ParamTypeBool:
		refuseAttributes(prefix, param, p, "pattern", "minLength", "maxLength", "min", "max", "generate", "options",
			"engine")
	case ParamTypeSelect:
		validateOptions(prefix, param.Options, p)
		refuseAttributes(prefix, param, p, "pattern", "minLength", "maxLength", "min", "max", "generate", "engine")
	case ParamTypeVolume:
		if param.Default != nil {
			p.add("%s: a volume has no default", prefix)
		}
		refuseAttributes(prefix, param, p, "pattern", "minLength", "maxLength", "min", "max", "generate", "options",
			"engine")
	case ParamTypeDomain:
		if param.Default != nil {
			p.add("%s: a domain has no default", prefix)
		}
		if !param.Optional {
			p.add("%s: a domain is optional, because an app can be created before it has one", prefix)
		}
		refuseAttributes(prefix, param, p, "pattern", "minLength", "maxLength", "min", "max", "generate", "options",
			"engine")
	case ParamTypeApp:
		if param.Default != nil {
			p.add("%s: an app parameter has no default; it names an app of this installation", prefix)
		}
		if param.Engine != "" && !engineNamePattern.MatchString(param.Engine) {
			p.add("%s.engine %q must be lowercase letters, digits and dashes", prefix, param.Engine)
		}
		refuseAttributes(prefix, param, p, "pattern", "minLength", "maxLength", "min", "max", "generate", "options")
	default:
		p.add("%s.type %q is not one of %v", prefix, param.Type, AllParamTypes)
	}
}

func validateLengths(prefix string, param *Parameter, p *problems) {
	if param.MinLength != nil && *param.MinLength < 0 || param.MaxLength != nil && *param.MaxLength < 0 {
		p.add("%s.minLength and maxLength must not be negative", prefix)
	}
	if param.MinLength != nil && param.MaxLength != nil && *param.MinLength > *param.MaxLength {
		p.add("%s.minLength must not exceed maxLength", prefix)
	}
}

func validateBounds(prefix string, param *Parameter, p *problems, kind string, parse func(any) (int64, bool)) {
	minValue, minOK := parse(param.Min)
	maxValue, maxOK := parse(param.Max)
	if param.Min != nil && !minOK || param.Max != nil && !maxOK {
		p.add("%s.min and max must be %s", prefix, kind)
		return
	}
	if param.Min != nil && param.Max != nil && minValue > maxValue {
		p.add("%s.min must not exceed max", prefix)
	}
}

func validateOptions(prefix string, options []*SelectOption, p *problems) {
	if len(options) == 0 {
		p.add("%s.options must list at least one option", prefix)
		return
	}
	seen := map[string]bool{}
	for _, option := range options {
		if option == nil || option.Value == "" || option.Title == "" {
			p.add("%s.options: every option needs a value and a title", prefix)
			continue
		}
		if seen[option.Value] {
			p.add("%s.options: %q is listed twice", prefix, option.Value)
		}
		seen[option.Value] = true
	}
}

func refuseAttributes(prefix string, param *Parameter, p *problems, names ...string) {
	set := map[string]bool{
		"pattern":   param.Pattern != "",
		"minLength": param.MinLength != nil,
		"maxLength": param.MaxLength != nil,
		"min":       param.Min != nil,
		"max":       param.Max != nil,
		"generate":  param.Generate != nil,
		"options":   len(param.Options) > 0,
		"engine":    param.Engine != "",
	}
	for _, name := range names {
		if set[name] {
			p.add("%s.%s does not apply to type %s", prefix, name, param.Type)
		}
	}
}

func validateVariants(variants []*Variant, p *problems) map[string]bool {
	names := map[string]bool{}
	defaults := 0
	for i, variant := range variants {
		if variant == nil {
			p.add("variants[%d] is empty", i)
			continue
		}
		if !choiceNamePattern.MatchString(variant.Name) {
			p.add("variants[%d].name %q is invalid", i, variant.Name)
		}
		if names[variant.Name] {
			p.add("variant %q is declared twice", variant.Name)
		}
		names[variant.Name] = true
		if variant.Title == "" {
			p.add("variant %q needs a title", variant.Name)
		}
		if variant.Default {
			defaults++
		}
	}
	if len(variants) > 0 && defaults != 1 {
		p.add("exactly one variant must be the default")
	}
	return names
}

func validateVersions(t *Template, versions []*Version, variants map[string]bool, p *problems) {
	if len(versions) == 0 {
		p.add("versions must not be empty")
		return
	}
	seen := map[string]bool{}
	defaults := 0
	for i, version := range versions {
		if version == nil {
			p.add("versions[%d] is empty", i)
			continue
		}
		prefix := fmt.Sprintf("versions[%s]", version.Name)
		if !choiceNamePattern.MatchString(version.Name) {
			p.add("versions[%d].name %q is invalid", i, version.Name)
		}
		if seen[version.Name] {
			p.add("%s is declared twice", prefix)
		}
		seen[version.Name] = true
		if version.Release == "" {
			p.add("%s.release is required", prefix)
		}
		if version.Default {
			defaults++
			if version.Deprecated {
				p.add("%s: the default version cannot be deprecated", prefix)
			}
		}
		validateVersionImages(t, prefix, version, variants, p)
		if version.Override != nil && len(version.Override.App) == 0 {
			p.add("%s.override.app is empty", prefix)
		}
	}
	if defaults != 1 {
		p.add("exactly one version must be the default")
	}
}

func validateVersionImages(t *Template, prefix string, version *Version, variants map[string]bool, p *problems) {
	if t.HasComponents() {
		validateVersionComponentImages(t, prefix, version, p)
		return
	}
	if len(version.Components) > 0 {
		p.add("%s: without components, set image and not components", prefix)
	}
	if len(variants) == 0 {
		if version.Image == "" || len(version.Images) > 0 {
			p.add("%s: without variants, set image and not images", prefix)
		}
		checkPinned(prefix+".image", version.Image, p)
		return
	}
	if version.Image != "" || len(version.Images) == 0 {
		p.add("%s: with variants, set images and not image", prefix)
	}
	for _, variant := range slices.Sorted(maps.Keys(version.Images)) {
		if !variants[variant] {
			p.add("%s.images: variant %q is not declared", prefix, variant)
		}
		checkPinned(fmt.Sprintf("%s.images[%s]", prefix, variant), version.Images[variant], p)
	}
}

// validateVersionComponentImages checks a version's image table against the
// components the template declares: one pinned image each, and nothing named
// that is not a component. A component with no image in some version is a stack
// that half deploys, found on the day somebody picks that version.
func validateVersionComponentImages(t *Template, prefix string, version *Version, p *problems) {
	if version.Image != "" || len(version.Images) > 0 {
		p.add("%s: with components, set components and not image or images", prefix)
	}
	if version.Override != nil {
		// An override patches one app tree, and a template with components has
		// several. Which one it meant would have to be guessed.
		p.add("%s: a version of a template with components cannot override app", prefix)
	}
	for _, name := range slices.Sorted(maps.Keys(version.Components)) {
		if t.FindComponent(name) == nil {
			p.add("%s.components: %q is not a declared component", prefix, name)
		}
	}
	for _, component := range t.Components {
		if component == nil {
			continue
		}
		field := fmt.Sprintf("%s.components[%s].image", prefix, component.Name)
		image := version.ComponentImage(component.Name)
		if image == "" {
			p.add("%s is required", field)
			continue
		}
		checkPinned(field, image, p)
	}
}

// isPinnedTag reports whether a tag names one release. The version part is
// everything before the first separator: `18.6` of `18.6-alpine3.24`.
func isPinnedTag(tag string) bool {
	if pinnedDatePattern.MatchString(tag) || pinnedPrefixedVersionPattern.MatchString(tag) {
		return true
	}
	return pinnedVersionPattern.MatchString(tagVersion(tag))
}

func checkPinned(field, image string, p *problems) {
	if image != "" && !IsPinnedImage(image) {
		p.add("%s %q must name an exact release, not a moving tag", field, image)
	}
}

// IsPinnedImage reports whether an image reference names one release: a digest,
// or a tag carrying at least major.minor. A moving tag changes what runs without
// any template revision changing - on a redeploy, or when swarm reschedules a
// task onto another node.
func IsPinnedImage(image string) bool {
	ref := imageref.Parse(image)
	if ref.Digest != "" {
		return true
	}
	if ref.Tag == "" || ref.Tag == latestTag {
		return false
	}
	return isPinnedTag(ref.Tag)
}

// ToInt64 accepts the integer shapes a YAML or JSON decode produces.
func ToInt64(v any) (int64, bool) {
	switch n := v.(type) {
	case int:
		return int64(n), true
	case int64:
		return n, true
	case uint64:
		if n > uint64(1<<63-1) {
			return 0, false
		}
		return int64(n), true
	case float64:
		if n != float64(int64(n)) {
			return 0, false
		}
		return int64(n), true
	default:
		return 0, false
	}
}

func sizeBytes(v any) (int64, bool) {
	text, ok := v.(string)
	if !ok {
		return 0, false
	}
	size, err := unit.ParseDataSizeString(text)
	if err != nil {
		return 0, false
	}
	return size.Bytes(), true
}

// IntBounds returns an int parameter's min and max, nil where unset. It assumes
// Validate passed.
func (p *Parameter) IntBounds() (minValue, maxValue *int64) {
	if v, ok := ToInt64(p.Min); ok {
		minValue = &v
	}
	if v, ok := ToInt64(p.Max); ok {
		maxValue = &v
	}
	return minValue, maxValue
}

// SizeBounds returns a size parameter's min and max, nil where unset. It assumes
// Validate passed.
func (p *Parameter) SizeBounds() (minValue, maxValue *unit.DataSize) {
	if v, ok := sizeBytes(p.Min); ok {
		size := unit.DataSize(v)
		minValue = &size
	}
	if v, ok := sizeBytes(p.Max); ok {
		size := unit.DataSize(v)
		maxValue = &size
	}
	return minValue, maxValue
}
