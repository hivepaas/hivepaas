package templaterepo

import (
	"crypto/sha256"
	"encoding/hex"
	"maps"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
)

const categoriesYAML = `
apiVersion: hivepaas.com/v1
kind: TemplateCategories
categories:
  - id: databases
    title: Databases
    children:
      - {id: sql, title: SQL}
      - {id: cache, title: Cache & Queues}
`

const tagsYAML = `
apiVersion: hivepaas.com/v1
kind: TemplateTags
tags:
  - {id: sql, title: SQL}
`

const demoTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: demo
  title: Demo
  tagline: A demo database
  description: Demo.
  categories: [databases/sql]
  tags: [sql]
  icon: icons/demo.svg
  license: Apache-2.0
  requires: {versionCode: v000001}
parameters:
  - {name: password, title: Password, type: secret, generate: {length: 16}}
  - {name: dataVolume, title: Data volume, type: volume}
versions:
  - {name: "2", release: "2.1", default: true, image: "demo:2.1.0"}
  - {name: "1", release: "1.9", deprecated: true, image: "demo:1.9.3"}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
    storage:
      mounts:
        /data: {type: volume, source: "${{ params.dataVolume }}"}
  settings:
    kind: {category: database, engine: demo, database: {password: "${{ params.password }}"}}
`

const demoIcon = `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1 1"/>`

func validRepoFS() fstest.MapFS {
	return fstest.MapFS{
		CategoriesFile:        {Data: []byte(categoriesYAML)},
		TagsFile:              {Data: []byte(tagsYAML)},
		"templates/demo.yaml": {Data: []byte(demoTemplateYAML)},
		"icons/demo.svg":      {Data: []byte(demoIcon)},
	}
}

func withFile(fsys fstest.MapFS, path, content string) fstest.MapFS {
	out := maps.Clone(fsys)
	out[path] = &fstest.MapFile{Data: []byte(content)}
	return out
}

func loadAndLint(t *testing.T, fsys fstest.MapFS) (*Repo, []Problem) {
	t.Helper()
	repo, problems, err := Load(fsys)
	assert.NoError(t, err)
	return repo, append(problems, Lint(repo)...)
}

func TestLoadAndLintAValidRepository(t *testing.T) {
	repo, problems := loadAndLint(t, validRepoFS())

	assert.Empty(t, problems)
	assert.Len(t, repo.Templates, 1)
	assert.Equal(t, "demo", repo.FindTemplate("demo").Template.Metadata.Name)
	assert.Equal(t, demoIcon, string(repo.Icons["icons/demo.svg"].Content))
}

func TestLintFindsProblems(t *testing.T) {
	cases := map[string]struct {
		fsys fstest.MapFS
		want string
	}{
		"unknown category": {
			withFile(validRepoFS(), "templates/demo.yaml",
				strings.Replace(demoTemplateYAML, "databases/sql", "databases/graph", 1)),
			`category "databases/graph" is not in categories.yaml`,
		},
		"unknown tag": {
			withFile(validRepoFS(), "templates/demo.yaml", strings.Replace(demoTemplateYAML, "[sql]", "[nosql]", 1)),
			`tag "nosql" is not in tags.yaml`,
		},
		"missing icon": {
			func() fstest.MapFS { fsys := validRepoFS(); delete(fsys, "icons/demo.svg"); return fsys }(),
			"icons/demo.svg",
		},
		"floating image": {
			withFile(validRepoFS(), "templates/demo.yaml", strings.Replace(demoTemplateYAML, "demo:2.1.0", "demo:2", 1)),
			"exact release",
		},
		"undefined placeholder": {
			withFile(validRepoFS(), "templates/demo.yaml",
				strings.Replace(demoTemplateYAML, "params.dataVolume", "params.volume", 1)),
			`placeholder "params.volume" refers to nothing`,
		},
		"unknown field": {
			withFile(validRepoFS(), "templates/demo.yaml", demoTemplateYAML+"\nextra: 1\n"),
			"extra",
		},
		"file name mismatch": {
			withFile(validRepoFS(), "templates/other.yaml", demoTemplateYAML),
			`must equal the file name "other"`,
		},
		"oversized template": {
			withFile(validRepoFS(), "templates/huge.yaml", strings.Repeat("#", MaxTemplateSize+1)),
			"templates/huge.yaml",
		},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			_, problems := loadAndLint(t, tc.fsys)
			var all []string
			for _, problem := range problems {
				all = append(all, problem.String())
			}
			assert.Contains(t, strings.Join(all, "\n"), tc.want)
		})
	}
}

