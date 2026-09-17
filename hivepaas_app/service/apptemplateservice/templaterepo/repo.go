// Package templaterepo reads a checkout of an app templates repository: it
// loads the files, lints them and builds index.json. It works over an fs.FS and
// nothing else, so tools/apptemplate and the development source read a checkout
// exactly the same way.
package templaterepo

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"io/fs"
	"sort"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const (
	CategoriesFile = "categories.yaml"
	TagsFile       = "tags.yaml"
	IndexFile      = "index.json"
	TemplatesDir   = "templates"

	MaxIndexSize    = 1 << 20
	MaxTemplateSize = 256 << 10
	MaxIconSize     = 256 << 10
)

type File struct {
	Path    string
	Content []byte
	SHA256  string
}

type TemplateFile struct {
	File
	Template *templatemodel.Template
}

type Repo struct {
	Categories *templatemodel.Categories
	Tags       *templatemodel.Tags
	// Templates are sorted by name.
	Templates []*TemplateFile
	// Icons holds the icons templates name, by path.
	Icons map[string]*File
}

// Problem is one thing wrong with a repository, for a person to fix.
type Problem struct {
	Path    string
	Message string
}

func (p Problem) String() string {
	return p.Path + ": " + p.Message
}

func (r *Repo) FindTemplate(name string) *TemplateFile {
	for _, file := range r.Templates {
		if file.Template.Metadata.Name == name {
			return file
		}
	}
	return nil
}

// Load reads a checkout. The two vocabularies are required, so failing to read
// either is an error. A template that cannot be read or decoded is a problem
// rather than an error: one broken file must not hide what is wrong with the rest.
func Load(fsys fs.FS) (*Repo, []Problem, error) {
	categoriesData, err := readLimited(fsys, CategoriesFile, MaxTemplateSize)
	if err != nil {
		return nil, nil, err
	}
	categories, err := templatemodel.DecodeCategories(categoriesData)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	tagsData, err := readLimited(fsys, TagsFile, MaxTemplateSize)
	if err != nil {
		return nil, nil, err
	}
	tags, err := templatemodel.DecodeTags(tagsData)
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}

	repo := &Repo{Categories: categories, Tags: tags, Icons: map[string]*File{}}
	var problems []Problem

	paths, err := fs.Glob(fsys, TemplatesDir+"/*.yaml")
	if err != nil {
		return nil, nil, hperrors.Wrap(err)
	}
	for _, path := range paths {
		data, err := readLimited(fsys, path, MaxTemplateSize)
		if err != nil {
			problems = append(problems, Problem{Path: path, Message: ErrorText(err)})
			continue
		}
		tmpl, err := templatemodel.DecodeTemplate(data)
		if err != nil {
			problems = append(problems, Problem{Path: path, Message: ErrorText(err)})
			continue
		}
		repo.Templates = append(repo.Templates, &TemplateFile{File: *newFile(path, data), Template: tmpl})
		problems = append(problems, repo.loadIcon(fsys, tmpl.Metadata.Icon)...)
	}

	sort.Slice(repo.Templates, func(i, j int) bool {
		return repo.Templates[i].Template.Metadata.Name < repo.Templates[j].Template.Metadata.Name
	})
	return repo, problems, nil
}

func (r *Repo) loadIcon(fsys fs.FS, path string) []Problem {
	if path == "" || r.Icons[path] != nil {
		return nil
	}
	if !fs.ValidPath(path) {
		return []Problem{{Path: path, Message: "an icon path must stay inside the repository"}}
	}
	data, err := readLimited(fsys, path, MaxIconSize)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil // Lint reports a missing icon against the template that names it
		}
		return []Problem{{Path: path, Message: ErrorText(err)}}
	}
	r.Icons[path] = newFile(path, data)
	return nil
}

func newFile(path string, data []byte) *File {
	sum := sha256.Sum256(data)
	return &File{Path: path, Content: data, SHA256: hex.EncodeToString(sum[:])}
}

// readLimited reads one file, refusing anything over its limit. The limits are
// equal today, but they are separate parts of the format's contract - a bigger
// icon budget must not silently become a bigger template budget.
//
//nolint:unparam
func readLimited(fsys fs.FS, name string, limit int64) ([]byte, error) {
	file, err := fsys.Open(name)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	defer func() { _ = file.Close() }()

	data, err := io.ReadAll(io.LimitReader(file, limit+1))
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	if int64(len(data)) > limit {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateFileTooLarge).
			WithExtraDetail("%s is larger than %d bytes", name, limit)
	}
	return data, nil
}

// ErrorText renders an error for a person reading lint output: the translated
// message and its detail, rather than an error code.
func ErrorText(err error) string {
	var hpErr hperrors.HPError
	if errors.As(err, &hpErr) {
		return strings.Join(strings.Fields(hpErr.Build("en").Detail), " ")
	}
	return err.Error()
}
