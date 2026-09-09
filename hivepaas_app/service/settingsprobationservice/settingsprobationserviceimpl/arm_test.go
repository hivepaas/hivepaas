package settingsprobationserviceimpl

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingsprobationservice"
)

// noTasksRepo answers "nothing on trial" and nothing else.
//
// The embedded nil interface is deliberate: any method this test does not expect
// Arm to call panics instead of quietly returning a zero value.
type noTasksRepo struct {
	repository.TaskRepo
}

func (noTasksRepo) ListByTarget(_ context.Context, _ database.IDB, _ string, _ *basedto.Paging,
	_ ...bunex.SelectQueryOption) ([]*entity.Task, *basedto.PagingMeta, error) {
	return nil, nil, nil
}

// A settle delay is a correctness bound before it is a caller's preference. A
// caller that passes nothing - or something shorter than the bound - must still
// get it, or a change becomes confirmable before the configuration it applied is
// the one answering, and the confirmation vouches for the configuration it
// replaced.
func TestArmFloorsTheSettleDelay(t *testing.T) {
	s := &service{taskRepo: noTasksRepo{}}
	setting := &entity.Setting{ID: "s1", Type: base.SettingTypeAppRouting, UpdateVer: 7}

	for name, given := range map[string]time.Duration{
		"unset":      0,
		"too short":  time.Second,
		"negative":   -time.Hour,
		"long stays": entity.SettingsProbationAppRestartSettleDelay,
	} {
		t.Run(name, func(t *testing.T) {
			out := &settingsprobationservice.ArmResult{}
			before := time.Now().UTC()
			err := s.Arm(context.Background(), database.Tx{}, &basedto.Auth{}, &settingsprobationservice.ArmReq{
				AppID:       "app-1",
				Setting:     setting,
				Snapshot:    entity.SettingSnapshot{Data: "{}", Version: 1},
				Window:      10 * time.Minute,
				SettleDelay: given,
			}, out, func(*entity.Task) {})
			assert.NoError(t, err)

			args, err := out.Probation.ArgsAsSettingsRevert()
			assert.NoError(t, err)
			if args == nil {
				t.Fatal("Arm produced no probation args")
			}

			want := max(given, entity.SettingsProbationSettleDelay)
			assert.False(t, args.ConfirmableFrom().Before(before.Add(want)),
				"a %v settle delay must not be confirmable before %v has passed", given, want)
		})
	}
}

// A confirmation cancels the trial, and after that there is no getting the old
// configuration back. Granting one without checking that the change is still what
// is running is exactly what the rollback case turns into a silent divergence, so
// a caller that forgot the check has to fail here rather than succeed quietly.
func TestConfirmRefusesWithoutALivenessCheck(t *testing.T) {
	s := &service{}

	err := s.Confirm(context.Background(), &basedto.Auth{}, &settingsprobationservice.AnswerReq{
		AppID:       "app-1",
		SettingType: base.SettingTypeTraefikConfig,
	})

	assert.ErrorIs(t, err, hperrors.ErrInternal)
}
