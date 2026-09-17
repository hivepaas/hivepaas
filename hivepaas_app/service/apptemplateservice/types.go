package apptemplateservice

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
)

type IndexResp struct {
	Source   string
	Revision string
	Index    *templatemodel.Index
}

type TemplateResp struct {
	Source   string
	Revision string
	Entry    *templatemodel.IndexEntry
	Template *templatemodel.Template
}

type IconResp struct {
	Content     []byte
	ContentType string
}

type RenderReq struct {
	Name    string
	Version string
	Variant string
	Params  map[string]any
}

type RenderResp struct {
	TemplateResp
	Result *templaterender.Result
}

type ImageTagsReq struct {
	Name string
	// Version and Variant are the user's choice; empty means the template's default.
	Version string
	Variant string
}

type ImageTag struct {
	Tag   string
	Class templatemodel.ImageOverrideClass
	Newer bool
}

type ImageTagsResp struct {
	// Repository is where the tags came from, host included, so the dashboard can
	// say what it scanned.
	Repository string
	CurrentTag string
	// Truncated says the registry publishes more tags than were read.
	Truncated bool
	Tags      []*ImageTag
}
