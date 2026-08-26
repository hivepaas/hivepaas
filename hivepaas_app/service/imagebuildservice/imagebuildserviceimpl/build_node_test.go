package imagebuildserviceimpl

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/services/docker"
)

type mockRedisForBuildNode struct {
	redis.UniversalClient
	zsetData map[string]float64 // member -> expireScore
	zaddErr  error
	pipeErr  error
}

func newMockRedisForBuildNode() *mockRedisForBuildNode {
	return &mockRedisForBuildNode{
		zsetData: make(map[string]float64),
	}
}

func (m *mockRedisForBuildNode) Pipeline() redis.Pipeliner {
	return &mockPipelinerForBuildNode{mock: m}
}

func (m *mockRedisForBuildNode) ZAdd(ctx context.Context, key string, members ...redis.Z) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	if m.zaddErr != nil {
		cmd.SetErr(m.zaddErr)
		return cmd
	}
	for _, z := range members {
		m.zsetData[fmt.Sprint(z.Member)] = z.Score
	}
	cmd.SetVal(int64(len(members)))
	return cmd
}

func (m *mockRedisForBuildNode) ZRem(ctx context.Context, key string, members ...any) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	count := int64(0)
	for _, member := range members {
		memStr := fmt.Sprint(member)
		if _, ok := m.zsetData[memStr]; ok {
			delete(m.zsetData, memStr)
			count++
		}
	}
	cmd.SetVal(count)
	return cmd
}

func (m *mockRedisForBuildNode) Close() error {
	return nil
}

type mockPipelinerForBuildNode struct {
	redis.Pipeliner
	mock *mockRedisForBuildNode
	cmds []redis.Cmder
}

func (p *mockPipelinerForBuildNode) ZRemRangeByScore(ctx context.Context, key, min, max string) *redis.IntCmd {
	cmd := redis.NewIntCmd(ctx)
	maxVal, _ := strconv.ParseFloat(max, 64)
	removed := int64(0)
	for mem, score := range p.mock.zsetData {
		if score <= maxVal {
			delete(p.mock.zsetData, mem)
			removed++
		}
	}
	cmd.SetVal(removed)
	p.cmds = append(p.cmds, cmd)
	return cmd
}

func (p *mockPipelinerForBuildNode) ZRange(ctx context.Context, key string, start, stop int64) *redis.StringSliceCmd {
	cmd := redis.NewStringSliceCmd(ctx)
	var members []string
	for mem := range p.mock.zsetData {
		members = append(members, mem)
	}
	cmd.SetVal(members)
	p.cmds = append(p.cmds, cmd)
	return cmd
}

func (p *mockPipelinerForBuildNode) Exec(ctx context.Context) ([]redis.Cmder, error) {
	if p.mock.pipeErr != nil {
		return nil, p.mock.pipeErr
	}
	return p.cmds, nil
}

type mockDockerManagerForBuildNode struct {
	docker.Manager
	currNodeID string
	nodes      []swarm.Node
}

func (m *mockDockerManagerForBuildNode) NodeCurrentID(ctx context.Context) (string, error) {
	return m.currNodeID, nil
}

func (m *mockDockerManagerForBuildNode) NodeList(ctx context.Context, options ...docker.NodeListOption) (
	*client.NodeListResult, error) {
	return &client.NodeListResult{Items: m.nodes}, nil
}

