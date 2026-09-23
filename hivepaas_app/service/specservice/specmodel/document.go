package specmodel

// DocHeader is repeated at the top of every payload file, so that a file
// extracted from a bundle and handed to somebody on its own still says what it
// is.
type DocHeader struct {
	APIVersion string `yaml:"apiVersion"`
	Kind       Kind   `yaml:"kind"`
	Scope      string `yaml:"scope"`
}

// NewDocHeader builds the header every payload file starts with.
func NewDocHeader(scope string) DocHeader {
	return DocHeader{APIVersion: APIVersion, Kind: KindSpec, Scope: scope}
}

// GlobalDoc is global.yaml: the global and hivepaas scope settings.
//
// It does not carry the hivepaas *project*, which holds HivePaaS's own stack -
// selectProjects drops that.
type GlobalDoc struct {
	DocHeader `yaml:",inline"`

	// Settings is the assembled block map: singleton types under their block
	// name, collection types as keyed maps.
	Settings map[string]any `yaml:"settings,omitempty"`
}

// ProjectDoc is projects/<key>/project.yaml.
type ProjectDoc struct {
	DocHeader `yaml:",inline"`

	Project string `yaml:"project"`
	// ID is the project's id on the installation that exported it. An import
	// into that installation matches on it; anywhere else it matches nothing,
	// and the key is used instead.
	ID   string `yaml:"id,omitempty"`
	Name string `yaml:"name"`
	Note string `yaml:"note,omitempty"`
	// Owner is the user who owns the project, as import can find them again.
	Owner *ProjectOwner `yaml:"owner,omitempty"`
	// Envs names the env files belonging to this project, so a reader of one
	// project file knows what else there is without listing the archive.
	Envs []string `yaml:"envs,omitempty"`

	Settings map[string]any `yaml:"settings,omitempty"`
}

// ProjectOwner names the user who owns a project. Users do not travel in a
// bundle, so import finds the owner again: by id on the installation that
// exported it - which still works after the owner changed their email - and by
// email anywhere else.
type ProjectOwner struct {
	ID    string `yaml:"id,omitempty"`
	Email string `yaml:"email,omitempty"`
}

// EnvDoc is projects/<key>/envs/<env>.yaml: the env's own settings and every
// app in it.
type EnvDoc struct {
	DocHeader `yaml:",inline"`

	Project string `yaml:"project"`
	Env     string `yaml:"env"`
	Name    string `yaml:"name,omitempty"`
	Color   string `yaml:"color,omitempty"`
	Index   int    `yaml:"index,omitempty"`

	// Labels are env-scope service labels, already filtered.
	Labels map[string]string `yaml:"labels,omitempty"`

	// Apps is keyed by app key, which is unique within an env.
	Apps map[string]*AppDoc `yaml:"apps,omitempty"`

	Settings map[string]any `yaml:"settings,omitempty"`
}

// AppDoc is one app inside an env document.
//
// Deployment is a pointer and is omitted entirely for an app that has never
// been deployed - a real state rather than an edge case: two of five user apps
// in a development installation have an empty ServiceID.
type AppDoc struct {
	App string `yaml:"app"`
	// ID is the app's id on the installation that exported it; see ProjectDoc.ID.
	ID     string `yaml:"id,omitempty"`
	Name   string `yaml:"name"`
	Status string `yaml:"status,omitempty"`
	Note   string `yaml:"note,omitempty"`

	Deployment *Deployment `yaml:"deployment,omitempty"`

	Settings map[string]any `yaml:"settings,omitempty"`
}

// ExternalRef stands in for a reference whose target was outside the export.
//
// The scope rule produces these constantly - exporting one project whose apps
// use a global certificate is the ordinary case - so they are a designed part
// of the format rather than an error path.
type ExternalRef struct {
	Type string `yaml:"type"`
	Name string `yaml:"name"`
	Kind string `yaml:"kind,omitempty"`
	// ID is the source identifier, which lets a re-import into the same
	// installation match exactly instead of by name.
	ID string `yaml:"id,omitempty"`
}

// CollectionEntryIDKey is where export writes the id of the setting a collection
// entry was exported from: at the top level of the entry, beside the setting's
// own data. An import into the installation that exported it matches on it, so
// a setting renamed since is still recognized. No exported collection type may
// have a top-level field of that name - TestNoExportedCollectionTypeHasATopLevelID
// holds that.
const CollectionEntryIDKey = "id"

// Bundle is what the exporter hands the bundle writer.
type Bundle struct {
	Manifest *Manifest
	// Files maps bundle-relative path to already-serialized content.
	Files map[string][]byte
	// Report carries everything export skipped or could not resolve.
	Report *Report
}
