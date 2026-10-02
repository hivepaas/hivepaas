package internal

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/moby/moby/client"
	"go.uber.org/fx"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/resourcesampler"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/safego"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appservice"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	// resourceSampleInterval is how often an app container's usage is written.
	resourceSampleInterval = 15 * time.Second
	// loggingCheckInterval is how often the agent reads whether logs are stored:
	// the rows are written only then, as nothing else would read them.
	loggingCheckInterval = time.Minute
)

// hostPrefix is where the agent sees the node's filesystem: the stack mounts /
// at /host; an agent run on the node itself reads the node's own.
func hostPrefix() string {
	if _, err := os.Stat("/host/sys/fs/cgroup"); err == nil {
		return "/host"
	}
	return ""
}

// ResourceSamplerOnAgent writes, every resourceSampleInterval, a row of each
// app container's usage on the node to the agent's stdout, while stored logs
// are on. See resourcesampler.
func ResourceSamplerOnAgent(
	lc fx.Lifecycle,
	cfg *config.Config,
	db *database.DB,
	dockerManager docker.Manager,
	settingRepo repository.SettingRepo,
	logger logging.Logger,
) {
	if cfg.RunMode != config.RunModeAgent {
		return
	}
	prefix := hostPrefix()
	sampler := &resourcesampler.Sampler{
		CgroupRoot: prefix + "/sys/fs/cgroup",
		ProcRoot:   prefix + "/proc",
		Out:        os.Stdout,
		Now:        time.Now,
	}
	lister := &appContainers{docker: dockerManager, cache: map[string]resourcesampler.Container{}}
	ctx, cancel := context.WithCancel(context.Background())

	lc.Append(fx.Hook{
		OnStart: func(context.Context) error {
			safego.Go("resource sampler", func() {
				runResourceSampler(ctx, db, settingRepo, sampler, lister, logger)
			})
			return nil
		},
		OnStop: func(context.Context) error {
			cancel()
			return nil
		},
	})
}

func runResourceSampler(
	ctx context.Context,
	db *database.DB,
	settingRepo repository.SettingRepo,
	sampler *resourcesampler.Sampler,
	lister *appContainers,
	logger logging.Logger,
) {
	ticker := time.NewTicker(resourceSampleInterval)
	defer ticker.Stop()
	var on bool
	var checkedAt time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
		if time.Since(checkedAt) >= loggingCheckInterval {
			on, checkedAt = logsStored(ctx, db, settingRepo), time.Now()
		}
		if !on {
			sampler.Reset()
			continue
		}
		containers, err := lister.list(ctx)
		if err != nil {
			logger.Warnf("resource sampler: failed to list the app containers: %v", err)
			continue
		}
		sampler.Tick(ctx, containers)
	}
}

// logsStored is whether the node's log collector stores the agent's lines.
func logsStored(ctx context.Context, db *database.DB, settingRepo repository.SettingRepo) bool {
	setting, err := settingRepo.GetSingle(ctx, db, entity.NewObjectScopeGlobal(), base.SettingTypeLogging, true)
	if err != nil {
		return false
	}
	cfg, err := setting.AsLoggingSettings()
	return err == nil && cfg.Enabled && (cfg.Sources.Apps || cfg.Sources.HivePaaS)
}

// appContainers lists the node's running app containers, by the label every
// app container carries; a container's pid and cgroup parent are kept, as
// they do not change while it runs.
type appContainers struct {
	docker docker.Manager
	mu     sync.Mutex
	cache  map[string]resourcesampler.Container
}

func (l *appContainers) list(ctx context.Context) ([]resourcesampler.Container, error) {
	resp, err := l.docker.ContainerList(ctx, func(opts *client.ContainerListOptions) {
		docker.FilterAdd(&opts.Filters, "label", appservice.LabelLogAppID)
		docker.FilterAdd(&opts.Filters, "status", "running")
	})
	if err != nil {
		return nil, err //nolint:wrapcheck // logged by the caller
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]resourcesampler.Container, 0, len(resp.Items))
	seen := make(map[string]bool, len(resp.Items))
	for _, item := range resp.Items {
		seen[item.ID] = true
		c, ok := l.cache[item.ID]
		if !ok {
			inspect, err := l.docker.ContainerInspect(ctx, item.ID)
			if err != nil || inspect.Container.State == nil {
				continue
			}
			c = resourcesampler.Container{ID: item.ID, AppID: item.Labels[appservice.LabelLogAppID],
				Pid: inspect.Container.State.Pid}
			if inspect.Container.HostConfig != nil {
				c.CgroupParent = inspect.Container.HostConfig.CgroupParent
			}
			l.cache[item.ID] = c
		}
		out = append(out, c)
	}
	for id := range l.cache {
		if !seen[id] {
			delete(l.cache, id)
		}
	}
	return out, nil
}
