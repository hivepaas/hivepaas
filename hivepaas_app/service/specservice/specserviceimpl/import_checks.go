package specserviceimpl

import (
	"context"
	"errors"
	"maps"
	"slices"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/clusterservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/volumeservice"
)

// writingApps are the selected app nodes import writes, in plan order.
func (p *planner) writingApps() []*specmodel.PlanNode {
	var out []*specmodel.PlanNode
	for _, node := range p.nodes {
		if node.Kind == specmodel.NodeKindApp && node.Selected && writes(node) {
			out = append(out, node)
		}
	}
	return out
}

// currentApp is the target's own export of the app a node matched.
func (p *planner) currentApp(node *specmodel.PlanNode) *specmodel.AppDoc {
	place := p.apps[node.Path]
	env := p.current.Envs[place.project][place.env]
	if env == nil {
		return nil
	}
	return env.Apps[node.Key]
}

// checkPermissions skips the apps whose import would grant what the operator
// may not, by the gates template creation applies to the same grants.
func (p *planner) checkPermissions(ctx context.Context) error {
	for _, node := range p.writingApps() {
		if err := p.checkCapabilities(ctx, node); err != nil {
			return err
		}
		if node.Action == specmodel.ActionSkip {
			continue
		}
		if err := p.checkSharedMounts(ctx, node); err != nil {
			return err
		}
	}
	return nil
}

// checkCapabilities skips an app created with capabilities, or updated so that
// what it grants changes, when the operator may not grant them.
func (p *planner) checkCapabilities(ctx context.Context, node *specmodel.PlanNode) error {
	doc := p.apps[node.Path].doc
	granted := specmodel.GrantedCapabilities(doc)
	if len(granted) == 0 || !writesBlock(node, "deployment.resources") {
		return nil
	}
	if node.Action == specmodel.ActionUpdate &&
		sameYAML(doc.Deployment.Resources.Capabilities, currentCapabilities(p.currentApp(node))) {
		return nil
	}
	if p.capabilitiesAllowed == nil {
		allowed := true
		if p.req.MayGrantCapabilities != nil {
			var err error
			if allowed, err = p.req.MayGrantCapabilities(ctx); err != nil {
				return hperrors.Wrap(err)
			}
		}
		p.capabilitiesAllowed = &allowed
	}
	if *p.capabilitiesAllowed {
		return nil
	}
	p.skipNode(node, specmodel.Issue{
		Severity: specmodel.SeveritySkipped, Code: specmodel.CodeCapabilityNotPermitted, Path: node.Path,
		Detail: map[string]any{"capabilities": granted},
		Action: "not imported: granting this needs Write permission on the Cluster module",
	})
	return nil
}

func currentCapabilities(doc *specmodel.AppDoc) *specmodel.Capabilities {
	if doc == nil || doc.Deployment == nil || doc.Deployment.Resources == nil {
		return nil
	}
	return doc.Deployment.Resources.Capabilities
}

// checkSharedMounts skips an app whose storage is written with a mount into
// the directory of an app the import does not write, when the operator may not
// write to that app. An app the import writes is exempt, as an app of the same
// template request is: the operator is writing both.
func (p *planner) checkSharedMounts(ctx context.Context, node *specmodel.PlanNode) error {
	doc := p.apps[node.Path].doc
	if doc.Deployment == nil || doc.Deployment.Storage == nil || !writesBlock(node, "deployment.storage") {
		return nil
	}
	scope := p.lookupScope[node.Path]
	if scope == nil || scope.ScopeType != base.ObjectScopeProjectEnv || p.req.MayWriteApp == nil {
		// An env being created has no app of its own yet to reach into.
		return nil
	}
	envPath := strings.TrimSuffix(node.Path, "/apps/"+node.Key)
	mounts := doc.Deployment.Storage.Mounts
	for _, target := range slices.Sorted(maps.Keys(mounts)) {
		source := mounts[target].SourceApp
		if source == nil || source.App == "" || source.App == node.Key {
			continue
		}
		if other := p.byPath[envPath+"/apps/"+source.App]; other != nil && other.Selected && writes(other) {
			continue
		}
		app, err := p.envApp(ctx, scope, source.App)
		if err != nil {
			return err
		}
		if app == nil {
			// Nothing to reach into: REF_NOT_FOUND already says so.
			continue
		}
		allowed, err := p.req.MayWriteApp(ctx, app)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if !allowed {
			p.skipNode(node, specmodel.Issue{
				Severity: specmodel.SeveritySkipped, Code: specmodel.CodeSharedMountNotPermitted, Path: node.Path,
				Detail: map[string]any{refInMount: target, "app": source.App},
				Action: "not imported: reaching this app's storage needs Write permission on it",
			})
			return nil
		}
	}
	return nil
}

// envApp is the app an env has on this installation under a key, or nil.
func (p *planner) envApp(ctx context.Context, scope *entity.ObjectScope, key string) (*entity.App, error) {
	apps, _, err := p.s.appRepo.List(ctx, p.db, scope.ProjectID, nil,
		bunex.SelectWhere("app.project_env_id = ?", scope.ProjectEnvID),
		bunex.SelectWhere("app.key = ?", key),
	)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	for _, app := range apps {
		if app.ProjectEnvID == scope.ProjectEnvID && app.Key == key {
			return app, nil
		}
	}
	return nil, nil
}

