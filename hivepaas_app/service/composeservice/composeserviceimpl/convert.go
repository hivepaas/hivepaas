package composeserviceimpl

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"maps"
	"regexp"
	"slices"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/projecthelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// The keys of an issue's detail.
const (
	detailField    = "field"
	detailPath     = "path"
	detailTarget   = "target"
	detailSource   = "source"
	detailServices = "services"
	detailImage    = "image"
	detailApp      = "app"
	detailName     = "name"
)

// The parts of a mounted setting, and an entry file's key for its part.
const (
	partContent = "content"
	partValue   = "value"
	filePart    = "part"
)

type service struct{}

func New() composeservice.Service {
	return &service{}
}

func (s *service) Convert(
	ctx context.Context,
	req *composeservice.ConvertReq,
) (*composeservice.ConvertResp, error) {
	r, err := readCompose(ctx, req)
	if err != nil {
		return nil, err
	}
	c := &converter{req: req, r: r, resp: &composeservice.ConvertResp{
		Issues: map[string][]specmodel.Issue{}, Profiles: fileProfiles(r.raw),
	}}
	c.resp.Variables = c.variableViews()
	if len(r.missing) > 0 || len(r.missingFiles) > 0 {
		c.resp.Needs = r.missingFiles
		return c.resp, nil
	}
	if len(r.project.Services) == 0 {
		return nil, hperrors.Wrap(hperrors.ErrComposeNoServices)
	}
	if r.project.Name != placeholderName {
		c.resp.FileName = r.project.Name
	}
	c.convert()
	if err = c.assemble(); err != nil {
		return nil, err
	}
	return c.resp, nil
}

// converter makes the env's document of a compose project: an app per
// service, and the env's own secrets and config files.
type converter struct {
	req  *composeservice.ConvertReq
	r    *read
	resp *composeservice.ConvertResp
	env  *specmodel.EnvDoc

	// keys are the app key of each service, by its name.
	keys map[string]string
	// taken are the names the existing env's apps answer to, each with its
	// app's key; existing the app a service's name or key is, by service; used
	// the existing app a service is used as.
	taken, existing, used map[string]string
	// secretNames and configNames are the env setting each of the file's
	// secrets and configs is, by its name in the file.
	secretNames, configNames map[string]string
	// names are the services, in depends_on order.
	names []string
	// aliases are the names each service is reached by beside its key.
	aliases map[string][]string
	// owners are the service whose directory a shared volume is, by volume.
	owners map[string]string
	// jobs are the services run to completion.
	jobs map[string]bool
	// envSecrets and envConfigs are the env's settings.
	envSecrets map[string]any
	envConfigs map[string]any
	// secretVariables are the secret variables an environment refers to.
	secretVariables map[string]bool
	needs           map[string]*composeservice.FileNeed
}

func (c *converter) convert() {
	c.env = &specmodel.EnvDoc{Project: c.req.ProjectKey, Env: c.req.EnvKey, Name: c.req.EnvName,
		Color: c.req.EnvColor, Index: c.req.EnvIndex, Apps: map[string]*specmodel.AppDoc{}}
	c.envSecrets, c.envConfigs = map[string]any{}, map[string]any{}
	c.secretVariables, c.needs = map[string]bool{}, map[string]*composeservice.FileNeed{}

	c.nameApps()
	c.orderServices()
	c.findJobs()
	c.findAliases()
	c.findVolumeOwners()
	c.fileObjects()
	for _, name := range c.names {
		doc, view := c.app(name, c.r.project.Services[name])
		c.env.Apps[c.keys[name]] = doc
		c.resp.Services = append(c.resp.Services, view)
	}
	c.secretVariableSettings()
	c.emptyVariables()
	for _, block := range []map[string]any{c.envSecrets, c.envConfigs} {
		for _, body := range block {
			fields, _ := body.(map[string]any)
			delete(fields, fileSourceKey)
			delete(fields, variableKey)
		}
	}

	settings := map[string]any{}
	if len(c.envSecrets) > 0 {
		settings[blockSecrets] = c.envSecrets
	}
	if len(c.envConfigs) > 0 {
		settings[blockConfigFiles] = c.envConfigs
	}
	if len(settings) > 0 {
		c.env.Settings = settings
	}
	for _, p := range slices.Sorted(maps.Keys(c.needs)) {
		c.resp.Needs = append(c.resp.Needs, c.needs[p])
	}
}

