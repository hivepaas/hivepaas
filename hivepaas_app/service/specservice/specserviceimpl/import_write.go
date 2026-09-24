package specserviceimpl

import (
	"context"
	"maps"
	"slices"
	"strings"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/appprovisionservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/projectservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// writer writes a plan: every selected node that creates or updates something,
// in one call to PersistProjectData on the caller's transaction.
type writer struct {
	p          *planner
	operatorID string
	now        time.Time
	data       projectservice.PersistingProjectData

	// ids is the id each setting written gets - its own when it exists, a new
	// one otherwise - by its path in the bundle. It is complete before the first
	// setting is built, so settings of one import can name each other.
	ids map[string]string
	// rows are the settings of each scope written, as this installation has them
	// - or, for a project being created, the defaults it is created with.
	rows map[string]*scopeRows
	// projectIDs are the projects created, by their key.
	projectIDs map[string]string
	// written are the settings written, for what their types do after commit.
	written []*entity.Setting
	// appIDs is what each app node is here - the id chosen for an app created,
	// the matched one otherwise - by its path, and by the id it had in the bundle.
	appIDs         map[string]string
	appIDsByBundle map[string]string
	// appSettings are the settings of each app created, which are provisioned
	// with it rather than persisted with the scopes.
	appSettings map[string][]*entity.Setting
	// provisioned is what provisioning created, for Cleanup.
	provisioned *appprovisionservice.ProvisionAppsResp
	// tasks and deployments are what phase 1 queued, for phase 3.
	tasks       []*entity.Task
	deployments []*specservice.ImportDeployment
}

func newWriter(p *planner, operatorID string) *writer {
	return &writer{
		p: p, operatorID: operatorID, now: timeutil.NowUTC(),
		ids: map[string]string{}, rows: map[string]*scopeRows{},
		projectIDs: map[string]string{},
		appIDs:     map[string]string{}, appIDsByBundle: map[string]string{},
		appSettings: map[string][]*entity.Setting{},
	}
}

func (w *writer) write(ctx context.Context) error {
	if err := w.writeRecords(ctx); err != nil {
		return err
	}
	nodes := w.writtenNodes()
	w.chooseAppIDs(nodes)
	for _, node := range nodes {
		if err := w.chooseIDs(ctx, node); err != nil {
			return err
		}
	}
	for _, node := range nodes {
		if err := w.writeSettings(ctx, node); err != nil {
			return err
		}
	}
	if err := w.p.s.projectService.PersistProjectData(ctx, w.p.db, &w.data); err != nil {
		return hperrors.Wrap(err)
	}
	if err := w.writeUpdatedApps(ctx); err != nil {
		return err
	}
	if err := w.provisionApps(ctx); err != nil {
		return err
	}
	w.setOutcomes()
	return nil
}

// writeRecords writes the projects and envs themselves: a project created with
// the defaults the dashboard gives one, an existing one renamed or given its
// owner, an env's row.
func (w *writer) writeRecords(ctx context.Context) error {
	for _, node := range w.p.nodes {
		if !node.Selected || node.Action == specmodel.ActionSkip {
			continue
		}
		var err error
		switch node.Kind {
		case specmodel.NodeKindProject:
			err = w.writeProject(ctx, node)
		case specmodel.NodeKindEnv:
			err = w.writeEnv(ctx, node)
		case specmodel.NodeKindGlobal, specmodel.NodeKindSettings, specmodel.NodeKindApp:
		}
		if err != nil {
			return err
		}
	}
	return nil
}

func (w *writer) writeProject(ctx context.Context, node *specmodel.PlanNode) error {
	doc := w.p.bundle.Projects[node.Key]
	switch node.Action {
	case specmodel.ActionCreate:
		project := &entity.Project{
			ID:      gofn.Must(ulid.NewStringULID()),
			Key:     node.Key,
			Name:    doc.Name,
			Note:    doc.Note,
			Status:  base.ProjectStatusActive,
			OwnerID: gofn.Coalesce(w.p.owners[node.Path], w.operatorID),
		}
		w.projectIDs[node.Key] = project.ID
		err := w.p.s.projectService.PrepareNewProject(ctx,
			&projectservice.NewProjectReq{Project: project, TimeNow: w.now}, &w.data)
		return hperrors.Wrap(err)
	case specmodel.ActionUpdate:
		project, err := w.p.s.projectRepo.GetByID(ctx, w.p.db, node.TargetID)
		if err != nil {
			return hperrors.Wrap(err)
		}
		if slices.Contains(node.Changes, "name") {
			project.Name = doc.Name
		}
		if slices.Contains(node.Changes, "note") {
			project.Note = doc.Note
		}
		if slices.Contains(node.Changes, ownerField) {
			project.OwnerID = w.p.owners[node.Path]
		}
		project.UpdateVer++
		project.UpdatedAt = w.now
		w.data.UpsertingProjects = append(w.data.UpsertingProjects, project)
	case specmodel.ActionUnchanged, specmodel.ActionKeep, specmodel.ActionSkip:
	}
	return nil
}

func (w *writer) writeEnv(ctx context.Context, node *specmodel.PlanNode) error {
	place := w.p.envPlace(node.Path)
	doc := w.p.bundle.Envs[place.project][place.env]
	projectID := w.projectID(place.project)
	switch node.Action {
	case specmodel.ActionCreate:
		env := projectservice.NewProjectEnv(&entity.Project{ID: projectID},
			gofn.Coalesce(doc.Name, place.env), doc.Color, doc.Index, w.now)
		// The bundle's key is the env's key: a name edited in the bundle does not
		// make another env.
		env.Key, env.ID = place.env, projecthelper.CalcProjectEnvID(projectID, place.env)
		w.data.UpsertingProjectEnvs = append(w.data.UpsertingProjectEnvs, env)
	case specmodel.ActionUpdate:
		env, err := w.p.s.projectEnvRepo.GetByKey(ctx, w.p.db, projectID, place.env)
		if err != nil {
			return hperrors.Wrap(err)
		}
		env.Name = gofn.Coalesce(doc.Name, env.Name)
		env.Color, env.Index = doc.Color, doc.Index
		env.UpdateVer++
		env.UpdatedAt = w.now
		w.data.UpsertingProjectEnvs = append(w.data.UpsertingProjectEnvs, env)
	case specmodel.ActionUnchanged, specmodel.ActionKeep, specmodel.ActionSkip:
	}
	return nil
}

// projectID is what a project of the bundle is here: the one created for it, or
// the one it matched.
func (w *writer) projectID(key string) string {
	if id := w.projectIDs[key]; id != "" {
		return id
	}
	if node := w.p.byPath[projectsSegment+"/"+key]; node != nil && node.TargetID != "" {
		return node.TargetID
	}
	// An env route plans no project: its scope names it.
	for path, scope := range w.p.lookupScope {
		if strings.HasPrefix(path, projectsSegment+"/"+key+"/") && scope.ProjectID != "" {
			return scope.ProjectID
		}
	}
	return ""
}

// envPlace reads the project and env keys of an env node's path.
func (p *planner) envPlace(path string) appPlace {
	segments := strings.Split(path, "/")
	if len(segments) < 4 { //nolint:mnd // projects/<p>/envs/<e>
		return appPlace{}
	}
	return appPlace{project: segments[1], env: segments[3]}
}

// writtenNodes are the nodes whose settings the writer writes: the global one,
// a project's, an env's, an app's.
func (w *writer) writtenNodes() []*specmodel.PlanNode {
	var out []*specmodel.PlanNode
	for _, node := range w.p.nodes {
		if node.Selected && writes(node) && node.Kind != specmodel.NodeKindProject &&
			node.Kind != specmodel.NodeKindEnv {
			out = append(out, node)
		}
	}
	return out
}

// writtenNames are the settings a node writes: everything it holds when it is
// created - but a type import skips - and what changed otherwise.
func (w *writer) writtenNames(node *specmodel.PlanNode) []string {
	if place, isApp := w.p.apps[node.Path]; isApp {
		return appWrittenNames(node, place.doc)
	}
	if node.Action != specmodel.ActionCreate {
		return node.Changes
	}
	settings := w.p.settingsOf[node.Path]
	var names []string
	for _, name := range settingNames(settings) {
		block, key, _ := strings.Cut(name, "/")
		if _, typ, found := settingBody(settings, block, key); found && importPolicyFor(typ).skip == "" {
			names = append(names, name)
		}
	}
	return names
}

// settingPath is where a setting of a node sits in the bundle, the path a
// reference names it by.
func settingPath(node *specmodel.PlanNode, name string) string {
	return strings.TrimSuffix(node.Path, "/settings") + "/" + name
}

// scopeOf is the scope a settings node writes into, and the id of its object.
func (w *writer) scopeOf(node *specmodel.PlanNode) (base.ObjectScopeType, string) {
	switch node.Kind { //nolint:exhaustive // projects and envs hold no settings of their own
	case specmodel.NodeKindGlobal:
		return base.ObjectScopeGlobal, ""
	case specmodel.NodeKindApp:
		return base.ObjectScopeApp, w.appIDs[node.Path]
	}
	if strings.Contains(node.Path, "/envs/") {
		place := w.p.envPlace(strings.TrimSuffix(node.Path, "/settings"))
		return base.ObjectScopeProjectEnv, projecthelper.CalcProjectEnvID(w.projectID(place.project), place.env)
	}
	key := strings.TrimSuffix(strings.TrimPrefix(node.Path, projectsSegment+"/"), "/settings")
	return base.ObjectScopeProject, w.projectID(key)
}

// scopeRows are a scope's settings, found the way export keys them.
type scopeRows struct {
	byID   map[string]*entity.Setting
	byKey  map[string]*entity.Setting
	byType map[base.SettingType]*entity.Setting
}

func newScopeRows(settings []*entity.Setting) *scopeRows {
	rows := &scopeRows{
		byID: map[string]*entity.Setting{}, byKey: map[string]*entity.Setting{},
		byType: map[base.SettingType]*entity.Setting{},
	}
	keys := specmodel.DeriveSettingKeys(settings)
	for _, setting := range settings {
		rows.byID[setting.ID] = setting
		rows.byKey[string(setting.Type)+"/"+keys[setting.ID]] = setting
		if specmodel.IsSingletonType(setting.Type) {
			rows.byType[setting.Type] = setting
		}
	}
	return rows
}

// match is this installation's setting a bundle entry stands for: a singleton
// by its type; a collection entry by the id it was exported under, then by key.
func (r *scopeRows) match(typ base.SettingType, key string, body any) *entity.Setting {
	if specmodel.IsSingletonType(typ) {
		return r.byType[typ]
	}
	if row := r.byID[entryID(body)]; row != nil && row.Type == typ {
		return row
	}
	return r.byKey[string(typ)+"/"+key]
}

// chooseIDs gives every setting a node writes its id, before anything is built.
func (w *writer) chooseIDs(ctx context.Context, node *specmodel.PlanNode) error {
	scope, objectID := w.scopeOf(node)
	var owned []*entity.Setting
	switch {
	case scope == base.ObjectScopeProject && w.projectIDs[node.Key] != "":
		// A project being created has the defaults it is created with.
		for _, setting := range w.data.UpsertingSettings {
			if setting.ObjectID == objectID {
				owned = append(owned, setting)
			}
		}
	case w.scopeCreated(node):
		// A scope being created has nothing yet.
	default:
		scopes := []base.ObjectScopeType{scope}
		if scope == base.ObjectScopeGlobal {
			scopes = append(scopes, base.ObjectScopeHivepaas)
		}
		var err error
		if owned, err = w.p.s.loadOwned(ctx, w.p.db, scopes, objectID); err != nil {
			return hperrors.Wrap(err)
		}
	}
	rows := newScopeRows(owned)
	w.rows[node.Path] = rows

	for _, name := range w.writtenNames(node) {
		body, typ, key, found := w.settingOf(node, name)
		if !found {
			continue
		}
		id := gofn.Must(ulid.NewStringULID())
		if row := rows.match(typ, key, body); row != nil {
			id = row.ID
		}
		w.ids[settingPath(node, name)] = id
	}
	return nil
}

// scopeCreated reports whether the scope a settings node belongs to is being
// created, and so has no settings here yet.
func (w *writer) scopeCreated(node *specmodel.PlanNode) bool {
	owner := w.p.byPath[strings.TrimSuffix(node.Path, "/settings")]
	return owner != nil && owner.Action == specmodel.ActionCreate
}

// writeSettings builds each setting a node writes and adds it to what is
// persisted: over the default it replaces, for a project being created.
func (w *writer) writeSettings(ctx context.Context, node *specmodel.PlanNode) error {
	scope, objectID := w.scopeOf(node)
	for _, name := range w.writtenNames(node) {
		body, typ, key, found := w.settingOf(node, name)
		if !found {
			continue
		}
		setting, err := w.buildSetting(ctx, node, name, typ, key, body)
		if err != nil {
			return err
		}
		setting.Scope, setting.ObjectID = scope, objectID
		if row := w.rows[node.Path].match(typ, key, body); row != nil {
			setting.Scope, setting.ObjectID = row.Scope, row.ObjectID
			setting.CreatedAt, setting.UpdateVer = row.CreatedAt, row.UpdateVer+1
		}
		if node.Kind == specmodel.NodeKindApp && node.Action == specmodel.ActionCreate {
			// Provisioned with the app, which does not exist yet.
			w.written = append(w.written, setting)
			w.appSettings[node.Path] = append(w.appSettings[node.Path], setting)
			continue
		}
		w.persist(setting)
	}
	return nil
}

// persist adds a setting to what is written, in place of one of the same id
// already there - the default of a project being created.
func (w *writer) persist(setting *entity.Setting) {
	w.written = append(w.written, setting)
	for i, existing := range w.data.UpsertingSettings {
		if existing.ID == setting.ID {
			w.data.UpsertingSettings[i] = setting
			return
		}
	}
	w.data.UpsertingSettings = append(w.data.UpsertingSettings, setting)
}

// buildSetting builds one setting of a scope from its bundle body: references
// rewritten to this installation's ids, the fixes validate promised applied,
// and - from a bundle without secrets - the secrets of the setting it replaces
// kept.
func (w *writer) buildSetting(
	ctx context.Context, node *specmodel.PlanNode, name string, typ base.SettingType, key string, body any,
) (*entity.Setting, error) {
	var externals []*specmodel.ExternalRef
	setting, data, err := decodeImportedSetting(specmodel.Block(name), typ, key,
		withExternalPlaceholders(body, &externals))
	if err != nil {
		return nil, err
	}
	mapping, err := w.refMapping(ctx, node, data, externals)
	if err != nil {
		return nil, err
	}
	if err = entity.RemapRefs(data, mapping); err != nil {
		return nil, hperrors.Wrap(err)
	}

	row := w.rows[node.Path].match(typ, key, body)
	if err = w.keepAndGenerate(node, data, row); err != nil {
		return nil, err
	}
	holder := w.holderOf(node, name)
	for _, issue := range node.Issues {
		switch {
		case issue.Code == specmodel.CodeDomainInUse:
			if routing, ok := data.(*entity.AppRoutingSettings); ok {
				dropDomain(routing, issue.Detail["domain"])
			}
		case issue.Detail[refInSetting] != holder:
		case issue.Code == specmodel.CodeNodeNotFound:
			if volume, ok := data.(*entity.ClusterVolume); ok {
				volume.NodeID = ""
			}
		case issue.Code == specmodel.CodeRefNotSelected, issue.Code == specmodel.CodeRefNotFound,
			issue.Code == specmodel.CodeSecretOmitted:
			setting.Status = base.SettingStatusPending
		}
	}

	setting.ID = w.ids[settingPath(node, name)]
	setting.CreatedAt, setting.UpdatedAt, setting.UpdateVer = w.now, w.now, 1
	if err = setting.SetData(data); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return setting, nil
}

// refMapping maps each reference a setting makes to the id it names here. What
// resolves to nothing - what validate cleared - maps to nothing.
func (w *writer) refMapping(
	ctx context.Context, node *specmodel.PlanNode, data entity.SettingData, externals []*specmodel.ExternalRef,
) (map[string]string, error) {
	refs := data.GetRefObjectIDs()
	mapping := map[string]string{}
	if refs == nil {
		return mapping, nil
	}
	scope := w.p.lookupScope[node.Path]
	if scope == nil {
		scope = entity.NewObjectScopeGlobal()
	}
	for _, ref := range refs.RefAppIDs {
		id, err := w.appRefID(ctx, ref)
		if err != nil {
			return nil, err
		}
		if ref != "" {
			mapping[ref] = id
		}
	}
	for _, ref := range refs.RefSettingIDs {
		if ref == "" {
			continue
		}
		if index, ok := placeholderIndex(ref, len(externals)); ok {
			found, err := w.p.s.findRef(ctx, w.p.db, scope, externals[index])
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			mapping[ref] = ""
			if found != nil {
				mapping[ref] = found.ID
			}
			continue
		}
		if t, isPath := parseRefPath(ref); isPath {
			id, err := w.pathID(ctx, scope, ref, t)
			if err != nil {
				return nil, err
			}
			mapping[ref] = id
			continue
		}
		existing, err := w.p.s.loadByIDs(ctx, w.p.db, []string{ref})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if len(existing) == 0 {
			mapping[ref] = ""
		}
	}
	return mapping, nil
}

// pathID is the id a path of the bundle names here: the setting this import
// writes there, the one the target has there, or the one the target has under
// its name.
func (w *writer) pathID(ctx context.Context, scope *entity.ObjectScope, path string, t refTarget) (string, error) {
	if id := w.ids[path]; id != "" {
		return id, nil
	}
	body, _, inFull := settingBody(settingsAt(w.p.full, t), t.block, t.key)
	if t.isCollectionSetting {
		entries, _ := settingsAt(w.p.current, t)[t.block].(map[string]any)
		if current, found := currentEntry(entries, t.key, body); found {
			return entryID(current), nil
		}
	}
	if !inFull {
		return "", nil
	}
	found, err := w.p.s.findRef(ctx, w.p.db, scope, externalOf(t, body))
	if err != nil || found == nil {
		return "", hperrors.Wrap(err)
	}
	return found.ID, nil
}

// setOutcomes says what phase 1 did with each selected node. Phase 2 can still
// fail an app it updates.
func (w *writer) setOutcomes() {
	for _, node := range w.p.nodes {
		if !node.Selected {
			continue
		}
		switch {
		case node.Action == specmodel.ActionSkip:
			node.Outcome = specmodel.OutcomeSkipped
		case writes(node):
			node.Outcome = specmodel.OutcomeApplied
		default:
			node.Outcome = specmodel.OutcomeUnchanged
		}
	}
}

// afterCommit is phase 2: what the settings written take beyond their rows.
func (w *writer) afterCommit(ctx context.Context, db database.IDB) error {
	w.applyToServices(ctx, db)
	return w.writeFiles()
}

// writeFiles does what the settings written take beyond their rows, by type.
func (w *writer) writeFiles() error {
	byType := map[base.SettingType][]*entity.Setting{}
	for _, setting := range w.written {
		byType[setting.Type] = append(byType[setting.Type], setting)
	}
	for _, typ := range slices.Sorted(maps.Keys(byType)) {
		if policy := importPolicyFor(typ); policy.afterCommit != nil {
			if err := policy.afterCommit(w.p.s, byType[typ]); err != nil {
				return err
			}
		}
	}
	return nil
}
