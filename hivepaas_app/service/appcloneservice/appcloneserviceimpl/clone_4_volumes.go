package appcloneserviceimpl

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

func (s *service) cloneVolumes(
	ctx context.Context,
	data *appCloneData,
) (err error) {
	settings := data.CloneSettings
	if !settings.CloneVolumes {
		return nil
	}

	destApp, srcApp := data.DestApp, data.SrcApp
	destService, srcService := data.DestService, data.SrcService
	cloneFunc := data.OnCloneVolumes
	if cloneFunc == nil {
		cloneFunc = func(destApp, srcApp *entity.App, srcMounts []mount.Mount) ([]mount.Mount, error) {
			return s.onCloneVolumesDefault(ctx, srcMounts, data)
		}
	}

	if srcService.Spec.TaskTemplate.ContainerSpec == nil {
		return nil
	}

	destMounts, err := cloneFunc(destApp, srcApp, srcService.Spec.TaskTemplate.ContainerSpec.Mounts)
	if err != nil {
		return apperrors.Wrap(err)
	}

	destService.Spec.TaskTemplate.ContainerSpec.Mounts = destMounts
	return nil
}

//nolint:gocognit
func (s *service) onCloneVolumesDefault(
	ctx context.Context,
	srcMounts []mount.Mount,
	data *appCloneData,
) (destMounts []mount.Mount, err error) {
	settings := data.CloneSettings
	var mountsToCopyData [][]*mount.Mount

	for _, srcMount := range srcMounts {
		if len(settings.IncludedVolumes) > 0 && !gofn.Contain(settings.IncludedVolumes, srcMount.Source) {
			continue
		}
		if len(settings.ExcludedVolumes) > 0 && gofn.Contain(settings.ExcludedVolumes, srcMount.Source) {
			continue
		}

		switch srcMount.Type {
		case mount.TypeVolume, mount.TypeCluster:
			destMount := srcMount
			if !s.calcVolumeMountSubpath(&destMount, &srcMount, data) {
				continue
			}
			destMounts = append(destMounts, destMount)
			if settings.CloneVolumeData {
				mountsToCopyData = append(mountsToCopyData, []*mount.Mount{&destMount, &srcMount})
			}
		case mount.TypeBind: // Bind mounts are skipped as host paths cannot be automatically cloned
			continue
		case mount.TypeTmpfs, mount.TypeNamedPipe, mount.TypeImage: // Keep it as is
			destMounts = append(destMounts, srcMount)
		}
	}

	// Before cloning data, if configured, the src app must be stopped
	if len(mountsToCopyData) > 0 && !settings.LiveVolumeClone {
		restoreFunc, err := s.stopSrcAppBeforeCloningVolumes(ctx, data)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		if restoreFunc != nil {
			defer restoreFunc()
		}
	}

	// Clone volume data
	for _, mounts := range mountsToCopyData {
		destMount, srcMount := mounts[0], mounts[1]
		if data.LogStore != nil {
			srcPath := srcMount.Source
			if srcMount.VolumeOptions != nil && srcMount.VolumeOptions.Subpath != "" {
				srcPath = filepath.Join(srcPath, srcMount.VolumeOptions.Subpath)
			}
			dstPath := destMount.Source
			if destMount.VolumeOptions != nil && destMount.VolumeOptions.Subpath != "" {
				dstPath = filepath.Join(dstPath, destMount.VolumeOptions.Subpath)
			}
			_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
				fmt.Sprintf("Cloning volume data from '%s' to '%s'...", srcPath, dstPath),
				tasklog.TsNow,
			))
		}

		rsyncOpts := []volumeservice.RsyncOption{
			volumeservice.WithRsyncLogStore(data.LogStore),
			volumeservice.WithRsyncDelete(true),
		}
		if srcMount.VolumeOptions != nil && srcMount.VolumeOptions.Subpath != "" {
			rsyncOpts = append(rsyncOpts, volumeservice.WithSourceSubpath(srcMount.VolumeOptions.Subpath))
		}
		if destMount.VolumeOptions != nil && destMount.VolumeOptions.Subpath != "" {
			rsyncOpts = append(rsyncOpts, volumeservice.WithDestSubpath(destMount.VolumeOptions.Subpath))
		}

		err = s.volumeService.Rsync(ctx, srcMount, destMount, rsyncOpts...)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	return destMounts, nil
}

