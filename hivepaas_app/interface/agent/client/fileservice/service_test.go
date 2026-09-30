package fileservice

import (
	"context"
	"io"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	serverfile "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/fileservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/fileagentuc"
)

// agentFiles is the agent's file service over a test directory standing for the
// node's filesystem.
type agentFiles struct {
	agentproto.UnimplementedFileServiceServer
	uc *fileagentuc.UC
}

func (a *agentFiles) FileRead(req *agentproto.FileReq, s grpc.ServerStreamingServer[agentproto.FileChunk]) error {
	return serverfile.FileRead(a.uc, req, s)
}

func (a *agentFiles) FileWrite(s grpc.ClientStreamingServer[agentproto.FileWriteReq, agentproto.FileWriteResp]) error {
	return serverfile.FileWrite(a.uc, s)
}

func (a *agentFiles) FileRemove(_ context.Context, req *agentproto.FileReq) (*agentproto.FileRemoveResp, error) {
	return serverfile.FileRemove(a.uc, req)
}

func (a *agentFiles) FileStat(_ context.Context, req *agentproto.FileReq) (*agentproto.FileStatResp, error) {
	return serverfile.FileStat(a.uc, req)
}

// startAgent serves the file service on a local port, over host/.
func startAgent(t *testing.T) (FileServiceClient, string) {
	t.Helper()
	config.SetCurrent(&config.Config{})
	t.Cleanup(func() { config.SetCurrent(nil) })

	host := t.TempDir()
	if err := os.MkdirAll(filepath.Join(host, "srv/vol"), 0o755); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	// As the agent's own server does, errors cross as gRPC statuses.
	server := grpc.NewServer(
		grpc.ChainUnaryInterceptor(func(ctx context.Context, req any, _ *grpc.UnaryServerInfo,
			handler grpc.UnaryHandler) (any, error) {
			resp, err := handler(ctx, req)
			return resp, hperrors.ToGRPCError(err)
		}),
		grpc.ChainStreamInterceptor(func(srv any, ss grpc.ServerStream, _ *grpc.StreamServerInfo,
			handler grpc.StreamHandler) error {
			return hperrors.ToGRPCError(handler(srv, ss))
		}),
	)
	agentproto.RegisterFileServiceServer(server, &agentFiles{uc: fileagentuc.NewOnHost(nil, host)})
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	c, err := NewFileServiceClient(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c, host
}

// A file larger than one message goes to the agent and comes back whole.
func TestAFileGoesThroughTheAgentAndBack(t *testing.T) {
	c, host := startAgent(t)
	ctx := context.Background()
	content := strings.Repeat("0123456789abcdef", 20000) // 320 KB: five messages

	size, err := c.Write(ctx, "/srv/vol", ".hivepaas/job-output/prod/web/out.sql", strings.NewReader(content))
	assert.NoError(t, err)
	assert.Equal(t, int64(len(content)), size)
	onDisk, _ := os.ReadFile(filepath.Join(host, "srv/vol/.hivepaas/job-output/prod/web/out.sql"))
	assert.Equal(t, content, string(onDisk))

	stated, err := c.Stat(ctx, "/srv/vol", ".hivepaas/job-output/prod/web/out.sql")
	assert.NoError(t, err)
	assert.Equal(t, size, stated)

	reader, err := c.Read(ctx, "/srv/vol", ".hivepaas/job-output/prod/web/out.sql")
	if assert.NoError(t, err) {
		back, err := io.ReadAll(reader)
		assert.NoError(t, err)
		assert.NoError(t, reader.Close())
		assert.Equal(t, content, string(back))
	}

	assert.NoError(t, c.Remove(ctx, "/srv/vol", ".hivepaas/job-output/prod/web/out.sql"))
	_, err = c.Read(ctx, "/srv/vol", ".hivepaas/job-output/prod/web/out.sql")
	assert.ErrorIs(t, err, hperrors.ErrNotFound)
	_, err = c.Stat(ctx, "/srv/vol", ".hivepaas/job-output/prod/web/out.sql")
	assert.ErrorIs(t, err, hperrors.ErrNotFound)
}

// The agent refuses a path out of the volume, and the refusal reaches the caller.
func TestTheAgentRefusesAPathOutOfTheVolume(t *testing.T) {
	c, _ := startAgent(t)

	_, err := c.Write(context.Background(), "/srv/vol", "../../etc/cron.d/x", strings.NewReader("x"))

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "is not inside the directory")
}
