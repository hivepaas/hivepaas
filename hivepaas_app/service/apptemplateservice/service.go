package apptemplateservice

import "context"

type Service interface {
	Index(ctx context.Context) (*IndexResp, error)
	Template(ctx context.Context, name string) (*TemplateResp, error)
	Icon(ctx context.Context, sha256 string) (*IconResp, error)
	// Render loads a template and renders it for creating an app. A template that
	// needs a newer HivePaaS is refused, and so is a deprecated version.
	Render(ctx context.Context, req *RenderReq) (*RenderResp, error)
}
