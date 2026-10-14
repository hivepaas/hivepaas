// Package imagescanservice calls an agent's image scan service: a scan on the
// agent's node.
package imagescanservice

import (
	"context"
	"encoding/json"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagescanservice/imagescanrun"
)

// maxAnswerSize is how large a scan's answer may be: its report and the end of
// its log, over gRPC's default of 4 MB.
const maxAnswerSize = imagescanrun.OutputMax + imagescanrun.LogMax + 1<<20

type ImageScanServiceClient interface {
	// Run scans an image on the agent's node.
	Run(ctx context.Context, req *imagescanrun.RunReq) (*imagescanrun.RunResp, error)
	Close() error
}

type grpcImageScanServiceClient struct {
	protoClient agentproto.ImageScanServiceClient
	conn        *grpc.ClientConn
}

//nolint:ireturn // the client is used through its interface, which tests replace
func NewImageScanServiceClient(agentAddr string) (ImageScanServiceClient, error) {
	conn, err := grpc.NewClient(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &grpcImageScanServiceClient{conn: conn, protoClient: agentproto.NewImageScanServiceClient(conn)}, nil
}

func (c *grpcImageScanServiceClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (c *grpcImageScanServiceClient) Run(
	ctx context.Context,
	req *imagescanrun.RunReq,
) (*imagescanrun.RunResp, error) {
	volume, err := json.Marshal(req.Volume)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	resp, err := c.protoClient.ImageScanRun(client.CreateAuthCtx(ctx), &agentproto.ImageScanRunReq{
		TrivyImage:       req.TrivyImage,
		ImageId:          req.ImageID,
		ImageNode:        req.ImageNode,
		RemoteRef:        req.RemoteRef,
		Platform:         req.Platform,
		DockerConfig:     req.DockerConfig,
		VolumeMount:      volume,
		CacheDir:         req.CacheDir,
		DbRepository:     req.DBRepository,
		JavaDbRepository: req.JavaDBRepository,
		NanoCpus:         req.NanoCPUs,
		MemoryBytes:      req.MemoryBytes,
		TimeoutSeconds:   int64(req.Timeout / time.Second),
	}, grpc.MaxCallRecvMsgSize(maxAnswerSize))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &imagescanrun.RunResp{
		Source:       base.ImageScanSource(resp.GetSource()),
		Output:       resp.GetOutput(),
		Log:          resp.GetLog(),
		LogTruncated: resp.GetLogTruncated(),
		OOMKilled:    resp.GetOomKilled(),
		ExitCode:     resp.GetExitCode(),
	}, nil
}
