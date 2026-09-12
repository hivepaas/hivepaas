package appplacementsettingsdto

import (
	vld "github.com/tiendc/go-validator"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

// maxNodeLabelSelectors bounds a list that is ANDed into every app's placement.
const maxNodeLabelSelectors = 20

type UpdateAppPlacementSettingsReq struct {
	settings.UpdateUniqueSettingReq
	*AppPlacementSettingsBaseReq
}

type AppPlacementSettingsBaseReq struct {
	ExcludeManagerNodes bool     `json:"excludeManagerNodes"`
	ExcludeBuildNodes   bool     `json:"excludeBuildNodes"`
	RequireNodeLabels   []string `json:"requireNodeLabels"`
	ExcludeNodeLabels   []string `json:"excludeNodeLabels"`
}

func (req *AppPlacementSettingsBaseReq) ToEntity() *entity.AppPlacementSettings {
	if req == nil {
		return nil
	}
	return &entity.AppPlacementSettings{
		ExcludeManagerNodes: req.ExcludeManagerNodes,
		ExcludeBuildNodes:   req.ExcludeBuildNodes,
		RequireNodeLabels:   req.RequireNodeLabels,
		ExcludeNodeLabels:   req.ExcludeNodeLabels,
	}
}

func (req *AppPlacementSettingsBaseReq) validate(_ string) []vld.Validator {
	res := make([]vld.Validator, 0, maxNodeLabelSelectors)
	res = append(res, validateNodeLabelSelectors(req.RequireNodeLabels, "requireNodeLabels")...)
	res = append(res, validateNodeLabelSelectors(req.ExcludeNodeLabels, "excludeNodeLabels")...)
	return res
}

// validateNodeLabelSelectors refuses what could not become a constraint, so
// that a typo is answered here rather than by tasks that never schedule.
func validateNodeLabelSelectors(list []string, field string) []vld.Validator {
	res := basedto.ValidateSliceEx(list, true, 0, maxNodeLabelSelectors, nil, field)
	for _, item := range list {
		_, ok := dockerhelper.NodeLabelConstraint(item, "==")
		res = append(res, basedto.ValidateCond(ok, field)...)
	}
	return res
}

func NewUpdateAppPlacementSettingsReq() *UpdateAppPlacementSettingsReq {
	return &UpdateAppPlacementSettingsReq{}
}

// Validate implements interface basedto.ReqValidator
func (req *UpdateAppPlacementSettingsReq) Validate() hperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 10) //nolint:mnd
	validators = append(validators, req.UpdateUniqueSettingReq.Validate()...)
	validators = append(validators, req.validate("")...)
	return hperrors.NewValidationErrors(vld.Validate(validators...))
}

type UpdateAppPlacementSettingsResp struct {
	Meta *basedto.Meta `json:"meta"`
}
