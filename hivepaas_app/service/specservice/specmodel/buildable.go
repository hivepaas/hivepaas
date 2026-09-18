package specmodel

import (
	"maps"
	"reflect"
	"slices"
	"strings"

	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// Block names a part of an AppDoc that can be built into an app.
type Block string

const (
	BlockDeploymentSource     Block = "deployment.source"
	BlockDeploymentStorage    Block = "deployment.storage"
	BlockContainerHealthcheck Block = "deployment.container.healthcheck"
	BlockDeploymentResources  Block = "deployment.resources"
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
	// MaxSecretValueBytes is what a secret may hold: a password, a key, a
	// certificate - never a file somebody meant to mount.
	MaxSecretValueBytes = 64 << 10
	// MaxConfigFileBytes is what one config file may hold. Docker's own limit is
	// 500KB; a configuration file a template writes is far smaller than that, and
	// every byte here is copied into the swarm and into every task.
	MaxConfigFileBytes = 128 << 10
)

// BuildableBlocks is every block specservice.BuildApp builds, in the order it
// builds them. specserviceimpl's builder registry is tested against this list.
//
// TODO: app templates phase 3 - more blocks: repository builds, registry auth,
// bind mounts, config files, secrets, scheduled jobs, routing domains. See
// docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
var BuildableBlocks = []Block{BlockDeploymentSource, BlockDeploymentStorage, BlockContainerHealthcheck,
	BlockDeploymentResources, BlockSettingsKind, BlockSettingsEnvVars, BlockSettingsSecrets,
	BlockSettingsConfigFiles, BlockSettingsRouting}

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
		if d.Resources != nil {
			blocks = append(blocks, BlockDeploymentResources)
		}
	}
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
	// A collection block with no entries builds nothing, so it is not present.
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
	slices.SortFunc(blocks, func(a, b Block) int {
		return slices.Index(BuildableBlocks, a) - slices.Index(BuildableBlocks, b)
	})
	return blocks
}

func checkDeployment(d *Deployment) error {
	if d == nil {
		return nil
	}
	if err := onlyFields("deployment.", d, "source", "container", "resources", "storage"); err != nil {
		return err
	}
	if err := checkSource(d.Source); err != nil {
		return err
	}
	if d.Container != nil {
		if err := onlyFields("deployment.container.", d.Container, "healthcheck"); err != nil {
			return err
		}
	}
	if err := checkResources(d.Resources); err != nil {
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
	if err := onlyFields("deployment.resources.", r, "reservations", "limits"); err != nil {
		return err
	}
	if r.Reservations != nil {
		if err := onlyFields("deployment.resources.reservations.", r.Reservations, "cpus", "memory"); err != nil {
			return err
		}
	}
	if r.Limits != nil {
		return onlyFields("deployment.resources.limits.", r.Limits, "cpus", "memory", "pids")
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
		if err := onlyFields(path, &m, "type", "source", "readOnly", "volumeOptions"); err != nil {
			return err
		}
		if m.VolumeOptions != nil {
			if err := onlyFields(path+"volumeOptions.", m.VolumeOptions, "subpath"); err != nil {
				return err
			}
		}
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
			body, ok := settings[key].(map[string]any)
			if !ok {
				return unsupported("settings." + key)
			}
			for _, field := range slices.Sorted(maps.Keys(body)) {
				if field != "port" {
					return unsupported("settings." + key + "." + field)
				}
			}
		default:
			return unsupported("settings." + key)
		}
	}
	return nil
}

// onlyFields refuses the first non-zero field of a struct not named in allowed,
// naming it by its yaml key under prefix.
func onlyFields(prefix string, value any, allowed ...string) error {
	v := reflect.Indirect(reflect.ValueOf(value))
	if !v.IsValid() || v.Kind() != reflect.Struct {
		return nil
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
		return unsupported(prefix + name)
	}
	return nil
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
