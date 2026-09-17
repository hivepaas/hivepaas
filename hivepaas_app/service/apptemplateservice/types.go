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