func (s *service) stopSrcAppBeforeCloningVolumes(
	ctx context.Context,
	data *appCloneData,
) (restoreReplicasFunc func(), err error) {
	srcApp, srcService := data.SrcApp, data.SrcService

	origSvcMode := srcService.Spec.Mode
	if origSvcMode.Replicated == nil || *origSvcMode.Replicated.Replicas == 0 {
		return func() {}, nil
	}

	if data.LogStore != nil {
		_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(
			fmt.Sprintf("Stopping source app '%s' before cloning volume data...", srcApp.Key),
			tasklog.TsNow,
		))
	}

	// Scale down source app to 0 replicas
	stopSpec := srcService.Spec
	stopSpec.Mode = swarm.ServiceMode{
		Replicated: &swarm.ReplicatedService{Replicas: new(uint64(0))},
	}

	_, updateErr := s.dockerManager.ServiceUpdate(ctx, srcService.ID, &srcService.Version, &stopSpec)
	if updateErr != nil {
		return nil, apperrors.Wrap(updateErr)
	}

	// Always restore original replicas count on finish, regardless of success or error
	restoreReplicasFunc = func() { //nolint:contextcheck
		bgCtx, cancel := context.WithTimeout(context.Background(), 2*time.Minute) //nolint:mnd
		defer cancel()

		if data.LogStore != nil {
			_ = data.LogStore.Add(bgCtx, tasklog.NewOutFrame(
				fmt.Sprintf("Restoring source app '%s' instances...", srcApp.Key),
				tasklog.TsNow,
			))
		}

		reloadedResp, getErr := s.dockerManager.ServiceInspect(bgCtx, srcService.ID)
		if getErr == nil && reloadedResp != nil {
			srcSvc := &reloadedResp.Service
			restoreSpec := srcSvc.Spec
			restoreSpec.Mode = origSvcMode
			_, _ = s.dockerManager.ServiceUpdate(bgCtx, srcSvc.ID, &srcSvc.Version, &restoreSpec)
		}
	}
	defer func() {
		if err != nil || recover() != nil {
			restoreReplicasFunc()
		}
	}()

	// Wait until all source containers stop completely
	_, err = s.dockerManager.ServiceWaitUntilStopped(ctx, srcService.ID, 2*time.Second) //nolint:mnd
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return restoreReplicasFunc, apperrors.Wrap(err)
}

func (s *service) calcVolumeMountSubpath(
	destMount, srcMount *mount.Mount,
	data *appCloneData,
) bool {
	if srcMount.VolumeOptions == nil || srcMount.VolumeOptions.Subpath == "" {
		return false
	}
	if destMount.VolumeOptions == nil {
		destMount.VolumeOptions = &mount.VolumeOptions{}
	}
	destApp, srcApp := data.DestApp, data.SrcApp
	subpath := srcMount.VolumeOptions.Subpath
	subpath = "/" + strings.TrimSuffix(strings.TrimPrefix(subpath, "/"), "/") + "/"
	// Replace all /src-app/ with /dst-app/
	subpath = strings.ReplaceAll(subpath, "/"+srcApp.Key+"/", "/"+destApp.Key+"/")

	if !strings.HasSuffix(srcMount.VolumeOptions.Subpath, "/") {
		subpath = strings.TrimSuffix(subpath, "/")
	}
	if !strings.HasPrefix(srcMount.VolumeOptions.Subpath, "/") {
		subpath = strings.TrimPrefix(subpath, "/")
	}

	if subpath == srcMount.VolumeOptions.Subpath {
		return false
	}
	destMount.VolumeOptions.Subpath = subpath
	return true
}
