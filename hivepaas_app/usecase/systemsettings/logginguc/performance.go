package logginguc

import (
	"context"
	"errors"
	"sort"

	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/obi"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/systemsettings/logginguc/loggingdto"
)

// GetLoggingPerformance lists the swarm's nodes for the collection of apps'
// routes and calls by OBI: which run it, at which capacity, and what each
// node's agent last said of it.
func (uc *UC) GetLoggingPerformance(
	ctx context.Context,
	_ *basedto.Auth,
	_ *loggingdto.GetLoggingPerformanceReq,
) (*loggingdto.GetLoggingPerformanceResp, error) {
	setting, err := uc.SettingRepo.GetSingle(ctx, uc.DB, entity.NewObjectScopeGlobal(), currentSettingType, false)
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	cfg := &entity.LoggingSettings{}
	data := &loggingdto.LoggingPerformanceResp{Capacities: loggingdto.PerformanceCapacities()}
	if setting != nil {
		if cfg, err = setting.AsLoggingSettings(); err != nil {
			return nil, hperrors.Wrap(err)
		}
		data.Configured, data.UpdateVer = true, setting.UpdateVer
	}
	chosen := map[string]*entity.LoggingPerformanceNode{}
	if cfg.Performance != nil {
		data.Enabled = cfg.Performance.Enabled
		for _, node := range cfg.Performance.Nodes {
			if node != nil {
				chosen[node.ID] = node
			}
		}
	}
	data.LogsStored = cfg.Enabled && (cfg.Sources.Apps || cfg.Sources.HivePaaS)

	nodes, err := uc.dockerManager.NodeList(ctx)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	statuses := map[string]*loggingservice.PerformanceNodeStatus{}
	if !data.LogsStored {
		data.StatusReason = loggingdto.PerformanceStatusReasonLogsNotStored
	} else if statuses, err = uc.loggingService.PerformanceStatus(ctx, uc.DB); err != nil {
		// The settings are shown without: the nodes can be chosen while the
		// logs cannot be read.
		data.StatusReason, statuses = loggingdto.PerformanceStatusReasonUnreadable, nil
	}
	data.Nodes = make([]*loggingdto.PerformanceNodeResp, 0, len(nodes.Items))
	for i := range nodes.Items {
		data.Nodes = append(data.Nodes, performanceNode(&nodes.Items[i], chosen[nodes.Items[i].ID], statuses))
	}
	sort.SliceStable(data.Nodes, func(i, j int) bool { return data.Nodes[i].Hostname < data.Nodes[j].Hostname })
	return &loggingdto.GetLoggingPerformanceResp{Data: data}, nil
}

// performanceNode is a node as the settings show it: chosen nil when it runs
// no OBI.
func performanceNode(
	node *swarm.Node,
	chosen *entity.LoggingPerformanceNode,
	statuses map[string]*loggingservice.PerformanceNodeStatus,
) *loggingdto.PerformanceNodeResp {
	memory := node.Description.Resources.MemoryBytes
	out := &loggingdto.PerformanceNodeResp{
		ID:           node.ID,
		Hostname:     node.Description.Hostname,
		Role:         string(node.Spec.Role),
		State:        string(node.Status.State),
		Availability: string(node.Spec.Availability),
		MemoryBytes:  memory,
		Recommended:  string(obi.Recommended(int(memory >> 20))), //nolint:mnd // bytes to MB
		Capacity:     string(obi.CapacityAuto),
	}
	if chosen != nil {
		out.Enabled = true
		if c, ok := obi.ParseCapacity(chosen.Capacity); ok {
			out.Capacity = string(c)
		}
	}
	if status := statuses[node.ID]; status != nil {
		out.Status = loggingdto.PerformanceNodeStatus(status.Time, &status.Status)
	}
	return out
}

// UpdateLoggingPerformance saves which nodes run OBI, and at which capacity.
// It is saved with the logging settings, apart from them: nothing is applied,
// the agents read it on their own within 30 seconds.
func (uc *UC) UpdateLoggingPerformance(
	ctx context.Context,
	auth *basedto.Auth,
	req *loggingdto.UpdateLoggingPerformanceReq,
) (*loggingdto.UpdateLoggingPerformanceResp, error) {
	req.Type = currentSettingType
	req.Auth = auth
	next := req.ToEntity()
	if err := uc.checkPerformanceNodes(ctx, next); err != nil {
		return nil, hperrors.Wrap(err)
	}

	var current *entity.LoggingSettings
	_, err := uc.UpdateUniqueSetting(ctx, &req.UpdateUniqueSettingReq, &settings.UpdateUniqueSettingData{
		Name: loggingSettingName,
		Load: func(ctx context.Context, db database.Tx, data *settings.UpdateUniqueSettingData) error {
			setting, err := uc.SettingRepo.GetSingle(ctx, db, req.Scope, currentSettingType, false,
				bunex.SelectFor("UPDATE OF setting"))
			if errors.Is(err, hperrors.ErrNotFound) || (err == nil && setting == nil) {
				// Saved without the logging settings, these would make them,
				// unapplied and without their defaults.
				return hperrors.Wrap(hperrors.ErrLoggingNotConfigured)
			}
			if err != nil {
				return hperrors.Wrap(err)
			}
			data.Setting = setting
			current, err = setting.AsLoggingSettings()
			return hperrors.Wrap(err)
		},
		PrepareUpdate: func(
			_ context.Context,
			_ database.Tx,
			_ *settings.UpdateUniqueSettingData,
			pData *settings.PersistingSettingData,
		) error {
			updated := *current
			updated.Performance = next
			return hperrors.Wrap(pData.Setting.SetData(&updated))
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &loggingdto.UpdateLoggingPerformanceResp{}, nil
}

// checkPerformanceNodes refuses a node that is no node of the swarm.
func (uc *UC) checkPerformanceNodes(ctx context.Context, perf *entity.LoggingPerformance) error {
	if len(perf.Nodes) == 0 {
		return nil
	}
	nodes, err := uc.dockerManager.NodeList(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	known := make(map[string]bool, len(nodes.Items))
	for i := range nodes.Items {
		known[nodes.Items[i].ID] = true
	}
	for _, node := range perf.Nodes {
		if !known[node.ID] {
			return hperrors.Wrap(hperrors.ErrLoggingPerformanceNodeUnknown).WithParam("Name", node.ID)
		}
	}
	return nil
}