// assemble is the bundle: the project, its env, and a digest naming what it
// was read from - the same request reads into the same bundle.
func (c *converter) assemble() error {
	header := specmodel.NewDocHeader(string(base.ObjectScopeProject))
	c.env.DocHeader = header
	project := &specmodel.ProjectDoc{DocHeader: header, Project: c.req.ProjectKey, Name: c.req.ProjectName,
		Envs: []string{c.req.EnvKey}}
	if c.req.OwnerID != "" {
		project.Owner = &specmodel.ProjectOwner{ID: c.req.OwnerID}
	}
	data, err := json.Marshal([]any{project, c.env})
	if err != nil {
		return hperrors.Wrap(err)
	}
	digest := sha256.Sum256(data)
	c.resp.Bundle = &specmodel.ImportBundle{
		Manifest: &specmodel.Manifest{
			APIVersion: specmodel.APIVersion, Kind: specmodel.KindSpec, SourceAppVersion: "compose",
			Scope: string(base.ObjectScopeProject), SecretsMode: specmodel.SecretsModePlaintext,
		},
		Projects: map[string]*specmodel.ProjectDoc{c.req.ProjectKey: project},
		Envs:     map[string]map[string]*specmodel.EnvDoc{c.req.ProjectKey: {c.req.EnvKey: c.env}},
		Digest:   hex.EncodeToString(digest[:]),
	}
	return nil
}

func (c *converter) envPath() string {
	return "projects/" + c.req.ProjectKey + "/envs/" + c.req.EnvKey
}

func (c *converter) appPath(name string) string {
	return c.envPath() + "/apps/" + c.keys[name]
}

func (c *converter) add(path string, severity specmodel.Severity, code string, detail map[string]any, action string) {
	c.resp.Issues[path] = append(c.resp.Issues[path], specmodel.Issue{
		Severity: severity, Code: code, Path: path, Detail: detail, Action: action,
	})
}

// nameApps gives each service the app key its name makes, or the one the
// review chose. Two services whose names make one key cannot both be created:
// the file has to rename one. In an existing env, a service whose name or key
// an app there answers to waits for the review: that app used, or another key.
func (c *converter) nameApps() {
	c.keys, c.existing, c.used = map[string]string{}, map[string]string{}, map[string]string{}
	c.taken = existingNames(c.req.Existing)
	byKey := map[string]string{}
	var conflicts []string
	for _, name := range slices.Sorted(maps.Keys(c.r.project.Services)) {
		choice := c.req.Services[name]
		key, existing := projecthelper.CalcAppKey(name), ""
		if choice != nil && choice.App != "" {
			key = projecthelper.CalcAppKey(choice.App)
			existing = c.taken[key]
		} else {
			existing = gofn.Coalesce(c.taken[key], c.taken[name])
		}
		if existing != "" && choice != nil && choice.UseExisting {
			key, c.used[name] = existing, existing
		}
		if other, taken := byKey[key]; taken {
			c.add(c.envPath(), specmodel.SeverityBlocked, composeservice.CodeKeyConflict,
				map[string]any{detailServices: []string{other, name}, detailApp: key},
				"two services would be one app: rename one in the file")
			continue
		}
		byKey[key], c.keys[name] = name, key
		if existing != "" {
			c.existing[name] = existing
			if c.used[name] == "" {
				conflicts = append(conflicts, name)
			}
		}
	}
	for _, name := range conflicts {
		c.add(c.appPath(name), specmodel.SeverityBlocked, composeservice.CodeAppExists,
			map[string]any{"service": name, detailApp: c.existing[name]},
			"the env has an app by this name: use it, or give the service another key")
	}
}

// orderServices puts the services in depends_on order, each after what it
// depends on - the order the import creates the apps and their volumes'
// owners first. A cycle is the file's mistake.
func (c *converter) orderServices() {
	pending := map[string][]string{}
	for name := range c.keys {
		for dep := range c.r.project.Services[name].DependsOn {
			if _, ok := c.keys[dep]; ok && dep != name {
				pending[name] = append(pending[name], dep)
			}
		}
	}
	done := map[string]bool{}
	for len(c.names) < len(c.keys) {
		var ready []string
		for name := range c.keys {
			if !done[name] && allDone(pending[name], done) {
				ready = append(ready, name)
			}
		}
		if len(ready) == 0 {
			var rest []string
			for name := range c.keys {
				if !done[name] {
					rest = append(rest, name)
				}
			}
			slices.Sort(rest)
			c.add(c.envPath(), specmodel.SeverityBlocked, composeservice.CodeDependsCycle,
				map[string]any{detailServices: rest}, "these services depend on each other: none can start first")
			c.names = append(c.names, rest...)
			return
		}
		slices.Sort(ready)
		for _, name := range ready {
			done[name] = true
		}
		c.names = append(c.names, ready...)
	}
}

