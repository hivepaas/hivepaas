package startupserviceimpl

import (
	"context"
	"sync"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/entity"
)

type startupData struct {
	mu sync.RWMutex

	LocalPaaSServiceSetting *entity.Setting
}

var (
	gStartupData = &startupData{
		mu: sync.RWMutex{},
	}
)

func (s *service) LoadLocalPaaSServiceSetting(
	ctx context.Context,
) (*entity.Setting, error) {
	if gStartupData == nil {
		panic("startup service shutdown")
	}

	gStartupData.mu.Lock()
	defer gStartupData.mu.Unlock()

	if gStartupData.LocalPaaSServiceSetting != nil {
		return gStartupData.LocalPaaSServiceSetting, nil
	}

	setting, err := s.settingRepo.GetSingle(ctx, s.db, nil, base.SettingTypeLocalPaaSService, true)
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	gStartupData.LocalPaaSServiceSetting = setting

	return setting, nil
}

func (s *service) Shutdown() {
	if gStartupData == nil {
		panic("startup service shutdown")
	}

	gStartupData.mu.Lock()
	defer gStartupData.mu.Unlock()
	gStartupData = nil
}
