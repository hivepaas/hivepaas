package apptemplatedto

import (
	"fmt"
	"path"
	"regexp"

	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

// iconFilePattern reads postgres.3a18fec853.svg. A template name carries no dot,
// so the three parts cannot be confused with one another.
var iconFilePattern = regexp.MustCompile(
	fmt.Sprintf(`^([a-z0-9][a-z0-9-]{0,62})\.([0-9a-f]{%d})\.(svg|png)$`, apptemplateservice.IconHashLen))

type GetAppTemplateIconReq struct {
	// File is the last segment of the icon URL, as AppTemplateIconURL writes it.
	File string `json:"-"`

	Name         string `json:"-"`
	SHA256Prefix string `json:"-"`
	Ext          string `json:"-"`
}

func NewGetAppTemplateIconReq() *GetAppTemplateIconReq {
	return &GetAppTemplateIconReq{}
}

// ModifyRequest implements interface basedto.ReqModifier. A file name that does not
// have the icon URL's shape leaves the parts empty, which Validate refuses.
func (req *GetAppTemplateIconReq) ModifyRequest() error {
	if match := iconFilePattern.FindStringSubmatch(req.File); match != nil {
		req.Name, req.SHA256Prefix, req.Ext = match[1], match[2], match[3]
	}
	return nil
}

// Validate implements interface basedto.ReqValidator
func (req *GetAppTemplateIconReq) Validate() hperrors.ValidationErrors {
	return hperrors.NewValidationErrors(vld.Validate(
		basedto.ValidateStr(&req.Name, true, 1, templateNameMaxLen, "file")...))
}

// GetAppTemplateIconResp is written as the image itself, not as JSON.
type GetAppTemplateIconResp struct {
	Content     []byte
	ContentType string
}

// AppTemplateIconURL is where the dashboard loads a template's icon from:
// .../app-templates/icons/postgres.3a18fec853.svg.
//
// The name says which template the file is. The piece of hash changes whenever the
// icon does, which is what lets the response be cached as immutable. The route is
// public because an <img> cannot send the Authorization header.
func AppTemplateIconURL(entry *templatemodel.IndexEntry) string {
	sha := entry.Icon.SHA256
	return fmt.Sprintf("%v/app-templates/icons/%v.%v%v", config.Current().HTTPServer.BasePath,
		entry.Name, sha[:min(len(sha), apptemplateservice.IconHashLen)], path.Ext(entry.Icon.Path))
}
