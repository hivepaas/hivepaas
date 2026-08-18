package appcontaineruc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/containerservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerfileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appcontaineruc/appcontainerdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc/containeragentdto"
	"github.com/hivepaas/hivepaas/services/docker"
)

func (uc *UC) UploadFileToContainer(
	ctx context.Context,
	auth *basedto.Auth,
	req *appcontainerdto.UploadFileToContainerReq,
) (*appcontainerdto.UploadFileToContainerResp, error) {
	app, err := uc.appService.LoadApp(ctx, uc.db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
		bunex.SelectRelation("ProjectEnv"),
	)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	if app.ServiceID == "" {
		return nil, apperrors.NewUnavailable("App service").
			WithMsgLog("service not exist for app")
	}

	if req.ContainerID == "" {
		task, _, taskErr := uc.dockerManager.ServiceTaskGetRunning(ctx, app.ServiceID,
			0, 0, 0, nil)
		if taskErr != nil {
			return nil, apperrors.Wrap(taskErr)
		}
		if task == nil || task.Status.ContainerStatus == nil || task.Status.ContainerStatus.ContainerID == "" {
			return nil, apperrors.Wrap(apperrors.ErrActiveContainerNotFound).WithParam("App", app.Name)
		}

		req.ContainerID = task.Status.ContainerStatus.ContainerID
		req.NodeID = task.NodeID
	}

	prepResp, err := uc.containerFileService.PrepareUploadTarStream(ctx, &containerfileservice.PrepareUploadTarStreamReq{
		Path:              req.Path,
		FileName:          req.FileName,
		FileSize:          req.FileSize,
		Extract:           req.Extract,
		CompressionFormat: req.CompressionFormat,
		Content:           req.FileContent,
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	defer prepResp.TarStream.Close()

	overwrite := true
	if req.Overwrite != nil {
		overwrite = *req.Overwrite
	}

	currNodeID, err := uc.dockerManager.NodeCurrentID(ctx)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	isRemote := req.NodeID != "" && req.NodeID != currNodeID
	if isRemote {
		agentAddr, err := uc.agentService.GetAgentAddrForNode(ctx, req.NodeID)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}

		agentClient, err := containerservice.NewContainerServiceClient(agentAddr)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
		defer agentClient.Close()

		err = agentClient.ContainerCopyTo(ctx, &containeragentdto.UploadFileInput{
			ContainerID: req.ContainerID,
			DstPath:     prepResp.DestPath,
			TarReader:   prepResp.TarStream,
			Overwrite:   overwrite,
		})
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	} else {
		opts := make([]docker.ContainerCopyToOption, 0, 1)
		if overwrite {
			opts = append(opts, docker.ContainerCopyToWithAllowOverwriteDirWithFile(true))
		}

		_, err = uc.dockerManager.ContainerCopyTo(ctx, req.ContainerID, prepResp.DestPath, prepResp.TarStream, opts...)
		if err != nil {
			return nil, apperrors.Wrap(err)
		}
	}

	return &appcontainerdto.UploadFileToContainerResp{
		Data: &appcontainerdto.UploadFileToContainerDataResp{
			Path:    req.Path,
			Message: "File uploaded successfully",
		},
	}, nil
}
