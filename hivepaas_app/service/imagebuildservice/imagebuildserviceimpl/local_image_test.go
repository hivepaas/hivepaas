package imagebuildserviceimpl

import (
	"context"
	"testing"

	"github.com/moby/moby/api/types/registry"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/tasks/queue"
)

// A local image has its own name, and only that: no app names it, and it stays
// on the node it is built on.
func TestALocalImageIsNamedOnlyItsName(t *testing.T) {
	inputs := &imagebuildservice.BuildInputs{PushRegistry: &registry.AuthConfig{ServerAddress: "registry.example.com"}}

	refs, err := imageReferences(&imagebuildservice.ImageBuildReq{LocalImage: "hivepaas-function-libs:abc"}, inputs)

	assert.NoError(t, err)
	assert.Equal(t, []string{"hivepaas-function-libs:abc"}, refs)

	refs, err = imageReferences(&imagebuildservice.ImageBuildReq{App: buildApp(), CommitHash: "9f3c1de0ab"},
		&imagebuildservice.BuildInputs{})
	assert.NoError(t, err)
	assert.Equal(t, []string{"shop-api:dev-9f3c1de"}, refs)
}

func TestALocalImageIsNotPushed(t *testing.T) {
	data := &imageBuildData{
		ImageBuildReq: &imagebuildservice.ImageBuildReq{
			LocalImage:     "hivepaas-function-libs:abc",
			PushToRegistry: entity.ObjectID{ID: "registry-1"},
			TaskExecData:   &queue.TaskExecData{LogStore: tasklog.NewNullStore()},
		},
		Resp:   &imagebuildservice.ImageBuildResp{ImageTags: []string{"hivepaas-function-libs:abc"}},
		Inputs: &imagebuildservice.BuildInputs{},
	}
	// No docker manager: a push would panic.
	assert.NoError(t, (&service{}).imagePush(context.Background(), data))
}
