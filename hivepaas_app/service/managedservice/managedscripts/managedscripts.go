// Package managedscripts is what a managed database's template provides that
// its engine owns - its scripts and its artifacts - as a newer revision offers
// them and as the app has them, and whether the one can replace the other. See
// docs/superpowers/specs/2026-10-21-managed-scripts-upgrade-design.md.
package managedscripts

import (
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"maps"
	"slices"
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templatemodel"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/apptemplateservice/templaterender"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/artifactservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice/walgkit"
)

// The states of what a newer revision offers.
const (
	StateUpToDate        = "up-to-date"
	StateUpdateAvailable = "update-available"
	// StateIncompatible is a revision that cannot be applied to the app: another
	// engine, another engine API, or scripts that are not scripts.
	StateIncompatible = "incompatible"
	// StateUnavailable is an app the revision has nothing for: its template, or
	// its version, is no longer there, or the template is no longer a managed one.
	StateUnavailable = "unavailable"
)

// What changes, of a script by its path or an artifact by its name.
const (
	KindScript   = "script"
	KindArtifact = "artifact"

	ChangeAdded   = "added"
	ChangeChanged = "changed"
	ChangeRemoved = "removed"
)

// configFilesBlock is the settings block a template's config files are in.
const configFilesBlock = "configFiles"

// Script is one of the engine's scripts: a config file mounted under the
// script directory.
type Script struct {
	// Name is the config file's, which its setting is called.
	Name    string
	Path    string
	Mode    fileutil.FileMode
	Content string
	SHA256  string
}

// Set is the engine's scripts, by path, and the artifacts, by name.
type Set struct {
	Scripts   []*Script
	Artifacts *entity.AppArtifactsSettings
}

type Change struct {
	Kind string
	// Name is a script's path, or an artifact's name.
	Name   string
	Change string
	// From and To are an artifact's versions.
	From, To string
}

// Offer is what a revision offers an app, and whether it may be applied.
type Offer struct {
	State string
	// Reason says why a revision is incompatible or unavailable.
	Reason    string
	Available *Set
	Changes   []*Change
}

// FromTemplate is what the template offers an app of that version: the config
// files its version's tree mounts under the script directory, as an app made
// from it has them, and its artifacts.
func FromTemplate(tmpl *templatemodel.Template, version string) (*Set, error) {
	tree, err := templaterender.VersionTree(tmpl, version)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	app, _ := tree.(map[string]any)
	settings, _ := app["settings"].(map[string]any)
	static, err := templaterender.Static(settings[configFilesBlock])
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	entries, _ := static.(map[string]any)
	set := &Set{Artifacts: artifactservice.FromTemplate(tmpl.Artifacts)}
	for _, name := range slices.Sorted(maps.Keys(entries)) {
		script, err := scriptOf(name, entries[name])
		if err != nil {
			return nil, err
		}
		if script != nil {
			set.Scripts = append(set.Scripts, script)
		}
	}
	sortScripts(set.Scripts)
	return set, nil
}

// scriptOf is a template's config file entry as a script, nil for one that is
// not mounted under the script directory.
func scriptOf(name string, entry any) (*Script, error) {
	fields, _ := entry.(map[string]any)
	swarmRef, _ := fields["swarmRef"].(map[string]any)
	file, _ := swarmRef["file"].(map[string]any)
	path, _ := file["name"].(string)
	if !isScriptPath(path) {
		return nil, nil //nolint:nilnil // not a script, and no error
	}
	body := &entity.ConfigFile{Name: name}
	body.Content, _ = fields["content"].(string)
	body.Base64, _ = fields["base64"].(bool)
	script := &Script{Name: name, Path: path}
	if mode, ok := file["mode"]; ok {
		parsed, err := fileutil.ParseFileMode(fmt.Sprint(mode))
		if err != nil {
			return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
				WithExtraDetail("%s: mode %v is not a file mode", name, mode)
		}
		script.Mode = parsed
	}
	content, err := contentOf(body)
	if err != nil {
		return nil, err
	}
	script.Content, script.SHA256 = content, sha256Hex(content)
	return script, nil
}

