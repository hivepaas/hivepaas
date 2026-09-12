package volumeuc

import (
	"context"
	"fmt"
	"path/filepath"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/nodeexecservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/cluster/volumeuc/volumedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (uc *UC) CreateVolume(
	ctx context.Context,
	auth *basedto.Auth,
	req *volumedto.CreateVolumeReq,
) (*volumedto.CreateVolumeResp, error) {
	req.Type = currentSettingType
	req.Auth = auth

	nodeID, err := uc.resolveCurrentNode(ctx, req.NodeID)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	req.NodeID = nodeID

	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		// The settings framework refuses a duplicate name within the request's
		// scope for us. It replaces the VolumeInspect this used to do, which asked
		// the wrong daemon anyway - and it only catches what actually gets
		// persisted because req.Name is never rewritten before pData.Setting.Name
		// is set to it below: same string, checked once and stored unchanged.
		// ClusterVolume.GetRefObjectIDs always returns an empty RefObjectIDs - a
		// volume references no setting, app or user - so VerifyingRefIDs would be
		// a no-op here and is left unset.
		VerifyingName: req.Name,
		Version:       currentSettingVersion,
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			volEntity, err := uc.prepareVolumeSpec(ctx, req)
			if err != nil {
				return hperrors.Wrap(err)
			}
			// The setting's own id is the volume's name in docker. Nothing has
			// created it yet - docker does that when a task first mounts it - so
			// the name has to come from the only identity that exists at this
			// point.
			pData.Setting.RefID = pData.Setting.ID
			pData.Setting.Name = req.Name
			pData.Setting.Kind = volEntity.Driver
			return hperrors.Wrap(pData.Setting.SetData(volEntity))
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	return &volumedto.CreateVolumeResp{
		Data: resp.Data,
	}, nil
}

// prepareVolumeSpec works out the description that will be stored, doing the two
// things that have to happen on a real host: making the bind directory on the
// pinned node, and resolving the directory a bind volume points at.
func (uc *UC) prepareVolumeSpec(
	ctx context.Context,
	req *volumedto.CreateVolumeReq,
) (*entity.ClusterVolume, error) {
	isPinnedToNode := req.NodeID != "" || req.NodeLabel != ""

	if isPinnedToNode && req.BindOptions != nil && req.BindOptions.Directory != "" {
		if err := uc.createBindDirectoryInNode(ctx, req); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	bindDirectory := ""
	if req.Driver == docker.VolumeDriverLocal && req.BindOptions != nil {
		directory, err := uc.calcBindDirectory(ctx, req, req.BindOptions.Directory)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		bindDirectory = directory
	}

	return req.ToEntityWithDriverOpts(req.BuildDriverOpts(bindDirectory)), nil
}

func (uc *UC) calcBindDirectory(
	ctx context.Context,
	req *volumedto.CreateVolumeReq,
	directory string,
) (string, error) {
	subpath := ""
	if directory == "" {
		storageInHost := config.Current().Storage.BindSource
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

	targetDir := filepath.Join(volumeservice.HostPathPrefix, req.BindOptions.Directory)
	mkdirCmd := fmt.Sprintf("mkdir -p '%s' && chmod -R 777 '%s'", targetDir, targetDir)
	cmdReq := &nodeexecservice.CommandExecReq{
		NodeID:    req.NodeID,
		NodeLabel: req.NodeLabel,
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