func TestLoadRequiresTheVocabularies(t *testing.T) {
	fsys := validRepoFS()
	delete(fsys, CategoriesFile)
	_, _, err := Load(fsys)
	assert.Error(t, err)
}

func TestBuildIndex(t *testing.T) {
	repo, problems := loadAndLint(t, validRepoFS())
	assert.Empty(t, problems)

	index, err := BuildIndex(repo)
	assert.NoError(t, err)

	entry := index.FindTemplate("demo")
	templateSum := sha256.Sum256([]byte(demoTemplateYAML))
	iconSum := sha256.Sum256([]byte(demoIcon))
	assert.Equal(t, templatemodel.FileRef{Path: "templates/demo.yaml", SHA256: hex.EncodeToString(templateSum[:])},
		entry.File)
	assert.Equal(t, templatemodel.FileRef{Path: "icons/demo.svg", SHA256: hex.EncodeToString(iconSum[:])}, entry.Icon)
	assert.Equal(t, []*templatemodel.IndexVersion{
		{Name: "2", Release: "2.1", Default: true},
		{Name: "1", Release: "1.9", Deprecated: true},
	}, entry.Versions)
	assert.Equal(t, "Apache-2.0", entry.License, "the store lists the license without reading every template file")
	assert.Equal(t, "Cache & Queues", index.Categories[0].Children[1].Title)

	data, err := MarshalIndex(index)
	assert.NoError(t, err)
	assert.True(t, strings.HasSuffix(string(data), "}\n"))
	assert.Contains(t, string(data), "Cache & Queues", "no HTML escaping in a file people read")

	again, err := MarshalIndex(index)
	assert.NoError(t, err)
	assert.Equal(t, data, again)

	decoded, err := templatemodel.DecodeIndex(data)
	assert.NoError(t, err)
	assert.Equal(t, index, decoded)
}

// capsTemplateYAML is a template that asks the host for something, and
// appTemplateYAML one that only depends on it. The store marks both: the app is
// created together with the dependency, and whoever cannot grant IPC_LOCK to one
// cannot grant it to the other.
const capsTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: search
  title: Search
  tagline: A search engine
  description: Search.
  categories: [databases/sql]
  icon: icons/demo.svg
  requires: {versionCode: v000001}
versions:
  - {name: "3", release: "3.0", default: true, image: "search:3.0.0"}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
    resources:
      capabilities: {capabilityAdd: [IPC_LOCK]}
`

const appTemplateYAML = `
apiVersion: hivepaas.com/v1
kind: AppTemplate
metadata:
  name: site
  title: Site
  tagline: A site that searches
  description: Site.
  categories: [databases/sql]
  icon: icons/demo.svg
  requires: {versionCode: v000001}
dependencies:
  - {name: search, title: Search, template: search}
versions:
  - {name: "1", release: "1.0", default: true, image: "site:1.0.0"}
app:
  deployment:
    source: {activeMethod: image, imageSource: {image: "${{ image }}"}}
`

func TestBuildIndexMarksWhatNeedsCapabilities(t *testing.T) {
	fsys := withFile(withFile(validRepoFS(), "templates/search.yaml", capsTemplateYAML),
		"templates/site.yaml", appTemplateYAML)
	repo, problems := loadAndLint(t, fsys)
	assert.Empty(t, problems)

	index, err := BuildIndex(repo)
	assert.NoError(t, err)

	assert.True(t, index.FindTemplate("search").RequiresCapabilities)
	assert.True(t, index.FindTemplate("site").RequiresCapabilities,
		"a dependency's capabilities are granted by the same request")
	assert.False(t, index.FindTemplate("demo").RequiresCapabilities)
}
