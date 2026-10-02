package settinginitserviceimpl

import (
	"context"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

const (
	registryAuthRenewalSettingName = "Registry auth renewal settings"
	registryAuthRenewalJobName     = "Registry auth renewal job"
	registryAuthRenewalMaxRetry    = 1
	registryAuthRenewalRetryDelay  = timeutil.Duration(time.Second * 60)
)

// initDefaultRegistryAuthRenewal makes the setting and the scheduled job that
// renew the Amazon ECR tokens Swarm keeps. With no ECR credential a run does
// nothing, so it is made on every installation, active.
func (s *service) initDefaultRegistryAuthRenewal(
	ctx context.Context,
	db database.IDB,
	timeNow time.Time,
) (err error) {
	renewalSetting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Type:        base.SettingTypeRegistryAuthRenewal,
		Status:      base.SettingStatusActive,
		Name:        registryAuthRenewalSettingName,
		Inheritable: false,
		Version:     entity.CurrentRegistryAuthRenewalVersion,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}
	renewal := entity.NewRegistryAuthRenewal(timeNow)
	renewalSetting.MustSetData(renewal)

	jobSetting := &entity.Setting{
		ID:          gofn.Must(ulid.NewStringULID()),
		Type:        base.SettingTypeSchedJob,
		Kind:        string(base.SchedJobTypeRegistryAuthRenewal),
		Status:      base.SettingStatusActive,
		Name:        registryAuthRenewalJobName,
		Inheritable: false,
		Version:     entity.CurrentSchedJobVersion,
		CreatedAt:   timeNow,
		UpdatedAt:   timeNow,
	}
	jobSetting.MustSetData(&entity.SchedJob{
		JobType:       base.SchedJobTypeRegistryAuthRenewal,
		Schedule:      &renewal.Schedule,
		TargetSetting: entity.ObjectID{ID: renewalSetting.ID},
		MaxRetry:      registryAuthRenewalMaxRetry,
		RetryDelay:    registryAuthRenewalRetryDelay,
		Notification:  renewal.Notification,
	})

	err = s.settingRepo.InsertMulti(ctx, db, []*entity.Setting{renewalSetting, jobSetting})
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
