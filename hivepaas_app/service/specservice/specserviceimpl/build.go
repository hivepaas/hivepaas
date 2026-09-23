package specserviceimpl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/specservice/specmodel"
)

type blockBuilder func(ctx context.Context, state *buildState) error

// buildState is what the builders of one BuildApp call share.
type buildState struct {
	db       database.IDB
	req      *specservice.BuildAppReq
	settings []*entity.Setting
}

// builders is the registry of what can be built. It has to cover exactly
// specmodel.BuildableBlocks, which TestBuilderRegistryCoversEveryBuildableBlock
// holds it to.
//
// TODO: app templates phase 3 - more blocks. See
// docs/superpowers/specs/2026-09-17-app-templates-design.md §12.
func (s *service) builders() map[specmodel.Block]blockBuilder {
	return map[specmodel.Block]blockBuilder{
		specmodel.BlockDeploymentSource:     s.buildSource,
		specmodel.BlockDeploymentStorage:    s.buildStorage,
		specmodel.BlockContainerHealthcheck: s.buildHealthcheck,
		specmodel.BlockContainerInit:        s.buildInit,
		specmodel.BlockDeploymentResources:  s.buildResources,
		specmodel.BlockDeploymentNetworks:   s.buildNetworks,
		specmodel.BlockSettingsKind:         s.buildKind,
		specmodel.BlockSettingsEnvVars:      s.buildEnvVars,
		specmodel.BlockSettingsSecrets:      s.buildSecrets,
		specmodel.BlockSettingsConfigFiles:  s.buildConfigFiles,
		specmodel.BlockSettingsRouting:      s.buildRouting,
		specmodel.BlockContainer:            s.buildContainer,
		specmodel.BlockDeploymentService:    s.buildService,
		specmodel.BlockSettings:             s.buildImportedSettings,
	}
}

func (s *service) BuildApp(
	ctx context.Context,
	db database.IDB,
	req *specservice.BuildAppReq,
) (*specservice.BuildAppResp, error) {
	check, blocks := specmodel.CheckBuildable, specmodel.PresentBlocks
	if req.Import {
		check, blocks = specmodel.CheckImportable, specmodel.ImportBlocks
	}
	if err := check(req.Doc); err != nil {
		return nil, hperrors.Wrap(err)
	}

	builders := s.builders()
	state := &buildState{db: db, req: req}
	for _, block := range blocks(req.Doc) {
		build, found := builders[block]
		if !found {
			return nil, hperrors.Wrap(hperrors.ErrSpecBlockUnsupported).WithExtraDetail("%s", block)
		}
		if err := build(ctx, state); err != nil {
			return nil, hperrors.Wrap(err)
		}
	}
	return &specservice.BuildAppResp{Settings: state.settings}, nil
}

// addSetting creates a setting row the way the settings usecases create one.
//
// Every setting type carries its own current version. They all happen to be 1
// today, which is what makes the argument look constant; passing each type's own
// constant is what keeps a bump to one of them from silently writing rows
// claiming another type's version.
//
//nolint:unparam
func (state *buildState) addSetting(
	typ base.SettingType,
	version int,
	inheritable bool,
	data entity.SettingData,
) error {
	return state.addNamedSetting(typ, "", version, inheritable, data)
}

// addNamedSetting creates a setting that carries a name: the key a collection
// type is found and referred to by - a secret's key, a config file's name.
func (state *buildState) addNamedSetting(
	typ base.SettingType,
	name string,
	version int,
	inheritable bool,
	data entity.SettingData,
) error {
	setting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Scope:       base.ObjectScopeApp,
		ObjectID:    state.req.App.ID,
		Type:        typ,
		Name:        name,
		Status:      base.SettingStatusActive,
		Inheritable: inheritable,
		Version:     version,
		UpdateVer:   1,
		CreatedAt:   state.req.TimeNow,
		UpdatedAt:   state.req.TimeNow,
	}
	if err := setting.SetData(data); err != nil {
		return hperrors.Wrap(err)
	}
	state.settings = append(state.settings, setting)
	return nil
}

// decodeBlock decodes a settings-shaped block into its entity. Unknown fields
// are refused: a spec field the entity does not have is a field that would not
// be applied.
func decodeBlock(block specmodel.Block, body any, out any) error {
	encoded, err := json.Marshal(body)
	if err != nil {
		return invalidBlock(block, "%s", err.Error())
	}
	decoder := json.NewDecoder(bytes.NewReader(encoded))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(out); err != nil {
		return invalidBlock(block, "%s", err.Error())
	}
	return nil
}

func invalidBlock(block specmodel.Block, format string, args ...any) error {
	return hperrors.Wrap(hperrors.ErrSpecBlockInvalid).WithExtraDetail("%s: %s", block, fmt.Sprintf(format, args...))
}
