package imagebuildsettingsdto

import (
	"fmt"

	vld "github.com/tiendc/go-validator"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/copier"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
)

type GetImageBuildSettingsReq struct {
	settings.GetUniqueSettingReq
}

func NewGetImageBuildSettingsReq() *GetImageBuildSettingsReq {
	return &GetImageBuildSettingsReq{}
}

func (req *GetImageBuildSettingsReq) Validate() apperrors.ValidationErrors {
	validators := make([]vld.Validator, 0, 5) //nolint:mnd
	validators = append(validators, req.GetUniqueSettingReq.Validate()...)
	return apperrors.NewValidationErrors(vld.Validate(validators...))
}

type GetImageBuildSettingsResp struct {
	Meta *basedto.Meta           `json:"meta"`
	Data *ImageBuildSettingsResp `json:"data"`
}

type ImageBuildSettingsResp struct {
	*settings.BaseSettingResp
	Workers   *ImageBuildWorkerSettingsResp   `json:"workers"`
	Resources *ImageBuildResourceSettingsResp `json:"resources"`
	Sources   *ImageBuildSourceSettingsResp   `json:"sources"`
	NoCache   bool                            `json:"noCache"`
	NoVerbose bool                            `json:"noVerbose"`
}

type ImageBuildWorkerSettingsResp struct {
	Nodes          basedto.NamedObjectSliceResp `json:"nodes,omitempty"`
	NodeLabels     []string                     `json:"nodeLabels,omitempty"`
	MaxParallelism int                          `json:"maxParallelism,omitempty"`
}

type ImageBuildResourceSettingsResp struct {
	CPUs    uint          `json:"cpus"`
	Mem     unit.DataSize `json:"mem"`
	MemSwap unit.DataSize `json:"memSwap"`
	ShmSize unit.DataSize `json:"shmSize"`
}

type ImageBuildSourceSettingsResp struct {
	RepoCache bool `json:"repoCache"`
}

func TransformImageBuild(
	setting *entity.Setting,
	_ *entity.RefObjects,
	refClusterObjects *entity.RefClusterObjects,
) (resp *ImageBuildSettingsResp, err error) {
	config := setting.MustAsImageBuildSettings()
	if err = copier.Copy(&resp, config); err != nil {
		return nil, apperrors.Wrap(err)
	}

	resp.BaseSettingResp, err = settings.TransformSettingBase(setting)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	if refClusterObjects == nil {
		refClusterObjects = &entity.RefClusterObjects{}
	}

	if resp.Workers != nil {
		resp.Workers.Nodes = make(basedto.NamedObjectSliceResp, 0, len(config.Workers.NodeIDs))
		for _, nodeID := range config.Workers.NodeIDs {
			node := refClusterObjects.RefNodes[nodeID]
			nodeName := "<missing>"
			if node != nil {
				nodeName = gofn.Coalesce(node.Spec.Name, node.Description.Hostname)
			}
			resp.Workers.Nodes = append(resp.Workers.Nodes, &basedto.NamedObjectResp{
				ID:   nodeID,
				Name: fmt.Sprintf("%s (node: %s)", nodeID, nodeName),
			})
		}
	}

	return resp, nil
}
