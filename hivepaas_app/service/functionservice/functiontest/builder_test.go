package functiontest

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/functionservice/functionbuild"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
)

// fakeImageBuilds keeps the build it was asked for, and logs while it builds.
type fakeImageBuilds struct {
	imagebuildservice.Service
	req *imagebuildservice.ImageBuildReq
	log []string
}

func (f *fakeImageBuilds) ImageBuild(
	ctx context.Context, _ database.IDB, req *imagebuildservice.ImageBuildReq,
) (*imagebuildservice.ImageBuildResp, error) {
	f.req = req
	for _, line := range f.log {
		_ = req.LogStore.Add(ctx, tasklog.NewDebugFrame(line, tasklog.TsNow))
	}
	return &imagebuildservice.ImageBuildResp{ImageTags: []string{req.LocalImage}}, nil
}

// The libraries are built the way any build is - on this node, with the
// build's inputs and settings - as a local image from the Dockerfile written
// for them, and what the build wrote comes back.
func TestTheLibrariesAreBuiltAsALocalImage(t *testing.T) {
	builds := &fakeImageBuilds{log: []string{"#5 [deps 1/2] RUN hivepaas-runtime deps", "#5 DONE 3.1s"}}
	inputs := &imagebuildservice.BuildInputs{}

	log, err := NewLibrariesBuilder(builds).BuildLibraries(context.Background(), &LibrariesBuildReq{
		Image: "hivepaas-function-libs:abc", Dockerfile: "FROM node AS base\n",
		ContextDir: "/tmp/run/code", TempDir: "/tmp/run", Inputs: inputs,
	})

	assert.NoError(t, err)
	assert.Equal(t, "#5 [deps 1/2] RUN hivepaas-runtime deps\n#5 DONE 3.1s\n", log)
	req := builds.req
	assert.Equal(t, "hivepaas-function-libs:abc", req.LocalImage)
	assert.Equal(t, base.DockerfileSourceManual, req.Dockerfile.Source)
	assert.Equal(t, functionbuild.DockerfilePath, req.Dockerfile.Path)
	assert.Equal(t, "FROM node AS base\n", req.Dockerfile.Content)
	assert.Equal(t, "/tmp/run/code", req.CheckoutDir)
	assert.Same(t, inputs, req.Inputs)
	assert.Empty(t, req.PushToRegistry.ID)
}

// A long install keeps the end of its log, where the error is.
func TestALongInstallLogKeepsItsEnd(t *testing.T) {
	builds := &fakeImageBuilds{log: []string{strings.Repeat("x", int(LibrariesLogMax)), "npm error E404"}}

	log, err := NewLibrariesBuilder(builds).BuildLibraries(context.Background(), &LibrariesBuildReq{
		Image: "hivepaas-function-libs:abc", Inputs: &imagebuildservice.BuildInputs{},
	})

	assert.NoError(t, err)
	assert.Len(t, log, int(LibrariesLogMax))
	assert.True(t, strings.HasSuffix(log, "npm error E404\n"))
}
