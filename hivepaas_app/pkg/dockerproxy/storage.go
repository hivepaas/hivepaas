package dockerproxy

import (
	"context"
	"net/url"
	"path"
	"slices"
	"strings"
)

const (
	mountTypeVolume = "volume"
	mountTypeBind   = "bind"
	mountTypeTmpfs  = "tmpfs"

	// A bind is source:target, or source:target:modes.
	bindMinParts = 2
	bindMaxParts = 3
)

var (
	// bindModes are the modes a child may give a bind. Relabeling (z, Z) and
	// propagation act on the host's side of the mount.
	bindModes          = []string{"ro", "rw", "nocopy"}
	mountFields        = []string{"Type", "Source", "Target", "ReadOnly", "Consistency", "VolumeOptions", "TmpfsOptions"}
	volumeOptionFields = []string{"NoCopy", fieldLabels, "Subpath"}
)

// taskMount is one mount of the app's own task, as its container reports it.
type taskMount struct {
	Type          string
	Source        string
	Target        string
	VolumeOptions *struct {
		Subpath string
	}
}

// rewriteStorage turns every bind and mount of a child into one the app may give
// it, or refuses. A path becomes a mount of the directory the app itself has
// there; a named volume must be the app's.
func (p *Proxy) rewriteStorage(ctx context.Context, policy *Policy, host map[string]any) error {
	var mounts []any
	for _, raw := range list(host["Mounts"]) {
		mount, err := p.mount(ctx, policy, object(raw))
		if err != nil {
			return err
		}
		mounts = append(mounts, mount)
	}
	var binds []any
	for _, raw := range list(host["Binds"]) {
		spec := text(raw)
		mount, err := p.bind(ctx, policy, spec)
		if err != nil {
			return err
		}
		if mount == nil {
			binds = append(binds, spec)
		} else {
			mounts = append(mounts, mount)
		}
	}
	host["Binds"] = binds
	host["Mounts"] = mounts
	return nil
}

// bind judges one entry of Binds. A named volume stays a bind; a path comes back
// as the mount that replaces it.
func (p *Proxy) bind(ctx context.Context, policy *Policy, spec string) (map[string]any, error) {
	parts := strings.Split(spec, ":")
	if len(parts) < bindMinParts || len(parts) > bindMaxParts {
		return nil, refusef("bind %q is not source:target[:modes]", spec)
	}
	source, target := parts[0], parts[1]
	readOnly := false
	if len(parts) == bindMaxParts {
		for _, mode := range strings.Split(parts[2], ",") {
			if !slices.Contains(bindModes, mode) {
				return nil, refusef("bind mode %s is not allowed", mode)
			}
			readOnly = readOnly || mode == "ro"
		}
	}
	if strings.HasPrefix(source, "/") {
		return p.hostPath(ctx, policy, source, target, readOnly)
	}
	return nil, p.namedVolume(ctx, policy, source)
}

// mount judges one entry of Mounts, and returns what replaces it.
func (p *Proxy) mount(ctx context.Context, policy *Policy, m map[string]any) (map[string]any, error) {
	if err := checkFields("Mount", m, mountFields, nil); err != nil {
		return nil, err
	}
	switch kind := text(m["Type"]); kind {
	case mountTypeTmpfs:
		return m, nil
	case mountTypeVolume:
		options := object(m["VolumeOptions"])
		if err := checkFields("Mount.VolumeOptions", options, volumeOptionFields, nil); err != nil {
			return nil, err
		}
		if subpath := text(options["Subpath"]); subpath != "" && !localPath(subpath) {
			return nil, refusef("volume subpath %s leaves the volume", subpath)
		}
		if source := text(m["Source"]); source != "" {
			if err := p.namedVolume(ctx, policy, source); err != nil {
				return nil, err
			}
		}
		return m, nil
	case mountTypeBind:
		readOnly, _ := m["ReadOnly"].(bool)
		return p.hostPath(ctx, policy, text(m["Source"]), text(m["Target"]), readOnly)
	default:
		return nil, refusef("mount type %s is not allowed", kind)
	}
}

