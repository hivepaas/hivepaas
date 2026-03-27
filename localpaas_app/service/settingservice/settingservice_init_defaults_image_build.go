package settingservice

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/ulid"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	imageBuildSettingName = "Image build settings"
	imageBuildCPUDefault  = 2
	imageBuildCPUMin      = 1
	imageBuildCPUMax      = 8
	imageBuildMemDefault  = 2048      // 2GB
	imageBuildMemMin      = 1024      // MB
	imageBuildMemMax      = 16 * 1024 // MB
)

func (s *settingService) initDefaultImageBuild(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	imageBuildSetting := &entity.Setting{
		ID:              gofn.Must(ulid.NewStringULID()),
		Scope:           base.SettingScopeGlobal,
		Type:            base.SettingTypeImageBuild,
		Status:          base.SettingStatusActive,
		Name:            imageBuildSettingName,
		AvailInProjects: true,
		Default:         true,
		Version:         entity.CurrentImageBuildVersion,
		CreatedAt:       timeNow,
		UpdatedAt:       timeNow,
	}
	imageBuild := &entity.ImageBuild{
		Resources: &entity.ImageBuildResources{
			CPUs:  imageBuildCPUDefault,
			MemMB: imageBuildMemDefault,
		},
	}

	// Calculate the best values for resource settings
	nodes, err := s.dockerManager.NodeManagerList(ctx)
	if err != nil {
		return apperrors.Wrap(err)
	}
	//nolint
	if leaderNode, found := gofn.FindPtr(nodes, func(n *swarm.Node) bool {
		return n.ManagerStatus != nil && n.ManagerStatus.Leader
	}); found {
		// Use half of the leader node's resources for image building
		res := &leaderNode.Description.Resources
		cpus := max(min(res.NanoCPUs/docker.UnitCPUNano/2, imageBuildCPUMax), imageBuildCPUMin)
		memMB := max(min(res.MemoryBytes/docker.UnitMemMB/2, imageBuildMemMax), imageBuildMemMin)
		imageBuild.Resources.CPUs = int32(cpus)
		imageBuild.Resources.MemMB = memMB
	}

	imageBuildSetting.MustSetData(imageBuild)

	err = s.settingRepo.Insert(ctx, db, imageBuildSetting)
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
