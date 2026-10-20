package templaterender

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

// VersionTree is the template's app tree for a version, with the version's
// override applied and nothing substituted: what an update of an existing app
// reads the parts that carry no placeholder from. A deprecated version has one:
// an app that runs it is updated, never created.
func VersionTree(tmpl *templatemodel.Template, version string) (any, error) {
	if tmpl.HasComponents() {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: a template with components has no one app tree", tmpl.Metadata.Name)
	}
	found := tmpl.FindVersion(version)
	if found == nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateVersionNotFound).
			WithParam("Template", tmpl.Metadata.Name).WithParam("Version", version)
	}
	tree := deepCopy(tmpl.App)
	if found.Override != nil {
		tree = MergePatch(tree, found.Override.App)
	}
	return tree, nil
}

// Static is a part of a template that may hold no placeholder, as an app has
// it: escapes undone, and a placeholder refused.
func Static(node any) (any, error) {
	return Substitute(node, func(ref string) (any, bool, error) {
		return nil, false, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("${{ %s }} where no placeholder may be", ref)
	})
}
