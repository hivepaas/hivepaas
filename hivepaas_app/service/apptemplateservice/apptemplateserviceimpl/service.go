package apptemplateserviceimpl

import (
	"context"
	"maps"
	"path"
	"slices"
	"strings"
	"sync"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
	"github.com/hivepaas/hivepaas/services/registry"
)

const (
	contentTypeSVG = "image/svg+xml"
	contentTypePNG = "image/png"
)

// tagLister is the part of the registry client this service uses. It is an
// interface so a test can answer with a tag list instead of a fake registry.
type tagLister interface {
	ListTags(ctx context.Context, ref registry.Reference, maxTags int) (*registry.ListTagsResult, error)
}

func New(
	hpAppService hpappservice.Service,
	registryClient *registry.Client,
	logger logging.Logger,
) apptemplateservice.Service {
	return &service{
		official:       newOfficialSource(hpAppService),
		registryClient: registryClient,
		tagCache:       newImageTagsCache(),
		logger:         logger,
	}
}

type service struct {
	official       apptemplateservice.Source
	registryClient tagLister
	tagCache       *imageTagsCache
	logger         logging.Logger
	warnOnce       sync.Once
}

// source picks where templates come from. HP_TEMPLATES_DIR is read with no
// signature and no hashes, so it is honored only in development - anywhere else
// it would let whoever sets an environment variable choose what images run.
func (s *service) source() apptemplateservice.Source {
	cfg := config.Current()
	if cfg == nil || cfg.AppTemplates.Dir == "" {
		return s.official
	}
	if cfg.IsDevEnv() {
		return newLocalDirSource(cfg.AppTemplates.Dir)
	}
	s.warnOnce.Do(func() {
		s.logger.Warnf("app templates: HP_TEMPLATES_DIR is set but ignored outside the development environment")
	})
	return s.official
}

