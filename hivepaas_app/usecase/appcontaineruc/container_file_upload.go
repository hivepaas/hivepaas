package appcontaineruc

import (
	"context"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/containerservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
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
		return nil, hperrors.Wrap(err)
	}
	if app.ServiceID == "" {
		return nil, hperrors.NewUnavailable("App service").
			WithMsgLog("service not exist for app")
	}

	if req.ContainerID == "" {
		task, _, taskErr := uc.dockerManager.ServiceTaskGetRunning(ctx, app.ServiceID,
			0, 0, 0, nil)
		if taskErr != nil {
			return nil, hperrors.Wrap(taskErr)
		}
		if task == nil || task.Status.ContainerStatus == nil || task.Status.ContainerStatus.ContainerID == "" {
			return nil, hperrors.Wrap(hperrors.ErrActiveContainerNotFound).WithParam("App", app.Name)
		}

		req.ContainerID = task.Status.ContainerStatus.ContainerID
		req.NodeID = task.NodeID
	}

	// Before the upload, and its failure abandons it: pushing a file into a
	// running production container is exactly the kind of act that is worth
	// nothing to anybody unless there is a record of who did it.
	err = uc.recordAppAction(ctx, uc.db, auth, app, base.AuditLogSourceAPIAction, "container-file-upload",
		auditdetail.New().
			Set("path", req.Path).
			Set("fileName", req.FileName).
			Set("fileSize", req.FileSize).
			Set("extract", req.Extract).
			Set("containerId", req.ContainerID))
	if err != nil {
		return nil, hperrors.Wrap(err)
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
		return nil, hperrors.Wrap(err)
	}
	defer prepResp.TarStream.Close()

	overwrite := allowDirReplaced(req)

	currNodeID, err := uc.dockerManager.NodeCurrentID(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	isRemote := req.NodeID != "" && req.NodeID != currNodeID
	if isRemote {
		agentAddr, err := uc.agentService.GetAgentAddrForNode(ctx, req.NodeID)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}

		agentClient, err := containerservice.NewContainerServiceClient(agentAddr)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		defer agentClient.Close()

		err = agentClient.ContainerCopyTo(ctx, &containeragentdto.UploadFileInput{
			ContainerID: req.ContainerID,
			DstPath:     prepResp.DestPath,
			TarReader:   prepResp.TarStream,
			Overwrite:   overwrite,
		})
		if err != nil {
			return nil, copyToError(err, req)
		}
	} else {
		opts := make([]docker.ContainerCopyToOption, 0, 1)
		if overwrite {
			opts = append(opts, docker.ContainerCopyToWithAllowOverwriteDirWithFile(true))
		}

		_, err = uc.dockerManager.ContainerCopyTo(ctx, req.ContainerID, prepResp.DestPath, prepResp.TarStream, opts...)
		if err != nil {
			return nil, copyToError(err, req)
		}
	}

	return &appcontainerdto.UploadFileToContainerResp{
		Data: &appcontainerdto.UploadFileToContainerDataResp{
			Path:    req.Path,
			Message: "File uploaded successfully",
		},
	}, nil
}

// allowDirReplaced is whether the copy may put a file where a directory is, or
// a directory where a file is: what overwrite asks for an archive's entries. A
// single file never may: a path without its slash names the file, and /app
// given for the directory would put the file where the app's code was.
func allowDirReplaced(req *appcontainerdto.UploadFileToContainerReq) bool {
	if !req.Extract {
		return false
	}
	return req.Overwrite == nil || *req.Overwrite
}

// copyToError is the copy's failure as the client is told it: a single file
// docker would not put where a directory is says what to do instead. Docker
// says it in words only, the same through an agent.
func copyToError(err error, req *appcontainerdto.UploadFileToContainerReq) error {
	if !req.Extract && strings.Contains(err.Error(), "cannot overwrite directory") {
		return hperrors.Wrap(hperrors.ErrContainerPathIsDir).WithParam("Path", req.Path).WithCause(err)
	}
	return hperrors.Wrap(err)
}
