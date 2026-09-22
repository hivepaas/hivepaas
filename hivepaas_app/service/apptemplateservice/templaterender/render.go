package templaterender

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const yamlIndent = 2

type Request struct {
	Template *templatemodel.Template
	// Version and Variant are the user's choice; empty means the template's default.
	Version string
	Variant string
	Params  map[string]any
	// AllowDeprecated renders a deprecated version, which creating an app never
	// does. The linter and phase 2's updates of existing apps do.
	AllowDeprecated bool
	// ImageTag replaces the tag of the image this version pins, for this app only.
	// A tag rather than a reference: the repository is the template's and is never
	// the caller's to choose, so it is not asked for and cannot be got wrong.
	ImageTag string
	// ResolvedParams, when set, are used instead of resolving Params. A caller that
	// renders a template's dependencies from its parameters resolves them first,
	// and this render must use the same generated secrets.
	ResolvedParams map[string]*Value
	// Deps binds every dependency the template declares to its app.
	Deps map[string]*DepBinding
	// Component names which component of the template to render, for a template
	// that declares several. It is empty, and must be, for a template with one
	// app.
	Component string
	// Comps binds the components the template declares to their apps. Every
	// component is in it from the start, because their keys are known before any
	// of them is rendered; SharedVars fill in as each one renders.
	Comps map[string]*CompBinding
}

type Result struct {
	// Doc is what the app is built from, secrets included.
	Doc     *specmodel.AppDoc
	Params  map[string]*Value
	Version *templatemodel.Version
	// Variant is nil for a template without variants.
	Variant *templatemodel.Variant
	Image   string
	// ImageOverride is the image the user chose instead of Image, empty when the
	// template's own image is in use. Image stays what the template pinned, because
	// the version and release recorded on the app describe the template.
	ImageOverride      string
	ImageOverrideClass templatemodel.ImageOverrideClass

	// Base is the same render with every secret placeholder left in, as canonical
	// YAML: the base phase 2's three-way merge compares against. It is stored, so
	// it must never hold a secret, and it is hashed, so it must be byte-stable.
	Base       []byte
	BaseSHA256 string
}

// Render turns a template and a user's choices into an AppDoc.
//
// Every step is pure, in this order: validate, choose version and variant,
// resolve parameters, apply the version override, substitute placeholders,
// decode strictly as an AppDoc, refuse what cannot be built. The base render
// repeats the substitution with secrets kept as placeholders.
func Render(req *Request) (*Result, error) {
	tmpl := req.Template
	if err := tmpl.Validate(""); err != nil {
		return nil, hperrors.Wrap(err)
	}
	component, err := selectComponent(tmpl, req)
	if err != nil {
		return nil, err
	}
	version, variant, image, err := selectVersion(tmpl, req)
	if err != nil {
		return nil, err
	}
	params := req.ResolvedParams
	if params == nil {
		if params, err = ResolveParams(tmpl.Parameters, req.Params); err != nil {
			return nil, err
		}
	}
	for _, dep := range tmpl.Dependencies {
		if req.Deps[dep.Name] == nil {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
				WithExtraDetail("%s: dependency %q is not bound to an app", tmpl.Metadata.Name, dep.Name)
		}
	}

	appTree := tmpl.App
	if component != nil {
		appTree = component.App
	}
	tree := deepCopy(appTree)
	if version.Override != nil {
		// TODO: app templates phase 3 - apply the variant's override after the
		// version's. See docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
		tree = MergePatch(tree, version.Override.App)
	}
	vars := templateVars(version, variant, image)
	name := tmpl.Metadata.Name

	applied, err := Substitute(tree, resolver(name, vars, params, req, false))
	if err != nil {
		return nil, err
	}
	doc, err := decodeAppDoc(name, applied)
	if err != nil {
		return nil, err
	}
	if err = specmodel.CheckBuildable(doc); err != nil {
		return nil, hperrors.Wrap(err)
	}
	override, err := overrideImageFor(image, req.ImageTag)
	if err != nil {
		return nil, err
	}
	overrideClass, err := applyImageOverride(doc, name, version.Name, image, override)
	if err != nil {
		return nil, err
	}

	baseTree, err := Substitute(tree, resolver(name, vars, params, req, true))
	if err != nil {
		return nil, err
	}
	base, err := marshalCanonical(baseTree)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(base)

	return &Result{
		Doc:                doc,
		Params:             params,
		Version:            version,
		Variant:            variant,
		Image:              image,
		ImageOverride:      override,
		ImageOverrideClass: overrideClass,
		Base:               base,
		BaseSHA256:         hex.EncodeToString(sum[:]),
	}, nil
}

