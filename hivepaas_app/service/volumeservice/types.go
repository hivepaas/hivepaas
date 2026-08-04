package volumeservice

import "github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"

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
