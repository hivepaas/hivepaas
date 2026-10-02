package settingeventserviceimpl

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingeventservice"
)

// recordRegistryAuthRenewal records, in the event's transaction, a run of the
// registry auth renewal for the Amazon ECR credentials an edit changes how they
// sign in: a key auth's keys, an ECR credential's key auth, role or registry,
// or either turned back on. Their services get a token from the keys as they
// are now, rather than at the next scheduled run.
func (s *service) recordRegistryAuthRenewal(
	ctx context.Context,
	db database.IDB,
	event *settingeventservice.UpdateEvent,
) error {
	setting, old := event.Setting, event.OldSetting
	if old == nil || setting.Status != base.SettingStatusActive {
		return nil
	}
	turnedOn := old.Status != base.SettingStatusActive

	var task *entity.Task
	var err error
	switch setting.Type { //nolint:exhaustive // only these two sign in to ECR
	case base.SettingTypeKeyAuth:
		if !turnedOn && setting.Data == old.Data {
			return nil
		}
		task, err = s.registryAuthService.RecordRenewalForKeyAuth(ctx, db, setting.ID)
	case base.SettingTypeRegistryAuth:
		auth, e := setting.AsRegistryAuth()
		if e != nil {
			return hperrors.Wrap(e)
		}
		if auth.Kind != base.RegistryAuthKindAWSECR {
			return nil
		}
		oldAuth, e := old.AsRegistryAuth()
		if e != nil {
			return hperrors.Wrap(e)
		}
		if !turnedOn && auth.SameECRKeys(oldAuth) {
			return nil
		}
		task, err = s.registryAuthService.RecordRenewal(ctx, db, []string{setting.ID})
	default:
		return nil
	}
	if err != nil {
		return hperrors.Wrap(err)
	}
	if task != nil {
		event.Tasks = append(event.Tasks, task)
	}
	return nil
}
