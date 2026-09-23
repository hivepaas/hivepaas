package specmodel

import (
	"fmt"
	"maps"
	"reflect"
	"regexp"
	"slices"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/network"
	"github.com/moby/moby/api/types/swarm"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// Block names a part of an AppDoc that can be built into an app.
type Block string

const (
	BlockDeploymentSource     Block = "deployment.source"
	BlockDeploymentStorage    Block = "deployment.storage"
	BlockContainerHealthcheck Block = "deployment.container.healthcheck"
	BlockContainerInit        Block = "deployment.container.init"
	BlockDeploymentResources  Block = "deployment.resources"
	BlockDeploymentNetworks   Block = "deployment.networks"
	BlockSettingsKind         Block = "settings.kind"
	BlockSettingsEnvVars      Block = "settings.envVars"
	BlockSettingsSecrets      Block = "settings.secrets"
	BlockSettingsConfigFiles  Block = "settings.configFiles"
	BlockSettingsRouting      Block = "settings.routing"
)

const (
	// MaxSettingsPerBlock caps how many secrets or config files one app may be
	// built with. An app that needs more of either is being configured, not
	// provisioned.
	MaxSettingsPerBlock = 10
	// MaxRoutingDomains is what one app may answer at. A template names the
	// address an app is created with, not every address it will ever have.
	MaxRoutingDomains = 5
	// MaxSecretValueBytes is what a secret may hold: a password, a key, a
	// certificate - never a file somebody meant to mount.
	MaxSecretValueBytes = 64 << 10
	// MaxPublishedPorts is how many addresses on the cluster one app may claim.
	// A published port is taken from every other app on the installation, so a
	// template asks for the few the software answers at.
	MaxPublishedPorts = 5
	// maxPortNumber is the largest port there is.
	maxPortNumber = 65535
	// MaxCapabilityEntries is how many capabilities, sysctls or ulimits one
	// document may carry. A template asks for the few its software cannot run
	// without; a list longer than this is not that.
	MaxCapabilityEntries = 10
	// MaxConfigFileBytes is what one config file may hold. Docker's own limit is
	// 500KB; a configuration file a template writes is far smaller than that, and
	// every byte here is copied into the swarm and into every task.
	MaxConfigFileBytes = 128 << 10
)

// capabilityNamePattern is how docker names a capability in a service spec:
// the name without the CAP_ prefix, which is what the app's resource settings
// screen takes as well.
var capabilityNamePattern = regexp.MustCompile(`^[A-Z][A-Z0-9_]{0,31}$`)

// mountSourceAppKeyPattern is the shape projecthelper.CalcAppKey produces, which
// is what a `type: app` parameter of a template carries: hyphens now, underscores
// for apps created before keys became host names.
var mountSourceAppKeyPattern = regexp.MustCompile(`^[a-z0-9_-]{1,100}$`)

// publishedProtocols and publishModes are what docker takes for a published
// port. An empty protocol is tcp and an empty mode is ingress, which is why both
// lists leave the empty value out and the check allows it separately.
var (
	publishedProtocols = []network.IPProtocol{network.TCP, network.UDP, network.SCTP}
	publishModes       = []swarm.PortConfigPublishMode{
		swarm.PortConfigPublishModeIngress, swarm.PortConfigPublishModeHost,
	}
)

const (
	// capabilityAll is the wildcard docker accepts and a document may not use.
	capabilityAll = "ALL"
	// capabilityPrefix is the spelling docker's own API uses and this format does
	// not. The daemon takes either, and one spelling everywhere is what keeps the
	// app's resource settings screen showing what the template asked for.
	capabilityPrefix = "CAP_"
)

// Blocks only import builds. A template cannot ask for them - CheckBuildable
// refuses every field they carry - so PresentBlocks never names them.
const (
	BlockContainer         Block = "deployment.container"
	BlockDeploymentService Block = "deployment.service"
	// BlockSettings is every settings block of an exported document, built from
	// the rows export wrote beside the data.
	BlockSettings Block = "settings"
)

var ImportOnlyBlocks = []Block{BlockContainer, BlockDeploymentService, BlockSettings}

// buildOrder is the order blocks are built in. Storage replaces the mounts
// before resources sets the size of /dev/shm, which is one of them.
var buildOrder = []Block{
	BlockDeploymentSource, BlockDeploymentStorage, BlockContainerHealthcheck, BlockContainerInit,
	BlockContainer, BlockDeploymentResources, BlockDeploymentNetworks, BlockDeploymentService,
	BlockSettingsKind, BlockSettingsEnvVars, BlockSettingsSecrets, BlockSettingsConfigFiles,
	BlockSettingsRouting, BlockSettings,
}

// ImportBlocks lists the blocks an exported document is built with, in build
// order. Every deployment block is built when the document has a deployment at
// all: export writes a block only when it holds something, so a block missing
// from a deployment is one the app has none of, and building it clears whatever
// the service holds. The container block covers its healthcheck and init. The
// source is a setting rather than part of the service, and is built only when
// present; the settings are one block, built from their rows - import never
// deletes a setting.
func ImportBlocks(doc *AppDoc) []Block {
	var blocks []Block
	if doc == nil {
		return blocks
	}
	if d := doc.Deployment; d != nil {
		if d.Source != nil {
			blocks = append(blocks, BlockDeploymentSource)
		}
		blocks = append(blocks, BlockDeploymentStorage, BlockContainer, BlockDeploymentResources,
			BlockDeploymentNetworks, BlockDeploymentService)
	}
	if len(doc.Settings) > 0 {
		blocks = append(blocks, BlockSettings)
	}
	slices.SortFunc(blocks, func(a, b Block) int {
		return slices.Index(buildOrder, a) - slices.Index(buildOrder, b)
	})
	return blocks
}

// BuildableBlocks is every block specservice.BuildApp builds, in the order it
// builds them. specserviceimpl's builder registry is tested against this list.
//
// TODO: app templates phase 3 - more blocks: repository builds, registry auth,
// bind mounts, config files, secrets, scheduled jobs, routing domains. See
// docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
var BuildableBlocks = []Block{BlockDeploymentSource, BlockDeploymentStorage, BlockContainerHealthcheck,
	BlockContainerInit, BlockDeploymentResources, BlockDeploymentNetworks, BlockSettingsKind,
	BlockSettingsEnvVars, BlockSettingsSecrets, BlockSettingsConfigFiles, BlockSettingsRouting}

// CheckBuildable refuses any part of doc that phase 1 cannot build.
//
// Refusing is the point: a field that was skipped instead would provision an
// app with less than the document describes, and nothing would say which part
// went missing. The error's extra detail names the first offending path.
func CheckBuildable(doc *AppDoc) error {
	if doc == nil {
		return nil
	}
	if err := onlyFields("", doc, "deployment", "settings"); err != nil {
		return err
	}
	if err := checkDeployment(doc.Deployment); err != nil {
		return err
	}
	return checkSettings(doc.Settings)
}

// PresentBlocks lists the buildable blocks doc carries, in BuildableBlocks order.
// It assumes CheckBuildable passed.
func PresentBlocks(doc *AppDoc) []Block {
	var blocks []Block
	if doc == nil {
		return blocks
	}
	if d := doc.Deployment; d != nil {
		if d.Source != nil {
			blocks = append(blocks, BlockDeploymentSource)
		}
		if d.Storage != nil && len(d.Storage.Mounts) > 0 {
			blocks = append(blocks, BlockDeploymentStorage)
		}
		if d.Container != nil && d.Container.Healthcheck != nil {
			blocks = append(blocks, BlockContainerHealthcheck)
		}
		if d.Container != nil && d.Container.Init != nil {
			blocks = append(blocks, BlockContainerInit)
		}
		if d.Resources != nil {
			blocks = append(blocks, BlockDeploymentResources)
		}
		if d.Networks != nil {
			blocks = append(blocks, BlockDeploymentNetworks)
		}
	}
	blocks = append(blocks, presentSettingsBlocks(doc)...)
	slices.SortFunc(blocks, func(a, b Block) int {
		return slices.Index(BuildableBlocks, a) - slices.Index(BuildableBlocks, b)
	})
	return blocks
}

// presentSettingsBlocks lists the settings blocks doc carries. A collection
// block with no entries builds nothing, so it is not present.
func presentSettingsBlocks(doc *AppDoc) []Block {
	var blocks []Block
	for _, pair := range []struct {
		typ   base.SettingType
		block Block
	}{
		{base.SettingTypeAppKind, BlockSettingsKind},
		{base.SettingTypeEnvVar, BlockSettingsEnvVars},
		{base.SettingTypeAppRouting, BlockSettingsRouting},
	} {
		if _, ok := doc.Settings[SingletonBlockName(pair.typ)]; ok {
			blocks = append(blocks, pair.block)
		}
	}
	for _, pair := range []struct {
		typ   base.SettingType
		block Block
	}{
		{base.SettingTypeSecret, BlockSettingsSecrets},
		{base.SettingTypeConfigFile, BlockSettingsConfigFiles},
	} {
		if entries, ok := doc.Settings[CollectionBlockName(pair.typ)].(map[string]any); ok && len(entries) > 0 {
			blocks = append(blocks, pair.block)
		}
	}
	return blocks
}

func checkDeployment(d *Deployment) error {
	if d == nil {
		return nil
	}
	if err := onlyFields("deployment.", d, "source", "container", "resources", "storage", "networks"); err != nil {
		return err
	}
	if err := checkSource(d.Source); err != nil {
		return err
	}
	if d.Container != nil {
		if err := onlyFields("deployment.container.", d.Container, "healthcheck", "init"); err != nil {
			return err
		}
	}
	if err := checkResources(d.Resources); err != nil {
		return err
	}
	if err := checkNetworks(d.Networks); err != nil {
		return err
	}
	return checkStorage(d.Storage)
}

func checkSource(source map[string]any) error {
	for _, key := range slices.Sorted(maps.Keys(source)) {
		path := "deployment.source." + key
		switch key {
		case "activeMethod":
			if source[key] != string(base.DeploymentMethodImage) {
				return unsupported(path)
			}
		case "imageSource":
			imageSource, ok := source[key].(map[string]any)
			if !ok {
				return unsupported(path)
			}
			for _, field := range slices.Sorted(maps.Keys(imageSource)) {
				if field != "image" {
					return unsupported(path + "." + field)
				}
			}
		case "command", "workingDir":
		default:
			return unsupported(path)
		}
	}
	return nil
}

func checkResources(r *Resources) error {
	if r == nil {
		return nil
	}
	if err := onlyFields("deployment.resources.", r, "reservations", "limits", "capabilities"); err != nil {
		return err
	}
	if r.Reservations != nil {
		if err := onlyFields("deployment.resources.reservations.", r.Reservations, "cpus", "memory"); err != nil {
			return err
		}
	}
	if r.Limits != nil {
		if err := onlyFields("deployment.resources.limits.", r.Limits, "cpus", "memory", "pids"); err != nil {
			return err
		}
	}
	return CheckCapabilities(r.Capabilities)
}

// CheckCapabilities is the one part of a document that hands an app more of the
// host than a container ordinarily gets: kernel capabilities, sysctls, resource
// limits the daemon would otherwise cap, the GPU. Everything in the block is
// buildable, and everything in it is privileged - whoever provisions a document
// carrying one has to be allowed to change capabilities, which is what
// apptemplateuc checks before it creates anything.
func CheckCapabilities(c *Capabilities) error {
	if problem := CapabilitiesProblem(c); problem != "" {
		return hperrors.Wrap(hperrors.ErrSpecBlockUnsupported).WithExtraDetail("%s", problem)
	}
	return nil
}

// CapabilitiesProblem says what is wrong with a capabilities block, in words,
// and is empty when nothing is. A template linter prints it; CheckCapabilities
// turns it into the error the API answers with.
//
// What it checks is only that the block is well formed and small. A capability
// is named as docker names it, without the CAP_ prefix, and ALL is refused: a
// document that grants everything is not describing what it needs.
func CapabilitiesProblem(c *Capabilities) string {
	const prefix = "deployment.resources.capabilities."
	if c == nil {
		return ""
	}
	if field := extraField(prefix, c,
		"ulimits", "capabilityAdd", "capabilityDrop", "enableGPU", "oomScoreAdj", "sysctls"); field != "" {
		return field + " is not supported"
	}
	for _, counted := range []struct {
		field string
		count int
	}{
		{"ulimits", len(c.Ulimits)}, {"capabilityAdd", len(c.CapabilityAdd)},
		{"capabilityDrop", len(c.CapabilityDrop)}, {"sysctls", len(c.Sysctls)},
	} {
		if counted.count > MaxCapabilityEntries {
			return fmt.Sprintf("%s%s: at most %d, and this has %d",
				prefix, counted.field, MaxCapabilityEntries, counted.count)
		}
	}
	for _, listed := range []struct {
		field string
		names []string
	}{
		{"capabilityAdd", c.CapabilityAdd}, {"capabilityDrop", c.CapabilityDrop},
	} {
		for _, name := range listed.names {
			if !capabilityNamePattern.MatchString(name) || name == capabilityAll ||
				strings.HasPrefix(name, capabilityPrefix) {
				return fmt.Sprintf("%s%s: %q is not a capability such as NET_ADMIN", prefix, listed.field, name)
			}
		}
	}
	for _, ulimit := range c.Ulimits {
		if ulimit == nil || ulimit.Name == "" {
			return prefix + "ulimits: every entry needs a name"
		}
	}
	if _, unnamed := c.Sysctls[""]; unnamed {
		return prefix + "sysctls: every entry needs a name"
	}
	return ""
}

// checkNetworks accepts the ports an app publishes on the cluster, and nothing
// else in the block.
//
// A published port is how an app answers something that is not HTTP - a VPN, a
// DNS server, a game - and it is the one part of networking a template can know
// about. The rest of the block points at objects of the project: which networks
// the app joins, what its hosts file says, which resolver it uses. A template
// has no way to name those, and the app's network settings are where they are
// chosen.
func checkNetworks(n *Networks) error {
	if n == nil {
		return nil
	}
	if err := onlyFields("deployment.networks.", n, "endpointSpec"); err != nil {
		return err
	}
	if n.EndpointSpec == nil {
		return nil
	}
	if err := onlyFields("deployment.networks.endpointSpec.", n.EndpointSpec, "mode", "ports"); err != nil {
		return err
	}
	if len(n.EndpointSpec.Ports) > MaxPublishedPorts {
		return hperrors.Wrap(hperrors.ErrSpecBlockUnsupported).WithExtraDetail(
			"deployment.networks.endpointSpec.ports: at most %d, and this has %d",
			MaxPublishedPorts, len(n.EndpointSpec.Ports))
	}
	for i, port := range n.EndpointSpec.Ports {
		path := fmt.Sprintf("deployment.networks.endpointSpec.ports[%d]", i)
		if port == nil {
			return unsupported(path)
		}
		if err := onlyFields(path+".", port, "target", "published", "protocol", "publishMode"); err != nil {
			return err
		}
		// A published port of zero means docker picks one, which a template must not
		// do: nothing could be told where to connect, and the app's own description
		// of itself would be wrong.
		for _, numbered := range []struct {
			field string
			value uint32
		}{{"target", port.Target}, {"published", port.Published}} {
			if numbered.value < 1 || numbered.value > maxPortNumber {
				return hperrors.Wrap(hperrors.ErrSpecBlockUnsupported).WithExtraDetail(
					"%s.%s: %d is not a port", path, numbered.field, numbered.value)
			}
		}
		if port.Protocol != "" && !slices.Contains(publishedProtocols, port.Protocol) {
			return unsupported(path + ".protocol")
		}
		if port.PublishMode != "" && !slices.Contains(publishModes, port.PublishMode) {
			return unsupported(path + ".publishMode")
		}
	}
	return nil
}

func checkStorage(s *Storage) error {
	if s == nil {
		return nil
	}
	if err := onlyFields("deployment.storage.", s, "mounts"); err != nil {
		return err
	}
	for _, target := range slices.Sorted(maps.Keys(s.Mounts)) {
		m := s.Mounts[target]
		path := "deployment.storage.mounts." + target + "."
		if m.Type != mount.TypeVolume {
			return unsupported(path + "type")
		}
		if err := onlyFields(path, &m, "type", "source", "readOnly", "volumeOptions", "sourceApp"); err != nil {
			return err
		}
		if err := checkMountSourceApp(path, m.SourceApp); err != nil {
			return err
		}
		if m.VolumeOptions != nil {
			// noCopy alongside subpath: docker copies what an image holds at the mount
			// point into an empty volume the first time it is mounted there, ownership
			// included, which overwrites the permissions the volume was prepared with and
			// leaves an unprivileged process unable to write. A template whose image ships
			// a directory at that path says no to the copy.
			if err := onlyFields(path+"volumeOptions.", m.VolumeOptions, "subpath", "noCopy"); err != nil {
				return err
			}
		}
	}
	return nil
}

// checkMountSourceApp allows a mount to name the app whose directory it reaches.
//
// It is the one thing a template can ask for that reaches outside the app being
// created, so what it may say is narrow: an app of this environment, by key, and
// whether it may write there. Whether that app exists and whether the person
// creating this one may have its data are questions for provisioning, which has
// the database and the session; nothing here can answer either.
func checkMountSourceApp(path string, src *MountSourceApp) error {
	if src == nil {
		return nil
	}
	if err := onlyFields(path+"sourceApp.", src, "app", "write"); err != nil {
		return err
	}
	if !mountSourceAppKeyPattern.MatchString(src.App) {
		return unsupported(path + "sourceApp.app")
	}
	return nil
}

func checkSettings(settings map[string]any) error {
	kind := SingletonBlockName(base.SettingTypeAppKind)
	envVars := SingletonBlockName(base.SettingTypeEnvVar)
	routing := SingletonBlockName(base.SettingTypeAppRouting)

	secrets := CollectionBlockName(base.SettingTypeSecret)
	configFiles := CollectionBlockName(base.SettingTypeConfigFile)

	for _, key := range slices.Sorted(maps.Keys(settings)) {
		switch key {
		case kind, envVars:
		case secrets:
			if err := checkSecrets(settings[key]); err != nil {
				return err
			}
		case configFiles:
			if err := checkConfigFiles(settings[key]); err != nil {
				return err
			}
		case routing:
			if err := checkRouting(settings[key]); err != nil {
				return err
			}
		default:
			return unsupported("settings." + key)
		}
	}
	return nil
}

// checkRouting accepts the port an app listens on and the domains it answers at.
//
// A domain entry carries only what names the address: everything else an
// AppDomain can hold - basic authentication, rate limits, path rewrites, a
// certificate chosen by hand - points at another object, and choosing those
// belongs in the app's routing settings rather than in a template.
func checkRouting(body any) error {
	routing := SingletonBlockName(base.SettingTypeAppRouting)
	fields, ok := body.(map[string]any)
	if !ok {
		return unsupported("settings." + routing)
	}
	for _, field := range slices.Sorted(maps.Keys(fields)) {
		switch field {
		case "port", "exposePublicly":
		case "domains":
			entries, ok := fields[field].([]any)
			if !ok {
				return unsupported("settings." + routing + ".domains")
			}
			if len(entries) > MaxRoutingDomains {
				return hperrors.Wrap(hperrors.ErrSpecBlockUnsupported).WithExtraDetail(
					"settings.%s.domains: at most %d, and this has %d", routing, MaxRoutingDomains, len(entries))
			}
			for i, entry := range entries {
				domain, ok := entry.(map[string]any)
				if !ok {
					return unsupported(fmt.Sprintf("settings.%s.domains[%d]", routing, i))
				}
				for _, name := range slices.Sorted(maps.Keys(domain)) {
					switch name {
					case "domain", "enabled", "protocol", "containerPort", "forceHttps":
					default:
						return unsupported(fmt.Sprintf("settings.%s.domains[%d].%s", routing, i, name))
					}
				}
			}
		default:
			return unsupported("settings." + routing + "." + field)
		}
	}
	return nil
}

// onlyFields refuses the first non-zero field of a struct not named in allowed,
// naming it by its yaml key under prefix.
func onlyFields(prefix string, value any, allowed ...string) error {
	if field := extraField(prefix, value, allowed...); field != "" {
		return unsupported(field)
	}
	return nil
}

// extraField names the first non-zero field of a struct not in allowed, by its
// yaml key under prefix, and is empty when every field is allowed.
func extraField(prefix string, value any, allowed ...string) string {
	v := reflect.Indirect(reflect.ValueOf(value))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return ""
	}
	typ := v.Type()
	for i := range typ.NumField() {
		field := typ.Field(i)
		name, _, _ := strings.Cut(field.Tag.Get("yaml"), ",")
		if name == "" {
			name = field.Name
		}
		if slices.Contains(allowed, name) || v.Field(i).IsZero() {
			continue
		}
		return prefix + name
	}
	return ""
}