func allDone(names []string, done map[string]bool) bool {
	for _, name := range names {
		if !done[name] {
			return false
		}
	}
	return true
}

// findJobs finds the services run to completion: one that does not restart,
// and that another waits on to complete - a migration, a setup step.
func (c *converter) findJobs() {
	c.jobs = map[string]bool{}
	for _, svc := range c.r.project.Services {
		for dep, dependency := range svc.DependsOn {
			target, ok := c.r.project.Services[dep]
			if ok && dependency.Condition == types.ServiceConditionCompletedSuccessfully && !restarts(target) {
				c.jobs[dep] = true
			}
		}
	}
}

// restarts says whether compose restarts a service's container once it exits.
func restarts(svc types.ServiceConfig) bool {
	if svc.Deploy != nil && svc.Deploy.RestartPolicy != nil {
		return svc.Deploy.RestartPolicy.Condition != "none"
	}
	return svc.Restart != "" && svc.Restart != types.RestartPolicyNo
}

// aliasPattern is a name a container may be reached by on a network.
var aliasPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]{0,62}$`)

// findAliases are the names each service is reached by in compose beside its
// key: its own name when the key is not it, its container's, what its
// networks call it, and what another's links call it.
func (c *converter) findAliases() {
	c.aliases = map[string][]string{}
	taken := map[string][]string{}
	add := func(name, alias string) {
		if _, ok := c.keys[name]; !ok || c.used[name] != "" || alias == c.keys[name] ||
			!aliasPattern.MatchString(alias) || slices.Contains(c.aliases[name], alias) {
			return
		}
		if c.taken[alias] != "" {
			if !slices.Contains(taken[name], alias) {
				taken[name] = append(taken[name], alias)
			}
			return
		}
		c.aliases[name] = append(c.aliases[name], alias)
	}
	for _, name := range slices.Sorted(maps.Keys(c.keys)) {
		svc := c.r.project.Services[name]
		add(name, name)
		add(name, svc.ContainerName)
		for _, network := range slices.Sorted(maps.Keys(svc.Networks)) {
			if config := svc.Networks[network]; config != nil {
				for _, alias := range config.Aliases {
					add(name, alias)
				}
			}
		}
		for _, link := range svc.Links {
			target, alias, found := cutLink(link)
			if found {
				add(target, alias)
			}
		}
	}
	for _, name := range slices.Sorted(maps.Keys(taken)) {
		c.add(c.appPath(name), specmodel.SeverityWarning, composeservice.CodeAliasTaken,
			map[string]any{"aliases": taken[name]}, "not added: the env's apps answer to these names already")
	}
}

// cutLink reads a link, `service` or `service:alias`.
func cutLink(link string) (string, string, bool) {
	for i := range len(link) {
		if link[i] == ':' {
			return link[:i], link[i+1:], true
		}
	}
	return link, "", false
}

// fileProfiles are every profile the file names, from the file as written:
// the loaded project holds only the services of those asked for.
func fileProfiles(raw map[string]any) []string {
	services, _ := raw["services"].(map[string]any)
	var out []string
	for _, svc := range services {
		fields, _ := svc.(map[string]any)
		profiles, _ := fields["profiles"].([]any)
		for _, profile := range profiles {
			if name, ok := profile.(string); ok && !slices.Contains(out, name) {
				out = append(out, name)
			}
		}
	}
	slices.Sort(out)
	return out
}

func (c *converter) variableViews() []*composeservice.VariableView {
	out := make([]*composeservice.VariableView, 0, len(c.r.variables))
	for _, name := range slices.Sorted(maps.Keys(c.r.variables)) {
		v := c.r.variables[name]
		out = append(out, &composeservice.VariableView{
			Name: name, Default: v.DefaultValue, Required: v.Required,
			Given: c.r.values[name] != "", Secret: c.r.secret[name],
		})
	}
	return out
}

// emptyVariables warns of what the file uses with no value and no default:
// empty, as compose leaves it.
func (c *converter) emptyVariables() {
	var empty []string
	for _, name := range slices.Sorted(maps.Keys(c.r.variables)) {
		v := c.r.variables[name]
		if _, given := c.r.values[name]; !given && v.DefaultValue == "" && !v.Required {
			empty = append(empty, name)
		}
	}
	if len(empty) > 0 {
		c.add(c.envPath(), specmodel.SeverityWarning, composeservice.CodeVariableEmpty,
			map[string]any{"variables": empty}, "these are empty where the file uses them")
	}
}
