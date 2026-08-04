package apphelper

import (
	"strings"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func CalcMountSubpath(
	app *entity.App,
	pathTemplate string, // something like `project_data/{{project}}/{{env}}/{{app}}`
) string {
	path := strings.NewReplacer("{{project}}", app.Project.Key, "{{env}}", app.ProjectEnv.Key,
		"{{app}}", app.Key).Replace(pathTemplate)
	path = strings.ReplaceAll(path, "//", "/")
	return path
}
