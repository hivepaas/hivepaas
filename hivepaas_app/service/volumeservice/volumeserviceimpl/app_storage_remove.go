package volumeserviceimpl

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/placementservice"
)

// volumeHelperTarget is where the helper container sees the storage, whole.
const volumeHelperTarget = "/mnt/vol"

// storageTarget is one directory to delete: the mount that reaches the storage
// it is in, the path of the directory inside that mount, and the volume setting
// that storage belongs to - which is what says whose directory it is.
type storageTarget struct {
	mount   mount.Mount
	subpath string
	volume  *entity.Setting
}

func (s *service) RemoveAppStorage(
	ctx context.Context,
	db database.IDB,
	app *entity.App,
	mounts []mount.Mount,
) error {
	if len(mounts) == 0 {
		return nil
	}

	// The volumes the app was allowed to mount, which is what a mount has to be
	// matched against to learn which node its data is on - and, for a bind, to
	// learn that it is a volume's directory at all.
	volumes, _, err := s.settingRepo.List(ctx, db, app.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterVolume),
	)
	if err != nil {
		return hperrors.Wrap(err)
	}

	pins, err := placementservice.VolumePinsForMounts(mounts, volumes)
	if err != nil {
		return hperrors.Wrap(err)
	}
	constraint, conflict := placementservice.VolumePinConstraint(pins)
	if conflict != nil {
		// The app's mounts name two nodes, so there is no answer to which node
		// holds the data. Deleting on the wrong one removes nothing at best, and
		// at worst a directory of the same name belonging to something else.
		return hperrors.Wrap(hperrors.ErrActionFailed).WithMsgLog(
			"cannot remove the storage of app %s: %s", app.ID, conflict.Error())
	}
	local := s.storageIsOnThisNode(ctx, pins)

	targets := ownStorageTargets(app, mounts, volumes)
	if len(targets) == 0 {
		// Nothing matched, and the caller asked for data to be deleted: saying so
		// is the difference between a storage removal that had nothing to do and
		// one that quietly did nothing. The second is what a mount whose source
		// docker reports differently than HivePaaS wrote it looks like.
		if s.logger != nil {
			s.logger.Warn("no storage of app was removed: none of its mounts matched a volume it owns",
				"app", app.ID, "mounts", len(mounts), "volumes", len(volumes),
				"sources", mountSources(mounts))
		}
		return nil
	}
	for _, target := range targets {
		if err := s.removeStorageTarget(ctx, &target, constraint, local); err != nil {
			return hperrors.Wrap(err)
		}
	}
	return nil
}

// mountSources is what the warning above prints: enough to compare against the
// volumes by hand, and nothing a path could hide in.
func mountSources(mounts []mount.Mount) []string {
	sources := make([]string, 0, len(mounts))
	for i := range mounts {
		sources = append(sources, string(mounts[i].Type)+":"+mounts[i].Source)
	}
	return sources
}

// ownStorageTargets is what deleting this app may remove: the directories its
// mounts reach, minus the ones belonging to somebody else.
//
// An app may have been given a directory of another app - a file manager over
// the database beside it. Seeing it is one thing; taking it along when this app
// is deleted is another. The test is the path rather than any marker on the
// mount, so it holds for mounts written before such a thing was possible, for a
// service spec somebody restored by hand, and for one edited outside HivePaaS.
func ownStorageTargets(app *entity.App, mounts []mount.Mount, volumes []*entity.Setting) []storageTarget {
	targets := make([]storageTarget, 0, len(mounts))
	for i := range mounts {
		target, ok := appStorageTarget(&mounts[i], volumes)
		if !ok {
			continue
		}
		if !appOwnsSubpath(app, target.volume.Scope, target.subpath) {
			continue
		}
		targets = append(targets, target)
	}
	return targets
}

// storageIsOnThisNode reports whether the data the pins describe is reachable
// from the node this process runs on - the only case in which the host and a
// plain container can be used at all.
//
// A pin by node label is not something this node can answer for itself, so it
// counts as elsewhere: swarm places the helper by the same label instead.
func (s *service) storageIsOnThisNode(ctx context.Context, pins []placementservice.VolumePin) bool {
	currNodeID := ""
	for _, pin := range pins {
		if pin.NodeLabel != "" {
			return false
		}
		if pin.NodeID == "" {
			continue
		}
		if currNodeID == "" {
			id, err := s.dockerManager.NodeCurrentID(ctx)
			if err != nil || id == "" {
				return false
			}
			currNodeID = id
		}
		if pin.NodeID != currNodeID {
			return false
		}
	}
	return true
}

