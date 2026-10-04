package composeserviceimpl

import (
	"context"
	"os"
	"path"
	"path/filepath"
	"slices"
	"strings"
	"sync"

	"github.com/compose-spec/compose-go/v2/consts"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// sandbox is the scratch directory compose-go reads the file in, and the only
// resource loader it has: it accepts every path, so that neither a remote
// loader nor compose-go's own local one is asked, and loads none but a file of
// the request's.
//
// A file of the request is written to the directory only once compose-go asks
// for it - an included or extended compose file, the .env beside one, an
// include's env file - since compose-go reads those from the disk. Every other
// file the converter reads from the request itself: most compose files have
// nothing of theirs written at all, and what they read is often a secret.
type sandbox struct {
	dir   string
	files map[string][]byte

	mu      sync.Mutex
	written map[string]bool
	// loaded is where each path compose-go asked for, as it wrote it, was
	// found the last time it asked: extends asks Dir with that path right
	// after.
	loaded map[string]string
	// missing are the compose files compose-go asked for that the request
	// lacks.
	missing []string
}

func newSandbox(dir string, files map[string][]byte) *sandbox {
	return &sandbox{dir: dir, files: files, written: map[string]bool{}, loaded: map[string]string{}}
}

// write writes a file of the request to the directory, once.
func (s *sandbox) write(rel string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.written[rel] {
		return nil
	}
	target := filepath.Join(s.dir, filepath.FromSlash(rel))
	if err := os.MkdirAll(filepath.Dir(target), 0o700); err != nil { //nolint:mnd // owner only
		return hperrors.Wrap(err)
	}
	if err := os.WriteFile(target, s.files[rel], 0o600); err != nil { //nolint:mnd // owner only
		return hperrors.Wrap(err)
	}
	s.written[rel] = true
	return nil
}

func (s *sandbox) given(rel string) bool {
	_, found := s.files[rel]
	return found
}

func (s *sandbox) Accept(string) bool { return true }

// Load is a compose file an include or extends names: a path relative to the
// file that names it, as compose reads one, or an absolute one inside the
// directory. One leaving it is refused; one the request lacks is missing, and
// the review asks for it.
func (s *sandbox) Load(ctx context.Context, p string) (string, error) {
	rel, ok := s.resolve(ctx, p)
	if !ok {
		return "", hperrors.Wrap(hperrors.ErrComposeFilePath).WithExtraDetail("%s", p)
	}
	if !s.given(rel) {
		s.mu.Lock()
		if !slices.Contains(s.missing, rel) {
			s.missing = append(s.missing, rel)
		}
		s.mu.Unlock()
		return "", hperrors.Wrap(hperrors.ErrComposeInvalid).WithExtraDetail("%s is not among the files given", rel)
	}
	if err := s.write(rel); err != nil {
		return "", err
	}
	// compose-go reads the .env beside an included file itself.
	if env := path.Join(path.Dir(rel), ".env"); s.given(env) {
		if err := s.write(env); err != nil {
			return "", err
		}
	}
	target := filepath.Join(s.dir, filepath.FromSlash(rel))
	s.mu.Lock()
	s.loaded[p] = target
	s.mu.Unlock()
	return target, nil
}

// resolve is a path compose-go asks for as one of the request's files: an
// absolute one relative to the directory, a relative one to the directory of
// the file naming it - which compose-go puts in the context. A remote one, a
// URL, is none: nothing is fetched.
func (s *sandbox) resolve(ctx context.Context, p string) (string, bool) {
	if strings.Contains(p, "://") {
		return "", false
	}
	if filepath.IsAbs(p) {
		rel, err := filepath.Rel(s.dir, p)
		if err != nil {
			return "", false
		}
		return cleanPath(rel)
	}
	return cleanPath(path.Join(s.base(ctx), filepath.ToSlash(strings.TrimSpace(p))))
}

// base is the directory of the file compose-go is reading, relative to the
// scratch directory: empty for the compose file itself.
func (s *sandbox) base(ctx context.Context) string {
	file, _ := ctx.Value(consts.ComposeFileKey{}).(string)
	if !filepath.IsAbs(file) {
		return ""
	}
	rel, err := filepath.Rel(s.dir, filepath.Dir(file))
	if err != nil {
		return ""
	}
	cleaned, ok := cleanPath(rel)
	if !ok {
		return ""
	}
	return cleaned
}

// Dir is the directory the relative paths of a file compose-go loaded are
// read from: the file's own, absolute. compose-go asks it with the path it
// loaded for an include, and with the path as written right after loading it
// for extends. Anything else is the scratch directory itself: never a
// directory outside it.
func (s *sandbox) Dir(p string) string {
	if !filepath.IsAbs(p) {
		s.mu.Lock()
		target, found := s.loaded[p]
		s.mu.Unlock()
		if !found {
			return s.dir
		}
		p = target
	}
	if rel, ok := s.inside(p); ok {
		return filepath.Join(s.dir, filepath.FromSlash(path.Dir(rel)))
	}
	return s.dir
}

// inside is an absolute path inside the directory, relative to it.
func (s *sandbox) inside(p string) (string, bool) {
	rel, err := filepath.Rel(s.dir, p)
	if err != nil {
		return "", false
	}
	return cleanPath(rel)
}

// loadedFiles are the files of the request's compose-go loaded.
func (s *sandbox) loadedFiles() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []string
	for _, target := range s.loaded {
		if rel, ok := s.inside(target); ok && !slices.Contains(out, rel) {
			out = append(out, rel)
		}
	}
	slices.Sort(out)
	return out
}

// missingFiles are the compose files compose-go asked for that the request
// lacks, in the order it asked.
func (s *sandbox) missingFiles() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return slices.Clone(s.missing)
}
