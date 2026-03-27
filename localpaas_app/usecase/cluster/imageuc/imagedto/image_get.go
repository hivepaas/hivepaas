package imagedto

import (
	"time"

	"github.com/docker/docker/api/types/image"
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
)

type GetImageReq struct {
	ImageID string `json:"-"`
}

func NewGetImageReq() *GetImageReq {
	return &GetImageReq{}
}

func (req *GetImageReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	// NOTE: image id is docker id, it's not ULID
	validators = append(validators, basedto.ValidateStr(&req.ImageID, true, 1, imageIDMaxLen, "imageID")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetImageResp struct {
	Meta *basedto.Meta `json:"meta"`
	Data *ImageResp    `json:"data"`
}

type ImageResp struct {
	ID        string            `json:"id"`
	Labels    map[string]string `json:"labels"`
	Size      int64             `json:"size"`
	RepoTags  []string          `json:"repoTags"`
	CreatedAt time.Time         `json:"createdAt"`
	UpdateVer int               `json:"updateVer"`
}

func TransformImage(img *image.Summary, _ bool) *ImageResp {
	resp := &ImageResp{
		ID:        img.ID,
		Labels:    img.Labels,
		Size:      img.Size,
		RepoTags:  img.RepoTags,
		CreatedAt: time.Unix(img.Created, 0),
	}
	return resp
}

func TransformImageFromResp(img *image.InspectResponse, _ bool) *ImageResp {
	resp := &ImageResp{
		ID:        img.ID,
		Size:      img.Size,
		RepoTags:  img.RepoTags,
		CreatedAt: transformImageCreatedAt(img.Created),
	}
	return resp
}

func transformImageCreatedAt(createdAt string) time.Time {
	t, err := time.Parse(time.RFC3339Nano, createdAt)
	if err == nil {
		return t
	}
	return time.Time{}
}
