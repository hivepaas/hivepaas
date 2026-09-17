package apptemplateservice

import "context"

type Service interface {
	Index(ctx context.Context) (*IndexResp, error)
	Template(ctx context.Context, name string) (*TemplateResp, error)
	// Icon serves a template's icon only while the index still lists it with that
	// hash and extension, so a URL never starts answering with different bytes.
	Icon(ctx context.Context, req *IconReq) (*IconResp, error)
	// Render loads a template and renders it for creating an app. A template that
	// needs a newer HivePaaS is refused, and so is a deprecated version.
	Render(ctx context.Context, req *RenderReq) (*RenderResp, error)

	// ImageTags lists the tags a user could use instead of the one this template
	// version pins. It reads the registry, so it is slow, it can fail, and it is
	// called only when somebody asks for it.
	ImageTags(ctx context.Context, req *ImageTagsReq) (*ImageTagsResp, error)
}
