package specserviceimpl

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"maps"
	"reflect"
	"slices"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// settingsNode plans a node that holds a scope's own settings. A scope being
// created creates its settings; otherwise the node changes when any setting the
// bundle carries differs from this installation's.
func (p *planner) settingsNode(
	node *specmodel.PlanNode,
	prefix string,
	bundle, current map[string]any,
	scopeExists bool,
) {
	changes := p.settingsChanges(node, prefix, bundle, current)
	if !scopeExists {
		node.Action = specmodel.ActionCreate
		return
	}
	setChanges(node, changes)
}

// settingsChanges names the settings of a block map that differ from this
// installation's, and records the issues a setting raises on its own: a block
// no type is known by, and data from a newer HivePaaS. Only what the bundle
// carries is compared: import never deletes a setting the bundle leaves out.
func (p *planner) settingsChanges(
	node *specmodel.PlanNode,
	prefix string,
	bundle, current map[string]any,
) []string {
	var changes []string
	for _, block := range slices.Sorted(maps.Keys(bundle)) {
		if typ, ok := specmodel.SingletonTypeOf(block); ok {
			if newerSetting(typ, bundle[block]) {
				node.Issues = append(node.Issues, versionNewer(node.Path, prefix+block))
			}
			currentBody, found := current[block]
			if !found {
				p.addMissing(node, prefix+block)
			}
			if !found || !sameBody(bundle[block], currentBody) {
				changes = append(changes, prefix+block)
			}
			continue
		}
		typ, ok := specmodel.CollectionTypeOf(block)
		if !ok {
			node.Issues = append(node.Issues, specmodel.Issue{
				Severity: specmodel.SeveritySkipped, Code: specmodel.CodeTypeNotImportable,
				Path: node.Path, Detail: map[string]any{"block": prefix + block},
				Action: "not imported: no setting type is known by this name",
			})
			continue
		}
		entries, _ := bundle[block].(map[string]any)
		currentEntries, _ := current[block].(map[string]any)
		for _, key := range slices.Sorted(maps.Keys(entries)) {
			if newerSetting(typ, entries[key]) {
				node.Issues = append(node.Issues, versionNewer(node.Path, prefix+block+"/"+key))
			}
			currentBody, found := currentEntry(currentEntries, key, entries[key])
			if !found {
				p.addMissing(node, prefix+block+"/"+key)
			}
			if !found || !sameBody(entries[key], currentBody) {
				changes = append(changes, prefix+block+"/"+key)
			}
		}
	}
	return changes
}

// currentEntry is the target's entry for a collection entry of the bundle: the
// one under the same key, or else the one exported from the same setting - a
// rename since the export, which §4 matches by id.
func currentEntry(entries map[string]any, key string, body any) (any, bool) {
	if current, found := entries[key]; found {
		return current, true
	}
	id := entryID(body)
	if id == "" {
		return nil, false
	}
	for _, current := range entries {
		if entryID(current) == id {
			return current, true
		}
	}
	return nil, false
}

func entryID(body any) string {
	fields, _ := body.(map[string]any)
	id, _ := fields[specmodel.CollectionEntryIDKey].(string)
	return id
}

// addMissing records a setting the target does not have at all, which keep
// still creates.
func (p *planner) addMissing(node *specmodel.PlanNode, setting string) {
	if p.missing == nil {
		p.missing = map[string][]string{}
	}
	p.missing[node.Path] = append(p.missing[node.Path], setting)
}

// newerSetting reports whether a setting's row names a version of its data this
// installation cannot read: Migrate is what would read it, and it says so.
func newerSetting(typ base.SettingType, body any) bool {
	fields, _ := body.(map[string]any)
	meta, _ := fields[specmodel.SettingMetaKey].(map[string]any)
	version, ok := meta["version"].(int)
	if !ok || version == 0 {
		return false
	}
	setting := &entity.Setting{Type: typ, Version: version, Data: "{}"}
	_, err := setting.Migrate()
	return errors.Is(err, hperrors.ErrDataVerNewerThanSystemVer)
}

func versionNewer(path, block string) specmodel.Issue {
	return specmodel.Issue{
		Severity: specmodel.SeverityBlocked, Code: specmodel.CodeSettingVersionNewer, Path: path,
		Detail: map[string]any{"setting": block},
		Action: "nothing can be imported: this setting was written by a newer HivePaaS",
	}
}

// sameBody compares two setting bodies as import would write them: the id a
// collection entry was exported under, and the id inside an external reference,
// name the object on the installation that wrote the bundle and are not
// configuration.
func sameBody(a, b any) bool {
	return reflect.DeepEqual(normalizeBody(a, true), normalizeBody(b, true))
}

func normalizeBody(node any, top bool) any {
	switch typed := node.(type) {
	case map[string]any:
		out := make(map[string]any, len(typed))
		for key, value := range typed {
			if top && key == specmodel.CollectionEntryIDKey {
				continue
			}
			if key == externalRefKey {
				if ref, ok := value.(map[string]any); ok {
					ref = maps.Clone(ref)
					delete(ref, "id")
					out[key] = ref
					continue
				}
			}
			out[key] = normalizeBody(value, false)
		}
		return out
	case []any:
		out := make([]any, len(typed))
		for i, value := range typed {
			out[i] = normalizeBody(value, false)
		}
		return out
	}
	return node
}

