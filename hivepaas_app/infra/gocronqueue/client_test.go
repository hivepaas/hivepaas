package gocronqueue

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/logging/mocks"
)

type mockRedisClient struct {
	redis.UniversalClient
	rpushKey    string
	rpushValues []any
}

func (m *mockRedisClient) RPush(ctx context.Context, key string, values ...any) *redis.IntCmd {
	m.rpushKey = key
	m.rpushValues = values
	cmd := redis.NewIntCmd(ctx)
	cmd.SetVal(1)
	return cmd
}

func TestClient_ScheduleTask(t *testing.T) {
	mr := &mockRedisClient{}
	logger := &mocks.Logger{}
	client, _ := NewClient(mr, logger)

	ctx := context.Background()
	task := &entity.Task{ID: "task-1"}

	err := client.ScheduleTask(ctx, task)
	assert.NoError(t, err)

	assert.Equal(t, taskQueueCtrlKey, mr.rpushKey)
	assert.Len(t, mr.rpushValues, 1)

	// The value is marshaled as JSON string by redishelper.RPush
	jsonStr, ok := mr.rpushValues[0].(string)
	assert.True(t, ok)

	var msg Message
	err = json.Unmarshal([]byte(jsonStr), &msg)
	assert.NoError(t, err)
	assert.Len(t, msg.SchedTasks, 1)
	assert.Equal(t, task.ID, msg.SchedTasks[0].ID)
}

func TestClient_UnscheduleTask(t *testing.T) {
	mr := &mockRedisClient{}
	logger := &mocks.Logger{}
	client, _ := NewClient(mr, logger)

	ctx := context.Background()
	taskID := "task-1"

	err := client.UnscheduleTask(ctx, taskID)
	assert.NoError(t, err)

	assert.Equal(t, taskQueueCtrlKey, mr.rpushKey)
	assert.Len(t, mr.rpushValues, 1)

	jsonStr, ok := mr.rpushValues[0].(string)
	assert.True(t, ok)

	var msg Message
	err = json.Unmarshal([]byte(jsonStr), &msg)
	assert.NoError(t, err)
	assert.Equal(t, []string{taskID}, msg.UnschedTaskIDs)
}
