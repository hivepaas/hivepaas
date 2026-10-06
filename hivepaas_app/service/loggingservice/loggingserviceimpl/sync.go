package loggingserviceimpl

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/loggingservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

// obiStatusWithin is how far back a node's status about OBI is looked for: an
// agent writes one every minute while the feature is on.
const obiStatusWithin = 15 * time.Minute

func (s *service) Sync(ctx context.Context, db database.IDB) (*loggingservice.SyncResp, error) {
	resp := &loggingservice.SyncResp{}
	setting, err := s.settingRepo.GetSingle(ctx, db, entity.NewObjectScopeGlobal(), base.SettingTypeLogging, true,
		bunex.SelectFor("UPDATE OF setting"))
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	cfg := &entity.LoggingSettings{}
	if setting != nil {
		if cfg, err = setting.AsLoggingSettings(); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	apps := []*systemappservice.SyncedApp{
		{Key: backendAppKey, Name: backendAppName, Wanted: cfg.Enabled && cfg.Backend.Managed},
		{Key: collectorAppKey, Name: collectorAppName, Wanted: cfg.Enabled && cfg.Collector.Managed},
	}
	for _, app := range apps {
		if err = systemappservice.RemoveIfServiceGone(ctx, db, s.systemAppService, app); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}

	if setting == nil {
		// Never configured, or its row is gone: no app is wanted, and Apply
		// would return before looking. The collector goes first, as Apply has
		// it: nothing ships into a store that is going away.
		for i := len(apps) - 1; i >= 0; i-- {
			if app := apps[i]; app.Before != nil {
				if err = s.systemAppService.Remove(ctx, db, app.Before, false); err != nil {
					return nil, hperrors.Wrap(err)
				}
			}
		}
	} else {
		applied, applyErr := s.Apply(ctx, db, &loggingservice.SettingApplyReq{Setting: setting, RemoveApp: true})
		if applied != nil {
			resp.Tasks, resp.Cleanup = applied.Tasks, applied.Cleanup
		}
		if applyErr != nil {
			return resp, hperrors.Wrap(applyErr)
		}
	}

	backendFresh := false
	for _, app := range apps {
		out, fresh, err := systemappservice.SyncOutcome(ctx, db, s.systemAppService, app, resp.Tasks)
		if err != nil {
			return resp, hperrors.Wrap(err)
		}
		if out != nil {
			resp.Apps = append(resp.Apps, out)
		}
		backendFresh = backendFresh || (app.Key == backendAppKey && fresh)
	}

	resp.OBI, resp.OBISettingsChanged, err = s.checkOBI(ctx, db, setting, cfg, backendFresh)
	return resp, hperrors.Wrap(err)
}

// checkOBI checks OBI on the nodes against the settings: a node no longer in
// the cluster leaves them; one whose agent does not do what they say has the
// agents read them again; the rest is reported. It says whether the agents are
// to read the settings again.
func (s *service) checkOBI(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
	cfg *entity.LoggingSettings,
	backendFresh bool,
) (*entity.OBISyncOutput, bool, error) {
	perf := cfg.Performance
	if setting == nil || perf == nil || len(perf.Nodes) == 0 {
		return nil, false, nil
	}
	nodes, err := s.dockerManager.NodeList(ctx)
	if err != nil {
		return nil, false, hperrors.Wrap(err)
	}
	hostnames := make(map[string]string, len(nodes.Items))
	for i := range nodes.Items {
		hostnames[nodes.Items[i].ID] = nodes.Items[i].Description.Hostname
	}

	out := &entity.OBISyncOutput{}
	changed, err := s.removeGoneOBINodes(ctx, db, setting, cfg, hostnames, out)
	if err != nil {
		return nil, false, hperrors.Wrap(err)
	}
	switch {
	case !perf.Enabled || len(perf.Nodes) == 0:
		return outputOrNil(out), changed, nil
	case !cfg.OBIOn():
		out.Problem = "Performance is on, but logging is off or stores neither the apps' logs nor HivePaaS's: " +
			"OBI does not run"
		return out, changed, nil
	case backendFresh:
		out.Unknown = "the log store was just created: the nodes' status is read from it"
		return out, changed, nil
	}
	statuses, err := s.PerformanceStatus(ctx, db, obiStatusWithin)
	if err != nil {
		// Said, not failed: the store's own trouble is the backend's to report.
		out.Unknown = "the nodes' status could not be read from the logs: " + err.Error()
		return out, changed, nil //nolint:nilerr
	}
	agents, err := s.agentNodes(ctx)
	if err != nil {
		return nil, false, hperrors.Wrap(err)
	}

	expected := map[string]bool{}
	for _, node := range perf.Nodes {
		if node == nil {
			continue
		}
		expected[node.ID] = true
		check := judgeOBINode(statuses[node.ID], agents[node.ID])
		check.NodeID, check.Hostname, check.Expected = node.ID, hostnames[node.ID], true
		changed = changed || check.Action == entity.OBISyncCacheCleared
		out.Nodes = append(out.Nodes, check)
	}
	for id, status := range statuses {
		if expected[id] || !status.Running {
			continue
		}
		changed = true
		out.Nodes = append(out.Nodes, &entity.OBINodeSyncOutput{NodeID: id, Hostname: hostnames[id],
			LastSeen: &status.Time, Wanted: status.Wanted, Running: true, Action: entity.OBISyncCacheCleared,
			Problem: "OBI runs there, though the settings do not list the node"})
	}
	sort.SliceStable(out.Nodes, func(i, j int) bool { return out.Nodes[i].Hostname < out.Nodes[j].Hostname })
	return out, changed, nil
}

// judgeOBINode says what a node the settings list is doing about OBI, by its
// latest status, and whether its agent has a task running there.
func judgeOBINode(status *loggingservice.PerformanceNodeStatus, agentRunning bool) *entity.OBINodeSyncOutput {
	out := &entity.OBINodeSyncOutput{Action: entity.OBISyncNone}
	switch {
	case status == nil && !agentRunning:
		out.Action, out.Problem = entity.OBISyncReported, "its agent is not running"
		return out
	case status == nil:
		out.Action, out.Problem = entity.OBISyncReported, "its agent wrote no status in 15 minutes"
		return out
	}
	out.LastSeen, out.Wanted, out.Running = &status.Time, status.Wanted, status.Running
	switch {
	case !status.Wanted:
		out.Action, out.Problem = entity.OBISyncCacheCleared, "its agent did not know it runs OBI"
	case status.Running:
	case !status.Preflight.OK:
		out.Action = entity.OBISyncReported
		out.Problem = "OBI cannot run there: " + strings.Join(status.Preflight.Reasons, ", ")
	case status.Apps == 0:
		// No app asks for its routes and calls: nothing to run.
	default:
		out.Action, out.Problem = entity.OBISyncReported, "OBI is not running"
	}
	return out
}

// removeGoneOBINodes takes the nodes no longer in the cluster off the ones
// that run OBI, and saves the settings when it did, as rememberApps does.
func (s *service) removeGoneOBINodes(
	ctx context.Context,
	db database.IDB,
	setting *entity.Setting,
	cfg *entity.LoggingSettings,
	hostnames map[string]string,
	out *entity.OBISyncOutput,
) (bool, error) {
	perf := cfg.Performance
	kept := make([]*entity.LoggingPerformanceNode, 0, len(perf.Nodes))
	for _, node := range perf.Nodes {
		if node == nil {
			continue
		}
		if _, ok := hostnames[node.ID]; ok {
			kept = append(kept, node)
			continue
		}
		out.Nodes = append(out.Nodes, &entity.OBINodeSyncOutput{NodeID: node.ID, Expected: true,
			Action: entity.OBISyncNodeRemoved, Problem: "the node is no longer in the cluster"})
	}
	if len(kept) == len(perf.Nodes) {
		return false, nil
	}
	s.logger.Info("taking nodes no longer in the cluster off OBI's", "removed", len(perf.Nodes)-len(kept))
	perf.Nodes = kept
	if err := setting.SetData(cfg); err != nil {
		return false, hperrors.Wrap(err)
	}
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()
	return true, hperrors.Wrap(s.settingRepo.Upsert(ctx, db, setting,
		entity.SettingUpsertingConflictCols, entity.SettingUpsertingUpdateCols))
}

// agentNodes are the nodes the HivePaaS agent has a task running on.
func (s *service) agentNodes(ctx context.Context) (map[string]bool, error) {
	agent, err := s.dockerManager.ServiceGetByName(ctx, base.HivepaasAgentServiceName, false)
	if err != nil {
		if errors.Is(err, hperrors.ErrNotFound) {
			return map[string]bool{}, nil
		}
		return nil, hperrors.Wrap(err)
	}
	tasks, err := s.dockerManager.TaskList(ctx, func(opts *client.TaskListOptions) {
		docker.FilterAdd(&opts.Filters, "service", agent.ID)
		docker.FilterAdd(&opts.Filters, "desired-state", string(swarm.TaskStateRunning))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	out := map[string]bool{}
	for i := range tasks.Items {
		if tasks.Items[i].Status.State == swarm.TaskStateRunning {
			out[tasks.Items[i].NodeID] = true
		}
	}
	return out, nil
}

func outputOrNil(out *entity.OBISyncOutput) *entity.OBISyncOutput {
	if out.Problem == "" && out.Unknown == "" && len(out.Nodes) == 0 {
		return nil
	}
	return out
}
