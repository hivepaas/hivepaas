package volumeuc

import (
	"context"
	"strings"

	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/volume"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/localpaas/localpaas/services/docker"
)

const (
	localpaasVolumeLabel = "localpaas.volume.managed"
)

func (uc *UC) ListVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.ListVolumeReq,
) (_ *volumedto.ListVolumeResp, err error) {
	var project *entity.Project
	if req.ProjectID != "" {
		project, err = uc.projectService.LoadProject(ctx, uc.db, req.ProjectID, true)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	volumes, err := uc.dockerManager.VolumeList(ctx, func(opts *volume.ListOptions) {
		if opts.Filters.Len() == 0 {
			opts.Filters = filters.NewArgs()
		}
		if !req.ListAll {
			opts.Filters.Add("label", localpaasVolumeLabel)
		}
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	filterVolumes := volumes.Volumes
	if req.ProjectID != "" {
		filterVolumes = gofn.Filter(filterVolumes, func(vol *volume.Volume) bool {
			label := vol.Labels[docker.StackLabelNamespace]
			return label == "" || label == project.Key
		})
	}
	if req.Search != "" {
		keyword := strings.ToLower(req.Search)
		filterVolumes = gofn.Filter(filterVolumes, func(vol *volume.Volume) bool {
			return strings.Contains(vol.Name, keyword) || strings.Contains(vol.Mountpoint, keyword)
		})
	}
	if len(auth.AllowObjectIDs) > 0 {
		filterVolumes = gofn.Filter(filterVolumes, func(vol *volume.Volume) bool {
			volID := vol.Name
			if vol.ClusterVolume != nil {
				volID = vol.ClusterVolume.ID
			}
			return gofn.Contain(auth.AllowObjectIDs, volID)
		})
	}

	return &volumedto.ListVolumeResp{
		Meta: &basedto.ListMeta{Page: &basedto.PagingMeta{
			Offset: 0,
			Limit:  req.Paging.Limit,
			Total:  len(filterVolumes),
		}},
		Data: volumedto.TransformVolumes(filterVolumes, false),
	}, nil
}
