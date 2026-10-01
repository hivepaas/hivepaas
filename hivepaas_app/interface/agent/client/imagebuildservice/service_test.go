package imagebuildservice

import (
	"context"
	"crypto/rand"
	"errors"
	"io"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/moby/moby/api/types/registry"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	agentproto "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/proto"
	serverbuild "github.com/hivepaas/hivepaas/hivepaas_app/interface/agent/server/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging/mocks"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/srcpack"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc/imagebuildagentdto"
)

// builds stands for the agent's build: it sees the request and the directory the
// source was unpacked in.
type builds struct {
	imagebuildservice.Service
	mu    sync.Mutex
	calls int
	build func(ctx context.Context, req *imagebuildservice.ImageBuildReq) (*imagebuildservice.ImageBuildResp, error)
}

func (b *builds) ImageBuild(
	ctx context.Context, _ database.IDB, req *imagebuildservice.ImageBuildReq,
) (*imagebuildservice.ImageBuildResp, error) {
	b.mu.Lock()
	b.calls++
	b.mu.Unlock()
	return b.build(ctx, req)
}

func (b *builds) called() int {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.calls
}

type agentBuilds struct {
	agentproto.UnimplementedImageBuildServiceServer
	uc *imagebuildagentuc.UC
}

func (a *agentBuilds) ImageBuildFromSource(
	stream grpc.BidiStreamingServer[agentproto.ImageBuildMsg, agentproto.ImageBuildResp],
) error {
	return serverbuild.ImageBuildFromSource(a.uc, stream)
}

// startAgent serves a grpc server on a local port, with the agent's build
// service when svc is given, and says where the agent keeps its sources.
func startAgent(t *testing.T, svc *builds) (ImageBuildServiceClient, string) {
	t.Helper()
	config.SetCurrent(&config.Config{})
	t.Cleanup(func() { config.SetCurrent(nil) })

	tempBase := t.TempDir()
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	server := grpc.NewServer()
	if svc != nil {
		agentproto.RegisterImageBuildServiceServer(server, &agentBuilds{
			uc: imagebuildagentuc.New(&mocks.Logger{}, nil, nil, nil, svc).WithTempBaseDir(tempBase),
		})
	}
	go func() { _ = server.Serve(listener) }()
	t.Cleanup(server.Stop)

	c, err := NewImageBuildServiceClient(listener.Addr().String())
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = c.Close() })
	return c, tempBase
}

// checkout is a checked-out source: a Dockerfile and a file larger than one
// message, which does not compress.
func checkout(t *testing.T) (dir string, big []byte) {
	t.Helper()
	dir = t.TempDir()
	big = make([]byte, 3<<20)
	if _, err := rand.Read(big); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "Dockerfile"), []byte("FROM scratch\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "assets.bin"), big, 0o644); err != nil {
		t.Fatal(err)
	}
	return dir, big
}

// withInputs is a build request as the app sends it: with its inputs resolved.
func withInputs() *imagebuildagentdto.ImageBuildReq {
	return &imagebuildagentdto.ImageBuildReq{
		TaskID:        "t1",
		ImageBuildReq: imagebuildservice.ImageBuildReq{Inputs: &imagebuildservice.BuildInputs{}},
	}
}

func sendDir(dir string) func(io.Writer) error {
	return func(w io.Writer) error {
		_, err := srcpack.Pack(context.Background(), dir, w)
		return err
	}
}

func sourcesLeft(t *testing.T, tempBase string) []string {
	t.Helper()
	found, err := filepath.Glob(filepath.Join(tempBase, "*", "*"))
	if err != nil {
		t.Fatal(err)
	}
	return found
}

