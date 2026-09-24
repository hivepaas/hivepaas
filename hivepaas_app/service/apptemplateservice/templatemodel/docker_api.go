package templatemodel

import (
	"fmt"
	"strings"

	"gopkg.in/yaml.v3"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

// dockerAPIPath is where a template gives its app the Docker API: the spec's own
// path, because a template's app block is a spec document.
var dockerAPIPath = []string{"settings", "dockerApi"}

// DockerAPI is the Docker API provisioning this template gives its app, and is
// nil for a template that gives none. Like capabilities, it is read from the
// template rather than from a render, and a version may not override it: what
// the store and the deploy dialog show is what is granted.
func (t *Template) DockerAPI() (*entity.AppDockerAPISettings, error) {
	access, problem := dockerAPIIn(t.App)
	if problem != "" {
		return nil, hperrors.Wrap(hperrors.ErrAppTemplateInvalid).WithExtraDetail("%s", problem)
	}
	return access, nil
}

// RequiresDockerAPI reports whether any app of the template is given the Docker
// API. One whose block cannot be read counts, as for capabilities.
func (t *Template) RequiresDockerAPI() bool {
	for _, app := range t.appTrees() {
		access, problem := dockerAPIIn(app)
		if problem != "" || access != nil {
			return true
		}
	}
	return false
}

// dockerAPIIn reads the block out of an app tree, and says what is wrong with it
// instead when it cannot. Both are empty for a tree that gives none.
func dockerAPIIn(app map[string]any) (*entity.AppDockerAPISettings, string) {
	where := "app." + strings.Join(dockerAPIPath, ".")
	settings, _ := app[dockerAPIPath[0]].(map[string]any)
	body, found := settings[dockerAPIPath[1]]
	if !found {
		return nil, ""
	}
	encoded, err := yaml.Marshal(body)
	if err != nil {
		return nil, fmt.Sprintf("%s: %s", where, err.Error())
	}
	if strings.Contains(string(encoded), placeholderMark) {
		return nil, where + ": a placeholder here would make what is granted depend on what a person fills in"
	}
	access, err := specmodel.DockerAPIIn(settings)
	if err != nil {
		return nil, fmt.Sprintf("%s: %s", where, err.Error())
	}
	return access, ""
}

// validateDockerAPI checks the block a template declares, and refuses one a
// version declares: the Docker API an app is given is declared once.
func validateDockerAPI(t *Template, p *problems) {
	for i, app := range t.appTrees() {
		where := "app"
		if t.HasComponents() {
			where = fmt.Sprintf("components[%s].app", t.Components[i].Name)
		}
		access, problem := dockerAPIIn(app)
		if problem != "" {
			p.add("%s", problem)
		}
		if problem = specmodel.DockerAPIProblem(access); problem != "" {
			p.add("%s.%s", where, problem)
		}
		if access.IsHostMode() {
			p.add("%s.%s", where, specmodel.HostModeFromTemplate)
		}
	}
	for _, version := range t.Versions {
		if version == nil || version.Override == nil {
			continue
		}
		overridden, problem := dockerAPIIn(version.Override.App)
		if problem != "" || overridden != nil {
			p.add("versions[%s].override.app.%s: the Docker API is declared once, in app",
				version.Name, strings.Join(dockerAPIPath, "."))
		}
	}
}
