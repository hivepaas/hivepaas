package appcontaineruc

import (
	"context"
	"io"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
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
		return nil, apperrors.Wrap(err)
	}
	if app.ServiceID == "" {
		return nil, apperrors.NewUnavailable("App service").
			WithMsgLog("service not exist for app")
	}

	var (
		resultFileName    string
		resultFileSize    int64
		resultContentType string
		resultReader      io.ReadCloser
	)

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

		remoteRes, err := agentClient.ContainerCopyFrom(ctx, &containeragentdto.DownloadFileInput{
			ContainerID:       req.ContainerID,
			Path:              req.Path,
			IsDir:             req.IsDir,
			CompressionFormat: req.CompressionFormat,
		})
		if err != nil {
			_ = agentClient.Close()
			return nil, apperrors.Wrap(err)
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
			return nil, apperrors.Wrap(err)
		}

		resp, err := uc.containerFileService.StreamFile(ctx, &containerfileservice.StreamFileReq{
			Path:              req.Path,
			IsDir:             req.IsDir,
			CompressionFormat: req.CompressionFormat,
			Content:           res.Content,
			Stat:              &res.Stat,
		})
		if err != nil {
			_ = res.Content.Close()
			return nil, apperrors.Wrap(err)
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

func (r *clientReadCloser) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if err != nil {
		if err == io.EOF {
			return n, io.EOF
		}
		return n, apperrors.Wrap(err)
	}
	return n, nil
}

func (r *clientReadCloser) Close() error {
	var errs []error
	if r.reader != nil {
		if err := r.reader.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if r.client != nil {
		if err := r.client.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	if len(errs) > 0 {
		return apperrors.Wrap(errs[0])
	}
	return nil
}