// The agent builds the source the app sends with the request the app made, and
// its logs and the image's tags come back.
func TestAnAgentBuildsTheSourceItIsSent(t *testing.T) {
	dir, big := checkout(t)
	var got *imagebuildservice.ImageBuildReq
	var gotBig []byte
	svc := &builds{build: func(ctx context.Context, req *imagebuildservice.ImageBuildReq) (
		*imagebuildservice.ImageBuildResp, error) {
		got = req
		gotBig, _ = os.ReadFile(filepath.Join(req.CheckoutDir, "assets.bin"))
		_ = req.LogStore.Add(ctx, tasklog.NewOutFrame("step 1", tasklog.TsNow))
		return &imagebuildservice.ImageBuildResp{ImageTags: []string{"registry/app:9f3c1de"}}, nil
	}}
	c, tempBase := startAgent(t, svc)

	npmToken, nodeEnv := "tok-123", "production"
	inputs := &imagebuildservice.BuildInputs{
		EnvVars:       map[string]*string{"NODE_ENV": &nodeEnv},
		SecretEnvVars: map[string]string{"NPM_TOKEN": npmToken},
		RegistryAuths: map[string]registry.AuthConfig{
			"docker.io": {Username: "puller", Password: "pull-pass", ServerAddress: "docker.io"},
		},
		PushRegistry: &registry.AuthConfig{
			Username: "hivepaas", Password: "push-pass", ServerAddress: "registry.example.com",
		},
		Secrets: []string{"tok-123", "pull-pass", "push-pass"},
	}

	var logs []string
	resp, err := c.ImageBuild(context.Background(), &imagebuildagentdto.ImageBuildReq{
		TaskID: "t1",
		ImageBuildReq: imagebuildservice.ImageBuildReq{
			CommitHash: "9f3c1de0ab",
			Dockerfile: entity.DeploymentDockerfile{Source: base.DockerfileSourceManual, Path: "Dockerfile"},
			ImageTags:  []string{"v1.4.0"},
			BuildID:    "b1",
			// The app's own directory, which the agent must not be told to use.
			CheckoutDir: dir,
			Inputs:      inputs,
		},
		SendLog: func(frames []*tasklog.LogFrame) error {
			for _, f := range frames {
				logs = append(logs, f.Data)
			}
			return nil
		},
	}, sendDir(dir))

	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, []string{"registry/app:9f3c1de"}, resp.ImageTags)
	}
	assert.Equal(t, []string{"step 1"}, logs)
	if assert.NotNil(t, got) {
		assert.Equal(t, "9f3c1de0ab", got.CommitHash)
		assert.Equal(t, []string{"v1.4.0"}, got.ImageTags)
		assert.Equal(t, "Dockerfile", got.Dockerfile.Path)
		assert.Equal(t, "b1", got.BuildID)
		assert.NotEqual(t, dir, got.CheckoutDir)
		// What the app resolved arrives whole: the agent opens no secret itself.
		assert.Equal(t, inputs, got.Inputs)
	}
	assert.Equal(t, big, gotBig)
	assert.Empty(t, sourcesLeft(t, tempBase))
}

// A build the agent fails gives the agent's reason.
func TestAFailedBuildGivesTheAgentsReason(t *testing.T) {
	dir, _ := checkout(t)
	svc := &builds{build: func(context.Context, *imagebuildservice.ImageBuildReq) (
		*imagebuildservice.ImageBuildResp, error) {
		return nil, hperrors.NewNotFound("Dockerfile")
	}}
	c, tempBase := startAgent(t, svc)

	_, err := c.ImageBuild(context.Background(), withInputs(), sendDir(dir))

	assert.ErrorContains(t, err, "Dockerfile")
	assert.NotErrorIs(t, err, io.EOF)
	assert.Empty(t, sourcesLeft(t, tempBase))
}