// checkAvailability reports what this installation cannot give an imported
// object as it stands: every finding, not the first.
func (p *planner) checkAvailability(ctx context.Context) error {
	apps := p.writingApps()
	if err := p.checkDomains(ctx, apps); err != nil {
		return err
	}
	if err := p.checkPorts(ctx, apps); err != nil {
		return err
	}
	if err := p.checkNodes(ctx); err != nil {
		return err
	}
	return p.checkStorage(ctx, apps)
}

// checkDomains reports a domain an imported app answers at that another app
// holds. The apps whose routing the import rewrites hold nothing it is checked
// against - what they hold is what the bundle says - and two apps of the import
// asking for one domain are a finding on the second.
func (p *planner) checkDomains(ctx context.Context, apps []*specmodel.PlanNode) error {
	var rewritten []string
	for _, node := range apps {
		if target := p.targetApps[node.Path]; target != nil && writesBlock(node, "settings.routing") {
			rewritten = append(rewritten, target.ID)
		}
	}
	claimedBy := map[string]string{}
	for _, node := range apps {
		if node.Action == specmodel.ActionSkip || !writesBlock(node, "settings.routing") {
			continue
		}
		domains, err := specmodel.ActiveDomains(p.apps[node.Path].doc)
		if err != nil {
			return hperrors.Wrap(hperrors.ErrSpecBundleInvalid).WithCause(err).
				WithExtraDetail("%s: settings.routing does not read as routing", node.Path)
		}
		for _, domain := range domains {
			detail := map[string]any{"domain": domain}
			if by, claimed := claimedBy[domain]; claimed {
				detail["with"] = by
				p.fixable(node, specmodel.CodeDomainInUse, detail, "the domain is dropped")
				continue
			}
			claimedBy[domain] = node.Path
			err = p.s.domainService.VerifyDomainsAvailable(ctx, p.db, []string{domain}, rewritten)
			switch {
			case errors.Is(err, hperrors.ErrDomainInUse):
				p.fixable(node, specmodel.CodeDomainInUse, detail, "the domain is dropped")
			case err != nil:
				return hperrors.Wrap(err)
			}
		}
	}
	return nil
}

// checkPorts reports a port an imported app publishes that another service
// holds, the way checkDomains does for domains.
func (p *planner) checkPorts(ctx context.Context, apps []*specmodel.PlanNode) error {
	var rewritten []string
	for _, node := range apps {
		if target := p.targetApps[node.Path]; target != nil && target.ServiceID != "" &&
			writesBlock(node, "deployment.networks") {
			rewritten = append(rewritten, target.ServiceID)
		}
	}
	claimedBy := map[clusterservice.PortRef]string{}
	for _, node := range apps {
		if node.Action == specmodel.ActionSkip || !writesBlock(node, "deployment.networks") {
			continue
		}
		for _, port := range specmodel.PublishedPorts(p.apps[node.Path].doc) {
			ref := clusterservice.PortRef{Published: port.Published, Protocol: port.Protocol}
			if ref.Protocol == "" {
				ref.Protocol = network.TCP
			}
			detail := map[string]any{"port": describePortConfig(&port)}
			if by, claimed := claimedBy[ref]; claimed {
				detail["with"] = by
				p.fixable(node, specmodel.CodePortInUse, detail, "the port is not published")
				continue
			}
			claimedBy[ref] = node.Path
			err := p.s.clusterService.VerifyPortsAvailable(ctx, []clusterservice.PortRef{ref}, rewritten)
			switch {
			case errors.Is(err, hperrors.ErrPortInUse):
				p.fixable(node, specmodel.CodePortInUse, detail, "the port is not published")
			case err != nil:
				return hperrors.Wrap(err)
			}
		}
	}
	return nil
}

// checkNodes reports a volume written pinned to a node this installation does
// not have. A node removed here and one that never existed are the same finding.
func (p *planner) checkNodes(ctx context.Context) error {
	block := specmodel.CollectionBlockName(base.SettingTypeClusterVolume)
	for _, node := range p.nodes {
		settings, holds := p.settingsOf[node.Path]
		if !holds || node.Kind == specmodel.NodeKindApp || !node.Selected || !writes(node) {
			continue
		}
		volumes, _ := settings[block].(map[string]any)
		for _, key := range slices.Sorted(maps.Keys(volumes)) {
			name := block + "/" + key
			if !writesBlock(node, name) {
				continue
			}
			nodeID, err := volumeNodeID(key, volumes[key])
			if err != nil {
				return err
			}
			if nodeID == "" {
				continue
			}
			found, err := p.s.nodeExists(ctx, p.db, nodeID)
			if err != nil {
				return hperrors.Wrap(err)
			}
			if !found {
				p.fixable(node, specmodel.CodeNodeNotFound, map[string]any{refInSetting: name, "node": nodeID},
					"the volume is not pinned to a node")
			}
		}
	}
	return nil
}