// overrideImageFor is the reference a chosen tag means: that tag on the
// repository the template pinned. Empty when no tag was chosen.
//
// Composing it here rather than taking it from the caller is what makes a
// different repository unrepresentable instead of merely refused.
func overrideImageFor(templateImage, tag string) (string, error) {
	if tag == "" {
		return "", nil
	}
	if !templatemodel.ValidImageTag(tag) {
		return "", hperrors.Wrap(hperrors.ErrAppTemplateImageNotAllowed).
			WithParam("Image", tag).WithParam("Repository", imageref.Parse(templateImage).Repository).
			WithExtraDetail("%s: this is a tag, and a tag is letters, digits, dots, dashes and underscores", tag)
	}
	return imageref.Parse(templateImage).Repository + ":" + tag, nil
}

// applyImageOverride swaps the image in the document the app is built from.
//
// It runs after CheckBuildable, so the document has already been accepted as
// something this HivePaaS can build; all that changes is which release of the same
// repository it deploys. It does not touch the tree the base render walks - the
// base is the template's own output, and phase 2 needs to see the override as the
// user's change rather than as something the template did.
func applyImageOverride(
	doc *specmodel.AppDoc,
	templateName, versionName, templateImage, override string,
) (templatemodel.ImageOverrideClass, error) {
	if override == "" {
		return "", nil
	}
	class, err := templatemodel.ClassifyImageOverride(versionName, templateImage, override)
	if err != nil {
		return "", hperrors.Wrap(err)
	}

	var imageSource map[string]any
	if doc.Deployment != nil && doc.Deployment.Source != nil {
		imageSource, _ = doc.Deployment.Source["imageSource"].(map[string]any)
	}
	if imageSource == nil {
		return "", hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: the template deploys no image to replace", templateName)
	}
	imageSource["image"] = override
	return class, nil
}

// selectComponent is the component a request renders, nil for a template with
// one app. A template with components renders one of them at a time: each is its
// own app, with its own image, its own document and its own base.
func selectComponent(tmpl *templatemodel.Template, req *Request) (*templatemodel.Component, error) {
	name := tmpl.Metadata.Name
	if !tmpl.HasComponents() {
		if req.Component != "" {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
				WithExtraDetail("%s: has no components, and %q was asked for", name, req.Component)
		}
		return nil, nil
	}
	if req.Component == "" {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: creates several apps, and no component was named", name)
	}
	component := tmpl.FindComponent(req.Component)
	if component == nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: has no component %q", name, req.Component)
	}
	return component, nil
}

