package specserviceimpl

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/moby/moby/api/types/network"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// generatedSecretBytes is how much randomness a secret HivePaaS generates for an
// imported object gets, as hex.
const generatedSecretBytes = 24

// changeDeploymentSource names an app's deployment source among its changes.
const changeDeploymentSource = "deployment.source"

// sourceSetting is how the writer names an app's deployment source among its
// settings: it travels as deployment.source and is stored as the app-deployment
// setting.
var sourceSetting = specmodel.SingletonBlockName(base.SettingTypeAppDeployment)

// appWrittenNames are the settings an app node writes: all of an app created,
// what changed of one updated - its deployment source among them.
func appWrittenNames(node *specmodel.PlanNode, doc *specmodel.AppDoc) []string {
	hasSource := doc.Deployment != nil && doc.Deployment.Source != nil
	if node.Action == specmodel.ActionCreate {
		names := settingNames(doc.Settings)
		if hasSource {
			names = append(names, sourceSetting)
		}
		return names
	}
	var names []string
	for _, change := range node.Changes {
		if name, ok := strings.CutPrefix(change, "settings."); ok {
			names = append(names, name)
		}
		if change == changeDeploymentSource && hasSource {
			names = append(names, sourceSetting)
		}
	}
	return names
}

// settingOf is one setting a node writes: its body in the bundle, its type,
// and its key among the entries of a collection.
func (w *writer) settingOf(node *specmodel.PlanNode, name string) (any, base.SettingType, string, bool) {
	if place, isApp := w.p.apps[node.Path]; isApp && name == sourceSetting {
		if place.doc.Deployment == nil || place.doc.Deployment.Source == nil {
			return nil, "", "", false
		}
		return place.doc.Deployment.Source, base.SettingTypeAppDeployment, "", true
	}
	block, key, _ := strings.Cut(name, "/")
	body, typ, found := settingBody(w.p.settingsOf[node.Path], block, key)
	return body, typ, key, found
}

// holderOf is how validate's issues name a setting of a node.
func (w *writer) holderOf(node *specmodel.PlanNode, name string) string {
	if _, isApp := w.p.apps[node.Path]; !isApp {
		return name
	}
	if name == sourceSetting {
		return changeDeploymentSource
	}
	return "settings." + name
}

// chooseAppIDs gives every app written its id here before anything is built:
// a new one for an app created, the matched one otherwise.
func (w *writer) chooseAppIDs(nodes []*specmodel.PlanNode) {
	for _, node := range nodes {
		place, isApp := w.p.apps[node.Path]
		if !isApp {
			continue
		}
		id := node.TargetID
		if node.Action == specmodel.ActionCreate {
			id = gofn.Must(ulid.NewStringULID())
		}
		w.appIDs[node.Path] = id
		if place.doc.ID != "" {
			w.appIDsByBundle[place.doc.ID] = id
		}
	}
}

// appRefID is the id here of an app a setting names by its id in the bundle: an
// app this import writes, the one an app of the bundle matched, or the app itself
// when this installation has it. Otherwise the reference is cleared.
func (w *writer) appRefID(ctx context.Context, ref string) (string, error) {
	if ref == "" {
		return "", nil
	}
	if id := w.appIDsByBundle[ref]; id != "" {
		return id, nil
	}
	if path := appPathByID(w.p.full, ref); path != "" {
		if app := w.p.targetApps[path]; app != nil {
			return app.ID, nil
		}
	}
	_, err := w.p.s.appRepo.GetByID(ctx, w.p.db, "", ref)
	switch {
	case err == nil:
		return ref, nil
	case errors.Is(err, hperrors.ErrNotFound):
		return "", nil
	}
	return "", hperrors.Wrap(err)
}

// keepAndGenerate decides the secrets of a setting written over row, the one it
// replaces (nil for a setting created):
//   - a bundle without secrets keeps every secret of row it holds nothing in;
//   - an app matched by key keeps the target's credential, which its data was
//     initialized with;
//   - a secret's or config file's swarm ids are row's, or none: another
//     installation's ids name nothing here;
//   - a setting created from a bundle without secrets gets the values HivePaaS
//     owns generated.
func (w *writer) keepAndGenerate(node *specmodel.PlanNode, data entity.SettingData, row *entity.Setting) error {
	mode := w.p.bundle.Manifest.SecretsMode
	var kept entity.SettingData
	if row != nil {
		var err error
		if kept, err = row.Parse(); err != nil {
			return hperrors.Wrap(err)
		}
		if mode == specmodel.SecretsModeOmit {
			entity.KeepSecrets(data, kept)
		}
	}
	switch typed := data.(type) {
	case *entity.Secret:
		keptSecret, _ := kept.(*entity.Secret)
		if typed.SwarmRef != nil {
			typed.SwarmRef.SecretID, typed.SwarmRef.SecretName = "", ""
			if keptSecret != nil && keptSecret.SwarmRef != nil {
				typed.SwarmRef.SecretID, typed.SwarmRef.SecretName =
					keptSecret.SwarmRef.SecretID, keptSecret.SwarmRef.SecretName
			}
		}
	case *entity.ConfigFile:
		keptConfig, _ := kept.(*entity.ConfigFile)
		if typed.SwarmRef != nil {
			typed.SwarmRef.ConfigID, typed.SwarmRef.ConfigName = "", ""
			if keptConfig != nil && keptConfig.SwarmRef != nil {
				typed.SwarmRef.ConfigID, typed.SwarmRef.ConfigName =
					keptConfig.SwarmRef.ConfigID, keptConfig.SwarmRef.ConfigName
			}
		}
	case *entity.AppKindSettings:
		if keptKind, ok := kept.(*entity.AppKindSettings); ok && mode.RevealsSecrets() &&
			node.MatchedBy == specmodel.MatchedByKey {
			keepCredentials(typed, keptKind)
		}
	}
	if row == nil && mode == specmodel.SecretsModeOmit && slices.Contains(generatedSecretTypes, data.GetType()) {
		generateOwnedSecrets(data)
	}
	return nil
}