// deploymentChanges names the deployment blocks that differ. The container's
// image is left out, since it records what was running and import never writes
// it, and so is the id inside a volume's external reference.
func deploymentChanges(bundle *specmodel.Deployment, current *specmodel.AppDoc) []string {
	if bundle == nil {
		return nil
	}
	have := &specmodel.Deployment{}
	if current != nil && current.Deployment != nil {
		have = current.Deployment
	}
	var changes []string
	for _, block := range []struct {
		name       string
		want, have any
	}{
		{"deployment.source", bundle.Source, have.Source},
		{"deployment.container", containerSansImage(bundle.Container), containerSansImage(have.Container)},
		{"deployment.resources", bundle.Resources, have.Resources},
		{"deployment.storage", storageSansIDs(bundle.Storage), storageSansIDs(have.Storage)},
		{"deployment.networks", bundle.Networks, have.Networks},
		{"deployment.service", bundle.Service, have.Service},
	} {
		if !sameYAML(block.want, block.have) {
			changes = append(changes, block.name)
		}
	}
	return changes
}

func containerSansImage(c *specmodel.Container) *specmodel.Container {
	if c == nil {
		return nil
	}
	out := *c
	out.Image = ""
	return &out
}

func storageSansIDs(s *specmodel.Storage) *specmodel.Storage {
	if s == nil {
		return nil
	}
	out := &specmodel.Storage{DockerMounts: s.DockerMounts, Mounts: make(map[string]specmodel.Mount, len(s.Mounts))}
	for target, m := range s.Mounts {
		if m.External != nil {
			ref := *m.External
			ref.ID = ""
			m.External = &ref
		}
		out.Mounts[target] = m
	}
	return out
}

// sameYAML compares two values the way they are written: absent and empty are
// the same, as they are in a bundle.
func sameYAML(a, b any) bool {
	encodedA, errA := yaml.Marshal(a)
	encodedB, errB := yaml.Marshal(b)
	return errA == nil && errB == nil && string(encodedA) == string(encodedB)
}

// setChanges records what differs and the action it makes.
func setChanges(node *specmodel.PlanNode, changes []string) {
	node.Changes = changes
	if node.Action == specmodel.ActionSkip {
		return
	}
	node.Action = specmodel.ActionUnchanged
	if len(changes) > 0 {
		node.Action = specmodel.ActionUpdate
	}
}

func keyMismatch(path, targetKey string) *specmodel.Issue {
	return &specmodel.Issue{
		Severity: specmodel.SeveritySkipped, Code: specmodel.CodeKeyMismatch, Path: path,
		Detail: map[string]any{"targetKey": targetKey},
		Action: "not imported: the id names an object with another key here, so following it would " +
			"rename and overwrite that object",
	}
}

func (p *planner) skipNode(node *specmodel.PlanNode, issue specmodel.Issue) {
	node.Action, node.Restart, node.Deploy = specmodel.ActionSkip, false, false
	node.Issues = append(node.Issues, issue)
}

// skipBelow skips a node and every node below it: nothing inside an object that
// is not imported can be.
func (p *planner) skipBelow(path string, issue specmodel.Issue) {
	for _, node := range p.nodes {
		if node.Path == path || strings.HasPrefix(node.Path, path+"/") {
			p.skipNode(node, issue)
		}
	}
}

// applyExisting leaves what the installation has alone when the operator chose
// to: only what is missing is created, and nothing restarts. A setting is an
// object of its own, so a scope that exists still gains the settings it lacks;
// an app's settings are part of the app, and stay as the app does.
func (p *planner) applyExisting(node *specmodel.PlanNode) {
	if p.req.Options.Existing != specmodel.ExistingKeep {
		return
	}
	if node.Action != specmodel.ActionUpdate && node.Action != specmodel.ActionUnchanged {
		return
	}
	if missing := p.missing[node.Path]; node.Kind != specmodel.NodeKindApp && len(missing) > 0 {
		node.Action, node.Changes = specmodel.ActionUpdate, missing
		return
	}
	node.Action, node.Restart, node.Deploy = specmodel.ActionKeep, false, false
}

// summarize counts what the selected nodes will do.
func summarize(nodes []*specmodel.PlanNode) map[string]int {
	summary := map[string]int{}
	for _, node := range nodes {
		if !node.Selected {
			continue
		}
		summary[string(node.Action)]++
		if node.Restart {
			summary["restart"]++
		}
		if node.Deploy {
			summary["deploy"]++
		}
		for _, issue := range node.Issues {
			summary[string(issue.Severity)]++
		}
	}
	return summary
}

// planHash binds apply to the plan the operator saw: the bundle, the selection,
// the options, and what each selected node will do. It leaves out this
// installation's values, so an edit that does not change what import does does
// not change the hash.
func planHash(digest string, req *specservice.ValidateImportReq, nodes []*specmodel.PlanNode) string {
	type hashedNode struct {
		Path    string
		Action  specmodel.NodeAction
		Restart bool
		Deploy  bool
		Issues  []string
	}
	hashed := make([]hashedNode, 0, len(nodes))
	for _, node := range nodes {
		if !node.Selected {
			continue
		}
		codes := make([]string, 0, len(node.Issues))
		for _, issue := range node.Issues {
			codes = append(codes, issue.Code)
		}
		sort.Strings(codes)
		hashed = append(hashed, hashedNode{node.Path, node.Action, node.Restart, node.Deploy, codes})
	}
	sort.Slice(hashed, func(i, j int) bool { return hashed[i].Path < hashed[j].Path })
	encoded, _ := json.Marshal(struct {
		Digest    string
		Selection specmodel.Selection
		Options   specmodel.ImportOptions
		Nodes     []hashedNode
	}{digest, req.Selection, req.Options, hashed})
	sum := sha256.Sum256(encoded)
	return hex.EncodeToString(sum[:])
}