// FromSettings is what an app has: its config files mounted under the script
// directory, and its artifacts.
func FromSettings(settings []*entity.Setting) (*Set, error) {
	files := map[string]*entity.ConfigFile{}
	set := &Set{}
	for _, setting := range settings {
		if !setting.IsActive() {
			continue
		}
		switch setting.Type { //nolint:exhaustive // what holds scripts and artifacts
		case base.SettingTypeConfigFile:
			file, err := setting.AsConfigFile()
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			files[setting.ID] = file
		case base.SettingTypeAppArtifacts:
			artifacts, err := setting.AsAppArtifactsSettings()
			if err != nil {
				return nil, hperrors.Wrap(err)
			}
			set.Artifacts = artifacts
		}
	}
	for _, setting := range settings {
		if !setting.IsActive() || setting.Type != base.SettingTypeAppSettingMount {
			continue
		}
		mount, err := setting.AsAppSettingMount()
		if err != nil {
			return nil, hperrors.Wrap(err)
		}
		file := files[mount.Source.ID]
		if file == nil {
			continue
		}
		for _, target := range mount.Files {
			if !isScriptPath(target.Path) {
				continue
			}
			content, err := contentOf(file)
			if err != nil {
				return nil, err
			}
			set.Scripts = append(set.Scripts, &Script{Name: file.Name, Path: target.Path, Mode: target.Mode,
				Content: content, SHA256: sha256Hex(content)})
		}
	}
	sortScripts(set.Scripts)
	return set, nil
}

func isScriptPath(path string) bool {
	return strings.HasPrefix(path, walgkit.ScriptDir+"/")
}

// contentOf is a config file's content as its file holds it.
func contentOf(file *entity.ConfigFile) (string, error) {
	if !file.Base64 {
		return file.Content, nil
	}
	decoded, err := base64.StdEncoding.DecodeString(file.Content)
	if err != nil {
		return "", hperrors.Wrap(hperrors.ErrAppTemplateInvalid).
			WithExtraDetail("%s: its content is not base64", file.Name)
	}
	return string(decoded), nil
}

func sortScripts(scripts []*Script) {
	slices.SortFunc(scripts, func(a, b *Script) int { return strings.Compare(a.Path, b.Path) })
}

func sha256Hex(content string) string {
	sum := sha256.Sum256([]byte(content))
	return hex.EncodeToString(sum[:])
}

// Fingerprint is one hash of the scripts - path, mode, content - and the
// artifacts - name, version, mount path, each build's hash - in a fixed order.
func (s *Set) Fingerprint() string {
	var b strings.Builder
	for _, script := range s.Scripts {
		fmt.Fprintf(&b, "script\t%s\t%o\t%s\n", script.Path, uint32(script.Mode), script.SHA256)
	}
	for _, artifact := range s.artifacts() {
		fmt.Fprintf(&b, "artifact\t%s\t%s\t%s", artifact.Name, artifact.Version, artifact.MountPath)
		for _, arch := range slices.Sorted(maps.Keys(artifact.Files)) {
			if file := artifact.Files[arch]; file != nil {
				fmt.Fprintf(&b, "\t%s=%s", arch, file.SHA256)
			}
		}
		b.WriteString("\n")
	}
	return sha256Hex(b.String())
}

// artifacts are the set's artifacts by name.
func (s *Set) artifacts() []*entity.AppArtifact {
	if s.Artifacts == nil {
		return nil
	}
	out := slices.DeleteFunc(slices.Clone(s.Artifacts.Artifacts), func(a *entity.AppArtifact) bool { return a == nil })
	slices.SortFunc(out, func(a, b *entity.AppArtifact) int { return strings.Compare(a.Name, b.Name) })
	return out
}