func TestSelectBuildWorkerNode_RedisZSet(t *testing.T) {
	t.Run("select least loaded node and auto purge expired slots", func(t *testing.T) {
		mr := newMockRedisForBuildNode()
		future := float64(time.Now().Add(1 * time.Hour).Unix())
		past := float64(time.Now().Add(-1 * time.Hour).Unix())

		// node-1 has 3 active slots
		mr.zsetData["node-1:slot-1"] = future
		mr.zsetData["node-1:slot-2"] = future
		mr.zsetData["node-1:slot-3"] = future

		// node-2 has 1 active slot, and 1 expired slot (should be purged automatically!)
		mr.zsetData["node-2:slot-1"] = future
		mr.zsetData["node-2:slot-leaked"] = past

		// node-3 has 5 active slots (over maxParallelism 4)
		mr.zsetData["node-3:slot-1"] = future
		mr.zsetData["node-3:slot-2"] = future
		mr.zsetData["node-3:slot-3"] = future
		mr.zsetData["node-3:slot-4"] = future
		mr.zsetData["node-3:slot-5"] = future

		md := &mockDockerManagerForBuildNode{
			currNodeID: "node-1",
			nodes: []swarm.Node{
				{ID: "node-1", Status: swarm.NodeStatus{State: swarm.NodeStateReady}},
				{ID: "node-2", Status: swarm.NodeStatus{State: swarm.NodeStateReady}},
				{ID: "node-3", Status: swarm.NodeStatus{State: swarm.NodeStateReady}},
			},
		}

		svc := &service{
			redisClient:   mr,
			dockerManager: md,
		}

		setting := &entity.ImageBuildSettings{
			Workers: entity.ImageBuildWorkerSettings{
				NodeIDs:        []string{"node-1", "node-2", "node-3"},
				MaxParallelism: 4,
			},
		}

		resp, err := svc.SelectBuildWorkerNode(context.Background(), setting)
		assert.NoError(t, err)
		assert.NotNil(t, resp.Node)

		// Leaked slot was purged from node-2, so node-2 has activeCount = 1 (least loaded, compared to node-1 with 3)
		assert.Equal(t, "node-2", resp.Node.ID)
		assert.NotContains(t, mr.zsetData, "node-2:slot-leaked")

		// Verify new slot was allocated in ZSET for node-2
		var allocatedSlot string
		for mem := range mr.zsetData {
			if strings.HasPrefix(mem, "node-2:") && mem != "node-2:slot-1" {
				allocatedSlot = mem
				break
			}
		}
		assert.NotEmpty(t, allocatedSlot)

		// Release node
		assert.NotNil(t, resp.ReleaseNodeFunc)
		resp.ReleaseNodeFunc()
		assert.NotContains(t, mr.zsetData, allocatedSlot)
	})

	t.Run("all nodes at max parallelism returns nil node", func(t *testing.T) {
		mr := newMockRedisForBuildNode()
		future := float64(time.Now().Add(1 * time.Hour).Unix())
		mr.zsetData["node-1:slot-1"] = future
		mr.zsetData["node-1:slot-2"] = future

		md := &mockDockerManagerForBuildNode{
			currNodeID: "node-1",
			nodes: []swarm.Node{
				{ID: "node-1", Status: swarm.NodeStatus{State: swarm.NodeStateReady}},
			},
		}

		svc := &service{
			redisClient:   mr,
			dockerManager: md,
		}

		setting := &entity.ImageBuildSettings{
			Workers: entity.ImageBuildWorkerSettings{
				NodeIDs:        []string{"node-1"},
				MaxParallelism: 2,
			},
		}

		resp, err := svc.SelectBuildWorkerNode(context.Background(), setting)
		assert.NoError(t, err)
		assert.Nil(t, resp.Node)
		assert.Equal(t, "node-1", resp.CurrentNodeID)
	})

	t.Run("pipeline exec error returns error", func(t *testing.T) {
		mr := newMockRedisForBuildNode()
		mr.pipeErr = assert.AnError

		md := &mockDockerManagerForBuildNode{
			currNodeID: "node-1",
			nodes: []swarm.Node{
				{ID: "node-1", Status: swarm.NodeStatus{State: swarm.NodeStateReady}},
			},
		}

		svc := &service{
			redisClient:   mr,
			dockerManager: md,
		}

		setting := &entity.ImageBuildSettings{
			Workers: entity.ImageBuildWorkerSettings{
				NodeIDs: []string{"node-1"},
			},
		}

		_, err := svc.SelectBuildWorkerNode(context.Background(), setting)
		assert.Error(t, err)
	})

	t.Run("ZAdd error returns error", func(t *testing.T) {
		mr := newMockRedisForBuildNode()
		mr.zaddErr = assert.AnError

		md := &mockDockerManagerForBuildNode{
			currNodeID: "node-1",
			nodes: []swarm.Node{
				{ID: "node-1", Status: swarm.NodeStatus{State: swarm.NodeStateReady}},
			},
		}

		svc := &service{
			redisClient:   mr,
			dockerManager: md,
		}

		setting := &entity.ImageBuildSettings{
			Workers: entity.ImageBuildWorkerSettings{
				NodeIDs: []string{"node-1"},
			},
		}

		_, err := svc.SelectBuildWorkerNode(context.Background(), setting)
		assert.Error(t, err)
	})
}
