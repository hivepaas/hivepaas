package registryserviceimpl

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"

	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

const (
	// registryImage is pinned here rather than in the setting: the configuration
	// this package writes is the configuration this version of zot accepts, and
	// the two have to move together. registryVersion is the same release's minor
	// line, which is what the app's kind settings record.
	registryImage   = "ghcr.io/project-zot/zot:v2.1.21"
	registryVersion = "2.1"
)

//go:embed app.yaml.tmpl
var appDocTemplate string

// appDocInput is everything the document needs that is decided elsewhere.
type appDocInput struct {
	Name        string
	Key         string
	Domain      string
	MemoryLimit string
	// VolumeName is empty for a registry on S3, which mounts nothing.
	VolumeName string
	ZotConfig  string
	Htpasswd   string
}

// renderAppDoc produces the document BuildApp turns into the app's settings.
//
// The document carries deployment and settings and nothing else: CheckBuildable
// refuses any other top-level field, because an AppDoc built for provisioning is
// the same thing a template's `app:` block is, and the app's name and key come
// from the provisioning request rather than from the document.
//
// It goes through YAML rather than building specmodel structs directly because
// AppDoc.Settings is an untyped tree - the settings blocks are decoded from it -
// so the typed route would mean hand-writing the same maps with none of the
// readability. This is also the exact path a template takes, which means the
// registry cannot drift away from what templates can express.
func renderAppDoc(in appDocInput) (*specmodel.AppDoc, error) {
	tmpl, err := template.New("registry-app").Funcs(template.FuncMap{
		"indent": indentBlock,
	}).Parse(appDocTemplate)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		appDocInput
		Image   string
		Version string
		RootDir string
		ConfDir string
	}{
		appDocInput: in,
		Image:       registryImage,
		Version:     registryVersion,
		RootDir:     registryRootDir,
		ConfDir:     registryConfDir,
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	doc := &specmodel.AppDoc{}
	decoder := yaml.NewDecoder(bytes.NewReader(buf.Bytes()))
	decoder.KnownFields(true)
	if err = decoder.Decode(doc); err != nil {
		return nil, hperrors.Wrap(err)
	}
	if err = specmodel.CheckBuildable(doc); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return doc, nil
}

// indentBlock puts a multi-line value under a YAML block scalar. Every line is
// indented, and a trailing newline is dropped so that the block does not end in a
// blank line the decoder would keep.
func indentBlock(spaces int, value string) string {
	pad := strings.Repeat(" ", spaces)
	lines := strings.Split(strings.TrimRight(value, "\n"), "\n")
	for i, line := range lines {
		if line == "" {
			continue
		}
		lines[i] = pad + line
	}
	return strings.Join(lines, "\n")
}
