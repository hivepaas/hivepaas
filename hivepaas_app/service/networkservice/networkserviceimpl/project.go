package networkserviceimpl

import (
	"context"
	"errors"

	"github.com/moby/moby/api/types/network"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/dockerhelper"
)

func (s *service) ListProjectNetworks(
	ctx context.Context,
	db database.IDB,
	project *entity.Project,
) (settings []*entity.Setting, networks map[string]*network.Summary, err error) {
	settings, _, err = s.settingRepo.List(ctx, db, project.GetObjectScope(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeClusterNetwork),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
	)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}
	if len(settings) == 0 {
		return nil, nil, nil
	}

	netIDs := make([]string, 0, len(settings))
	for _, setting := range settings {
		netIDs = append(netIDs, dockerhelper.ParseID(setting.ID))
	}

	netList, err := s.dockerManager.NetworkListByIDs(ctx, netIDs)
	if err != nil {
		return nil, nil, apperrors.Wrap(err)
	}

	networks = make(map[string]*network.Summary, len(settings))
	for _, netID := range netIDs {
		net, found := gofn.FindPtr(netList.Items, func(net *network.Summary) bool {
			return net.ID == netID
		})
		if found {
			networks[net.ID] = &net
		}
	}

	return settings, networks, nil
}

func (s *service) RemoveAllProjectNetworks(
	ctx context.Context,
	db database.IDB,
	project *entity.Project,
) error {
	settings, networks, err := s.ListProjectNetworks(ctx, db, project)
	if err != nil {
		return apperrors.Wrap(err)
	}

	for _, setting := range settings {
		if setting.ObjectID != project.ID { // imported/inherited network, skip it
			continue
		}
		net := networks[dockerhelper.ParseID(setting.ID)]
		if net == nil {
			continue
		}
		_, e := s.dockerManager.NetworkRemove(ctx, net.ID)
		if e != nil && !errors.Is(e, apperrors.ErrNotFound) {
			err = errors.Join(err, e)
		}
	}
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
