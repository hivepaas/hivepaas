package settinginitserviceimpl

import (
	"context"
	"errors"
	"time"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

const (
	managedArtifactsJobName  = "Managed artifacts job"
	managedArtifactsInterval = timeutil.Duration(10 * time.Minute)
)

// initDefaultManagedArtifacts makes the job that has every app's artifacts on
// the nodes its tasks may start on, unless there is one. With no managed app a
// run does nothing, so it is made on every installation, active. It has no
// retry and no notification: the next run is minutes away.
func (s *service) initDefaultManagedArtifacts(ctx context.Context, db database.Tx, timeNow time.Time) error {
	_, err := s.settingRepo.GetSingle(ctx, db, entity.NewObjectScopeGlobal(), base.SettingTypeSchedJob, false,
		bunex.SelectWhere("setting.kind = ?", base.SchedJobTypeManagedArtifacts))
	if err == nil {
		return nil
	}
	if !errors.Is(err, hperrors.ErrNotFound) {
		return hperrors.Wrap(err)
	}
	jobSetting := &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeGlobal,
		Type:      base.SettingTypeSchedJob,
		Kind:      string(base.SchedJobTypeManagedArtifacts),
		Status:    base.SettingStatusActive,
		Name:      managedArtifactsJobName,
		Version:   entity.CurrentSchedJobVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	jobSetting.MustSetData(&entity.SchedJob{
		JobType:  base.SchedJobTypeManagedArtifacts,
		Schedule: &entity.SchedJobSchedule{Interval: managedArtifactsInterval, InitialTime: timeNow},
	})
	return hperrors.Wrap(s.settingRepo.Insert(ctx, db, jobSetting))
}