// keepCredentials puts the target's credential in place of the bundle's.
func keepCredentials(dst, src *entity.AppKindSettings) {
	keep := func(to *entity.EncryptedField, from entity.EncryptedField) {
		if !from.IsEmpty() {
			*to = from
		}
	}
	if dst.Database != nil && src.Database != nil {
		keep(&dst.Database.Password, src.Database.Password)
		keep(&dst.Database.RootPassword, src.Database.RootPassword)
	}
	if dst.Cache != nil && src.Cache != nil {
		keep(&dst.Cache.Password, src.Cache.Password)
	}
	if dst.Storage != nil && src.Storage != nil {
		keep(&dst.Storage.Secret, src.Storage.Secret)
	}
}

// generateOwnedSecrets gives a value HivePaaS owns to each such secret a setting
// holds nothing in. Nothing depends on them before the object exists: an app
// initializes its storage with the credential it is given.
func generateOwnedSecrets(data entity.SettingData) {
	generate := func(field *entity.EncryptedField) {
		if field.IsEmpty() {
			*field = entity.NewEncryptedField(gofn.RandTokenAsHex(generatedSecretBytes))
		}
	}
	switch typed := data.(type) {
	case *entity.RepoWebhook:
		generate(&typed.Secret)
	case *entity.AppKindSettings:
		if typed.Database != nil {
			generate(&typed.Database.Password)
			generate(&typed.Database.RootPassword)
		}
		if typed.Cache != nil {
			generate(&typed.Cache.Password)
		}
		if typed.Storage != nil {
			generate(&typed.Storage.Secret)
		}
	}
}

// dropDomain removes a domain another app holds from an app's routing.
func dropDomain(routing *entity.AppRoutingSettings, domain any) {
	routing.Domains = slices.DeleteFunc(routing.Domains, func(d *entity.AppDomain) bool {
		return d != nil && d.Domain == domain
	})
}

// preparedDeployment is the deployment an app is built from: the bundle's,
// without its source - written as a setting - with each managed mount's volume
// named by its id here, and without what validate said would be dropped: a mount
// whose volume or app nothing here satisfies, a port another service holds.
func (w *writer) preparedDeployment(ctx context.Context, node *specmodel.PlanNode) (*specmodel.Deployment, error) {
	doc := w.p.apps[node.Path].doc
	if doc.Deployment == nil {
		return nil, nil
	}
	out := *doc.Deployment
	out.Source = nil

	droppedMounts, droppedPorts := map[any]bool{}, map[any]bool{}
	for _, issue := range node.Issues {
		switch issue.Code {
		case specmodel.CodeRefNotSelected, specmodel.CodeRefNotFound:
			if target, ok := issue.Detail[refInMount]; ok {
				droppedMounts[target] = true
			}
		case specmodel.CodePortInUse:
			droppedPorts[issue.Detail["port"]] = true
		}
	}

	if storage := doc.Deployment.Storage; storage != nil {
		prepared := &specmodel.Storage{DockerMounts: storage.DockerMounts}
		scope := w.p.lookupScope[node.Path]
		for _, target := range slices.Sorted(maps.Keys(storage.Mounts)) {
			if droppedMounts[target] {
				continue
			}
			m := storage.Mounts[target]
			id, err := w.volumeID(ctx, scope, m)
			if err != nil {
				return nil, err
			}
			if id == "" {
				continue
			}
			m.Source, m.External = id, nil
			prepared.Mounts = withMount(prepared.Mounts, target, m)
		}
		out.Storage = prepared
	}

	if networks := doc.Deployment.Networks; networks != nil && networks.EndpointSpec != nil && len(droppedPorts) > 0 {
		prepared, endpoint := *networks, *networks.EndpointSpec
		endpoint.Ports = slices.DeleteFunc(slices.Clone(endpoint.Ports), func(port *specmodel.PortConfig) bool {
			return port != nil && droppedPorts[describePortConfig(port)]
		})
		prepared.EndpointSpec = &endpoint
		out.Networks = &prepared
	}
	return &out, nil
}

// volumeID is the id here of the volume a managed mount names.
func (w *writer) volumeID(ctx context.Context, scope *entity.ObjectScope, m specmodel.Mount) (string, error) {
	if scope == nil {
		scope = entity.NewObjectScopeGlobal()
	}
	if m.External != nil {
		found, err := w.p.s.findRef(ctx, w.p.db, scope, m.External)
		if err != nil || found == nil {
			return "", hperrors.Wrap(err)
		}
		return found.ID, nil
	}
	t, isPath := parseRefPath(m.Source)
	if !isPath {
		existing, err := w.p.s.loadByIDs(ctx, w.p.db, []string{m.Source})
		if err != nil || len(existing) == 0 {
			return "", hperrors.Wrap(err)
		}
		return m.Source, nil
	}
	return w.pathID(ctx, scope, m.Source, t)
}

// describePortConfig names a published port the way validate's issues do.
func describePortConfig(port *specmodel.PortConfig) string {
	protocol := port.Protocol
	if protocol == "" {
		protocol = network.TCP
	}
	return fmt.Sprintf("%d/%s", port.Published, strings.ToLower(string(protocol)))
}