func (s *service) Index(ctx context.Context) (*apptemplateservice.IndexResp, error) {
	src := s.source()
	index, err := src.Index(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	revision, err := src.Revision(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplateservice.IndexResp{Source: src.ID(), Revision: revision, Index: index}, nil
}

func (s *service) Template(ctx context.Context, name string) (*apptemplateservice.TemplateResp, error) {
	src := s.source()
	index, err := src.Index(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	revision, err := src.Revision(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	entry, tmpl, err := loadTemplate(ctx, src, index, name)
	if err != nil {
		return nil, err
	}
	resp := &apptemplateservice.TemplateResp{Source: src.ID(), Revision: revision, Entry: entry, Template: tmpl}
	// The same index, so the same revision: a template and what it depends on are
	// never read from two different pins.
	for _, dep := range tmpl.Dependencies {
		depEntry, depTmpl, err := loadTemplate(ctx, src, index, dep.Template)
		if err != nil {
			return nil, err
		}
		resp.Dependencies = append(resp.Dependencies,
			&apptemplateservice.DependencyTemplate{Dependency: dep, Entry: depEntry, Template: depTmpl})
	}
	return resp, nil
}

func loadTemplate(
	ctx context.Context,
	src apptemplateservice.Source,
	index *templatemodel.Index,
	name string,
) (*templatemodel.IndexEntry, *templatemodel.Template, error) {
	entry := index.FindTemplate(name)
	if entry == nil {
		return nil, nil, hperrors.Wrap(hperrors.ErrAppTemplateNotFound).WithParam("Name", name)
	}
	data, err := src.TemplateFile(ctx, entry)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	tmpl, err := templatemodel.DecodeTemplate(data)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	return entry, tmpl, nil
}

func (s *service) Icon(ctx context.Context, req *apptemplateservice.IconReq) (*apptemplateservice.IconResp, error) {
	src := s.source()
	index, err := src.Index(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	entry := index.FindTemplate(req.Name)
	if entry == nil || len(req.SHA256Prefix) != apptemplateservice.IconHashLen ||
		!strings.HasPrefix(entry.Icon.SHA256, req.SHA256Prefix) || path.Ext(entry.Icon.Path) != "."+req.Ext {
		return nil, hperrors.NewNotFound("Icon")
	}
	content, err := src.Icon(ctx, entry.Icon.SHA256)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	contentType := contentTypePNG
	if req.Ext == "svg" {
		contentType = contentTypeSVG
	}
	return &apptemplateservice.IconResp{Content: content, ContentType: contentType}, nil
}

func (s *service) Render(
	ctx context.Context,
	req *apptemplateservice.RenderReq,
) (*apptemplateservice.RenderResp, error) {
	loaded, err := s.Template(ctx, req.Name)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if !templatemodel.IsCompatible(loaded.Template.Metadata.Requires, base.CurrentVersion) {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateIncompatible).WithParam("Name", req.Name)
	}
	if len(loaded.Dependencies) > 0 {
		return renderWithDependencies(loaded, req)
	}
	if len(req.DependencyParams) > 0 {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateParamInvalid).
			WithExtraDetail("%s has no dependencies to give parameters to", req.Name)
	}
	result, err := templaterender.Render(&templaterender.Request{
		Template: loaded.Template,
		Version:  req.Version,
		Variant:  req.Variant,
		Params:   req.Params,
		ImageTag: req.ImageTag,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplateservice.RenderResp{TemplateResp: *loaded, Result: result}, nil
}

// renderWithDependencies renders every app of the request before any of them
// exists: the dependencies first, because the template refers to what they share
// and their parameters can refer to the template's.
func renderWithDependencies(
	loaded *apptemplateservice.TemplateResp,
	req *apptemplateservice.RenderReq,
) (*apptemplateservice.RenderResp, error) {
	tmpl := loaded.Template
	name := tmpl.Metadata.Name
	if req.AppName == "" {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: its dependencies are named after the app, and no app name was given", name)
	}
	for _, depName := range slices.Sorted(maps.Keys(req.DependencyParams)) {
		if tmpl.FindDependency(depName) == nil {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateParamInvalid).
				WithExtraDetail("%s has no dependency %q", name, depName)
		}
	}
	owner, err := templaterender.ResolveParams(tmpl.Parameters, req.Params)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	resp := &apptemplateservice.RenderResp{TemplateResp: *loaded}
	bindings := make(map[string]*templaterender.DepBinding, len(loaded.Dependencies))
	keys := map[string]string{projecthelper.CalcAppKey(req.AppName): req.AppName}
	for _, dep := range loaded.Dependencies {
		rendered, binding, err := renderDependency(loaded, dep, req, owner, keys)
		if err != nil {
			return nil, err
		}
		bindings[dep.Dependency.Name] = binding
		resp.Dependencies = append(resp.Dependencies, rendered)
	}

	result, err := templaterender.Render(&templaterender.Request{
		Template:       tmpl,
		Version:        req.Version,
		Variant:        req.Variant,
		ResolvedParams: owner,
		Deps:           bindings,
		ImageTag:       req.ImageTag,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp.Result = result
	return resp, nil
}

// renderDependency renders one dependency and binds it to the key its app will
// have. keys holds the keys already taken by this request: slugifying truncates,
// so two long names can meet, and that is refused rather than left to the second
// provision to discover.
func renderDependency(
	loaded *apptemplateservice.TemplateResp,
	dep *apptemplateservice.DependencyTemplate,
	req *apptemplateservice.RenderReq,
	owner map[string]*templaterender.Value,
	keys map[string]string,
) (*apptemplateservice.RenderedDependency, *templaterender.DepBinding, error) {
	name := loaded.Template.Metadata.Name
	depTmpl := dep.Template
	if !templatemodel.IsCompatible(depTmpl.Metadata.Requires, base.CurrentVersion) {
		return nil, nil, hperrors.Wrap(hperrors.ErrAppTemplateIncompatible).WithParam("Name", depTmpl.Metadata.Name)
	}
	if len(depTmpl.Dependencies) > 0 {
		return nil, nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail(
			"%s: dependency %q has dependencies of its own", name, dep.Dependency.Name)
	}

	appName := templatemodel.DependencyAppName(req.AppName, dep.Dependency.Name)
	appKey := projecthelper.CalcAppKey(appName)
	if other, taken := keys[appKey]; taken {
		return nil, nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail(
			"%s: the apps %q and %q would have the same key; choose a shorter name", name, other, appName)
	}
	keys[appKey] = appName

	input, err := templaterender.DependencyParams(name, dep.Dependency, depTmpl, owner,
		req.DependencyParams[dep.Dependency.Name])
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	result, err := templaterender.Render(&templaterender.Request{
		Template: depTmpl, Version: dep.Dependency.Version, Variant: dep.Dependency.Variant, Params: input,
	})
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	rendered := &apptemplateservice.RenderedDependency{
		Name:    dep.Dependency.Name,
		AppName: appName,
		Render: &apptemplateservice.RenderResp{
			TemplateResp: apptemplateservice.TemplateResp{
				Source: loaded.Source, Revision: loaded.Revision, Entry: dep.Entry, Template: depTmpl,
			},
			Result: result,
		},
	}
	binding := &templaterender.DepBinding{AppKey: appKey, SharedVars: templaterender.SharedVarsOf(result.Doc)}
	return rendered, binding, nil
}
