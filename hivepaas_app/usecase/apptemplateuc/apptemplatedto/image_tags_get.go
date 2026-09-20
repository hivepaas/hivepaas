package apptemplatedto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice"
)

type GetAppTemplateImageTagsReq struct {
	Name string `json:"-"`

	Version string `json:"-" mapstructure:"version"`
	Variant string `json:"-" mapstructure:"variant"`
}

func NewGetAppTemplateImageTagsReq() *GetAppTemplateImageTagsReq {
	return &GetAppTemplateImageTagsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *GetAppTemplateImageTagsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 4) //nolint:mnd
	validators = append(validators, basedto.ValidateStr(&req.Name, true, 1, templateNameMaxLen, "name")...)
	validators = append(validators, basedto.ValidateStr(&req.Version, false, 1, choiceNameMaxLen, "version")...)
	validators = append(validators, basedto.ValidateStr(&req.Variant, false, 1, choiceNameMaxLen, "variant")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetAppTemplateImageTagsResp struct {
	Meta *basedto.Meta             `json:"meta"`
	Data *AppTemplateImageTagsResp `json:"data"`
}

// AppTemplateImageTagsResp is what the registry publishes for the image this
// template version pins.
type AppTemplateImageTagsResp struct {
	Repository string `json:"repository"`
	CurrentTag string `json:"currentTag"`
	// Truncated says the repository has more tags than were read.
	Truncated bool                       `json:"truncated"`
	Tags      []*AppTemplateImageTagResp `json:"tags"`
}

// AppTemplateImageTagResp is one tag to offer. Tag is what goes back as
// imageTag when the app is created: the repository is the template's, so there
// is no reference for a caller to assemble and none to get wrong.
type AppTemplateImageTagResp struct {
	Tag string `json:"tag"`
	// Class is one of same-line, other-major or moving, and decides which warning
	// the dashboard shows before creating the app.
	Class string `json:"class"`
	Newer bool   `json:"newer"`
}

func TransformAppTemplateImageTags(tags *apptemplateservice.ImageTagsResp) *AppTemplateImageTagsResp {
	resp := &AppTemplateImageTagsResp{
		Repository: tags.Repository,
		CurrentTag: tags.CurrentTag,
		Truncated:  tags.Truncated,
		Tags:       make([]*AppTemplateImageTagResp, 0, len(tags.Tags)),
	}
	for _, tag := range tags.Tags {
		resp.Tags = append(resp.Tags, &AppTemplateImageTagResp{
			Tag:   tag.Tag,
			Class: string(tag.Class),
			Newer: tag.Newer,
		})
	}
	return resp
}
