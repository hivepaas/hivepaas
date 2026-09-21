package templaterender

import (
	"maps"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// DepBinding ties a dependency a template declares to the app that serves it.
type DepBinding struct {
	// AppKey is the dependency app's key: what ${key.VAR} names.
	AppKey string
	// SharedVars are what that app shares. A reference to anything else is
	// refused, because it would resolve to nothing at deploy time.
	SharedVars []string
}

// resolveDep answers deps.<name>.key with the app key, and deps.<name>.ref.<VAR>
// with an environment reference to what that app shares.
func resolveDep(template, ref string, deps map[string]*DepBinding) (any, error) {
	rest := strings.TrimPrefix(ref, "deps.")
	name, field, found := strings.Cut(rest, ".")
	binding := deps[name]
	if !found || binding == nil {
		return nil, undefinedPlaceholder(template, ref)
	}
	if field == "key" {
		return binding.AppKey, nil
	}
	variable, isRef := strings.CutPrefix(field, "ref.")
	if !isRef || !slices.Contains(binding.SharedVars, variable) {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: placeholder %q names nothing dependency %q shares", template, ref, name)
	}
	return "${" + binding.AppKey + "." + variable + "}", nil
}

// DependencyParams builds a dependency's parameter input: the values the
// declaring template fixes, with its own parameter placeholders substituted, and
// what the person creating the app was asked. Anything else sent for the
// dependency is refused - the dialog and the template would disagree about what
// it needs.
func DependencyParams(
	owner string,
	ownerAppKey string,
	dep *templatemodel.Dependency,
	depTemplate *templatemodel.Template,
	ownerParams map[string]*Value,
	input map[string]any,
) (map[string]any, error) {
	asked := map[string]bool{}
	for _, param := range dep.AskedParams(depTemplate) {
		asked[param.Name] = true
	}
	for _, name := range slices.Sorted(maps.Keys(input)) {
		if !asked[name] {
			return nil, paramInvalid(dep.Name+"."+name, "the dependency does not ask for it")
		}
	}

	fixed, err := Substitute(dep.Params, ownerParamResolver(owner, ownerAppKey, ownerParams))
	if err != nil {
		return nil, err
	}
	params := map[string]any{}
	maps.Copy(params, input)
	fixedParams, _ := fixed.(map[string]any)
	maps.Copy(params, fixedParams)
	return params, nil
}

// ownerParamResolver answers params.* of the declaring template and app.key, and
// refuses a secret: a secret belongs to one app, and copying it into another's
// parameters would store it twice, possibly in the clear.
//
// app.key is the app being created from this template, which its dependencies
// cannot name any other way - a dependency template is written without knowing
// that an owner exists. It is what a dependency asked to work on its owner's
// storage is given, and it is known before either app exists because the key
// comes from the name.
func ownerParamResolver(owner, ownerAppKey string, ownerParams map[string]*Value) Resolver {
	return func(ref string) (any, bool, error) {
		if ref == ownerAppKeyRef {
			if ownerAppKey == "" {
				return nil, false, undefinedPlaceholder(owner, ref)
			}
			return ownerAppKey, false, nil
		}
		name, isParam := strings.CutPrefix(ref, "params.")
		value, found := ownerParams[name]
		if !isParam || !found {
			return nil, false, undefinedPlaceholder(owner, ref)
		}
		if value.Param.Type == templatemodel.ParamTypeSecret {
			return nil, false, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail(
				"%s: a dependency parameter cannot take the secret %q", owner, ref)
		}
		if value.Value == nil {
			return "", false, nil
		}
		return value.Value, false, nil
	}
}

// ownerAppKeyRef is how a dependency's parameters name the app they are created
// for.
const ownerAppKeyRef = "app.key"

// KindCategory is the kind a rendered document declares, empty when it declares
// none.
func KindCategory(doc *specmodel.AppDoc) base.AppCategory {
	kind, _ := doc.Settings[specmodel.SingletonBlockName(base.SettingTypeAppKind)].(map[string]any)
	category, _ := kind["category"].(string)
	return base.AppCategory(category)
}

// SharedVarsOf is everything an app built from doc shares with other apps.
func SharedVarsOf(doc *specmodel.AppDoc) []string {
	return append(slices.Clone(base.AppCommonSharedEnvVars), base.AppKindSharedEnvVars(KindCategory(doc))...)
}
