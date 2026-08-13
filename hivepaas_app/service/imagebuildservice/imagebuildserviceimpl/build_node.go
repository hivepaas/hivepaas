package imagebuildserviceimpl

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/imagebuildservice"
)

const (
	redisKeyBuildNodeActive = "build:node:%s:active"
	buildNodeSlotTTL        = 2 * time.Hour
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

	var available []*candidateNode
	maxParallelism := 0
	if buildSetting != nil {
		maxParallelism = buildSetting.Workers.MaxParallelism
	}

	for _, cand := range candidates {
		key := fmt.Sprintf(redisKeyBuildNodeActive, cand.node.ID)
		count, _ := s.redisClient.Get(ctx, key).Int()
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

	// Sort by activeCount ascending (least loaded node first)
	sort.SliceStable(available, func(i, j int) bool {
		if available[i].activeCount == available[j].activeCount {
			return available[i].node.ID < available[j].node.ID
		}
		return available[i].activeCount < available[j].activeCount
	})

	selected := available[0].node
	key := fmt.Sprintf(redisKeyBuildNodeActive, selected.ID)
	_ = s.redisClient.Incr(ctx, key).Err()
	_ = s.redisClient.Expire(ctx, key, buildNodeSlotTTL).Err()

	var once sync.Once
	releaseFunc := func() { //nolint:contextcheck
		once.Do(func() {
			releaseCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second) //nolint:mnd
			defer cancel()
			val, err := s.redisClient.Decr(releaseCtx, key).Result()
			if err == nil && val <= 0 {
				_ = s.redisClient.Del(releaseCtx, key).Err()
			}
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
