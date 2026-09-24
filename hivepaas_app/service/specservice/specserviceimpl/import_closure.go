package specserviceimpl

import (
	"context"
	"errors"
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// refWork is references still to resolve, and the node holding them.
type refWork struct {
	node *specmodel.PlanNode
	refs []bundleRef
}

// resolveRefs follows every reference of what import writes, and says what
// becomes of each. Selecting one app almost always selects too little - its
// certificate and its volume live at other scopes - so a setting the target
// lacks is pulled in with the app, and what that setting references is followed
// in turn, until nothing new is pulled. A reference nothing can satisfy is
// cleared, and the node says so.
func (p *planner) resolveRefs(ctx context.Context) error {
	var queue []refWork
	for _, node := range p.nodes {
		if !node.Selected || !writes(node) {
			continue
		}
		refs, err := p.nodeRefs(node)
		if err != nil {
			return err
		}
		queue = append(queue, refWork{node: node, refs: refs})
	}
	for len(queue) > 0 {
		work := queue[0]
		queue = queue[1:]
		for _, ref := range work.refs {
			more, err := p.resolveRef(ctx, work.node, ref)
			if err != nil {
				return err
			}
			if more != nil {
				queue = append(queue, *more)
			}
		}
	}
	p.narrowPulled()
	return nil
}

func writes(node *specmodel.PlanNode) bool {
	return node.Action == specmodel.ActionCreate || node.Action == specmodel.ActionUpdate
}

// writesBlock reports whether a node writes one of its blocks: all of them for
// an object being created, what changed of one being updated.
func writesBlock(node *specmodel.PlanNode, change string) bool {
	return node.Action == specmodel.ActionCreate ||
		node.Action == specmodel.ActionUpdate && slices.Contains(node.Changes, change)
}

// nodeRefs reads the references of what a node writes: everything a node being
// created holds, and what changed of one being updated.
func (p *planner) nodeRefs(node *specmodel.PlanNode) ([]bundleRef, error) {
	settings, holdsSettings := p.settingsOf[node.Path]
	if !holdsSettings {
		return nil, nil
	}
	prefix := ""
	var deployment *specmodel.Deployment
	if place, ok := p.apps[node.Path]; ok {
		prefix, deployment = "settings.", place.doc.Deployment
	}

	var refs []bundleRef
	if deployment != nil {
		if writesBlock(node, "deployment.storage") {
			refs = append(refs, mountRefs(deployment.Storage)...)
		}
		if deployment.Source != nil && writesBlock(node, "deployment.source") {
			sourceRefs, err := settingRefs(base.SettingTypeAppDeployment, node.Path, "deployment.source",
				deployment.Source)
			if err != nil {
				return nil, err
			}
			refs = append(refs, sourceRefs...)
		}
	}

	names := settingNames(settings)
	if node.Action != specmodel.ActionCreate {
		names = nil
		for _, change := range node.Changes {
			if name, ok := strings.CutPrefix(change, prefix); ok {
				names = append(names, name)
			}
		}
	}
	for _, name := range names {
		block, key, _ := strings.Cut(name, "/")
		body, typ, found := settingBody(settings, block, key)
		if !found {
			continue
		}
		more, err := settingRefs(typ, node.Path, prefix+name, body)
		if err != nil {
			return nil, err
		}
		refs = append(refs, more...)
	}
	return refs, nil
}

// settingNames names every setting of a settings map the way changes name them:
// a singleton by its block, a collection entry by its block and key.
func settingNames(settings map[string]any) []string {
	var names []string
	for _, block := range slices.Sorted(maps.Keys(settings)) {
		if _, ok := specmodel.SingletonTypeOf(block); ok {
			names = append(names, block)
			continue
		}
		entries, _ := settings[block].(map[string]any)
		for _, key := range slices.Sorted(maps.Keys(entries)) {
			names = append(names, block+"/"+key)
		}
	}
	return names
}

// settingBody finds one setting of a settings map, and its type.
func settingBody(settings map[string]any, block, key string) (any, base.SettingType, bool) {
	if typ, ok := specmodel.SingletonTypeOf(block); ok {
		body, found := settings[block]
		return body, typ, found && key == ""
	}
	typ, ok := specmodel.CollectionTypeOf(block)
	if !ok {
		return nil, "", false
	}
	entries, _ := settings[block].(map[string]any)
	body, found := entries[key]
	return body, typ, found
}

// settingsAt is the settings map a path's scope has in a bundle - the upload,
// or this installation's own export.
func settingsAt(b *specmodel.ImportBundle, t refTarget) map[string]any {
	switch {
	case t.project == "":
		if b.Global != nil {
			return b.Global.Settings
		}
	case t.env == "":
		if doc := b.Projects[t.project]; doc != nil {
			return doc.Settings
		}
	default:
		env := b.Envs[t.project][t.env]
		switch {
		case env == nil:
		case t.app == "":
			return env.Settings
		case env.Apps[t.app] != nil:
			return env.Apps[t.app].Settings
		}
	}
	return nil
}

func (p *planner) resolveRef(ctx context.Context, node *specmodel.PlanNode, ref bundleRef) (*refWork, error) {
	switch ref.kind {
	case refPath:
		return p.resolvePath(ctx, node, ref)
	case refExternal:
		return nil, p.resolveExternal(ctx, node, ref, ref.external, "")
	case refRawID:
		settings, err := p.s.loadByIDs(ctx, p.db, []string{ref.value})
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		if len(settings) == 0 {
			p.refIssue(node, ref, specmodel.CodeRefNotFound, "")
		}
	case refAppID:
		return nil, p.resolveAppID(ctx, node, ref)
	case refSourceApp:
		p.resolveSourceApp(node, ref)
	}
	return nil, nil
}

// resolvePath resolves a reference to a setting the bundle held: to the setting
// import writes, to the one the target has at the same place, to one closure
// pulls in, or - when it cannot be pulled in - to what the target has under
// its name.
func (p *planner) resolvePath(ctx context.Context, node *specmodel.PlanNode, ref bundleRef) (*refWork, error) {
	t, _ := parseRefPath(ref.value)
	holder := p.byPath[t.node]
	if holder != nil && holder.SelectedBy == selectedByUser && holder.Action != specmodel.ActionSkip {
		return nil, nil
	}
	body, _, inFull := settingBody(settingsAt(p.full, t), t.block, t.key)
	if p.targetHas(t, body) {
		return nil, nil
	}
	_, _, inRoute := settingBody(settingsAt(p.bundle, t), t.block, t.key)
	if inRoute && holder != nil && t.app == "" && holder.Action != specmodel.ActionSkip && !p.excluded(holder.Path) {
		return p.pull(holder, t, body)
	}
	if !inFull {
		p.refIssue(node, ref, specmodel.CodeRefNotFound, "")
		return nil, nil
	}
	return nil, p.resolveExternal(ctx, node, ref, externalOf(t, body), t.node)
}

// targetHas reports whether this installation has the setting a path names, at
// the same place: under the same key, or exported from the same setting.
func (p *planner) targetHas(t refTarget, body any) bool {
	current := settingsAt(p.current, t)
	if !t.isCollectionSetting {
		_, found := current[t.block]
		return found
	}
	entries, _ := current[t.block].(map[string]any)
	_, found := currentEntry(entries, t.key, body)
	return found
}

// excluded reports whether the operator deselected a node: an exclude pattern
// covers it. A node an include merely leaves out is not selected, and closure
// may still pull from it.
func (p *planner) excluded(path string) bool {
	return !specmodel.Selection{Exclude: p.req.Selection.Exclude}.Selects(path)
}

// pull brings one setting of a node the operator did not select into the
// import, and returns its own references to follow.
func (p *planner) pull(holder *specmodel.PlanNode, t refTarget, body any) (*refWork, error) {
	name := t.block
	if t.isCollectionSetting {
		name += "/" + t.key
	}
	if slices.Contains(p.pulled[holder.Path], name) {
		return nil, nil
	}
	p.pulled[holder.Path] = append(p.pulled[holder.Path], name)
	holder.Selected, holder.SelectedBy = true, selectedByDependency
	if holder.Action != specmodel.ActionCreate {
		holder.Action = specmodel.ActionUpdate
	}
	holder.Changes = slices.Sorted(slices.Values(p.pulled[holder.Path]))

	refs, err := settingRefs(t.settingType, holder.Path, name, body)
	if err != nil {
		return nil, err
	}
	return &refWork{node: holder, refs: refs}, nil
}

// narrowPulled leaves a node closure pulled from with the issues of what was
// pulled: the rest of it is not imported.
func (p *planner) narrowPulled() {
	for path, names := range p.pulled {
		node := p.byPath[path]
		kept := node.Issues[:0]
		for _, issue := range node.Issues {
			if setting, _ := issue.Detail[refInSetting].(string); slices.Contains(names, setting) {
				kept = append(kept, issue)
			}
		}
		node.Issues = kept
	}
}

// externalOf is the external reference a bundle's setting would have been
// written as, had the export not held it.
func externalOf(t refTarget, body any) *specmodel.ExternalRef {
	fields, _ := body.(map[string]any)
	meta, _ := fields[specmodel.SettingMetaKey].(map[string]any)
	name, _ := meta[settingMetaName].(string)
	kind, _ := meta["kind"].(string)
	return &specmodel.ExternalRef{
		Type: string(t.settingType), Name: gofn.Coalesce(name, t.key), Kind: kind, ID: entryID(body),
	}
}

// resolveExternal finds what an external reference names, as the node sees
// settings on this installation. availableIn is the node of the bundle that
// holds it, when one does.
func (p *planner) resolveExternal(
	ctx context.Context,
	node *specmodel.PlanNode,
	ref bundleRef,
	external *specmodel.ExternalRef,
	availableIn string,
) error {
	scope := p.lookupScope[node.Path]
	if scope == nil {
		scope = entity.NewObjectScopeGlobal()
	}
	found, err := p.s.findRef(ctx, p.db, scope, external)
	if err != nil {
		return hperrors.Wrap(err)
	}
	switch {
	case found != nil:
	case availableIn != "":
		p.refIssue(node, ref, specmodel.CodeRefNotSelected, availableIn)
	default:
		p.refIssue(node, ref, specmodel.CodeRefNotFound, "")
	}
	return nil
}

// resolveAppID resolves an app's id: an app the import writes, or one this
// installation has.
func (p *planner) resolveAppID(ctx context.Context, node *specmodel.PlanNode, ref bundleRef) error {
	if path := appPathByID(p.bundle, ref.value); path != "" {
		if app := p.byPath[path]; app != nil && app.Selected && app.Action != specmodel.ActionSkip {
			return nil
		}
	}
	_, err := p.s.appRepo.GetByID(ctx, p.db, "", ref.value)
	switch {
	case err == nil:
		return nil
	case !errors.Is(err, hperrors.ErrNotFound):
		return hperrors.Wrap(err)
	}
	if path := appPathByID(p.full, ref.value); path != "" {
		p.refIssue(node, ref, specmodel.CodeRefNotSelected, path)
		return nil
	}
	p.refIssue(node, ref, specmodel.CodeRefNotFound, "")
	return nil
}

func appPathByID(b *specmodel.ImportBundle, id string) string {
	for _, project := range slices.Sorted(maps.Keys(b.Envs)) {
		for _, env := range slices.Sorted(maps.Keys(b.Envs[project])) {
			for _, key := range slices.Sorted(maps.Keys(b.Envs[project][env].Apps)) {
				if b.Envs[project][env].Apps[key].ID == id {
					return projectsSegment + "/" + project + "/envs/" + env + "/apps/" + key
				}
			}
		}
	}
	return ""
}

// resolveSourceApp resolves the app whose directory a mount reaches: one of the
// same env the import writes, or one the env has here. Closure never pulls in an
// app: it follows references, not containment.
func (p *planner) resolveSourceApp(node *specmodel.PlanNode, ref bundleRef) {
	place, ok := p.apps[node.Path]
	if !ok {
		return
	}
	path := projectsSegment + "/" + place.project + "/envs/" + place.env + "/apps/" + ref.value
	if app := p.byPath[path]; app != nil && app.Selected && app.Action != specmodel.ActionSkip {
		return
	}
	if env := p.current.Envs[place.project][place.env]; env != nil && env.Apps[ref.value] != nil {
		return
	}
	if env := p.full.Envs[place.project][place.env]; env != nil && env.Apps[ref.value] != nil {
		p.refIssue(node, ref, specmodel.CodeRefNotSelected, path)
		return
	}
	p.refIssue(node, ref, specmodel.CodeRefNotFound, "")
}

// refIssue records a reference import clears. It names what held the reference
// and what it named, never a value.
func (p *planner) refIssue(node *specmodel.PlanNode, ref bundleRef, code, availableIn string) {
	detail := map[string]any{ref.in: ref.holder, "ref": ref.value}
	if ref.kind == refExternal && ref.external != nil {
		named := map[string]any{"type": ref.external.Type, "name": ref.external.Name}
		if ref.external.Kind != "" {
			named["kind"] = ref.external.Kind
		}
		detail["ref"] = named
	}
	cleared := "the reference is cleared"
	if ref.in == refInMount {
		cleared = "the mount is left out"
	}
	action := cleared + ": neither the bundle nor this installation has what it names"
	if code == specmodel.CodeRefNotSelected {
		action = cleared + ": what it names is in the bundle, and not imported"
	}
	issue := specmodel.Issue{
		Severity: specmodel.SeverityFixable, Code: code, Path: node.Path, Detail: detail,
		AvailableIn: availableIn, Action: action,
	}
	for _, have := range node.Issues {
		if reflect.DeepEqual(have, issue) {
			return
		}
	}
	node.Issues = append(node.Issues, issue)
}

// findRefInRepo is the production refFinder. Both lookups go through the
// repository's scope filters, so a reference finds only what the referencing
// scope could use: its own settings, what it inherits, what is shared with it.
func (s *service) findRefInRepo(
	ctx context.Context,
	db database.IDB,
	scope *entity.ObjectScope,
	ref *specmodel.ExternalRef,
) (*entity.Setting, error) {
	typ := base.SettingType(ref.Type)
	if ref.ID != "" {
		setting, err := s.settingRepo.GetByID(ctx, db, scope, typ, ref.ID, true)
		if err == nil {
			return setting, nil
		}
		if !errors.Is(err, hperrors.ErrNotFound) {
			return nil, hperrors.Wrap(err)
		}
	}
	var opts []bunex.SelectQueryOption
	if ref.Kind != "" {
		opts = append(opts, bunex.SelectWhere("setting.kind = ?", ref.Kind))
	}
	setting, err := s.settingRepo.GetByName(ctx, db, scope, typ, ref.Name, true, opts...)
	if errors.Is(err, hperrors.ErrNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return setting, nil
}