// hostPath turns a bind of a path into a mount of the directory the app itself
// has at that path, or of the app's socket. Nothing else of the host is
// reachable.
func (p *Proxy) hostPath(
	ctx context.Context, policy *Policy, source, target string, readOnly bool,
) (map[string]any, error) {
	clean := path.Clean(source)
	if clean == SocketPath {
		if !policy.allows(GroupNestedSocket) {
			return nil, refusef("mounting the app's socket is not allowed for this app")
		}
		return volumeMount(policy.SocketVolume, SocketFile, target, readOnly), nil
	}
	dir := longestCover(policy.SharedDirs, clean)
	if dir == "" {
		return nil, refusef("%s is not a shared directory of this app", source)
	}
	mounts, err := p.taskMounts(ctx, policy)
	if err != nil {
		return nil, err
	}
	var best *taskMount
	for i := range mounts {
		m := &mounts[i]
		if m.Type == mountTypeVolume && covers(m.Target, dir) && (best == nil || len(m.Target) > len(best.Target)) {
			best = m
		}
	}
	if best == nil {
		return nil, refusef("shared directory %s is not on a volume of the app", dir)
	}
	subpath := ""
	if best.VolumeOptions != nil {
		subpath = best.VolumeOptions.Subpath
	}
	rest := strings.TrimPrefix(clean, strings.TrimSuffix(best.Target, "/"))
	return volumeMount(best.Source, strings.TrimPrefix(path.Join(subpath, rest), "/"), target, readOnly), nil
}

// taskMounts are the mounts of the app's own task on this node. They are read
// from the task's container rather than from the service because a worker node
// cannot read services, and the child is being created on this node, beside
// the task.
func (p *Proxy) taskMounts(ctx context.Context, policy *Policy) ([]taskMount, error) {
	var tasks []struct {
		ID     string `json:"Id"`
		Labels map[string]string
	}
	query := "/containers/json?filters=" + labelFilter(serviceIDLabel, policy.ServiceID)
	if _, err := p.daemon.get(ctx, query, &tasks); err != nil {
		return nil, err
	}
	for _, task := range tasks {
		// A container carrying the owner label is a child, whatever else it says.
		if task.Labels[OwnerLabel] != "" {
			continue
		}
		var info struct {
			HostConfig struct {
				Mounts []taskMount
			}
		}
		found, err := p.daemon.get(ctx, "/containers/"+url.PathEscape(task.ID)+"/json", &info)
		if err != nil {
			return nil, err
		}
		if found {
			return info.HostConfig.Mounts, nil
		}
	}
	return nil, refusef("the app is not running on this node")
}

// namedVolume checks a volume a child names: the app's own, created for it when
// it does not exist yet, since act names volumes in Binds and never creates
// them.
func (p *Proxy) namedVolume(ctx context.Context, policy *Policy, name string) error {
	if name == "" {
		return refusef("a volume must be named")
	}
	if name == policy.SocketVolume {
		if !policy.allows(GroupNestedSocket) {
			return refusef("mounting the app's socket is not allowed for this app")
		}
		return nil
	}
	// Any other reserved name is another app's socket or network, and naming one
	// the node does not have yet would otherwise create it for this app.
	if err := refuseReserved(policy, name); err != nil {
		return err
	}
	if !policy.allows(GroupVolumes) {
		return refusef("%s is not allowed for this app", GroupVolumes)
	}
	info, found, err := p.inspect(ctx, kindVolumes, name)
	if err != nil {
		return err
	}
	if !found {
		return p.daemon.createVolume(ctx, name, map[string]string{OwnerLabel: policy.AppID})
	}
	if info.Labels[OwnerLabel] != policy.AppID {
		return refusef("volume %s is not one this app created", name)
	}
	return nil
}

// volumeMount is a mount of a directory of a volume. NoCopy keeps docker from
// copying what an image holds at the target into the app's directory.
func volumeMount(source, subpath, target string, readOnly bool) map[string]any {
	options := map[string]any{"NoCopy": true}
	if subpath != "" {
		options["Subpath"] = subpath
	}
	return map[string]any{
		"Type": mountTypeVolume, "Source": source, "Target": target, "ReadOnly": readOnly,
		"VolumeOptions": options,
	}
}

// covers reports whether p is dir or lies below it.
func covers(dir, p string) bool {
	if dir == "" {
		return false
	}
	return p == dir || strings.HasPrefix(p, strings.TrimSuffix(dir, "/")+"/")
}

func longestCover(dirs []string, p string) string {
	best := ""
	for _, dir := range dirs {
		if covers(dir, p) && len(dir) > len(best) {
			best = dir
		}
	}
	return best
}

// localPath reports whether a subpath stays inside its volume.
func localPath(p string) bool {
	clean := path.Clean(p)
	return !path.IsAbs(clean) && clean != ".." && !strings.HasPrefix(clean, "../")
}
