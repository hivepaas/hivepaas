package settingserviceimpl

import (
	"context"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/localpaas_app/pkg/unit"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	imageBuildSettingName      = "Image build settings"
	imageBuildCPUDefault       = 2
	imageBuildCPUMin           = 1
	imageBuildCPUMax           = 8
	imageBuildMemDefault       = 2 * unit.GB
	imageBuildMemMin           = 1 * unit.GB
	imageBuildMemMax           = 16 * unit.GB
	imageBuildCheckoutMaxDepth = 100
)

func (s *service) initDefaultImageBuildSettings(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	imageBuildSetting := &entity.Setting{
		ID:              gofn.Must(ulid.NewStringULID()),
		Scope:           base.ObjectScopeGlobal,
		Type:            base.SettingTypeImageBuildSettings,
		Status:          base.SettingStatusActive,
		Name:            imageBuildSettingName,
		AvailInProjects: true,
		Default:         true,
		Version:         entity.CurrentImageBuildSettingsVersion,
		CreatedAt:       timeNow,
		UpdatedAt:       timeNow,
	}
	imageBuild := &entity.ImageBuildSettings{
		Resources: entity.ImageBuildResourceSettings{
			CPUs: imageBuildCPUDefault,
			Mem:  imageBuildMemDefault,
		},
		Sources: entity.ImageBuildSourceSettings{
			CheckoutMaxDepth: imageBuildCheckoutMaxDepth,
		},
	}

	// Calculate the best values for resource settings
	listResp, err := s.dockerManager.NodeManagerList(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	//nolint
	if leaderNode, found := gofn.FindPtr(listResp.Items, func(n *swarm.Node) bool {
		return n.ManagerStatus != nil && n.ManagerStatus.Leader
	}); found {
		// Use half of the leader node's resources for image building
		res := &leaderNode.Description.Resources
		cpus := max(min(res.NanoCPUs/docker.UnitCPUNano/2, imageBuildCPUMax), imageBuildCPUMin)
		mem := unit.DataSize(res.MemoryBytes / 2).Truncate(32 * unit.MB)
		mem = max(min(mem, imageBuildMemMax), imageBuildMemMin)
		imageBuild.Resources.CPUs = uint(cpus)
		imageBuild.Resources.Mem = mem
	}

	imageBuildSetting.MustSetData(imageBuild)

	err = s.settingRepo.Insert(ctx, db, imageBuildSetting)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
