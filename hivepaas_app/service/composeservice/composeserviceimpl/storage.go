package composeserviceimpl

import (
	"maps"
	"path"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/compose-spec/compose-go/v2/types"
	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/unit"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/composeservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/dockerapiservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingmountservice"
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

// findBindRoots gives each directory of the compose file's that services
// mount the outermost one of those it is in - itself, when it is in none.
func (c *converter) findBindRoots() {
	var dirs []string
	for _, name := range c.names {
		for _, v := range c.r.project.Services[name].Volumes {
			if v.Type != types.VolumeTypeBind || path.IsAbs(v.Source) {
				continue
			}
			rel, ok := cleanPath(v.Source)
			if ok && !c.isFile(rel) && !slices.Contains(dirs, rel) {
				dirs = append(dirs, rel)
			}
		}
	}
	c.bindRoots = map[string]string{}
	for _, rel := range dirs {
		root := rel
		for _, other := range dirs {
			if strings.HasPrefix(rel, other+"/") && len(other) < len(root) {
				root = other
			}
		}
		c.bindRoots[rel] = root
	}
}

// isFile says whether a path of the compose file's is a file rather than a
// directory: the request carries it, or its name has an extension - but not
// `.d`, a directory of files - and the request carries no file under it.
func (c *converter) isFile(rel string) bool {
	if _, given := c.r.files[rel]; given {
		return true
	}
	if len(c.filesUnder(rel)) > 0 {
		return false
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
		addMount(mounts, path.Base(v.Target), c.envConfigPath(configName),
			map[string]any{filePart: partContent, detailPath: v.Target})
		c.viewVolume(view, v, v.Source, composeservice.VolumeKindFile, "")
		return
	}
	key, files := "bind:"+rel, c.filesUnder(rel)
	c.need(rel, composeservice.NeedDirectory, name, len(files) > 0)
	if root := c.bindRoots[rel]; root != "" && root != rel {
		// Were it that one's subdirectory, preparing it first would make that
		// one's directory, opened to no user but root: a service not running
		// as root could not write to it. Each is its own, and the review says
		// so.
		c.add(appPath, specmodel.SeverityWarning, composeservice.CodeDirectoryApart,
			map[string]any{detailSource: v.Source, detailTarget: v.Target, "in": root},
			"a directory of its own, not a subdirectory of "+root+"'s: a service mounting that one does not see "+
				"its files")
	}
	owner := c.owners[key]
	if len(files) > 0 && v.ReadOnly && (owner == "" || owner == name) {
		// Read only, it holds nothing the app writes: the files given are the
		// directory - and none could be mounted in a read-only volume, which
		// has no place for them.
		c.mountFiles(rel, v.Target, files, mounts)
		c.viewVolume(view, v, v.Source, composeservice.VolumeKindFiles, "")
		view.Volumes[len(view.Volumes)-1].Files = len(files)
		c.add(appPath, "", composeservice.CodeDirectoryFiles,
			map[string]any{detailSource: v.Source, detailTarget: v.Target, detailFiles: len(files)},
			"the files given, mounted read only, as the directory is")
		return
	}
	c.managedMount(name, key, v, subpathOf(rel), false, st, view, v.Source)
	switch {
	case len(files) == 0:
		c.add(appPath, specmodel.SeverityWarning, composeservice.CodeDirectoryEmpty,
			map[string]any{detailSource: v.Source, detailTarget: v.Target},
			"a directory of its own, which starts empty: give its files - open the compose file's folder - to have "+
				"them in it")
	case v.ReadOnly:
		c.add(appPath, specmodel.SeverityWarning, composeservice.CodeDirectoryEmpty,
			map[string]any{detailSource: v.Source, detailTarget: v.Target, "owner": c.keys[owner]},
			"the files given are not mounted in it: another app writes this directory, and this one reads it as "+
				"that one does")
	default:
		// The directory is the app's, as an empty one is - what the app writes
		// in it is kept - and each file given is mounted at its place in it.
		c.mountFiles(rel, v.Target, files, mounts)
		view.Volumes[len(view.Volumes)-1].Files = len(files)
		c.add(appPath, "", composeservice.CodeDirectoryFiles,
			map[string]any{detailSource: v.Source, detailTarget: v.Target, detailFiles: len(files)},
			"a directory of its own, with the files given mounted in it, read only: what the app writes beside "+
				"them is kept")
	}
}

// mountFiles mounts each file given under a directory of the compose file's at
// its place under target, read only: an env config file each - once, as a
// directory mounted inside another holds some of the same files.
func (c *converter) mountFiles(rel, target string, files []string, mounts map[string]any) {
	for _, file := range files {
		at := path.Join(target, strings.TrimPrefix(file, rel+"/"))
		if mountedAt(mounts, at) {
			continue
		}
		addMount(mounts, path.Base(file), c.envConfigPath(c.configFile(file)),
			map[string]any{filePart: partContent, detailPath: at})
	}
}

// mountedAt says whether an app's setting mounts already put a file at a path.
func mountedAt(mounts map[string]any, at string) bool {
	for _, body := range mounts {
		entry, _ := body.(map[string]any)
		list, _ := entry["files"].([]any)
		for _, f := range list {
			if file, _ := f.(map[string]any); file != nil && file[detailPath] == at {
				return true
			}
		}
	}
	return false
}

// filesUnder are the request's files under a directory of the compose
// file's, in order.
func (c *converter) filesUnder(rel string) []string {
	var out []string
	for _, name := range slices.Sorted(maps.Keys(c.r.files)) {
		if strings.HasPrefix(name, rel+"/") {
			out = append(out, name)
		}
	}
	return out
}

// hostMount is a bind of a host's path: the Docker socket is not one HivePaaS
// gives an app this way, and any other takes a caller who may.
func (c *converter) hostMount(
	appPath string, v types.ServiceVolumeConfig, st *specmodel.Storage, view *composeservice.ServiceView,
) {
	switch {
	case containsString(dockerSockets, v.Source):
		// Never given from a compose file, as from no template: the node's
		// socket is root on the node, and the proxy is not the socket the file
		// means. The operator chooses, once the app is made.
		view.DockerSocket = v.Target
		action := "not mounted: once the app is created, give it the Docker API in its Docker API settings - " +
			"through the proxy, as configured there, or the node's own socket, which takes an administrator " +
			"and privileged apps. The proxy's socket is at $DOCKER_HOST, the node's at " +
			dockerapiservice.HostSocketPath
		if v.Target != dockerapiservice.HostSocketPath {
			action += ", and neither at " + v.Target + ", where the file mounts it"
		}
		c.add(appPath, specmodel.SeverityFixable, composeservice.CodeDockerSocket,
			map[string]any{detailTarget: v.Target, detailSource: v.Source}, action)
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

// configFileOf is the env config file holding a file of the compose file's
// a service mounts, named after it: made once, however many services mount it.
func (c *converter) configFileOf(rel, service string) string {
	_, given := c.r.files[rel]
	c.need(rel, composeservice.NeedBind, service, given)
	return c.configFile(rel)
}

// configFile is the env config file holding a file of the compose file's:
// made once.
func (c *converter) configFile(rel string) string {
	content, given := c.r.files[rel]
	for name, body := range c.envConfigs {
		if fields, _ := body.(map[string]any); fields[fileSourceKey] == rel {
			return name
		}
	}
	name := c.uniqueSettingName(blockConfigFiles, c.envConfigs, path.Base(rel))
	c.envConfigs[name] = c.fileSetting(name, partContent, content, given, rel)
	return name
}

// fileSourceKey marks which file of the compose file's a setting holds, while
// the bundle is made; it is removed before it is written.
const fileSourceKey = "source"

// fileSetting is an env secret's or config file's body: its content - in
// base64 when it is not text - inheritable by the env's apps, and pending
// while the file is missing.
func (c *converter) fileSetting(name, field string, content []byte, given bool, source string) map[string]any {
	setting := map[string]any{detailName: name, "inheritable": true}
	if !given {
		setting["status"] = string(base.SettingStatusPending)
	}
	body := map[string]any{specmodel.SettingMetaKey: setting}
	if field == partContent {
		body[detailName] = name
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

// addMount adds a setting mount entry to an app's: one file of one source.
// Each mount is an entry of its own - an entry gives a part of its source
// once - keyed after hint as an entry key has to be: the app's entries with a
// key that is not one are never mounted (settingmountservice.ValidEntryKey).
func addMount(mounts map[string]any, hint, source string, file map[string]any) {
	mounts[uniqueEntryKey(mounts, hint)] = map[string]any{
		detailSource: map[string]any{"id": source}, "files": []any{file},
	}
}

// uniqueEntryKey is the entry key hint makes, numbered when the app has one
// by that key already.
func uniqueEntryKey(mounts map[string]any, hint string) string {
	key := settingmountservice.EntryKeyFor(hint)
	if key == "" {
		key = "file"
	}
	candidate := key
	for i := 2; ; i++ {
		if _, taken := mounts[candidate]; !taken {
			return candidate
		}
		suffix := "-" + itoa(i)
		candidate = strings.TrimRight(key[:min(len(key), entryKeyMaxLen-len(suffix))], "-") + suffix
	}
}

// entryKeyMaxLen is how long an entry key may be.
const entryKeyMaxLen = 20

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

// writableForFiles makes writable a read-only mount a file of the app's is
// mounted in: Docker has no place to make for the file in a read-only one,
// and the app would not start.
func (c *converter) writableForFiles(
	appPath string, st *specmodel.Storage, mounts map[string]any, view *composeservice.ServiceView,
) {
	if st == nil || len(mounts) == 0 {
		return
	}
	var files []string
	for _, body := range mounts {
		entry, _ := body.(map[string]any)
		list, _ := entry["files"].([]any)
		for _, f := range list {
			if file, _ := f.(map[string]any); file != nil {
				if p, ok := file[detailPath].(string); ok {
					files = append(files, p)
				}
			}
		}
	}
	holds := func(target string) bool {
		prefix := strings.TrimSuffix(target, "/") + "/"
		return slices.ContainsFunc(files, func(file string) bool { return strings.HasPrefix(file, prefix) })
	}
	for _, block := range []map[string]specmodel.Mount{st.Mounts, st.DockerMounts} {
		for _, target := range slices.Sorted(maps.Keys(block)) {
			m := block[target]
			if !m.ReadOnly || !holds(target) {
				continue
			}
			m.ReadOnly = false
			block[target] = m
			for _, volume := range view.Volumes {
				if volume.Target == target {
					volume.ReadOnly = false
				}
			}
			c.add(appPath, specmodel.SeverityWarning, composeservice.CodeMountWritable,
				map[string]any{detailTarget: target},
				"mounted writable: a file is mounted in it, which Docker cannot do in a read-only one")
		}
	}
}
