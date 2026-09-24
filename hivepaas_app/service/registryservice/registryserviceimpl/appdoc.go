package registryserviceimpl

import (
	"bytes"
	_ "embed"
	"strings"
	"text/template"

	"github.com/tiendc/gofn"
	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/imageref"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
)

// defaultRegistryImage is what runs when the release names no registry image -
// a release file from before the field existed. The release is what normally
// decides it: the configuration this package writes is the configuration that
// version of zot accepts, and the two move together.
const defaultRegistryImage = "ghcr.io/project-zot/zot:v2.1.21"

// registryImage is the image this release runs the registry on.
func registryImage() string {
	return gofn.Coalesce(systemappservice.CurrentRelease().RegistryImage, defaultRegistryImage)
}

// registryVersion is the image's minor line - "2.1" for v2.1.21 - which is what
// the app's kind settings record.
func registryVersion(image string) string {
	tag := strings.TrimPrefix(imageref.Parse(image).Tag, "v")
	parts := strings.SplitN(tag, ".", 3) //nolint:mnd // major, minor, the rest
	if len(parts) < 2 {                  //nolint:mnd
		return tag
	}
	return parts[0] + "." + parts[1]
}

//go:embed app.yaml.tmpl
var appDocTemplate string

// appDocInput is everything the document needs that is decided elsewhere.
type appDocInput struct {
	Name        string
	Key         string
	Domain      string
	MemoryLimit string
	// CPULimit is in cores, zero for no cap.
	CPULimit float64
	// OomScoreAdj keeps the registry alive over user apps when memory runs out.
	// The registry always has a memory limit, which is what makes that safe.
	OomScoreAdj int64
	// VolumeID is the cluster-volume setting a mount names. A mount's source is
	// the setting's id, not the volume's name: BuildAppMounts looks the id up and
	// derives everything else - the docker name, the node pin, the bind rewrite -
	// from the setting it finds. It is empty for a registry on S3, which mounts
	// nothing.
	VolumeID  string
	ZotConfig string
	Htpasswd  string
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

	image := registryImage()
	var buf bytes.Buffer
	err = tmpl.Execute(&buf, struct {
		appDocInput
		Image   string
		Version string
		RootDir string
		ConfDir string
	}{
		appDocInput: in,
		Image:       image,
		Version:     registryVersion(image),
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
