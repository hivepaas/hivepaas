package getstartedserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
)

func (s *service) Checklist(
	ctx context.Context,
	db database.IDB,
	hasTwoFactor bool,
) (*getstartedservice.Checklist, error) {
	dashboardCert, err := s.DashboardCert(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	githubApp, err := s.hasGithubApp(ctx, db)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	return &getstartedservice.Checklist{
		DashboardCert: *dashboardCert,
		TwoFactor:     doneIf(hasTwoFactor),
		GithubApp:     doneIf(githubApp),
	}, nil
}

func doneIf(done bool) getstartedservice.Item {
	if done {
		return getstartedservice.Item{Status: getstartedservice.ItemStatusDone}
	}
	return getstartedservice.Item{Status: getstartedservice.ItemStatusTodo}
}

func (s *service) hasGithubApp(ctx context.Context, db database.IDB) (bool, error) {
	settings, _, err := s.settingRepo.List(ctx, db, entity.NewObjectScopeGlobal(), nil,
		bunex.SelectWhere("setting.type = ?", base.SettingTypeGithubApp),
		bunex.SelectWhere("setting.status = ?", base.SettingStatusActive),
		bunex.SelectLimit(1),
	)
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	return len(settings) > 0, nil
}
