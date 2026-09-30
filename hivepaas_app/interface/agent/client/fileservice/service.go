// Package fileservice is the app's client of the agent's file service: files
// inside a volume's directory on the agent's node.
package fileservice

import (
	"context"
	"errors"
	"io"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/client"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
)

// writeChunkSize is what one message of a write carries.
const writeChunkSize = 64 * 1024

// FileServiceClient reaches the files inside root, a volume's directory on the
// agent's node. path is relative to root.
type FileServiceClient interface {
	Read(ctx context.Context, root, path string) (io.ReadCloser, error)
	Write(ctx context.Context, root, path string, content io.Reader) (int64, error)
	Remove(ctx context.Context, root, path string) error
	Stat(ctx context.Context, root, path string) (int64, error)
	Close() error
}

type grpcFileServiceClient struct {
	protoClient agentproto.FileServiceClient
	conn        *grpc.ClientConn
}

func NewFileServiceClient(agentAddr string) (FileServiceClient, error) {
	conn, err := grpc.NewClient(agentAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &grpcFileServiceClient{conn: conn, protoClient: agentproto.NewFileServiceClient(conn)}, nil
}

func (c *grpcFileServiceClient) Close() error {
	return hperrors.Wrap(c.conn.Close())
}

func (c *grpcFileServiceClient) Read(ctx context.Context, root, path string) (io.ReadCloser, error) {
	authCtx, cancel := client.CreateAuthCtxWithCancel(ctx)
	stream, err := c.protoClient.FileRead(authCtx, &agentproto.FileReq{Root: root, Path: path})
	if err != nil {
		cancel()
		return nil, fromAgent(err)
	}
	// The first message tells a missing file from a file, before the caller is
	// handed a reader.
	first, err := stream.Recv()
	if err != nil && !errors.Is(err, io.EOF) {
		cancel()
		return nil, fromAgent(err)
	}

	pr, pw := io.Pipe()
	go func() {
		var copyErr error
		defer func() { _ = pw.CloseWithError(copyErr) }()
		defer safego.RecoverTo(&copyErr)
		if first == nil {
			return
		}
		if _, copyErr = pw.Write(first.GetChunk()); copyErr != nil {
			return
		}
		for {
			msg, recvErr := stream.Recv()
			if recvErr != nil {
				if !errors.Is(recvErr, io.EOF) {
					copyErr = fromAgent(recvErr)
				}
				return
			}
			if _, copyErr = pw.Write(msg.GetChunk()); copyErr != nil {
				return
			}
		}
	}()
	return &cancelingReader{PipeReader: pr, cancel: cancel}, nil
}

func (c *grpcFileServiceClient) Write(ctx context.Context, root, path string, content io.Reader) (int64, error) {
	authCtx, cancel := client.CreateAuthCtxWithCancel(ctx)
	defer cancel()

	stream, err := c.protoClient.FileWrite(authCtx)
	if err != nil {
		return 0, fromAgent(err)
	}
	if err = stream.Send(&agentproto.FileWriteReq{
		Value: &agentproto.FileWriteReq_Target{Target: &agentproto.FileReq{Root: root, Path: path}},
	}); err != nil {
		return 0, fromAgent(sendError(stream, err))
	}

	buf := make([]byte, writeChunkSize)
	for {
		n, readErr := content.Read(buf)
		if n > 0 {
			if err = stream.Send(&agentproto.FileWriteReq{
				Value: &agentproto.FileWriteReq_Chunk{Chunk: buf[:n]},
			}); err != nil {
				return 0, fromAgent(sendError(stream, err))
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				break
			}
			// Canceling the stream makes the agent drop what it has written.
			return 0, hperrors.Wrap(readErr)
		}
	}
	resp, err := stream.CloseAndRecv()
	if err != nil {
		return 0, fromAgent(err)
	}
	return resp.GetSize(), nil
}

// sendError is why a send failed. A send on a stream the agent has ended
// returns a bare io.EOF; the agent's reason is what the stream then answers.
func sendError(
	stream grpc.ClientStreamingClient[agentproto.FileWriteReq, agentproto.FileWriteResp],
	err error,
) error {
	if !errors.Is(err, io.EOF) {
		return err
	}
	if _, recvErr := stream.CloseAndRecv(); recvErr != nil {
		return recvErr //nolint:wrapcheck // wrapped by the caller
	}
	return err
}

func (c *grpcFileServiceClient) Remove(ctx context.Context, root, path string) error {
	_, err := c.protoClient.FileRemove(client.CreateAuthCtx(ctx), &agentproto.FileReq{Root: root, Path: path})
	return fromAgent(err)
}

func (c *grpcFileServiceClient) Stat(ctx context.Context, root, path string) (int64, error) {
	resp, err := c.protoClient.FileStat(client.CreateAuthCtx(ctx), &agentproto.FileReq{Root: root, Path: path})
	if err != nil {
		return 0, fromAgent(err)
	}
	return resp.GetSize(), nil
}

// fromAgent is an agent's answer as HivePaaS's errors: a file the agent did not
// find is not found here too.
func fromAgent(err error) error {
	if err == nil {
		return nil
	}
	if st, ok := status.FromError(err); ok && st.Code() == codes.NotFound {
		return hperrors.NewNotFound("File").WithCause(err)
	}
	return hperrors.Wrap(err)
}

// cancelingReader ends the stream when the reader is closed early.
type cancelingReader struct {
	*io.PipeReader
	cancel context.CancelFunc
}

func (r *cancelingReader) Close() error {
	r.cancel()
	return hperrors.Wrap(r.PipeReader.Close())
}