func unsupported(path string) error {
	return hperrors.Wrap(hperrors.ErrSpecBlockUnsupported).WithExtraDetail("%s", path)
}

// checkSecrets and checkConfigFiles read the two collection blocks a template
// may write. Both are keyed maps, the key being the setting's name, and both
// refuse the swarm ids: those name docker objects that exist only once the app
// does, and a template that set them would point at somebody else's.
func checkSecrets(body any) error {
	entries, err := collectionEntries(BlockSettingsSecrets, body)
	if err != nil {
		return err
	}
	for _, name := range slices.Sorted(maps.Keys(entries)) {
		path := string(BlockSettingsSecrets) + "." + name
		entry, ok := entries[name].(map[string]any)
		if !ok {
			return unsupported(path)
		}
		for _, field := range slices.Sorted(maps.Keys(entry)) {
			switch field {
			case "key", "base64":
			case "value":
				if text, isText := entry[field].(string); isText && len(text) > MaxSecretValueBytes {
					return tooLarge(path+".value", len(text), MaxSecretValueBytes)
				}
			case "swarmRef":
				if err := checkSwarmRef(path+".swarmRef", entry[field]); err != nil {
					return err
				}
			default:
				return unsupported(path + "." + field)
			}
		}
	}
	return nil
}

func checkConfigFiles(body any) error {
	entries, err := collectionEntries(BlockSettingsConfigFiles, body)
	if err != nil {
		return err
	}
	for _, name := range slices.Sorted(maps.Keys(entries)) {
		path := string(BlockSettingsConfigFiles) + "." + name
		entry, ok := entries[name].(map[string]any)
		if !ok {
			return unsupported(path)
		}
		for _, field := range slices.Sorted(maps.Keys(entry)) {
			switch field {
			case "name", "base64":
			case "content":
				if text, isText := entry[field].(string); isText && len(text) > MaxConfigFileBytes {
					return tooLarge(path+".content", len(text), MaxConfigFileBytes)
				}
			case "swarmRef":
				if err := checkSwarmRef(path+".swarmRef", entry[field]); err != nil {
					return err
				}
			default:
				return unsupported(path + "." + field)
			}
		}
	}
	return nil
}

