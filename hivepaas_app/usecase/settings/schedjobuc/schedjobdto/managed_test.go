package schedjobdto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

func TestAManagedJobSaysWhoManagesIt(t *testing.T) {
	for want, job := range map[string]*entity.SchedJob{
		"managed-1": {JobType: base.SchedJobTypeContainerCommand, ManagedBy: entity.ObjectID{ID: "managed-1"}},
		"":          {JobType: base.SchedJobTypeContainerCommand},
	} {
		setting := &entity.Setting{ID: "job-1", Type: base.SettingTypeSchedJob, Version: entity.CurrentSchedJobVersion}
		setting.MustSetData(job)

		resp, err := TransformSchedJob(setting, entity.NewRefObjects(), true)

		assert.NoError(t, err)
		if want == "" {
			assert.Nil(t, resp.ManagedBy)
		} else {
			assert.Equal(t, &basedto.ObjectIDResp{ID: want}, resp.ManagedBy)
		}
	}
}