// Diff is what replacing current with available changes: each script by its
// path, then each artifact by its name.
func Diff(current, available *Set) []*Change {
	var changes []*Change
	have := map[string]*Script{}
	for _, script := range current.Scripts {
		have[script.Path] = script
	}
	offered := map[string]bool{}
	for _, script := range available.Scripts {
		offered[script.Path] = true
		was := have[script.Path]
		switch {
		case was == nil:
			changes = append(changes, &Change{Kind: KindScript, Name: script.Path, Change: ChangeAdded})
		case was.SHA256 != script.SHA256 || was.Mode != script.Mode:
			changes = append(changes, &Change{Kind: KindScript, Name: script.Path, Change: ChangeChanged})
		}
	}
	for _, script := range current.Scripts {
		if !offered[script.Path] {
			changes = append(changes, &Change{Kind: KindScript, Name: script.Path, Change: ChangeRemoved})
		}
	}
	slices.SortStableFunc(changes, func(a, b *Change) int { return strings.Compare(a.Name, b.Name) })

	var artifactChanges []*Change
	haveArtifacts := map[string]*entity.AppArtifact{}
	for _, artifact := range current.artifacts() {
		haveArtifacts[artifact.Name] = artifact
	}
	offeredArtifacts := map[string]bool{}
	for _, artifact := range available.artifacts() {
		offeredArtifacts[artifact.Name] = true
		was := haveArtifacts[artifact.Name]
		switch {
		case was == nil:
			artifactChanges = append(artifactChanges, &Change{Kind: KindArtifact, Name: artifact.Name,
				Change: ChangeAdded, To: artifact.Version})
		case artifactFingerprint(was) != artifactFingerprint(artifact):
			artifactChanges = append(artifactChanges, &Change{Kind: KindArtifact, Name: artifact.Name,
				Change: ChangeChanged, From: was.Version, To: artifact.Version})
		}
	}
	for _, artifact := range current.artifacts() {
		if !offeredArtifacts[artifact.Name] {
			artifactChanges = append(artifactChanges, &Change{Kind: KindArtifact, Name: artifact.Name,
				Change: ChangeRemoved, From: artifact.Version})
		}
	}
	return append(changes, artifactChanges...)
}

func artifactFingerprint(artifact *entity.AppArtifact) string {
	return (&Set{Artifacts: &entity.AppArtifactsSettings{Artifacts: []*entity.AppArtifact{artifact}}}).Fingerprint()
}

// Evaluate is what the template offers an app of that version whose engine
// setting is managed and whose scripts are current: whether it can be applied,
// and what it changes.
func Evaluate(managed *entity.AppManagedSettings, tmpl *templatemodel.Template, version string,
	current *Set) *Offer {
	switch {
	case tmpl.Managed == nil:
		return &Offer{State: StateUnavailable, Reason: "the template is not a managed one any more"}
	case string(tmpl.Managed.Engine) != managed.Engine:
		return &Offer{State: StateIncompatible,
			Reason: fmt.Sprintf("the template's engine is %s, this database's %s", tmpl.Managed.Engine, managed.Engine)}
	case tmpl.Managed.EngineAPI != managed.EngineAPI:
		return &Offer{State: StateIncompatible, Reason: fmt.Sprintf("the template's scripts are for engine API %d, "+
			"this database's are for %d, and there is no migration between them", tmpl.Managed.EngineAPI,
			managed.EngineAPI)}
	}
	if tmpl.FindVersion(version) == nil {
		return &Offer{State: StateUnavailable, Reason: fmt.Sprintf("the template has no version %s any more", version)}
	}
	available, err := FromTemplate(tmpl, version)
	if err != nil {
		return &Offer{State: StateIncompatible, Reason: "its scripts cannot be read: " +
			hperrors.GetErrorDetail(err, "")}
	}
	if len(available.Scripts) == 0 {
		return &Offer{State: StateIncompatible, Reason: "the template ships no scripts"}
	}
	offer := &Offer{State: StateUpToDate, Available: available}
	if available.Fingerprint() != current.Fingerprint() {
		offer.State, offer.Changes = StateUpdateAvailable, Diff(current, available)
	}
	return offer
}
