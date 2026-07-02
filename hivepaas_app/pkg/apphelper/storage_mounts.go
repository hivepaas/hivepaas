package apphelper

import (
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func CalcMountSubpath(
	project *entity.Project,
	app *entity.App,
	pathTemplate string, // something like `project_data/{{project}}/{{env}}/{{app}}`
) string {
	path := strings.NewReplacer("{{project}}", project.Key, "{{env}}", app.Env, "{{app}}", app.LocalKey).
		Replace(pathTemplate)
	path = strings.ReplaceAll(path, "//", "/")
	return path
}
