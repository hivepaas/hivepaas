package volumeuc

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
	"github.com/hivepaas/hivepaas/services/docker/dockerhelper"
)

func (uc *UC) CreateVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.CreateVolumeReq,
) (*volumedto.CreateVolumeResp, error) {
	req.Type = currentSettingType
	volEntity := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingRefIDs: volEntity.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			if req.Scope.IsProjectScope() {
				req.Name = req.Scope.Project.Key + "_" + req.Name
			}
			createResp, err := uc.createVolumeInDocker(ctx, req)
			if err != nil {
				return hperrors.Wrap(err)
			}
			vol := &createResp.Volume
			volEntity.RefID = dockerhelper.GetVolumeID(vol)
			if req.BindOptions != nil {
				volEntity.NodeID = req.BindOptions.NodeID
				volEntity.NodeLabel = req.BindOptions.NodeLabel
			}
			pData.Setting.Name = req.Name
			pData.Setting.Kind = vol.Driver
			if err := pData.Setting.SetData(volEntity); err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &volumedto.CreateVolumeResp{
		Data: resp.Data,
	}, nil
}

//nolint:gocognit
func (uc *UC) createVolumeInDocker(
	ctx context.Context,
	req *volumedto.CreateVolumeReq,
) (*client.VolumeCreateResult, error) {
	_, err := uc.dockerManager.VolumeInspect(ctx, req.Name)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if err == nil {
		return nil, hperrors.NewAlreadyExist("Cluster volume")
	}

	// If this is node-local directory bind, create the dir
	if req.BindOptions != nil && req.BindOptions.Directory != "" {
		err = uc.createBindDirectoryInNode(ctx, req)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	driverOpts := map[string]string{}
	if req.Driver == docker.VolumeDriverLocal {
		switch {
		case req.BindOptions != nil:
			directory, err := uc.calcBindDirectory(ctx, req, req.BindOptions.Directory)
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			driverOpts["type"] = "none"
			driverOpts["device"] = directory
			o := fmt.Sprintf("bind,%s", gofn.If(req.BindOptions.Readonly, "ro", "rw"))
			if req.BindOptions.Propagation != "" {
				o += "," + string(req.BindOptions.Propagation)
			}
			if req.BindOptions.ExtraOptions != "" {
				o += "," + req.BindOptions.ExtraOptions
			}
			driverOpts["o"] = o

		case req.NfsOptions != nil:
			driverOpts["type"] = "nfs"
			driverOpts["device"] = req.NfsOptions.Device
			o := fmt.Sprintf("addr=%s,%s", req.NfsOptions.Addr, gofn.If(req.NfsOptions.Readonly, "ro", "rw"))
			if req.NfsOptions.Version != "" {
				o += ",nfsvers=" + req.NfsOptions.Version
			}
			if req.NfsOptions.ExtraOptions != "" {
				o += "," + req.NfsOptions.ExtraOptions
			}
			driverOpts["o"] = o

		case req.TmpfsOptions != nil:
			driverOpts["type"] = "tmpfs"
			driverOpts["device"] = gofn.Coalesce(req.TmpfsOptions.Device, "tmpfs")
			bytes := req.TmpfsOptions.Size.Bytes() + int64(unit.MB) - 1
			o := fmt.Sprintf("size=%vm", bytes/int64(unit.MB))
			if req.TmpfsOptions.Mode > 0 {
				o += fmt.Sprintf(",mode=%v", req.TmpfsOptions.Mode)
			}
			if req.TmpfsOptions.UID > 0 {
				o += fmt.Sprintf(",uid=%v", req.TmpfsOptions.UID)
			}
			if req.TmpfsOptions.GID > 0 {
				o += fmt.Sprintf(",gid=%v", req.TmpfsOptions.GID)
			}
			driverOpts["o"] = o
		}
	}

	// Overwrite the driver opts with the extra values from the client
	for k, v := range req.Options {
		if _, ok := driverOpts[k]; ok && (k == "type" || k == "device") {
			continue
		}
		driverOpts[k] = v
	}

	createResp, err := uc.dockerManager.VolumeCreate(ctx, func(opts *client.VolumeCreateOptions) {
		opts.Driver = string(req.Driver)
		opts.DriverOpts = driverOpts
		opts.Labels = req.Labels
		opts.Name = req.Name
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return createResp, nil
}

func (uc *UC) calcBindDirectory(
	ctx context.Context,
	req *volumedto.CreateVolumeReq,
	directory string,
) (string, error) {
	subpath := ""
	if directory == "" {
		storageInHost := config.Current.Storage.BindSource
		if storageInHost == "" {
			return "", hperrors.Wrap(hperrors.ErrUnconfigured).
				WithParam("Name", "HP_STORAGE_BIND_SOURCE")
		}
		directory = storageInHost
		subpath = "project_data"
		switch req.Scope.ScopeType {
		case base.ObjectScopeProject:
			subpath = filepath.Join(subpath, req.Scope.Project.Key)
		case base.ObjectScopeProjectEnv:
			projectEnv := req.Scope.ProjectEnv
			subpath = filepath.Join(subpath, projectEnv.Project.Key, projectEnv.Key)
		case base.ObjectScopeApp:
			app := req.Scope.App
			subpath = filepath.Join(subpath, app.Project.Key, app.ProjectEnv.Key, app.Key)
		case base.ObjectScopeGlobal, base.ObjectScopeHivepaas:
		case base.ObjectScopeUser:
			fallthrough
		default:
			return "", hperrors.Wrap(hperrors.ErrObjectScopeInvalid)
		}
	}

	err := uc.volumeService.MakeSubDirInHost(ctx, directory, subpath, false)
	if err != nil {
		return "", hperrors.Wrap(err)
	}

	directory = filepath.Join(directory, subpath)
	return directory, nil
}

func (uc *UC) createBindDirectoryInNode(
	ctx context.Context,
	req *volumedto.CreateVolumeReq,
) (err error) {
	if req.BindOptions == nil || req.BindOptions.Directory == "" {
		return nil
	}

	nodeID := req.BindOptions.NodeID
	nodeLabel := req.BindOptions.NodeLabel

	if nodeID == "" && nodeLabel == "" {
		nodeID, err = uc.dockerManager.NodeCurrentID(ctx)
		if err != nil {
			return hperrors.Wrap(err)
		}
	}

	targetDir := filepath.Join("/host", req.BindOptions.Directory)
	mkdirCmd := fmt.Sprintf("mkdir -p '%s' && chmod -R 777 '%s'", targetDir, targetDir)
	cmdReq := &nodeexecservice.CommandExecReq{
		NodeID:    nodeID,
		NodeLabel: nodeLabel,
		CommandExecOpts: &nodeexecservice.CommandExecOpts{
			Command: []string{"sh", "-c", mkdirCmd},
		},
	}

	resp, err := uc.nodeExecService.ExecCommand(ctx, cmdReq)
	if err != nil {
		return hperrors.Wrap(err)
	}
	if resp != nil && resp.ExitCode != 0 {
		return hperrors.Wrap(hperrors.ErrDirNotCreated).WithParam("Name", req.BindOptions.Directory)
	}

	return nil
}
