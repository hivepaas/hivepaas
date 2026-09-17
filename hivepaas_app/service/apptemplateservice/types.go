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

// IconHashLen is how much of an icon's sha256 its file name carries. The name has
// to change whenever the icon does, so a response can be cached as immutable, and
// 40 bits of hash is far beyond what one template's icon history will ever collide
// on. The rest of the hash stays in the index, where verification needs it.
const IconHashLen = 10

// IconReq names an icon the way its URL does: postgres.3a18fec853.svg.
type IconReq struct {
	Name         string
	SHA256Prefix string
	Ext          string
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
	// ImageOverride is an image the user chose instead of the template's, empty to
	// use the template's own.
	ImageOverride string
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
