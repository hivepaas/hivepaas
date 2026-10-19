package templatemodel

import (
	"fmt"
	"maps"
	"path"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/robfig/cron/v3"
	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// A managed template is an ordinary template plus the engine that looks after
// what it creates. The template is data - images, artifacts, scripts, defaults -
// and the engine is code in HivePaaS: anything that would need a condition or a
// loop in YAML belongs to the engine. See
// docs/superpowers/specs/2026-10-17-managed-templates-design.md.

type ManagedEngine string

const ManagedEnginePostgres ManagedEngine = "postgres"

// EngineContract is what a template may expect of an engine this HivePaaS has,
// and what the engine expects of the template: the engine API versions it
// speaks, and the artifacts it mounts by name. A template outside the range is
// one this HivePaaS cannot look after.
type EngineContract struct {
	// Title is how the engine is named to a person: PostgreSQL.
	Title     string
	MinAPI    int
	MaxAPI    int
	Artifacts []string
}

// ManagedEngines is every engine a template may name. The engines themselves are
// managedservice's; its registry is tested against this table, as the spec
// builders are against BuildableBlocks.
var ManagedEngines = map[ManagedEngine]EngineContract{
	ManagedEnginePostgres: {Title: "PostgreSQL", MinAPI: 1, MaxAPI: 1, Artifacts: []string{"wal-g"}},
}

// Managed binds a template to its engine.
type Managed struct {
	Engine ManagedEngine `yaml:"engine"`
	// EngineAPI is the contract between this template and the engine: which
	// scripts it ships, their arguments and exit codes, the environment they read.
	EngineAPI int `yaml:"engineApi"`
	// Postgres holds the defaults of the create form, for engine postgres. Any of
	// them may be left out, and the engine's own default applies.
	Postgres *ManagedDefaults `yaml:"postgres,omitempty"`
}

// ManagedDefaults are the create form's defaults, the same for every engine.
// The YAML key they are under is the engine's, so that a template says which
// engine it means.
type ManagedDefaults struct {
	ArchiveTimeout      Duration `yaml:"archiveTimeout,omitempty"`
	BaseBackupSchedule  string   `yaml:"baseBackupSchedule,omitempty"`
	RetainFull          int      `yaml:"retainFull,omitempty"`
	RestoreTestSchedule string   `yaml:"restoreTestSchedule,omitempty"`
}

// Defaults are the create form's defaults the template gives for its engine,
// nil for none.
func (m *Managed) Defaults() *ManagedDefaults {
	if m == nil {
		return nil
	}
	return m.Postgres
}

const (
	minArchiveTimeout = 10 * time.Second
	maxArchiveTimeout = time.Hour
	maxRetainFull     = 365
)

// Duration is a Go duration written as a string, 60s or 5m. A bare number is
// refused rather than read as seconds or nanoseconds: either guess is wrong for
// someone.
type Duration struct {
	time.Duration
}

func (d *Duration) UnmarshalYAML(value *yaml.Node) error {
	if value.Kind != yaml.ScalarNode || value.Tag != "!!str" {
		return hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("line %d: a duration is written with its unit, as 60s or 5m", value.Line)
	}
	parsed, err := time.ParseDuration(value.Value)
	if err != nil {
		return hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("line %d: %s", value.Line, err.Error())
	}
	d.Duration = parsed
	return nil
}

func (d Duration) MarshalYAML() (any, error) {
	return d.String(), nil
}

// Artifact is one file the engine mounts into the app's container: a binary its
// project publishes, downloaded by each node that runs the app, checked against
// its hash and kept by that hash. The hash is as trusted as the template, which
// the signed release info pins.
type Artifact struct {
	Name      string                         `yaml:"name"`
	Version   string                         `yaml:"version"`
	MountPath string                         `yaml:"mountPath"`
	Files     map[base.CPUArch]*ArtifactFile `yaml:"files"`
}

type ArtifactFile struct {
	URL    string `yaml:"url"`
	SHA256 string `yaml:"sha256"`
	Size   int64  `yaml:"size"`
}

const (
	mib = 1 << 20
	// MaxArtifactSize bounds what a node downloads for one artifact.
	MaxArtifactSize = 256 * mib
	// ArtifactMountDir is the one directory artifacts go into. It is HivePaaS's,
	// so an artifact can neither shadow a file of the image nor sit where a
	// volume is mounted by convention.
	ArtifactMountDir = "/opt/hivepaas/bin"
)

var (
	artifactNamePattern = regexp.MustCompile(`^[a-z0-9][a-z0-9._-]{0,31}$`)
	featureNamePattern  = regexp.MustCompile(`^[a-z0-9][a-z0-9-]{0,62}$`)
	sha256Pattern       = regexp.MustCompile(`^[0-9a-f]{64}$`)
	cronParser          = cron.NewParser(cron.Minute | cron.Hour | cron.Dom | cron.Month | cron.Dow | cron.Descriptor)
)

func (t *Template) IsManaged() bool {
	return t.Managed != nil
}

func (t *Template) FindArtifact(name string) *Artifact {
	for _, artifact := range t.Artifacts {
		if artifact != nil && artifact.Name == name {
			return artifact
		}
	}
	return nil
}

// validateFeatures checks the features a template requires. They came with
// VersionCodeV2: an older HivePaaS decodes templates strictly and would refuse the
// field, so a template naming any has to say it needs this one.
func validateFeatures(t *Template, p *problems) {
	features := t.Metadata.Requires.Features
	if len(features) > 0 && t.Metadata.Requires.VersionCode < base.VersionCodeV2 {
		p.add("metadata.requires.features needs requires.versionCode %s or later", base.VersionCodeV2)
	}
	seen := map[string]bool{}
	for _, feature := range features {
		if !featureNamePattern.MatchString(feature) {
			p.add("metadata.requires.features: %q must be lowercase letters, digits and dashes", feature)
		}
		if seen[feature] {
			p.add("metadata.requires.features: %q is listed twice", feature)
		}
		seen[feature] = true
	}
}

func validateManaged(t *Template, p *problems) {
	managed := t.Managed
	if managed == nil {
		return
	}
	if t.Metadata.Requires.VersionCode < base.VersionCodeV2 {
		p.add("managed and artifacts need requires.versionCode %s or later", base.VersionCodeV2)
	}
	contract, known := ManagedEngines[managed.Engine]
	switch {
	case !known:
		p.add("managed.engine %q is not an engine this HivePaaS has", managed.Engine)
	case managed.EngineAPI < contract.MinAPI || managed.EngineAPI > contract.MaxAPI:
		p.add("managed.engineApi %d is outside what engine %s speaks (%d to %d)",
			managed.EngineAPI, managed.Engine, contract.MinAPI, contract.MaxAPI)
	}
	for _, name := range contract.Artifacts {
		if t.FindArtifact(name) == nil {
			p.add("managed: engine %s needs the artifact %q", managed.Engine, name)
		}
	}
	if managed.Postgres != nil && managed.Engine != ManagedEnginePostgres {
		p.add("managed.postgres is for engine postgres")
	}
	validatePostgresDefaults(managed.Postgres, p)

	// An engine looks after one app. A template of components is several, and
	// which of them is the database would be one more thing to declare.
	if t.HasComponents() {
		p.add("a managed template has one app, not components")
	}
	// Artifacts are built against glibc, and Alpine's libc is musl: the binary
	// would be there and would not run. Read from names and tags, so it is a
	// heuristic, but the one that matches how images are published.
	for _, variant := range t.Variants {
		if variant != nil && strings.Contains(variant.Name, "alpine") {
			p.add("variants[%s]: a managed template runs binaries built for glibc, which Alpine does not have",
				variant.Name)
		}
	}
	for _, version := range t.Versions {
		if version == nil {
			continue
		}
		images := append([]string{version.Image}, slices.Sorted(maps.Values(version.Images))...)
		for _, image := range images {
			if _, tag, found := strings.Cut(path.Base(image), ":"); found && strings.Contains(tag, "alpine") {
				p.add("versions[%s]: %s is Alpine, and a managed template runs binaries built for glibc",
					version.Name, image)
			}
		}
	}
}

func validatePostgresDefaults(defaults *ManagedDefaults, p *problems) {
	if defaults == nil {
		return
	}
	if timeout := defaults.ArchiveTimeout.Duration; timeout != 0 &&
		(timeout < minArchiveTimeout || timeout > maxArchiveTimeout) {
		p.add("managed.postgres.archiveTimeout must be from %s to %s", minArchiveTimeout, maxArchiveTimeout)
	}
	for _, schedule := range [][2]string{
		{"baseBackupSchedule", defaults.BaseBackupSchedule}, {"restoreTestSchedule", defaults.RestoreTestSchedule},
	} {
		if schedule[1] == "" {
			continue
		}
		if _, err := cronParser.Parse(schedule[1]); err != nil {
			p.add("managed.postgres.%s %q is not a cron expression", schedule[0], schedule[1])
		}
	}
	if defaults.RetainFull < 0 || defaults.RetainFull > maxRetainFull {
		p.add("managed.postgres.retainFull must be from 1 to %d", maxRetainFull)
	}
}

func validateArtifacts(t *Template, p *problems) {
	if len(t.Artifacts) == 0 {
		return
	}
	if t.Managed == nil {
		p.add("artifacts are mounted by a managed engine, and this template has none")
	}
	names := map[string]bool{}
	mountPaths := map[string]bool{}
	for i, artifact := range t.Artifacts {
		if artifact == nil {
			p.add("artifacts[%d] is empty", i)
			continue
		}
		prefix := fmt.Sprintf("artifacts[%s]", artifact.Name)
		if !artifactNamePattern.MatchString(artifact.Name) {
			p.add("artifacts[%d]: name %q must be lowercase letters, digits, dots and dashes", i, artifact.Name)
		}
		if names[artifact.Name] {
			p.add("%s is declared twice", prefix)
		}
		names[artifact.Name] = true
		if artifact.Version == "" {
			p.add("%s.version is required", prefix)
		}
		if path.Clean(artifact.MountPath) != artifact.MountPath || path.Dir(artifact.MountPath) != ArtifactMountDir {
			p.add("%s.mountPath %q must be a file directly in %s/", prefix, artifact.MountPath, ArtifactMountDir)
		} else if mountPaths[artifact.MountPath] {
			p.add("%s.mountPath %s is another artifact's", prefix, artifact.MountPath)
		}
		mountPaths[artifact.MountPath] = true
		validateArtifactFiles(prefix, artifact.Files, p)
	}
}

func validateArtifactFiles(prefix string, files map[base.CPUArch]*ArtifactFile, p *problems) {
	if files[base.CPUArchAMD64] == nil {
		p.add("%s.files must include %s", prefix, base.CPUArchAMD64)
	}
	for _, arch := range slices.Sorted(maps.Keys(files)) {
		file := files[arch]
		where := fmt.Sprintf("%s.files.%s", prefix, arch)
		if !slices.Contains(base.AllCPUArches, arch) {
			p.add("%s: architecture %q is not one HivePaaS knows", where, arch)
		}
		if file == nil {
			p.add("%s is empty", where)
			continue
		}
		if !isHTTPSURL(file.URL) {
			p.add("%s.url must be an https URL", where)
		}
		if !sha256Pattern.MatchString(file.SHA256) {
			p.add("%s.sha256 must be 64 lowercase hex characters", where)
		}
		if file.Size <= 0 || file.Size > MaxArtifactSize {
			p.add("%s.size must be from 1 byte to %d MiB", where, MaxArtifactSize/mib)
		}
	}
}