func volumeNodeID(key string, body any) (string, error) {
	block := specmodel.Block("settings." + specmodel.CollectionBlockName(base.SettingTypeClusterVolume))
	_, data, err := decodeImportedSetting(block, base.SettingTypeClusterVolume, key, body)
	if errors.Is(err, hperrors.ErrDataVerNewerThanSystemVer) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	volume, _ := data.(*entity.ClusterVolume)
	if volume == nil {
		return "", nil
	}
	return volume.NodeID, nil
}

// checkStorage warns about an app being created on storage that already holds
// something, as the template preflight does: the app would start on another's
// leftovers - a database initialized with a password it no longer has. Only an
// env this installation has can have storage already, and only a mount of the
// app's own directory, in a volume the target has, is asked about.
func (p *planner) checkStorage(ctx context.Context, apps []*specmodel.PlanNode) error {
	for _, node := range apps {
		scope := p.lookupScope[node.Path]
		doc := p.apps[node.Path].doc
		if node.Action != specmodel.ActionCreate || scope == nil || scope.ScopeType != base.ObjectScopeProjectEnv ||
			doc.Deployment == nil || doc.Deployment.Storage == nil {
			continue
		}
		place := p.apps[node.Path]
		planned := &entity.App{
			Key:        node.Key,
			Project:    &entity.Project{ID: scope.ProjectID, Key: place.project},
			ProjectEnv: &entity.ProjectEnv{ID: scope.ProjectEnvID, Key: place.env},
		}
		req := &volumeservice.InspectAppStorageReq{Scope: scope}
		mounts := doc.Deployment.Storage.Mounts
		for _, target := range slices.Sorted(maps.Keys(mounts)) {
			m := mounts[target]
			if m.SourceApp != nil || (m.Type != mount.TypeVolume && m.Type != mount.TypeCluster) {
				continue
			}
			volumeID, err := p.targetVolumeID(ctx, node, m)
			if err != nil {
				return err
			}
			if volumeID == "" {
				continue
			}
			req.Queries = append(req.Queries, &volumeservice.AppStorageQuery{
				AppKey: node.Key, App: planned, VolumeID: volumeID, Subpath: mountSubpath(&m),
			})
		}
		if len(req.Queries) == 0 {
			continue
		}
		resp, err := p.s.volumeService.InspectAppStorage(ctx, p.db, req)
		if err != nil {
			return hperrors.Wrap(err)
		}
		for _, state := range resp.States {
			detail := map[string]any{"volume": state.VolumeName, "path": state.Path}
			switch {
			case !state.Checked:
				p.warning(node, specmodel.CodeStorageUnchecked, detail,
					"created without knowing whether its storage already holds data")
			case state.HasData():
				p.warning(node, specmodel.CodeStorageNotEmpty, detail, "created on the data already there")
			}
		}
	}
	return nil
}

func mountSubpath(m *specmodel.Mount) string {
	switch {
	case m.VolumeOptions != nil:
		return m.VolumeOptions.Subpath
	case m.ClusterOptions != nil:
		return m.ClusterOptions.Subpath
	}
	return ""
}

// targetVolumeID is the volume setting on this installation a managed mount
// names: the target's entry for the mount's path, or what its external
// reference finds. It is empty for a volume the target does not have yet.
func (p *planner) targetVolumeID(ctx context.Context, node *specmodel.PlanNode, m specmodel.Mount) (string, error) {
	if m.External != nil {
		found, err := p.s.findRef(ctx, p.db, p.lookupScope[node.Path], m.External)
		if err != nil || found == nil {
			return "", hperrors.Wrap(err)
		}
		return found.ID, nil
	}
	t, ok := parseRefPath(m.Source)
	if !ok || !t.isCollectionSetting {
		return "", nil
	}
	body, _, _ := settingBody(settingsAt(p.full, t), t.block, t.key)
	entries, _ := settingsAt(p.current, t)[t.block].(map[string]any)
	current, found := currentEntry(entries, t.key, body)
	if !found {
		return "", nil
	}
	return entryID(current), nil
}

func (p *planner) fixable(node *specmodel.PlanNode, code string, detail map[string]any, cleared string) {
	node.Issues = append(node.Issues, specmodel.Issue{
		Severity: specmodel.SeverityFixable, Code: code, Path: node.Path, Detail: detail, Action: cleared,
	})
}

func (p *planner) warning(node *specmodel.PlanNode, code string, detail map[string]any, result string) {
	node.Issues = append(node.Issues, specmodel.Issue{
		Severity: specmodel.SeverityWarning, Code: code, Path: node.Path, Detail: detail, Action: result,
	})
}

// nodeExistsInRepo is the production nodeFinder: cluster sync keeps a
// cluster-node setting for every node, with the Docker id as its ref id.
func (s *service) nodeExistsInRepo(ctx context.Context, db database.IDB, nodeID string) (bool, error) {
	nodes, _, err := s.settingRepo.List(ctx, db, nil, nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterNode),
		bunex.SelectWhere("setting.ref_id = ?", nodeID),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectLimit(1),
	)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return len(nodes) > 0, nil
}
