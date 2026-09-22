package templaterender

import (
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// CompBinding ties a component a template declares to the app that runs it. It
// is the same shape as a dependency's binding, because a component is reached
// the same way: by the key its app answers to on the project network.
type CompBinding struct {
	// AppKey is the component app's key: what ${key.VAR} names.
	AppKey string
	// SharedVars are what that app shares. Empty until the component has been
	// rendered, because what an app shares comes from the kind it declares.
	SharedVars []string
	// Rendered says the component has been rendered, and so that SharedVars is
	// what it shares rather than what is not known yet. It is what tells a
	// reference to a component rendered later from a reference to nothing.
	Rendered bool
}

// resolveComp answers comp.<name>.key with the app key, and comp.<name>.ref.<VAR>
// with an environment reference to what that component's app shares.
//
// A component may name any other, including one declared after it: the keys of
// every component are known before any of them is rendered, because a component
// app's name is the app's name and the component's. Only ref.<VAR> needs the
// other component rendered first, and that is what needs orders.
func resolveComp(template, ref string, comps map[string]*CompBinding) (any, error) {
	rest := strings.TrimPrefix(ref, "comp.")
	name, field, found := strings.Cut(rest, ".")
	binding := comps[name]
	if !found || binding == nil {
		return nil, undefinedPlaceholder(template, ref)
	}
	if field == "key" {
		return binding.AppKey, nil
	}
	variable, isRef := strings.CutPrefix(field, "ref.")
	if !isRef {
		return nil, undefinedPlaceholder(template, ref)
	}
	if !binding.Rendered {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail(
			"%s: placeholder %q reads component %q, which is rendered later: name it in needs", template, ref, name)
	}
	if !slices.Contains(binding.SharedVars, variable) {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: placeholder %q names nothing component %q shares", template, ref, name)
	}
	return "${" + binding.AppKey + "." + variable + "}", nil
}
