package apptemplateserviceimpl

import (
	"context"
	"path"
	"sync"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/hpappservice"
)

const (
	contentTypeSVG = "image/svg+xml"
	contentTypePNG = "image/png"
)

func New(
	hpAppService hpappservice.Service,
	logger logging.Logger,
) apptemplateservice.Service {
	return &service{
		official: newOfficialSource(hpAppService),
		logger:   logger,
	}
}

type service struct {
	official apptemplateservice.Source
	logger   logging.Logger
	warnOnce sync.Once
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
	entry := index.FindTemplate(name)
	if entry == nil {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateNotFound).WithParam("Name", name)
	}
	data, err := src.TemplateFile(ctx, entry)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	tmpl, err := templatemodel.DecodeTemplate(data)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	revision, err := src.Revision(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplateservice.TemplateResp{Source: src.ID(), Revision: revision, Entry: entry, Template: tmpl}, nil
}

func (s *service) Icon(ctx context.Context, sha256Hex string) (*apptemplateservice.IconResp, error) {
	src := s.source()
	index, err := src.Index(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	entry := index.FindIcon(sha256Hex)
	if entry == nil {
		return nil, hperrors.NewNotFound("Icon")
	}
	content, err := src.Icon(ctx, sha256Hex)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	contentType := contentTypePNG
	if path.Ext(entry.Icon.Path) == ".svg" {
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
	result, err := templaterender.Render(&templaterender.Request{
		Template: loaded.Template,
		Version:  req.Version,
		Variant:  req.Variant,
		Params:   req.Params,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &apptemplateservice.RenderResp{TemplateResp: *loaded, Result: result}, nil
}
