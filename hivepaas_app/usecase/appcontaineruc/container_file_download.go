package appcontaineruc

import (
	"context"
	"io"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client/containerservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/containerfileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/appcontaineruc/appcontainerdto"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/containeragentuc/containeragentdto"
)

func (uc *UC) DownloadFileFromContainer(
	ctx context.Context,
	auth *basedto.Auth,
	req *appcontainerdto.DownloadFileFromContainerReq,
) (*appcontainerdto.DownloadFileFromContainerResp, error) {
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

	var (
		resultFileName    string
		resultFileSize    int64
		resultContentType string
		resultReader      io.ReadCloser
	)

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

		remoteRes, err := agentClient.ContainerCopyFrom(ctx, &containeragentdto.DownloadFileInput{
			ContainerID:       req.ContainerID,
			Path:              req.Path,
			IsDir:             req.IsDir,
			CompressionFormat: req.CompressionFormat,
		})
		if err != nil {
			_ = agentClient.Close()
			return nil, hperrors.Wrap(err)
		}

		resultFileName = remoteRes.FileName
		resultFileSize = remoteRes.FileSize
		resultContentType = remoteRes.ContentType
		resultReader = &clientReadCloser{
			reader: remoteRes.Reader,
			client: agentClient,
		}
	} else {
		res, err := uc.dockerManager.ContainerCopyFrom(ctx, req.ContainerID, req.Path)
		if err != nil {
			return nil, hperrors.Wrap(err)
		}

		resp, err := uc.containerFileService.PrepareDownloadStream(ctx, &containerfileservice.PrepareDownloadStreamReq{
			Path:              req.Path,
			IsDir:             req.IsDir,
			CompressionFormat: req.CompressionFormat,
			Content:           res.Content,
			Stat:              &res.Stat,
		})
		if err != nil {
			_ = res.Content.Close()
			return nil, hperrors.Wrap(err)
		}

		resultFileName = resp.FileName
		resultFileSize = resp.FileSize
		resultContentType = resp.ContentType
		resultReader = resp.Reader
	}

	return &appcontainerdto.DownloadFileFromContainerResp{
		FileName:    resultFileName,
		FileSize:    resultFileSize,
		ContentType: resultContentType,
		Reader:      resultReader,
	}, nil
}

type clientReadCloser struct {
	reader io.ReadCloser
	client containerservice.ContainerServiceClient
}

func (req *clientReadCloser) Read(p []byte) (int, error) {
	n, err := req.reader.Read(p)
	if err != nil {
		if err == io.EOF {
			return n, io.EOF
		}
		return n, hperrors.Wrap(err)
	}
	return n, nil
}

func (req *clientReadCloser) Close() error {
	var errs []error
	if req.reader != nil {
		if err := req.reader.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if req.client != nil {
		if err := req.client.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return hperrors.Wrap(errs[0])
	}
	return nil
}
