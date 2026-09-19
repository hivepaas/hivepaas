package dockerhelper

import "path"

// initEntrypoints are the init processes an image starts itself with. Each of
// them does what docker's own init does - reap what the app orphans, pass
// signals on - and each of them expects to be process 1 while doing it.
var initEntrypoints = map[string]struct{}{
	"tini":              {},
	"docker-init":       {},
	"dumb-init":         {},
	"catatonit":         {},
	"s6-svscan":         {},
	"s6-overlay-suexec": {},
	"init":              {}, // s6-overlay's /init
}

// ImageProvidesInit reports whether an image starts with an init of its own.
//
// Giving such an image docker's init as well puts two of them in the container:
// the image's runs as process 2 and says so - tini warns that it cannot reap
// anything from there, s6 refuses to run at all - while the outer one does the
// work. What the image ships is the one its author tested, so it is the one to
// keep.
func ImageProvidesInit(entrypoint []string) bool {
	if len(entrypoint) == 0 {
		return false
	}
	_, found := initEntrypoints[path.Base(entrypoint[0])]
	return found
}