// appStorageTarget is what of a mount belongs to the app, or false when the
// mount holds nothing that deleting the app should take with it.
func appStorageTarget(mnt *mount.Mount, volumes []*entity.Setting) (storageTarget, bool) {
	switch mnt.Type { //nolint:exhaustive
	case mount.TypeVolume, mount.TypeCluster:
		subpath := mountSubpath(mnt)
		if subpath == "" {
			return storageTarget{}, false
		}
		// A volume the app's scope does not account for is one nothing here can
		// say anything about, least of all whose directory this is.
		volume := volumeByRefID(volumes, mnt.Source)
		if volume == nil {
			return storageTarget{}, false
		}
		// The helper sees the volume whole: the path to delete is inside it, and
		// mounting with the subpath would put the helper in the directory it is
		// meant to remove.
		helper := *mnt
		helper.Target = volumeHelperTarget
		helper.ReadOnly = false
		if helper.VolumeOptions != nil {
			helper.VolumeOptions = new(*helper.VolumeOptions)
			helper.VolumeOptions.Subpath = ""
		}
		return storageTarget{mount: helper, subpath: subpath, volume: volume}, true

	case mount.TypeBind:
		return bindStorageTarget(mnt, volumes)

	default:
		return storageTarget{}, false
	}
}

// bindStorageTarget finds the app's directory in a bind mount.
//
// A managed `local` volume made of a host directory never reaches docker as a
// volume: useBindMountIfAppropriate rewrites it into a plain bind whose source
// is that directory plus the app's subpath, and that is all the mount carries.
// So the volume is found again by comparing the source against the directories
// the volume settings are made of. What is below one of them is the app's; the
// directory itself is the volume, and a bind pointing straight at it, or at a
// path no volume accounts for, is not the app's to delete.
func bindStorageTarget(mnt *mount.Mount, volumes []*entity.Setting) (storageTarget, bool) {
	for _, source := range bindSourceCandidates(mnt.Source) {
		if target, ok := bindStorageTargetOf(source, volumes); ok {
			return target, true
		}
	}
	return storageTarget{}, false
}

// bindSourceCandidates is the source as docker reports it, and then the same
// path with a runtime's own prefix taken off.
//
// Docker Desktop rewrites a bind source into the path its file-sharing layer
// serves it at - /Users/x becomes /host_mnt/Users/x - so the spec no longer
// carries the path HivePaaS wrote. A candidate is only ever accepted when it
// matches a volume this installation configured, so stripping a prefix cannot
// make a path match something it should not; on Linux, where nothing rewrites
// anything, the first candidate is the only one that is ever used.
func bindSourceCandidates(source string) []string {
	cleaned := filepath.Clean(source)
	candidates := []string{cleaned}
	for _, prefix := range bindSourcePrefixes {
		if rest, found := strings.CutPrefix(cleaned, prefix); found && strings.HasPrefix(rest, "/") {
			candidates = append(candidates, rest)
		}
	}
	return candidates
}

var bindSourcePrefixes = []string{"/host_mnt", "/host-mnt"}

func bindStorageTargetOf(source string, volumes []*entity.Setting) (storageTarget, bool) {
	device := ""
	var setting *entity.Setting
	for _, vol := range volumes {
		clusterVol, err := vol.AsClusterVolume()
		if err != nil || clusterVol == nil {
			continue
		}
		// The same call that decided the bind's source in the first place, asked
		// for the volume's own directory rather than an app's inside it.
		dir, _, ok := bindMountTarget(clusterVol, "")
		if !ok {
			continue
		}
		// A volume at /srv/data/pg is a more specific answer than one at
		// /srv/data for a source below both, and the specific one is the volume
		// the mount was actually built from.
		if dir = filepath.Clean(dir); strings.HasPrefix(source, dir+"/") && len(dir) > len(device) {
			device, setting = dir, vol
		}
	}
	if device == "" {
		return storageTarget{}, false
	}

	subpath := safeSubpath(strings.TrimPrefix(source, device+"/"))
	if subpath == "" {
		return storageTarget{}, false
	}
	return storageTarget{mount: bindMountWhole(device), subpath: subpath, volume: setting}, true
}

