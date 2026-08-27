package startupserviceimpl

import (
	"context"
	"sync"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

type startupData struct {
	mu sync.RWMutex

	HivePaaSServiceSetting *entity.Setting
}

var (
	gStartupData = &startupData{
		mu: sync.RWMutex{},
	}
)

func (s *service) LoadHivePaaSServiceSetting(
	ctx context.Context,
) (*entity.Setting, error) {
	if gStartupData == nil {
		panic("startup service shutdown")
	}

	gStartupData.mu.Lock()
	defer gStartupData.mu.Unlock()

	if gStartupData.HivePaaSServiceSetting != nil {
		return gStartupData.HivePaaSServiceSetting, nil
	}

	setting, err := s.settingRepo.GetSingle(ctx, s.db, nil, base.SettingTypeHivePaaSService, true)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	gStartupData.HivePaaSServiceSetting = setting

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