func selectVersion(tmpl *templatemodel.Template, req *Request) (
	*templatemodel.Version, *templatemodel.Variant, string, error) {
	name := tmpl.Metadata.Name
	version := tmpl.DefaultVersion()
	if req.Version != "" {
		version = tmpl.FindVersion(req.Version)
	}
	if version == nil {
		return nil, nil, "", hperrors.Wrap(hperrors.ErrAppTemplateVersionNotFound).
			WithParam("Template", name).WithParam("Version", req.Version)
	}
	if version.Deprecated && !req.AllowDeprecated {
		return nil, nil, "", hperrors.Wrap(hperrors.ErrAppTemplateVersionDeprecated).
			WithParam("Template", name).WithParam("Version", version.Name)
	}

	if tmpl.HasComponents() {
		if req.Variant != "" {
			return nil, nil, "", variantUnavailable(name, version.Name, req.Variant)
		}
		image := version.ComponentImage(req.Component)
		if image == "" {
			return nil, nil, "", hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail(
				"%s: version %q pins no image for component %q", name, version.Name, req.Component)
		}
		return version, nil, image, nil
	}

	if len(tmpl.Variants) == 0 {
		if req.Variant != "" {
			return nil, nil, "", variantUnavailable(name, version.Name, req.Variant)
		}
		return version, nil, version.Image, nil
	}

	variant := tmpl.DefaultVariant()
	if req.Variant != "" {
		variant = tmpl.FindVariant(req.Variant)
	}
	if variant == nil || version.ImageFor(variant.Name) == "" {
		wanted := req.Variant
		if wanted == "" && variant != nil {
			wanted = variant.Name
		}
		return nil, nil, "", variantUnavailable(name, version.Name, wanted)
	}
	return version, variant, version.ImageFor(variant.Name), nil
}

func variantUnavailable(template, version, variant string) error {
	return hperrors.Wrap(hperrors.ErrAppTemplateVariantUnavailable).
		WithParam("Template", template).WithParam("Version", version).WithParam("Variant", variant)
}

// templateVars are the placeholders a template can use besides its parameters.
//
// TODO: app templates phase 3 - an app namespace (app.name, app.key). See
// docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
func templateVars(version *templatemodel.Version, variant *templatemodel.Variant, image string) map[string]any {
	vars := map[string]any{
		"version.name":    version.Name,
		"version.release": version.Release,
		"variant.name":    "",
		"image":           image,
	}
	if variant != nil {
		vars["variant.name"] = variant.Name
	}
	for key, value := range version.Vars {
		vars["version.vars."+key] = value
	}
	return vars
}

// resolver answers placeholders from the version, the variant, the resolved
// parameters and the bindings of the dependencies and sibling components.
// keepSecrets leaves secret placeholders as they are - the base render.
func resolver(
	template string,
	vars map[string]any,
	params map[string]*Value,
	req *Request,
	keepSecrets bool,
) Resolver {
	return func(ref string) (any, bool, error) {
		if strings.HasPrefix(ref, "deps.") {
			value, err := resolveDep(template, ref, req.Deps)
			return value, false, err
		}
		if strings.HasPrefix(ref, "comp.") {
			value, err := resolveComp(template, ref, req.Comps)
			return value, false, err
		}
		if paramName, isParam := strings.CutPrefix(ref, "params."); isParam {
			value, found := params[paramName]
			if !found {
				return nil, false, undefinedPlaceholder(template, ref)
			}
			if keepSecrets && value.Param.Type == templatemodel.ParamTypeSecret {
				return nil, true, nil
			}
			if value.Value == nil {
				return "", false, nil
			}
			return value.Value, false, nil
		}
		value, found := vars[ref]
		if !found {
			return nil, false, undefinedPlaceholder(template, ref)
		}
		return value, false, nil
	}
}

func undefinedPlaceholder(template, ref string) error {
	return hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
		WithExtraDetail("%s: placeholder %q refers to nothing", template, ref)
}

func decodeAppDoc(template string, tree any) (*specmodel.AppDoc, error) {
	encoded, err := yaml.Marshal(tree)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	doc := &specmodel.AppDoc{}
	decoder := yaml.NewDecoder(bytes.NewReader(encoded))
	decoder.KnownFields(true)
	if err = decoder.Decode(doc); err != nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("%s: app: %s", template, err.Error())
	}
	return doc, nil
}

// marshalCanonical writes a tree as YAML. yaml.v3 sorts map keys, which is what
// makes the same render hash the same every time.
func marshalCanonical(tree any) ([]byte, error) {
	var buf bytes.Buffer
	encoder := yaml.NewEncoder(&buf)
	encoder.SetIndent(yamlIndent)
	if err := encoder.Encode(tree); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err := encoder.Close(); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return buf.Bytes(), nil
}
