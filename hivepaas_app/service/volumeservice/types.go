package volumeservice

import (
	"github.com/moby/moby/api/types/mount"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
)

const (
	HostPathPrefix = "/host"

	// VolumeRemovalRetryMax is what a caller that can afford to wait a few
	// seconds passes to RemoveVolume. It is smaller than its clusterservice
	// counterpart's budget per attempt because volume removal runs inside the
	// transaction that deletes the volume's setting.
	VolumeRemovalRetryMax = 4
)

type RsyncOptions struct {
	Image         string
	LogStore      *tasklog.Store
	Delete        bool
	Exclude       []string
	SourceSubpath string
	DestSubpath   string
}

type RsyncOption func(opts *RsyncOptions)

func WithRsyncImage(image string) RsyncOption {
	return func(opts *RsyncOptions) {
		opts.Image = image
	}
}

func WithRsyncLogStore(logStore *tasklog.Store) RsyncOption {
	return func(opts *RsyncOptions) {
		opts.LogStore = logStore
	}
}

func WithRsyncDelete(delete bool) RsyncOption {
	return func(opts *RsyncOptions) {
		opts.Delete = delete
	}
}

func WithRsyncExclude(exclude ...string) RsyncOption {
	return func(opts *RsyncOptions) {
		opts.Exclude = append(opts.Exclude, exclude...)
	}
}

func WithSourceSubpath(subpath string) RsyncOption {
	return func(opts *RsyncOptions) {
		opts.SourceSubpath = subpath
	}
}

func WithDestSubpath(subpath string) RsyncOption {
	return func(opts *RsyncOptions) {
		opts.DestSubpath = subpath
	}
}

type CloneVolumeReq struct {
}

type CloneVolumeResp struct {
}

// AppMountReq asks for one cluster-volume setting to be mounted into an app.
type AppMountReq struct {
	Type mount.Type
	// Source is the id of the cluster-volume setting, not a docker volume name.
	Source         string
	Target         string
	ReadOnly       bool
	Consistency    mount.Consistency
	VolumeOptions  *AppMountVolumeOptions
	ClusterOptions *AppMountVolumeOptions

	// OwnerApp is the app whose directory inside the volume this mount reaches,
	// when that is not the app doing the mounting - a file manager given the
	// files of the database beside it. Nil is the ordinary case and means the
	// mounting app's own directory.
	//
	// Whoever sets this has already decided the caller may have it: the service
	// builds the mount it is told to and checks no permission. It needs Project
	// and ProjectEnv loaded, for the same reason BuildAppMountsReq.App does.
	OwnerApp *entity.App
}

type AppMountVolumeOptions struct {
	Subpath      string
	NoCopy       bool
	Labels       map[string]string
	DriverConfig *mount.Driver
}

type BuildAppMountsReq struct {
	// App has Project and ProjectEnv loaded: the subpath a volume gets depends on them.
	App *entity.App
	// Kept are mounts the app already has and keeps as they are. They are not
	// rebuilt, but they take part in the pin conflict check - an unchanged mount
	// pinned to one node conflicts with a new one pinned to another.
	Kept []mount.Mount
	New  []*AppMountReq
}

type BuildAppMountsResp struct {
	// Mounts are Kept followed by the built New mounts.
	Mounts []mount.Mount
}

// AppMountDesc is what a mount turned out to reach: whose directory it is, and
// what is left of the path below it.
type AppMountDesc struct {
	// AppKey is the app the directory belongs to. It is empty when the path is
	// not an app's directory at all - a volume mounted whole, a bind HivePaaS did
	// not make, or a volume the app's scope does not account for.
	AppKey string
	// Own says AppKey is the app the mounts were read from, which is the ordinary
	// case and the only one that existed before an app could be given another's
	// storage.
	Own bool
	// Subpath is what the request had asked for below that app's directory.
	Subpath string
	// VolumeID is the cluster-volume setting whose directory the mount reaches.
	// It is set whenever AppKey is: a mount that is no app's directory names no
	// volume here either.
	VolumeID string
}

// InspectAppStorageReq asks about the directories a set of apps would be given.
//
// The apps need not exist: a template is answered for before anything is
// created, which is the whole point of asking.
type ResetAppStoragePermissionsReq struct {
	App   *entity.App
	Mount mount.Mount
	// Owner is who the directory and everything in it are given to. Without one,
	// every user is let read and write them instead.
	Owner *StorageOwner
}

// StorageOwner is a user and group by number, which is how a volume records
// them: a name means something only inside the image that defines it.
type StorageOwner struct {
	UID int
	GID int
}

type ResetAppStoragePermissionsResp struct {
	// Path is the directory that was reset, inside its volume.
	Path string
}

type InspectAppStorageReq struct {
	// Scope is what the volumes are looked up in - the env the apps belong to.
	Scope   *entity.ObjectScope
	Queries []*AppStorageQuery
}

type AppStorageQuery struct {
	// AppKey is what the answer is reported under.
	AppKey string
	// App carries the keys the directory's name is built from: the app's own, its
	// env's, and its project's. It need not be saved.
	App *entity.App
	// VolumeID is the cluster-volume setting the mount names.
	VolumeID string
	// Subpath is what the mount asked for below the app's own directory, if
	// anything.
	Subpath string
}

type InspectAppStorageResp struct {
	States []*AppStorageState
}

// AppStorageState is one app's directory in one volume.
//
// Checked says whether this node could look at all. It is false for storage
// pinned to another node, and a caller must not read Exists or Empty as an
// answer when it is: nothing was seen, which is not the same as nothing there.
type AppStorageState struct {
	AppKey     string
	VolumeID   string
	VolumeName string
	// Path is the directory inside the volume, as the app would be given it.
	Path    string
	Checked bool
	Exists  bool
	Empty   bool
}

// HasData reports the one thing callers act on: this directory is there and
// something is in it.
func (s *AppStorageState) HasData() bool {
	return s != nil && s.Checked && s.Exists && !s.Empty
}
