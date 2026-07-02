package gocronqueue

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestNewServer(t *testing.T) {
	config := &Config{
		Concurrency: 10,
	}
	server, err := NewServer(config)
	assert.NoError(t, err)
	assert.NotNil(t, server)
	assert.Equal(t, 10, server.config.Concurrency)
	assert.NotNil(t, server.scheduler)
	assert.NotNil(t, server.jobMap)
}

func TestNewServer_DefaultConcurrency(t *testing.T) {
	config := &Config{}
	server, err := NewServer(config)
	assert.NoError(t, err)
	assert.Equal(t, defaultConcurrency, server.config.Concurrency)
}

func TestServer_shouldSchedule(t *testing.T) {
	server := &Server{
		config: &Config{},
		jobMap: make(map[string]*jobData),
	}

	task := &entity.Task{ID: "task-1"}
	runAt := time.Now()

	// Case 1: Job doesn't exist
	assert.True(t, server.shouldSchedule(task, runAt))

	// Case 2: Job exists but with different time
	server.jobMap[task.ID] = &jobData{RunAt: runAt.Add(time.Hour)}
	assert.True(t, server.shouldSchedule(task, runAt))

	// Case 3: Job exists with same time
	server.jobMap[task.ID] = &jobData{RunAt: runAt}
	assert.False(t, server.shouldSchedule(task, runAt))
}
