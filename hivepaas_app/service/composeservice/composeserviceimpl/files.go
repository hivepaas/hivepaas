package composeserviceimpl

import (
	"encoding/base64"
	"maps"
	"path"
	"slices"
	"strconv"

	"github.com/compose-spec/compose-go/v2/types"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// secretsDir is where compose mounts a secret given no target.
const secretsDir = "/run/secrets"

// fileModeMax is the largest mode a mounted file may have: its permission
// bits, setuid, setgid and sticky.
const fileModeMax = 0o7777

// fileObjects are the file's secrets and configs the services mount: the
// env's own secrets and config files, by their names in the file.
func (c *converter) fileObjects() {
	secrets, configs := map[string][]string{}, map[string][]string{}
	for _, name := range c.names {
		svc := c.r.project.Services[name]
		for _, ref := range svc.Secrets {
			secrets[ref.Source] = append(secrets[ref.Source], name)
		}
		for _, ref := range svc.Configs {
			configs[ref.Source] = append(configs[ref.Source], name)
		}
	}
	c.secretNames, c.configNames = map[string]string{}, map[string]string{}
	for _, name := range slices.Sorted(maps.Keys(secrets)) {
		object := types.FileObjectConfig(c.r.project.Secrets[name])
		content, given := c.fileObject(composeservice.NeedSecret, object, secrets[name])
		setting := c.uniqueSettingName(blockSecrets, c.envSecrets, name)
		c.secretNames[name] = setting
		c.envSecrets[setting] = c.fileSetting(setting, partValue, content, given, "")
	}
	for _, name := range slices.Sorted(maps.Keys(configs)) {
		object := types.FileObjectConfig(c.r.project.Configs[name])
		content, given := c.fileObject(composeservice.NeedConfig, object, configs[name])
		setting := c.uniqueSettingName(blockConfigFiles, c.envConfigs, name)
		c.configNames[name] = setting
		c.envConfigs[setting] = c.fileSetting(setting, partContent, content, given, "")
	}
}

// uniqueSettingName is a name for a new env setting of a block: base, or
// base-2 and on when the env being made or the existing one has it. A setting
// the existing env has is never what a file mounts: what is mounted is what
// the request carries.
func (c *converter) uniqueSettingName(block string, made map[string]any, base string) string {
	name := base
	for i := 2; ; i++ {
		if _, taken := made[name]; !taken && !c.existingSetting(block, name) {
			break
		}
		name = base + "-" + itoa(i)
	}
	if name != base && c.existingSetting(block, base) {
		c.add(c.envPath(), "", composeservice.CodeSettingRenamed,
			map[string]any{"block": block, detailName: base, "as": name},
			"named otherwise: the env has a setting by this name, which is left as it is")
	}
	return name
}

// fileObject is a secret's or a config's content: its file, its inline
// content, or a variable's value. One the file holds elsewhere - external, or
// a file not given - is empty, and pending until it is filled.
func (c *converter) fileObject(as string, object types.FileObjectConfig, services []string) ([]byte, bool) {
	switch {
	case object.File != "":
		rel, ok := cleanPath(object.File)
		if !ok {
			c.add(c.envPath(), specmodel.SeverityFixable, composeservice.CodeFileMissing,
				map[string]any{detailPath: object.File, as: object.Name}, "empty: it leaves the compose file's directory")
			return nil, false
		}
		content, given := c.r.files[rel]
		for _, service := range services {
			c.need(rel, as, service, given)
		}
		if !given {
			c.add(c.envPath(), specmodel.SeverityFixable, composeservice.CodeFileMissing,
				map[string]any{detailPath: rel, as: object.Name}, "created empty, and pending until it is filled")
		}
		return content, given
	case object.Content != "":
		return []byte(c.plain(c.envPath(), as, object.Content)), true
	case object.Environment != "":
		value, given := c.r.values[object.Environment]
		return []byte(value), given
	}
	c.add(c.envPath(), specmodel.SeverityFixable, composeservice.CodeFileMissing,
		map[string]any{as: object.Name, "external": bool(object.External)},
		"created empty, and pending until it is filled: the file keeps it elsewhere")
	return nil, false
}

// fileMounts mount the secrets and configs a service names, where compose
// would: a secret under /run/secrets, a config at the root.
func (c *converter) fileMounts(svc types.ServiceConfig, mounts map[string]any) {
	for _, ref := range svc.Secrets {
		target := ref.Target
		switch {
		case target == "":
			target = path.Join(secretsDir, ref.Source)
		case !path.IsAbs(target):
			target = path.Join(secretsDir, target)
		}
		addMount(mounts, "secret-"+ref.Source, c.envSecretPath(c.secretNames[ref.Source]),
			mountFile(partValue, target, types.FileReferenceConfig(ref)))
	}
	for _, ref := range svc.Configs {
		target := ref.Target
		if target == "" {
			target = "/" + ref.Source
		}
		addMount(mounts, "config-"+ref.Source, c.envConfigPath(c.configNames[ref.Source]),
			mountFile(partContent, target, types.FileReferenceConfig(ref)))
	}
}

func mountFile(part, target string, ref types.FileReferenceConfig) map[string]any {
	file := map[string]any{filePart: part, detailPath: target}
	if ref.UID != "" {
		file["uid"] = ref.UID
	}
	if ref.GID != "" {
		file["gid"] = ref.GID
	}
	if ref.Mode != nil && *ref.Mode >= 0 && *ref.Mode <= fileModeMax {
		file["mode"] = fileutil.FileMode(*ref.Mode).String() //nolint:gosec // bounded above
	}
	return file
}

func encodeBase64(content []byte) string {
	return base64.StdEncoding.EncodeToString(content)
}

func itoa(i int) string {
	return strconv.Itoa(i)
}

func containsString(list []string, s string) bool {
	return slices.Contains(list, s)
}
