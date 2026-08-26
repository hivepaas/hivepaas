package imagebuildserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/redis/go-redis/v9"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/nanoid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
)

const (
	redisKeyBuildNodesSlots = "build:nodes:slots"
	buildNodeSlotTTL        = 30 * time.Minute
)

type candidateNode struct {
	node        *swarm.Node
	activeCount int
}

func (s *service) SelectBuildWorkerNode(
	ctx context.Context,
	buildSetting *entity.ImageBuildSettings,
) (resp imagebuildservice.BuildNodeResp, err error) {
	candidates, currNodeID, err := s.findBuildNodeCandidates(ctx, buildSetting)
	if err != nil {
		return resp, apperrors.Wrap(err)
	}
	resp.CurrentNodeID = currNodeID

	if len(candidates) == 0 {
		return resp, nil
	}

	// 1. Purge expired slots and fetch current active slots in 1 roundtrip
	now := time.Now().Unix()
	pipe := s.redisClient.Pipeline()
	pipe.ZRemRangeByScore(ctx, redisKeyBuildNodesSlots, "-inf", fmt.Sprint(now))
	membersCmd := pipe.ZRange(ctx, redisKeyBuildNodesSlots, 0, -1)
	if _, err = pipe.Exec(ctx); err != nil && !errors.Is(err, redis.Nil) {
		return resp, apperrors.Wrap(err)
	}

	activeCounts := make(map[string]int)
	for _, member := range membersCmd.Val() {
		if parts := strings.SplitN(member, ":", 2); len(parts) == 2 { //nolint:mnd
			activeCounts[parts[0]]++
		}
	}

	// 2. Filter available nodes based on maxParallelism
	var available []*candidateNode
	maxParallelism := 0
	if buildSetting != nil {
		maxParallelism = buildSetting.Workers.MaxParallelism
	}

	for _, cand := range candidates {
		count := activeCounts[cand.node.ID]
		cand.activeCount = count

		if maxParallelism > 0 && count >= maxParallelism {
			continue
		}
		available = append(available, cand)
	}

	if len(available) == 0 {
		// All candidate nodes (including current node) are at max parallelism
		return resp, nil
	}

	// 3. Sort by activeCount ascending (least loaded node first)
	sort.SliceStable(available, func(i, j int) bool {
		if available[i].activeCount == available[j].activeCount {
			return available[i].node.ID < available[j].node.ID
		}
		return available[i].activeCount < available[j].activeCount
	})

	// 4. Allocate a unique slot for the selected node
	selected := available[0].node
	slotID := nanoid.NewStandard16()
	slotMember := fmt.Sprintf("%s:%s", selected.ID, slotID)
	expireAt := float64(time.Now().Add(buildNodeSlotTTL).Unix())

	if err = s.redisClient.ZAdd(ctx, redisKeyBuildNodesSlots, redis.Z{
		Score:  expireAt,
		Member: slotMember,
	}).Err(); err != nil {
		return resp, apperrors.Wrap(err)
	}

	var once sync.Once
	releaseFunc := func() { //nolint:contextcheck
		once.Do(func() {
			releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //nolint:mnd
			defer cancel()
			_ = s.redisClient.ZRem(releaseCtx, redisKeyBuildNodesSlots, slotMember).Err()
		})
	}
	resp.Node = selected
	resp.ReleaseNodeFunc = releaseFunc
	return resp, nil
}

//nolint:gocognit
func (s *service) findBuildNodeCandidates(
	ctx context.Context,
	buildSetting *entity.ImageBuildSettings,
) (candidates []*candidateNode, currNodeID string, err error) {
	currNodeID, err = s.dockerManager.NodeCurrentID(ctx)
	if err != nil {
		return nil, "", apperrors.Wrap(err)
	}

	nodesResp, err := s.dockerManager.NodeList(ctx)
	if err != nil {
		return nil, "", apperrors.Wrap(err)
	}
	nodeList := nodesResp.Items

	hasWorkerFilter := buildSetting != nil &&
		(len(buildSetting.Workers.NodeIDs) > 0 || len(buildSetting.Workers.NodeLabels) > 0)

	if !hasWorkerFilter {
		// If no worker nodes configured, use the current node as the candidate
		var currNode *swarm.Node
		for _, node := range nodeList {
			if node.ID == currNodeID {
				currNode = &node
				break
			}
		}
		if currNode != nil {
			candidates = append(candidates, &candidateNode{node: currNode})
		} else if currNodeID != "" {
			candidates = append(candidates, &candidateNode{node: &swarm.Node{ID: currNodeID}})
		}
		return candidates, currNodeID, nil
	}

	for _, node := range nodeList {
		if node.Status.State != swarm.NodeStateReady {
			continue
		}
		matched := false
		// Match by NodeIDs
		if len(buildSetting.Workers.NodeIDs) > 0 {
			if gofn.Contain(buildSetting.Workers.NodeIDs, node.ID) ||
				gofn.Contain(buildSetting.Workers.NodeIDs, node.Description.Hostname) {
				matched = true
			}
		}
		// Match by NodeLabels
		if !matched && len(buildSetting.Workers.NodeLabels) > 0 {
			labelMatched := true
			for _, label := range buildSetting.Workers.NodeLabels {
				parts := strings.SplitN(label, "=", 2) //nolint:mnd
				key := strings.TrimSpace(parts[0])
				val := ""
				if len(parts) == 2 { //nolint:mnd
					val = strings.TrimSpace(parts[1])
				}
				if actualVal, ok := node.Spec.Labels[key]; !ok || (val != "" && actualVal != val) {
					labelMatched = false
					break
				}
			}
			if labelMatched {
				matched = true
			}
		}

		if matched {
			candidates = append(candidates, &candidateNode{node: &node})
		}
	}
	return candidates, currNodeID, nil
}