// When the deployment is canceled during the build, the agent's build is
// canceled too and its copy of the source goes.
func TestACanceledBuildLeavesNoSourceOnTheAgent(t *testing.T) {
	dir, _ := checkout(t)
	building := make(chan struct{})
	ended := make(chan struct{})
	svc := &builds{build: func(ctx context.Context, _ *imagebuildservice.ImageBuildReq) (
		*imagebuildservice.ImageBuildResp, error) {
		close(building)
		<-ctx.Done()
		close(ended)
		return nil, ctx.Err()
	}}
	c, tempBase := startAgent(t, svc)

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		<-building
		cancel()
	}()
	_, err := c.ImageBuild(ctx, withInputs(), sendDir(dir))

	assert.Error(t, err)
	select {
	case <-ended:
	case <-time.After(5 * time.Second):
		t.Fatal("the agent's build was not canceled")
	}
	assert.Eventually(t, func() bool { return len(sourcesLeft(t, tempBase)) == 0 },
		5*time.Second, 20*time.Millisecond)
}

// A source the app fails to pack fails the build with that reason, and the
// agent builds nothing from the part it received.
func TestASourceThatFailsToPackIsNotBuilt(t *testing.T) {
	svc := &builds{build: func(context.Context, *imagebuildservice.ImageBuildReq) (
		*imagebuildservice.ImageBuildResp, error) {
		return &imagebuildservice.ImageBuildResp{}, nil
	}}
	c, tempBase := startAgent(t, svc)

	_, err := c.ImageBuild(context.Background(), withInputs(),
		func(w io.Writer) error {
			_, _ = w.Write([]byte("half a source"))
			return errors.New("the checkout cannot be read")
		})

	assert.ErrorContains(t, err, "the checkout cannot be read")
	assert.Zero(t, svc.called())
	assert.Eventually(t, func() bool { return len(sourcesLeft(t, tempBase)) == 0 },
		5*time.Second, 20*time.Millisecond)
}

// An agent of an older image does not have the call, and says so at once.
func TestAnAgentWithoutTheCallSaysSo(t *testing.T) {
	dir, _ := checkout(t)
	c, _ := startAgent(t, nil)

	_, err := c.ImageBuild(context.Background(), withInputs(), sendDir(dir))

	assert.ErrorContains(t, err, "Unimplemented")
}

// A panic while the agent handles a build ends that build with an error: it
// does not take the agent down with every other build running on it.
func TestAPanicInABuildDoesNotTakeTheAgentDown(t *testing.T) {
	dir, _ := checkout(t)
	svc := &builds{build: func(context.Context, *imagebuildservice.ImageBuildReq) (
		*imagebuildservice.ImageBuildResp, error) {
		panic("a nil map in the build")
	}}
	c, tempBase := startAgent(t, svc)

	_, err := c.ImageBuild(context.Background(), withInputs(), sendDir(dir))

	assert.ErrorContains(t, err, "a nil map in the build")
	assert.Empty(t, sourcesLeft(t, tempBase))
}

// A build without a push registry arrives with none, not with an empty one the
// agent would try to push to.
func TestABuildThatIsNotPushedArrivesWithoutARegistry(t *testing.T) {
	dir, _ := checkout(t)
	var got *imagebuildservice.BuildInputs
	svc := &builds{build: func(_ context.Context, req *imagebuildservice.ImageBuildReq) (
		*imagebuildservice.ImageBuildResp, error) {
		got = req.Inputs
		return &imagebuildservice.ImageBuildResp{}, nil
	}}
	c, _ := startAgent(t, svc)

	_, err := c.ImageBuild(context.Background(), withInputs(), sendDir(dir))

	assert.NoError(t, err)
	if assert.NotNil(t, got) {
		assert.Nil(t, got.PushRegistry)
	}
}

// The app always resolves a build's inputs before sending it to an agent.
func TestABuildIsNotSentWithoutItsInputs(t *testing.T) {
	dir, _ := checkout(t)
	svc := &builds{build: func(context.Context, *imagebuildservice.ImageBuildReq) (
		*imagebuildservice.ImageBuildResp, error) {
		return &imagebuildservice.ImageBuildResp{}, nil
	}}
	c, _ := startAgent(t, svc)

	_, err := c.ImageBuild(context.Background(), &imagebuildagentdto.ImageBuildReq{TaskID: "t1"}, sendDir(dir))

	assert.ErrorIs(t, err, hperrors.ErrMissing)
	assert.Zero(t, svc.called())
}
