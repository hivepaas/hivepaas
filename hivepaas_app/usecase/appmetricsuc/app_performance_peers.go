package appmetricsuc

import (
	"context"
	"net"
	"net/netip"
	"slices"
	"strings"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/obi"
	"github.com/hivepaas/hivepaas/services/docker"
)

var mysqlEngines = []string{"mysql", "mariadb"}

// dbEngines are the engines of the apps a database system OBI names may be,
// by that name: the protocol it speaks. A system not listed is its own
// engine's name.
var dbEngines = map[string][]string{
	"postgresql": {"postgres"},
	"mysql":      mysqlEngines,
	"mariadb":    mysqlEngines,
	"redis":      {"redis", "valkey", "dragonfly", "keydb"},
	"mongodb":    {"mongodb", "mongo"},
}

// peerResolver names the env's app behind what an app called, as it is
// today: by the name the app called it - its key, which is its alias in the
// env's network, or its service's name - or by a task's or service's address;
// a database by its engine and database.
type peerResolver struct {
	uc      *UC
	ctx     context.Context //nolint:containedctx // the request's, for the addresses read once if at all
	project *entity.Project
	apps    []*entity.App
	byName  map[string]*entity.App
	byIP    map[netip.Addr]*entity.App
	dbs     []*peerDatabase
}

// peerDatabase is an app of the env that is a database or a cache, by its
// kind.
type peerDatabase struct {
	app    *entity.App
	engine string
	dbName string
}

func (uc *UC) newPeerResolver(ctx context.Context, app *entity.App) (*peerResolver, error) {
	apps, _, err := uc.appRepo.List(ctx, uc.db, app.ProjectID, nil,
		bunex.SelectWhere("app.project_env_id = ?", app.ProjectEnvID),
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectRelation("Settings", bunex.SelectWhere("setting.type = ?", base.SettingTypeAppKind)),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	r := &peerResolver{uc: uc, ctx: ctx, project: app.Project, apps: apps, byName: map[string]*entity.App{}}
	for _, a := range apps {
		for _, name := range []string{a.Key, a.GlobalKey} {
			if name != "" {
				r.byName[strings.ToLower(name)] = a
			}
		}
		if a.GlobalKey != "" {
			r.byName["tasks."+strings.ToLower(a.GlobalKey)] = a
		}
		if setting := a.GetSettingByType(base.SettingTypeAppKind); setting != nil {
			kind, err := setting.AsAppKindSettings()
			if err == nil && kind.Engine != "" {
				db := &peerDatabase{app: a, engine: strings.ToLower(kind.Engine)}
				if kind.Database != nil {
					db.dbName = kind.Database.DbName
				}
				r.dbs = append(r.dbs, db)
			}
		}
	}
	return r, nil
}

// resolve is the app behind a peer, nil when none is known.
func (r *peerResolver) resolve(kind, peer string) *entity.App {
	if peer == "" {
		return nil
	}
	if kind == obi.KindDB {
		return r.database(peer)
	}
	host := strings.TrimSuffix(strings.ToLower(hostOf(peer)), ".")
	if app := r.byName[host]; app != nil {
		return app
	}
	if ip, err := netip.ParseAddr(host); err == nil {
		return r.addresses()[ip.Unmap()]
	}
	return nil
}

// hostOf is a peer's host, without its port.
func hostOf(peer string) string {
	if host, _, err := net.SplitHostPort(peer); err == nil {
		return host
	}
	return peer
}

// database is the only app of the env of a system's engines - and of its
// database, when the apps name theirs - as OBI names it: system/database.
func (r *peerResolver) database(peer string) *entity.App {
	system, namespace, _ := strings.Cut(peer, "/")
	engines, ok := dbEngines[system]
	if !ok {
		engines = []string{system}
	}
	var candidates, named []*peerDatabase
	for _, db := range r.dbs {
		if !slices.Contains(engines, db.engine) {
			continue
		}
		candidates = append(candidates, db)
		if db.dbName != "" && db.dbName == namespace {
			named = append(named, db)
		}
	}
	switch {
	case len(named) == 1:
		return named[0].app
	case len(candidates) == 1 && candidates[0].dbName == "":
		// A cache names no database: its engine is enough, when it is the only.
		return candidates[0].app
	}
	return nil
}

// addresses are the env's apps by their tasks' and services' addresses today,
// read once, when a peer is an address. Unread, every address stays a host.
func (r *peerResolver) addresses() map[netip.Addr]*entity.App {
	if r.byIP != nil {
		return r.byIP
	}
	r.byIP = map[netip.Addr]*entity.App{}
	byService := map[string]*entity.App{}
	for _, a := range r.apps {
		if a.ServiceID != "" {
			byService[a.ServiceID] = a
		}
	}
	if len(byService) == 0 || r.project == nil {
		return r.byIP
	}
	tasks, err := r.uc.dockerManager.TaskList(r.ctx, func(o *client.TaskListOptions) {
		for id := range byService {
			docker.FilterAdd(&o.Filters, "service", id)
		}
		docker.FilterAdd(&o.Filters, "desired-state", string(swarm.TaskStateRunning))
	})
	if err == nil {
		for i := range tasks.Items {
			task := &tasks.Items[i]
			for _, attachment := range task.NetworksAttachments {
				r.addAddresses(byService[task.ServiceID], attachment.Addresses...)
			}
		}
	}
	services, err := r.uc.dockerManager.ServiceListByStack(r.ctx, r.project.Key)
	if err == nil {
		for i := range services.Items {
			service := &services.Items[i]
			for _, vip := range service.Endpoint.VirtualIPs {
				r.addAddresses(byService[service.ID], vip.Addr)
			}
		}
	}
	return r.byIP
}

// addAddresses maps addresses, as Docker gives them - with their network's
// prefix - to an app.
func (r *peerResolver) addAddresses(app *entity.App, addresses ...netip.Prefix) {
	if app == nil {
		return
	}
	for _, address := range addresses {
		if address.IsValid() {
			r.byIP[address.Addr().Unmap()] = app
		}
	}
}
