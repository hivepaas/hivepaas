package composeserviceimpl

import (
	"path"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// dockerSockets are where a compose file mounts the Docker socket from.
var dockerSockets = []string{"/var/run/docker.sock", "/run/docker.sock"}

// findVolumeOwners gives each volume several services mount the service whose
// directory it is: the first, in depends_on order, that writes to it - or
// reads it, when none does. Every other mounts the owner's.
func (c *converter) findVolumeOwners() {
	c.owners = map[string]string{}
	readers := map[string]string{}
	for _, name := range c.names {
		for _, v := range c.r.project.Services[name].Volumes {
			key, shared := c.volumeKey(v)
			if !shared {
				continue
			}
			if _, owned := c.owners[key]; !owned && !v.ReadOnly {
				c.owners[key] = name
			}
			if _, read := readers[key]; !read {
				readers[key] = name
			}
		}
	}
	for key, name := range readers {
		if _, owned := c.owners[key]; !owned {
			c.owners[key] = name
		}
	}
}

// volumeKey names what a mount reaches, as several services may share it: a
// named volume, or a directory of the compose file's. false for what is the
// service's alone.
func (c *converter) volumeKey(v types.ServiceVolumeConfig) (string, bool) {
	switch v.Type {
	case types.VolumeTypeVolume:
		return "volume:" + v.Source, v.Source != ""
	case types.VolumeTypeBind:
		rel, ok := cleanPath(v.Source)
		if !ok || c.isFile(rel) {
			return "", false
		}
		return "bind:" + rel, true
	}
	return "", false
}

// isFile says whether a path of the compose file's is a file rather than a
// directory: the request carries it, or its name has an extension - but not
// `.d`, a directory of files.
func (c *converter) isFile(rel string) bool {
	if _, given := c.r.files[rel]; given {
		return true
	}
	ext := path.Ext(strings.TrimPrefix(path.Base(rel), "."))
	return ext != "" && ext != ".d"
}

// storage is a service's mounts: its volumes and directories on the project's
// volume, its files from env settings, its tmpfs.
func (c *converter) storage(
	appPath, name string, svc types.ServiceConfig, view *composeservice.ServiceView, mounts map[string]any,
) *specmodel.Storage {
	st := &specmodel.Storage{Mounts: map[string]specmodel.Mount{}, DockerMounts: map[string]specmodel.Mount{}}
	for _, v := range svc.Volumes {
		switch v.Type {
		case types.VolumeTypeVolume:
			c.volumeMount(appPath, name, v, st, view)
		case types.VolumeTypeBind:
			c.bindMount(appPath, name, v, st, view, mounts)
		case types.VolumeTypeTmpfs:
			st.DockerMounts[v.Target] = tmpfsMount(v.Tmpfs)
			c.viewVolume(view, v, v.Target, composeservice.VolumeKindTmpfs, "")
		default:
			c.add(appPath, specmodel.SeverityFixable, composeservice.CodeMountDropped,
				map[string]any{detailTarget: v.Target, "type": v.Type}, "not mounted: HivePaaS mounts no such volume")
			c.viewVolume(view, v, v.Source, composeservice.VolumeKindDropped, "")
		}
	}
	for _, entry := range svc.Tmpfs {
		target, _, _ := strings.Cut(entry, ":")
		if path.IsAbs(target) {
			st.DockerMounts[target] = tmpfsMount(nil)
			view.Volumes = append(view.Volumes, &composeservice.VolumeView{Target: target, Kind: composeservice.VolumeKindTmpfs})
		}
	}
	if len(st.Mounts) == 0 && len(st.DockerMounts) == 0 {
		return nil
	}
	return st
}

func (c *converter) viewVolume(
	view *composeservice.ServiceView, v types.ServiceVolumeConfig, source string, kind composeservice.VolumeKind,
	owner string,
) {
	view.Volumes = append(view.Volumes, &composeservice.VolumeView{
		Target: v.Target, Source: source, Kind: kind, ReadOnly: v.ReadOnly, Owner: owner,
	})
}

// volumeMount is a named volume - or an anonymous one, the service's alone -
// as a directory on the project's volume: the app's own, or the owner's.
func (c *converter) volumeMount(
	appPath, name string, v types.ServiceVolumeConfig, st *specmodel.Storage, view *composeservice.ServiceView,
) {
	volume, key := v.Source, "volume:"+v.Source
	if volume == "" {
		volume, key = "anonymous"+strings.ReplaceAll(v.Target, "/", "-"), ""
	}
	subpath := subpathOf(volume)
	if v.Volume != nil && v.Volume.Subpath != "" {
		if sub, ok := cleanPath(v.Volume.Subpath); ok {
			subpath += "/" + sub
		}
	}
	noCopy := v.Volume != nil && v.Volume.NoCopy
	c.managedMount(name, key, v, subpath, noCopy, st, view, volume)

	if declared, ok := c.r.project.Volumes[volume]; ok && key != "" {
		if bool(declared.External) {
			c.add(appPath, specmodel.SeverityWarning, composeservice.CodeVolumeExternal,
				map[string]any{"volume": volume}, "a new directory: the volume the file names is not here")
		}
		if declared.Driver != "" && declared.Driver != "local" || len(declared.DriverOpts) > 0 {
			c.add(appPath, specmodel.SeverityWarning, composeservice.CodeVolumeDriver,
				map[string]any{"volume": volume, "driver": declared.Driver},
				"on the project's volume: its driver is not used - a cluster volume with it is the way")
		}
	}
}

// managedMount mounts a directory of the project's volume: the app's own, or
// the owner's when the volume is shared and this app is not its owner.
func (c *converter) managedMount(
	name, key string, v types.ServiceVolumeConfig, subpath string, noCopy bool,
	st *specmodel.Storage, view *composeservice.ServiceView, source string,
) {
	m := specmodel.Mount{Type: mount.TypeVolume, ReadOnly: v.ReadOnly,
		VolumeOptions: &specmodel.VolumeOptions{Subpath: subpath, NoCopy: noCopy}}
	if c.req.Volume != nil {
		external := *c.req.Volume
		m.External = &external
	} else {
		m.Source = "projects/" + c.req.ProjectKey + "/volumes/default"
	}
	kind, owner := composeservice.VolumeKindVolume, ""
	if ownerName := c.owners[key]; key != "" && ownerName != "" && ownerName != name {
		m.SourceApp = &specmodel.MountSourceApp{App: c.keys[ownerName], Write: !v.ReadOnly}
		kind, owner = composeservice.VolumeKindShared, c.keys[ownerName]
	}
	st.Mounts[v.Target] = m
	c.viewVolume(view, v, source, kind, owner)
}

// subpathPattern is what a directory name keeps of a volume's or a path's.
var subpathPattern = regexp.MustCompile(`[^A-Za-z0-9_.-]+`)

// subpathOf is the directory a volume or a path of the compose file's is, in
// the app's directory: one segment, nothing a path could escape with.
func subpathOf(name string) string {
	out := strings.Trim(subpathPattern.ReplaceAllString(strings.ReplaceAll(name, "/", "-"), "-"), "-.")
	if out == "" {
		return "data"
	}
	return out
}

// bindMount is a bind of the compose file's: a file it reads as an env config
// file, a directory of it as a directory on the project's volume, a host's
// path as it is - for a caller who may.
func (c *converter) bindMount(
	appPath, name string, v types.ServiceVolumeConfig, st *specmodel.Storage, view *composeservice.ServiceView,
	mounts map[string]any,
) {
	if path.IsAbs(v.Source) {
		c.hostMount(appPath, v, st, view)
		return
	}
	rel, ok := cleanPath(v.Source)
	if !ok {
		c.add(appPath, specmodel.SeverityFixable, composeservice.CodeMountDropped,
			map[string]any{detailTarget: v.Target, detailSource: v.Source},
			"not mounted: it leaves the compose file's directory")
		c.viewVolume(view, v, v.Source, composeservice.VolumeKindDropped, "")
		return
	}
	if c.isFile(rel) {
		configName := c.configFileOf(rel, name)
		if _, given := c.r.files[rel]; !given {
			c.add(appPath, specmodel.SeverityFixable, composeservice.CodeFileMissing,
				map[string]any{detailPath: rel, detailTarget: v.Target}, "mounted empty until it is filled")
		}
		entry := mountEntry(mounts, "file"+strings.ReplaceAll(v.Target, "/", "-"), c.envConfigPath(configName))
		addMountFile(entry, map[string]any{"part": "content", detailPath: v.Target})
		c.viewVolume(view, v, v.Source, composeservice.VolumeKindFile, "")
		return
	}
	c.managedMount(name, "bind:"+rel, v, subpathOf(rel), false, st, view, v.Source)
	c.add(appPath, specmodel.SeverityWarning, composeservice.CodeDirectoryEmpty,
		map[string]any{detailSource: v.Source, detailTarget: v.Target},
		"a directory of its own, which starts empty: the files of the compose file's are not copied")
}

// hostMount is a bind of a host's path: the Docker socket is not one HivePaaS
// gives an app this way, and any other takes a caller who may.
func (c *converter) hostMount(
	appPath string, v types.ServiceVolumeConfig, st *specmodel.Storage, view *composeservice.ServiceView,
) {
	switch {
	case containsString(dockerSockets, v.Source):
		c.add(appPath, specmodel.SeverityFixable, composeservice.CodeMountDropped,
			map[string]any{detailTarget: v.Target, detailSource: v.Source},
			"not mounted: give the app Docker API access in its settings instead")
	case !c.req.MayBindHost:
		c.add(appPath, specmodel.SeverityFixable, composeservice.CodeMountDropped,
			map[string]any{detailTarget: v.Target, detailSource: v.Source},
			"not mounted: a host's directory takes privileged apps and an administrator")
	default:
		m := specmodel.Mount{Type: mount.TypeBind, Source: v.Source, ReadOnly: v.ReadOnly}
		if v.Bind != nil && v.Bind.Propagation != "" {
			m.BindOptions = &specmodel.BindOptions{Propagation: mount.Propagation(v.Bind.Propagation)}
		}
		st.DockerMounts[v.Target] = m
		c.viewVolume(view, v, v.Source, composeservice.VolumeKindHost, "")
		return
	}
	c.viewVolume(view, v, v.Source, composeservice.VolumeKindDropped, "")
}

func tmpfsMount(t *types.ServiceVolumeTmpfs) specmodel.Mount {
	m := specmodel.Mount{Type: mount.TypeTmpfs}
	if t != nil && (t.Size > 0 || t.Mode != 0) {
		m.TmpfsOptions = &specmodel.TmpfsOptions{Size: unit.DataSize(t.Size), Mode: fileutil.FileMode(t.Mode)}
	}
	return m
}

// configFileOf is the env config file holding a file of the compose file's,
// named after it: made once, however many services mount it.
func (c *converter) configFileOf(rel, service string) string {
	content, given := c.r.files[rel]
	c.need(rel, "bind", service, given)
	for name, body := range c.envConfigs {
		if fields, _ := body.(map[string]any); fields[fileSourceKey] == rel {
			return name
		}
	}
	name := c.uniqueConfigName(path.Base(rel))
	c.envConfigs[name] = c.fileSetting(name, "content", content, given, rel)
	return name
}

// fileSourceKey marks which file of the compose file's a setting holds, while
// the bundle is made; it is removed before it is written.
const fileSourceKey = "source"

func (c *converter) uniqueConfigName(base string) string {
	name := base
	for i := 2; ; i++ {
		if _, taken := c.envConfigs[name]; !taken {
			return name
		}
		name = base + "-" + itoa(i)
	}
}

// fileSetting is an env secret's or config file's body: its content - in
// base64 when it is not text - inheritable by the env's apps, and pending
// while the file is missing.
func (c *converter) fileSetting(name, field string, content []byte, given bool, source string) map[string]any {
	setting := map[string]any{"name": name, "inheritable": true}
	if !given {
		setting["status"] = string(base.SettingStatusPending)
	}
	body := map[string]any{specmodel.SettingMetaKey: setting}
	if field == "content" {
		body["name"] = name
	} else {
		body["key"] = name
	}
	if utf8.Valid(content) {
		body[field] = string(content)
	} else {
		body[field], body["base64"] = encodeBase64(content), true
	}
	if source != "" {
		body[fileSourceKey] = source
	}
	return body
}

func (c *converter) envConfigPath(name string) string {
	return c.envPath() + "/configFiles/" + name
}

func (c *converter) envSecretPath(name string) string {
	return c.envPath() + "/secrets/" + name
}

// mountEntry is a setting mount entry of an app's, from one source: made once.
func mountEntry(mounts map[string]any, key, source string) map[string]any {
	if entry, ok := mounts[key].(map[string]any); ok {
		return entry
	}
	entry := map[string]any{detailSource: map[string]any{"id": source}, "files": []any{}}
	mounts[key] = entry
	return entry
}

func addMountFile(entry map[string]any, file map[string]any) {
	files, _ := entry["files"].([]any)
	entry["files"] = append(files, file)
}

func (c *converter) need(rel, as, service string, given bool) {
	need := c.needs[rel]
	if need == nil {
		need = &composeservice.FileNeed{Path: rel, As: as, Given: given}
		c.needs[rel] = need
	}
	if service != "" && !containsString(need.By, service) {
		need.By = append(need.By, service)
	}
}
