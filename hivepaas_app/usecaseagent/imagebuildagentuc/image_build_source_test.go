package imagebuildagentuc

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging/mocks"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/srcpack"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc/imagebuildagentdto"
)

// packedSource is a source tree with one file, as the app sends it.
func packedSource(t *testing.T) *bytes.Buffer {
	t.Helper()
	src := t.TempDir()
	if err := os.WriteFile(filepath.Join(src, "Dockerfile"), []byte("FROM scratch\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	var packed bytes.Buffer
	if _, err := srcpack.Pack(context.Background(), src, &packed); err != nil {
		t.Fatal(err)
	}
	return &packed
}

// givenInputs is a build request as the app sends it: with its inputs resolved.
func givenInputs(taskID string) *imagebuildagentdto.ImageBuildReq {
	return &imagebuildagentdto.ImageBuildReq{
		TaskID:        taskID,
		ImageBuildReq: imagebuildservice.ImageBuildReq{Inputs: &imagebuildservice.BuildInputs{}},
	}
}

// leftovers are the build directories still in the agent's temporary directory.
func leftovers(t *testing.T, tempBase string) []string {
	t.Helper()
	found, err := filepath.Glob(filepath.Join(tempBase, "*", "*"))
	if err != nil {
		t.Fatal(err)
	}
	return found
}

// The build runs on the source it was sent, unpacked in a directory of the
// agent's own under the day's temporary directory, and removed afterwards.
func TestTheBuildRunsOnTheSourceItWasSent(t *testing.T) {
	tempBase := t.TempDir()
	var checkoutDir, dockerfile string
	mockSvc := &mockImageBuildService{
		buildFunc: func(_ context.Context, _ database.IDB, req *imagebuildservice.ImageBuildReq) (
			*imagebuildservice.ImageBuildResp, error) {
			checkoutDir = req.CheckoutDir
			content, err := os.ReadFile(filepath.Join(req.CheckoutDir, "Dockerfile"))
			dockerfile = string(content)
			return &imagebuildservice.ImageBuildResp{ImageTags: []string{"app:1"}}, err
		},
	}
	uc := New(&mocks.Logger{}, nil, nil, nil, mockSvc).WithTempBaseDir(tempBase)

	resp, err := uc.ImageBuildFromSource(context.Background(),
		givenInputs("t1"), packedSource(t))

	assert.NoError(t, err)
	if assert.NotNil(t, resp) {
		assert.Equal(t, []string{"app:1"}, resp.ImageTags)
	}
	assert.Equal(t, "FROM scratch\n", dockerfile)
	assert.True(t, strings.HasPrefix(checkoutDir, tempBase+string(filepath.Separator)), checkoutDir)
	assert.Equal(t, "checkout", filepath.Base(checkoutDir))
	assert.Empty(t, leftovers(t, tempBase))
}

// A build that fails leaves no source behind either.
func TestTheSourceIsRemovedAfterAFailedBuild(t *testing.T) {
	tempBase := t.TempDir()
	mockSvc := &mockImageBuildService{
		buildFunc: func(context.Context, database.IDB, *imagebuildservice.ImageBuildReq) (
			*imagebuildservice.ImageBuildResp, error) {
			return nil, errors.New("the Dockerfile does not build")
		},
	}
	uc := New(&mocks.Logger{}, nil, nil, nil, mockSvc).WithTempBaseDir(tempBase)

	_, err := uc.ImageBuildFromSource(context.Background(),
		givenInputs("t1"), packedSource(t))

	assert.ErrorContains(t, err, "the Dockerfile does not build")
	assert.Empty(t, leftovers(t, tempBase))
}

// A source that does not unpack is not built, and leaves nothing.
func TestASourceThatDoesNotUnpackIsNotBuilt(t *testing.T) {
	tempBase := t.TempDir()
	built := false
	mockSvc := &mockImageBuildService{
		buildFunc: func(context.Context, database.IDB, *imagebuildservice.ImageBuildReq) (
			*imagebuildservice.ImageBuildResp, error) {
			built = true
			return &imagebuildservice.ImageBuildResp{}, nil
		},
	}
	uc := New(&mocks.Logger{}, nil, nil, nil, mockSvc).WithTempBaseDir(tempBase)

	_, err := uc.ImageBuildFromSource(context.Background(),
		givenInputs("t1"), strings.NewReader("not a packed source"))

	assert.Error(t, err)
	assert.False(t, built)
	assert.Empty(t, leftovers(t, tempBase))
}

// An agent has no key to open a stored secret with. A build that comes without
// its inputs is refused before anything is unpacked, rather than failing later
// on the first secret it would have to open.
func TestABuildWithoutItsInputsIsRefused(t *testing.T) {
	tempBase := t.TempDir()
	built := false
	mockSvc := &mockImageBuildService{
		buildFunc: func(context.Context, database.IDB, *imagebuildservice.ImageBuildReq) (
			*imagebuildservice.ImageBuildResp, error) {
			built = true
			return &imagebuildservice.ImageBuildResp{}, nil
		},
	}
	uc := New(&mocks.Logger{}, nil, nil, nil, mockSvc).WithTempBaseDir(tempBase)

	_, err := uc.ImageBuildFromSource(context.Background(),
		&imagebuildagentdto.ImageBuildReq{TaskID: "t1"}, packedSource(t))

	assert.ErrorIs(t, err, hperrors.ErrMissing)
	assert.False(t, built)
	assert.Empty(t, leftovers(t, tempBase))
}