// volumeByRefID finds the volume setting a mount names. A volume mount carries
// the docker volume name, which is the setting's RefID rather than its id.
func volumeByRefID(volumes []*entity.Setting, refID string) *entity.Setting {
	if refID == "" {
		return nil
	}
	for _, vol := range volumes {
		if vol.RefID == refID {
			return vol
		}
	}
	return nil
}

// bindMountWhole is how the helper sees a host directory it has to delete
// something inside of.
func bindMountWhole(directory string) mount.Mount {
	return mount.Mount{Type: mount.TypeBind, Source: directory, Target: volumeHelperTarget}
}

// subpathPattern is what a directory inside a volume is allowed to look like:
// the names HivePaaS gives an app's storage, and nothing else. The path ends up
// in an rm -rf, so the rule is narrow on purpose.
var subpathPattern = regexp.MustCompile(`^[A-Za-z0-9._-]+(/[A-Za-z0-9._-]+)*$`)

// mountSubpath is the directory inside a volume an app was given. It is empty for
// a bind mount, for a volume mounted whole, and for anything that does not read
// as a plain relative path - a traversal, or a name carrying a character a shell
// would act on. Such a subpath is left in place rather than guessed at: what is
// above it belongs to somebody else.
func mountSubpath(mnt *mount.Mount) string {
	switch {
	case mnt.Type == mount.TypeVolume && mnt.VolumeOptions != nil:
		return safeSubpath(mnt.VolumeOptions.Subpath)
	case mnt.Type == mount.TypeCluster && mnt.VolumeOptions != nil:
		return safeSubpath(mnt.VolumeOptions.Subpath)
	default:
		return ""
	}
}

func safeSubpath(subpath string) string {
	trimmed := strings.Trim(subpath, "/")
	if !subpathPattern.MatchString(trimmed) {
		return ""
	}
	for _, segment := range strings.Split(trimmed, "/") {
		if segment == "." || segment == ".." {
			return ""
		}
	}
	return trimmed
}

// removeStorageTarget deletes one directory, the same three ways
// EnsureVolumePermissions creates it: on the host when the storage is a
// directory this node can reach, then a container, then a swarm task - which is
// the only one of the three that reaches another node, and so the only way to
// delete anything belonging to storage pinned to one.
func (s *service) removeStorageTarget(
	ctx context.Context,
	target *storageTarget,
	constraint string,
	local bool,
) error {
	if local {
		if hostPath, isDirect := s.getDirectHostPath(ctx, &target.mount, ""); isDirect && hostPath != "" {
			if err := os.RemoveAll(filepath.Join(hostPath, target.subpath)); err == nil {
				return nil
			}
		}
	}

	image := gofn.Coalesce(s.hpAppService.GetHpAgentImage(ctx), rsyncDefaultImage)
	// No shell: the path is an argument of its own, so nothing in it can be read
	// as a command, and -- keeps a name starting with a dash from becoming a flag.
	rmCmd := []string{"rm", "-rf", "--", path.Join(volumeHelperTarget, target.subpath)}

	// Docker creates a volume it does not know rather than refusing the mount, so
	// a container started here for storage that lives elsewhere would delete
	// nothing out of an empty volume and report that it had worked. It is only
	// worth trying while the data is on this node.
	if local && s.isVolumeAccessibleLocally(ctx, &target.mount) {
		_, statusCode, err := s.dockerManager.ContainerCreateToExec(ctx, image, rmCmd,
			func(opts *client.ContainerCreateOptions) {
				opts.HostConfig.Mounts = []mount.Mount{target.mount}
			})
		if err == nil && statusCode == 0 {
			return nil
		}
	}

	_, statusCode, err := s.dockerManager.ServiceCreateToExec(ctx, image, rmCmd, 0, 0,
		func(opts *client.ServiceCreateOptions) {
			opts.Spec.TaskTemplate.ContainerSpec.Mounts = []mount.Mount{target.mount}
			if constraint != "" {
				opts.Spec.TaskTemplate.Placement = &swarm.Placement{Constraints: []string{constraint}}
			}
		})
	if err != nil {
		return hperrors.Wrap(err)
	}
	if statusCode != 0 {
		return hperrors.Wrap(hperrors.ErrActionFailed).WithMsgLog(
			"removing storage directory %q exited with status code %d", target.subpath, statusCode)
	}
	return nil
}
