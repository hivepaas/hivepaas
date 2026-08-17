package commandserviceimpl

import (
	"context"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/assets"
	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

func (s *service) GetCommand(
	ctx context.Context,
	cmdType string,
	cmdName string,
) (cmd *entity.Setting, err error) {
	fileName := cmdName + ".json"
	data, err := fs.ReadFile(assets.GetTemplatesFS(), path.Join("commands", cmdType, fileName))
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	cmdTemplate := &entity.CommandTemplate{}
	if err := json.Unmarshal(data, cmdTemplate); err != nil {
		return nil, apperrors.Wrap(err)
	}

	cmdNameBase, cmdKind, found := strings.Cut(cmdName, ".")
	if !found {
		cmdNameBase = cmdName
	}

	cmd = &entity.Setting{
		ID:      gofn.Must(ulid.NewStringULID()),
		Type:    base.SettingTypeCommandTemplate,
		Kind:    cmdType,
		Name:    cmdNameBase + gofn.If(cmdKind != "", fmt.Sprintf(" (%s)", cmdKind), ""),
		Status:  base.SettingStatusActive,
		Version: entity.CurrentCommandTemplateVersion,
	}
	if err := cmd.SetData(cmdTemplate); err != nil {
		return nil, apperrors.Wrap(err)
	}

	return cmd, nil
}
