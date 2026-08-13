package imagebuildagentuc

import (
	"context"
	"errors"
	"sync"
	"testing"

	"github.com/moby/moby/api/types/swarm"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging/mocks"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecaseagent/imagebuildagentuc/imagebuildagentdto"
)

type mockImageBuildService struct {
	buildFunc func(ctx context.Context, db database.IDB, req *imagebuildservice.ImageBuildReq) (
		*imagebuildservice.ImageBuildResp, error)
}

func (m *mockImageBuildService) ImageBuild(
	ctx context.Context,
	db database.IDB,
	req *imagebuildservice.ImageBuildReq,
) (*imagebuildservice.ImageBuildResp, error) {
	if m.buildFunc != nil {
		return m.buildFunc(ctx, db, req)
	}
	return &imagebuildservice.ImageBuildResp{}, nil
}

func (m *mockImageBuildService) SelectBuildWorkerNode(
	ctx context.Context,
	buildSetting *entity.ImageBuildSettings,
) (*swarm.Node, error) {
	return nil, nil
}

func TestImageBuild_Success_SendLog(t *testing.T) {
	mockSvc := &mockImageBuildService{
		buildFunc: func(ctx context.Context, db database.IDB, req *imagebuildservice.ImageBuildReq) (
			*imagebuildservice.ImageBuildResp, error) {
			assert.NotNil(t, req.LogStore)
			_ = req.LogStore.Add(ctx, tasklog.NewOutFrame("step 1: downloading", tasklog.TsNow))
			_ = req.LogStore.Add(ctx, tasklog.NewOutFrame("step 2: building", tasklog.TsNow))
			return &imagebuildservice.ImageBuildResp{
				ImageTags: []string{"myimage:latest"},
			}, nil
		},
	}

	uc := New(&mocks.Logger{}, nil, nil, nil, mockSvc)

	var receivedLogs []*tasklog.LogFrame
	var mu sync.Mutex

	resp, err := uc.ImageBuild(context.Background(), &imagebuildagentdto.ImageBuildReq{
		TaskID: "task-123",
		SendLog: func(frames []*tasklog.LogFrame) error {
			mu.Lock()
			receivedLogs = append(receivedLogs, frames...)
			mu.Unlock()
			return nil
		},
	})

	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, []string{"myimage:latest"}, resp.ImageTags)

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, receivedLogs, 2)
	assert.Equal(t, "step 1: downloading", receivedLogs[0].Data)
	assert.Equal(t, "step 2: building", receivedLogs[1].Data)
}

func TestImageBuild_Error(t *testing.T) {
	expectedErr := errors.New("docker daemon error")
	mockSvc := &mockImageBuildService{
		buildFunc: func(ctx context.Context, db database.IDB, req *imagebuildservice.ImageBuildReq) (
			*imagebuildservice.ImageBuildResp, error) {
			_ = req.LogStore.Add(ctx, tasklog.NewErrFrame("build failed", tasklog.TsNow))
			return nil, expectedErr
		},
	}

	uc := New(&mocks.Logger{}, nil, nil, nil, mockSvc)

	var receivedLogs []*tasklog.LogFrame
	var mu sync.Mutex

	resp, err := uc.ImageBuild(context.Background(), &imagebuildagentdto.ImageBuildReq{
		TaskID: "task-456",
		SendLog: func(frames []*tasklog.LogFrame) error {
			mu.Lock()
			receivedLogs = append(receivedLogs, frames...)
			mu.Unlock()
			return nil
		},
	})

	assert.Error(t, err)
	assert.Nil(t, resp)

	mu.Lock()
	defer mu.Unlock()
	assert.Len(t, receivedLogs, 1)
	assert.Equal(t, "build failed", receivedLogs[0].Data)
}
