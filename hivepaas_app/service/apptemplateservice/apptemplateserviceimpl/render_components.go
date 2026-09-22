package apptemplateserviceimpl

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
)

// renderComponents renders every app of a template that creates several, in the
// order its needs put them, and returns the primary component's render as the
// response's own Result.
//
// The primary component is the app the person named; the others are that name
// with the component's added, exactly as a dependency's app is. So somebody who
// types "supabase" gets an app called supabase, which is the one they meant, and
// supabase-auth beside it.
//
// The parameters are resolved once, before any of this, and passed to every
// component: that is what makes one generated secret reach ten apps, which is
// the whole reason components exist.
func renderComponents(
	loaded *apptemplateservice.TemplateResp,
	req *apptemplateservice.RenderReq,
	owner map[string]*templaterender.Value,
	deps map[string]*templaterender.DepBinding,
	keys map[string]string,
	resp *apptemplateservice.RenderResp,
) (*apptemplateservice.RenderResp, error) {
	tmpl := loaded.Template
	name := tmpl.Metadata.Name
	ordered, err := tmpl.ComponentOrder()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// Every component's key before any of them renders, so that one can name
	// another whatever order they are in.
	comps := make(map[string]*templaterender.CompBinding, len(ordered))
	appNames := make(map[string]string, len(ordered))
	for _, component := range tmpl.Components {
		appName := componentAppName(req.AppName, component)
		appKey := projecthelper.CalcAppKey(appName)
		if other, taken := keys[appKey]; taken && other != appName {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail(
				"%s: the apps %q and %q would have the same key; choose a shorter name", name, other, appName)
		}
		keys[appKey] = appName
		appNames[component.Name] = appName
		comps[component.Name] = &templaterender.CompBinding{AppKey: appKey}
	}

	rendered := make([]*apptemplateservice.RenderedComponent, 0, len(ordered))
	for _, component := range ordered {
		result, renderErr := templaterender.Render(&templaterender.Request{
			Template:       tmpl,
			Version:        req.Version,
			ResolvedParams: owner,
			Deps:           deps,
			Component:      component.Name,
			Comps:          comps,
			ImageTag:       imageTagFor(req, component),
		})
		if renderErr != nil {
			return nil, hperrors.Wrap(renderErr)
		}
		binding := comps[component.Name]
		binding.SharedVars = templaterender.SharedVarsOf(result.Doc)
		binding.Rendered = true

		rendered = append(rendered, &apptemplateservice.RenderedComponent{
			Name:    component.Name,
			Title:   component.Title,
			AppName: appNames[component.Name],
			Primary: component.Primary,
			Result:  result,
		})
		if component.Primary {
			resp.Result = result
		}
	}
	if resp.Result == nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: no component is primary", name)
	}
	resp.Components = rendered
	return resp, nil
}

// componentAppName is what the app of one component is called: the name the
// person gave for the primary component, and that name with the component's
// added for the rest.
func componentAppName(appName string, component *templatemodel.Component) string {
	if component.Primary {
		return appName
	}
	return templatemodel.ComponentAppName(appName, component.Name)
}

// imageTagFor applies a chosen tag to the primary component only. A stack is
// released as a set - the version table pins every component of one release
// together - so a tag the person chose is about the application, which is the
// app they picked it on.
func imageTagFor(req *apptemplateservice.RenderReq, component *templatemodel.Component) string {
	if component.Primary {
		return req.ImageTag
	}
	return ""
}
