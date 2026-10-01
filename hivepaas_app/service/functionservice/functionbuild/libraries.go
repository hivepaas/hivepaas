package functionbuild

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"io/fs"
	"os"
	"path/filepath"
	"slices"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// LibrariesRepo is the repository of the images a function's test runs start
// from. They stay on the node they were built on, and are never pushed: the
// cluster cleanup's image prune removes them like any image no container uses.
const LibrariesRepo = "hivepaas-function-libs"

// librariesTagLen is how much of the hash an image's tag keeps.
const librariesTagLen = 24

// LibrariesResp is the image a function's test runs start from: the runtime's,
// with the function's Debian packages and libraries installed.
type LibrariesResp struct {
	// Dockerfile builds the image, empty when there is nothing to install and
	// Image is the runtime's own.
	Dockerfile string
	// Image is the image's name: LibrariesRepo, tagged after a hash of what it
	// is built from, so that the same libraries are installed once per node.
	Image string
	Notes []string
}

// Libraries writes the Dockerfile of a function's libraries: the function's
// own, up to its install. The code goes into the test run's container rather
// than the image, and the limits into its environment.
func Libraries(req *DockerfileReq) (*LibrariesResp, error) {
	st, err := writeStages(req, "The libraries of a function, for its test runs")
	if err != nil {
		return nil, err
	}
	if st.last == "base" && len(req.Source.SystemPackages) == 0 {
		return &LibrariesResp{Image: st.baseImage, Notes: st.notes}, nil
	}

	dockerfile := st.w.String()
	files := st.installedFrom
	if st.codeCopied {
		if files, err = allFiles(req.SourceDir); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	h := sha256.New()
	write := func(value string) {
		_ = binary.Write(h, binary.BigEndian, uint64(len(value)))
		_, _ = h.Write([]byte(value))
	}
	write(dockerfile)
	for _, name := range files {
		content, _ := readFile(req.SourceDir, name)
		write(name)
		write(content)
	}
	tag := hex.EncodeToString(h.Sum(nil))[:librariesTagLen]
	return &LibrariesResp{Dockerfile: dockerfile, Image: LibrariesRepo + ":" + tag, Notes: st.notes}, nil
}

// LockFiles are the files a runtime writes the versions it resolved into: what
// its install makes when there is none, and for Go what its build tidies.
func LockFiles(runtime base.FunctionRuntime) []string {
	m := manifests[runtime]
	if runtime.Compiled() {
		return []string{m.file, m.lock}
	}
	return []string{m.lock}
}

// allFiles are the paths of the files under dir, sorted, with "/" between their
// parts.
func allFiles(dir string) ([]string, error) {
	var files []string
	err := filepath.WalkDir(dir, func(path string, entry fs.DirEntry, err error) error {
		if err != nil || entry.IsDir() {
			return err
		}
		rel, err := filepath.Rel(dir, path)
		if err != nil {
			return err //nolint:wrapcheck
		}
		files = append(files, filepath.ToSlash(rel))
		return nil
	})
	if err != nil && !os.IsNotExist(err) {
		return nil, err //nolint:wrapcheck
	}
	slices.Sort(files)
	return files, nil
}
