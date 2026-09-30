// Package fileservice is the agent's gRPC face of fileagentuc.
package fileservice

import (
	"errors"
	"io"

	"google.golang.org/grpc"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/fileagentuc"
)

// chunkSize is what one message of a read carries.
const chunkSize = 64 * 1024

func FileRead(
	uc *fileagentuc.UC,
	req *agentproto.FileReq,
	stream grpc.ServerStreamingServer[agentproto.FileChunk],
) error {
	reader, _, err := uc.Open(req.GetRoot(), req.GetPath())
	if err != nil {
		return hperrors.Wrap(err)
	}
	defer reader.Close()

	buf := make([]byte, chunkSize)
	for {
		n, readErr := reader.Read(buf)
		if n > 0 {
			if err := stream.Send(&agentproto.FileChunk{Chunk: buf[:n]}); err != nil {
				return hperrors.Wrap(err)
			}
		}
		if readErr != nil {
			if errors.Is(readErr, io.EOF) {
				return nil
			}
			return hperrors.Wrap(readErr)
		}
	}
}

func FileWrite(
	uc *fileagentuc.UC,
	stream grpc.ClientStreamingServer[agentproto.FileWriteReq, agentproto.FileWriteResp],
) error {
	first, err := stream.Recv()
	if err != nil {
		return hperrors.Wrap(err)
	}
	target := first.GetTarget()
	if target == nil {
		return hperrors.NewMissing("File target")
	}

	pr, pw := io.Pipe()
	go feedWrite(stream, pw)

	size, err := uc.Write(stream.Context(), target.GetRoot(), target.GetPath(), pr)
	_ = pr.CloseWithError(err)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return hperrors.Wrap(stream.SendAndClose(&agentproto.FileWriteResp{Size: size}))
}

// feedWrite copies the chunks the client streams into pw, and ends it: cleanly
// when the client has sent everything, with the error otherwise.
func feedWrite(
	stream grpc.ClientStreamingServer[agentproto.FileWriteReq, agentproto.FileWriteResp],
	pw *io.PipeWriter,
) {
	var err error
	defer func() { _ = pw.CloseWithError(err) }()
	defer safego.RecoverTo(&err)
	for {
		msg, recvErr := stream.Recv()
		if recvErr != nil {
			if !errors.Is(recvErr, io.EOF) {
				err = hperrors.Wrap(recvErr)
			}
			return
		}
		if _, err = pw.Write(msg.GetChunk()); err != nil {
			return
		}
	}
}

func FileRemove(uc *fileagentuc.UC, req *agentproto.FileReq) (*agentproto.FileRemoveResp, error) {
	if err := uc.Remove(req.GetRoot(), req.GetPath()); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &agentproto.FileRemoveResp{}, nil
}

func FileStat(uc *fileagentuc.UC, req *agentproto.FileReq) (*agentproto.FileStatResp, error) {
	size, err := uc.Stat(req.GetRoot(), req.GetPath())
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &agentproto.FileStatResp{Size: size}, nil
}