func collectionEntries(block Block, body any) (map[string]any, error) {
	entries, ok := body.(map[string]any)
	if !ok {
		return nil, unsupported(string(block))
	}
	if len(entries) > MaxSettingsPerBlock {
		return nil, hperrors.Wrap(hperrors.ErrSpecBlockUnsupported).WithExtraDetail(
			"%s: at most %d, and this has %d", block, MaxSettingsPerBlock, len(entries))
	}
	return entries, nil
}

// checkSwarmRef allows the file a secret or config is mounted as, and nothing
// else - which is what refuses the swarm ids, since they are fields of the ref
// rather than of the file. The path has to be absolute: docker reads a relative one against the
// container's working directory, which the image chooses and a template does not.
func checkSwarmRef(path string, body any) error {
	ref, ok := body.(map[string]any)
	if !ok {
		return unsupported(path)
	}
	for _, field := range slices.Sorted(maps.Keys(ref)) {
		if field != "file" {
			return unsupported(path + "." + field)
		}
	}
	file, ok := ref["file"].(map[string]any)
	if !ok {
		return unsupported(path + ".file")
	}
	for _, field := range slices.Sorted(maps.Keys(file)) {
		switch field {
		case "uid", "gid", "mode":
		case "name":
			target, _ := file[field].(string)
			if !strings.HasPrefix(target, "/") {
				return hperrors.Wrap(hperrors.ErrSpecBlockUnsupported).WithExtraDetail(
					"%s.file.name %q must be an absolute path", path, target)
			}
		default:
			return unsupported(path + ".file." + field)
		}
	}
	return nil
}

func tooLarge(path string, size, limit int) error {
	return hperrors.Wrap(hperrors.ErrSpecBlockUnsupported).WithExtraDetail(
		"%s is %d bytes, and at most %d is allowed", path, size, limit)
}
