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
