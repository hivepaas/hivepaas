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
	BlockSettingsRouting      Block = "settings.routing"
)

// BuildableBlocks is every block specservice.BuildApp builds, in the order it
// builds them. specserviceimpl's builder registry is tested against this list.
//
// TODO: app templates phase 3 - more blocks: repository builds, registry auth,
// bind mounts, config files, secrets, scheduled jobs, routing domains. See
// docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
var BuildableBlocks = []Block{BlockDeploymentSource, BlockDeploymentStorage, BlockContainerHealthcheck,
	BlockDeploymentResources, BlockSettingsKind, BlockSettingsEnvVars, BlockSettingsRouting}

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

	for _, key := range slices.Sorted(maps.Keys(settings)) {
		switch key {
		case kind, envVars:
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
