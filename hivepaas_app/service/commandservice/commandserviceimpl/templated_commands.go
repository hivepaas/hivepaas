package commandserviceimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/assets"
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

var (
	commandTemplateKindMap = map[string]base.CommandTemplateKind{
		"pg_dump":               base.CommandTemplateDatabase,
		"pg_restore":            base.CommandTemplateDatabase,
		"mysqldump":             base.CommandTemplateDatabase,
		"mysql":                 base.CommandTemplateDatabase,
		"mariadb-dump":          base.CommandTemplateDatabase,
		"mariadb":               base.CommandTemplateDatabase,
		"mongodump":             base.CommandTemplateDatabase,
		"mongorestore":          base.CommandTemplateDatabase,
		"redis-dump":            base.CommandTemplateDatabase,
		"redis-restore":         base.CommandTemplateDatabase,
		"clickhouse-dump":       base.CommandTemplateDatabase,
		"clickhouse-restore":    base.CommandTemplateDatabase,
		"sqlite-dump":           base.CommandTemplateDatabase,
		"sqlite-restore":        base.CommandTemplateDatabase,
		"sqlserver-dump":        base.CommandTemplateDatabase,
		"sqlserver-restore":     base.CommandTemplateDatabase,
		"influx-dump":           base.CommandTemplateDatabase,
		"influx-restore":        base.CommandTemplateDatabase,
		"elasticsearch-dump":    base.CommandTemplateDatabase,
		"elasticsearch-restore": base.CommandTemplateDatabase,
		"dolt-dump":             base.CommandTemplateDatabase,
		"dolt-restore":          base.CommandTemplateDatabase,
		"neon-dump":             base.CommandTemplateDatabase,
		"neon-restore":          base.CommandTemplateDatabase,
	}
)

func (s *service) GetCommand(
	ctx context.Context,
	name string,
	kind string,
) (cmd *entity.Setting, err error) {
	fileName := name
	if kind != "" {
		fileName += "." + kind
	}
	fileName += ".json"

	data, err := fs.ReadFile(assets.GetTemplatesFS(), path.Join("commands", fileName))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	cmdTemplate := &entity.CommandTemplate{}
	if err := json.Unmarshal(data, cmdTemplate); err != nil {
		return nil, apperrors.Wrap(err)
	}

	cmd = &entity.Setting{
		ID:      gofn.Must(ulid.NewStringULID()),
		Type:    base.SettingTypeCommandTemplate,
		Kind:    gofn.Coalesce(string(commandTemplateKindMap[name]), name),
		Name:    name + gofn.If(kind != "", fmt.Sprintf(" (%s)", kind), ""),
		Status:  base.SettingStatusActive,
		Version: entity.CurrentCommandTemplateVersion,
	}
	if err := cmd.SetData(cmdTemplate); err != nil {
		return nil, apperrors.Wrap(err)
	}

	return cmd, nil
}
