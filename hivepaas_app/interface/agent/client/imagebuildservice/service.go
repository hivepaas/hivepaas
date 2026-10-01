package imagebuildservice

import (
	"bufio"
	"context"
	"errors"
	"io"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc/imagebuildagentdto"
)

// sourceChunkSize is how much of the packed source one message carries.
const sourceChunkSize = 1 << 20

type ImageBuildServiceClient interface {
	// ImageBuild builds an image on the agent's node from the source sendSource
	// writes: the checkout, packed (see srcpack.Pack). The agent keeps its copy
	// only as long as the call lasts.
	ImageBuild(
		ctx context.Context,
		req *imagebuildagentdto.ImageBuildReq,
		sendSource func(w io.Writer) error,
	) (*imagebuildagentdto.ImageBuildResp, error)
	Close() error
}

type grpcImageBuildServiceClient struct {
	protoClient agentproto.ImageBuildServiceClient
	conn        *grpc.ClientConn
}

func NewImageBuildServiceClient(agentAddr string) (ImageBuildServiceClient, error) {
	conn, err := grpc.NewClient(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &grpcImageBuildServiceClient{
		conn:        conn,
		protoClient: agentproto.NewImageBuildServiceClient(conn),
	}, nil
}

func (c *grpcImageBuildServiceClient) Close() error {
	if c.conn != nil {
		if err := c.conn.Close(); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

func (c *grpcImageBuildServiceClient) ImageBuild(
	ctx context.Context,
	req *imagebuildagentdto.ImageBuildReq,
	sendSource func(w io.Writer) error,
) (*imagebuildagentdto.ImageBuildResp, error) {
	// The agent cannot open a stored secret: what the build reads from settings
	// goes with the request, resolved here.
	if req.Inputs == nil {
		return nil, hperrors.NewMissing("Image build inputs")
	}

	// Canceling ends the call on the agent, which then drops the source it has.
	authCtx, cancel := client.CreateAuthCtxWithCancel(ctx)
	defer cancel()

	var protoDockerfile *agentproto.DeploymentDockerfile
	if req.Dockerfile.Source != "" || req.Dockerfile.Path != "" || req.Dockerfile.Content != "" ||
		req.Dockerfile.ScanPath != "" {
		protoDockerfile = &agentproto.DeploymentDockerfile{
			Source:   string(req.Dockerfile.Source),
			Path:     req.Dockerfile.Path,
			Content:  req.Dockerfile.Content,
			ScanPath: req.Dockerfile.ScanPath,
		}
	}

	protoBuildSettings := BuildSettingsToProto(req.ImageBuildSettings)

	appID := req.AppID
	if appID == "" && req.App != nil {
		appID = req.App.ID
	}

	protoReq := &agentproto.ImageBuildReq{
		TaskId:             req.TaskID,
		AppId:              appID,
		CommitHash:         req.CommitHash,
		Dockerfile:         protoDockerfile,
		ImageTags:          req.ImageTags,
		PushToRegistryId:   req.PushToRegistry.ID,
		ImageBuildSettings: protoBuildSettings,
		NoCache:            req.NoCache,
		BuildId:            req.BuildID,
		Inputs:             InputsToProto(req.Inputs),
	}

	stream, err := c.protoClient.ImageBuildFromSource(authCtx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	// The source is sent while the agent's answers are read: the agent may end
	// the call, with its reason, before the whole source has left.
	sent := make(chan error, 1)
	go func() {
		var sendErr error
		defer func() {
			if sendErr != nil {
				cancel()
			}
			sent <- sendErr
		}()
		defer safego.RecoverTo(&sendErr)
		sendErr = sendBuild(stream, protoReq, sendSource)
	}()

	respDTO, recvErr := c.receive(ctx, stream, req)
	cancel()
	if sendErr := <-sent; sendErr != nil {
		// Why the source could not be sent is the cause; the agent only saw the call end.
		return nil, hperrors.Wrap(sendErr)
	}
	if recvErr != nil {
		return nil, hperrors.Wrap(recvErr)
	}
	return respDTO, nil
}

// BuildSettingsToProto is an app's build settings as the agent takes them, nil
// for none.
func BuildSettingsToProto(settings *entity.ImageBuildSettings) *agentproto.ImageBuildSettings {
	if settings == nil {
		return nil
	}
	//nolint:gosec
	return &agentproto.ImageBuildSettings{
		NoCache:   settings.NoCache,
		NoVerbose: settings.NoVerbose,
		Workers: &agentproto.ImageBuildWorkerSettings{
			NodeIds:        settings.Workers.NodeIDs,
			NodeLabels:     settings.Workers.NodeLabels,
			MaxParallelism: uint32(settings.Workers.MaxParallelism),
		},
		Resources: &agentproto.ImageBuildResourceSettings{
			Cpus:    uint32(settings.Resources.CPUs),
			Mem:     uint64(settings.Resources.Mem),
			MemSwap: uint64(settings.Resources.MemSwap),
			ShmSize: uint64(settings.Resources.ShmSize),
		},
		Sources: &agentproto.ImageBuildSourceSettings{
			RepoCache: settings.Sources.RepoCache,
		},
	}
}

// InputsToProto is what the app resolved for a build, as the agent takes it.
func InputsToProto(inputs *imagebuildservice.BuildInputs) *agentproto.ImageBuildInputs {
	out := &agentproto.ImageBuildInputs{
		EnvVars:       make(map[string]string, len(inputs.EnvVars)),
		SecretEnvVars: inputs.SecretEnvVars,
		Secrets:       inputs.Secrets,
	}
	for key, value := range inputs.EnvVars {
		if value != nil {
			out.EnvVars[key] = *value
		}
	}
	for _, auth := range inputs.RegistryAuths {
		out.RegistryAuths = append(out.RegistryAuths, &agentproto.ImageBuildRegistryAuth{
			Address: auth.ServerAddress, Username: auth.Username, Password: auth.Password,
		})
	}
	if push := inputs.PushRegistry; push != nil {
		out.PushRegistry = &agentproto.ImageBuildRegistryAuth{
			Address: push.ServerAddress, Username: push.Username, Password: push.Password,
		}
	}
	return out
}

// sendBuild sends the request, then the source in chunks, then ends the sending
// side. When the stream itself refuses a message the agent has ended the call,
// and its reason is what the receiving side reads: that is not an error here.
func sendBuild(
	stream grpc.BidiStreamingClient[agentproto.ImageBuildMsg, agentproto.ImageBuildResp],
	req *agentproto.ImageBuildReq,
	sendSource func(w io.Writer) error,
) error {
	if err := stream.Send(&agentproto.ImageBuildMsg{Value: &agentproto.ImageBuildMsg_Req{Req: req}}); err != nil {
		return nil //nolint:nilerr // the receiving side reports why the call ended
	}
	chunks := &chunkWriter{stream: stream}
	buffered := bufio.NewWriterSize(chunks, sourceChunkSize)
	err := sendSource(buffered)
	if err == nil {
		err = buffered.Flush()
	}
	if chunks.ended {
		return nil
	}
	if err != nil {
		return hperrors.Wrap(err)
	}
	_ = stream.CloseSend()
	return nil
}

// chunkWriter sends what it is given as one chunk of the source.
type chunkWriter struct {
	stream grpc.BidiStreamingClient[agentproto.ImageBuildMsg, agentproto.ImageBuildResp]
	// ended is set once the stream refused a chunk: the call is over.
	ended bool
}

func (w *chunkWriter) Write(p []byte) (int, error) {
	// A copy: the message may still be in use after Send returns, and p is reused.
	chunk := append([]byte(nil), p...)
	err := w.stream.Send(&agentproto.ImageBuildMsg{Value: &agentproto.ImageBuildMsg_SourceChunk{SourceChunk: chunk}})
	if err != nil {
		w.ended = true
		return 0, err //nolint:wrapcheck
	}
	return len(p), nil
}

// receive reads the agent's answers to the end of the call: its logs, passed on
// as they come, and the result.
func (c *grpcImageBuildServiceClient) receive(
	ctx context.Context,
	stream grpc.BidiStreamingClient[agentproto.ImageBuildMsg, agentproto.ImageBuildResp],
	req *imagebuildagentdto.ImageBuildReq,
) (*imagebuildagentdto.ImageBuildResp, error) {
	respDTO := &imagebuildagentdto.ImageBuildResp{}

	for {
		resp, err := stream.Recv()
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return nil, hperrors.Wrap(err)
		}

		if logFrame := resp.GetLog(); logFrame != nil {
			frame := &tasklog.LogFrame{
				Type: tasklog.LogType(logFrame.GetType()),
				Data: logFrame.GetData(),
				Ts:   time.Unix(0, logFrame.GetTs()),
			}
			if req.SendLog != nil {
				_ = req.SendLog([]*tasklog.LogFrame{frame})
			} else if req.LogStore != nil {
				_ = req.LogStore.Add(ctx, frame)
			}
		}

		if result := resp.GetResult(); result != nil {
			respDTO.ImageTags = result.GetImageTags()
		}
	}

	return respDTO, nil
}
