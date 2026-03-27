package githubappdto

import (
	"github.com/google/go-github/v79/github"
	vld "github.com/tiendc/go-validator"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/pkg/copier"
)

type ListAppInstallationReq struct {
	*GithubAppBaseReq

	Paging basedto.Paging `json:"-"`
}

func NewListAppInstallationReq() *ListAppInstallationReq {
	return &ListAppInstallationReq{}
}

func (req *ListAppInstallationReq) Validate() apperrors.ValidationErrors {
	var validators []vld.Validator
	validators = append(validators, req.validate("")...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type ListAppInstallationResp struct {
	Meta *basedto.ListMeta      `json:"meta"`
	Data []*AppInstallationResp `json:"data"`
}

type AppInstallationResp struct {
	ID      int64  `json:"id"`
	NodeID  string `json:"nodeID"`
	AppID   int64  `json:"appID"`
	AppSlug string `json:"appSlug"`
}

func TransformAppInstallation(installation *github.Installation) (resp *AppInstallationResp, err error) {
	if err = copier.Copy(&resp, &installation); err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}

func TransformAppInstallations(installations []*github.Installation) ([]*AppInstallationResp, error) {
	resp, err := basedto.TransformObjectSlice(installations, TransformAppInstallation)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	return resp, nil
}
