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
	// Dependencies are the templates this one depends on, loaded from the same
	// revision, in declaration order.
	Dependencies []*DependencyTemplate
}

type DependencyTemplate struct {
	Dependency *templatemodel.Dependency
	Entry      *templatemodel.IndexEntry
	Template   *templatemodel.Template
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
	Name string
	// AppName is the name the app will have. A template with dependencies needs
	// it: each dependency's app is named after it, and referred to by key.
	AppName string
	Version string
	Variant string
	Params  map[string]any
	// DependencyParams are what the person was asked for each dependency, by the
	// dependency's name.
	DependencyParams map[string]map[string]any
	// ImageTag is a tag the user chose instead of the one the template's version
	// pins, empty to use the template's own. It is a tag and not a reference: the
	// repository comes from the template.
	ImageTag string
}

type RenderResp struct {
	TemplateResp
	Result *templaterender.Result
	// Dependencies are rendered, and created, before the app that needs them. The
	// field shadows TemplateResp.Dependencies, the templates they were rendered
	// from, which stay reachable through the embedded TemplateResp.
	Dependencies []*RenderedDependency
}

type RenderedDependency struct {
	// Name is the dependency's role in the template that declares it.
	Name    string
	AppName string
	Render  *RenderResp
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
