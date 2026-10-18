package schedjobuc

import (
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
)

// refuseManagedJob refuses to edit, pause or delete a job an app's managed
// engine made: its schedule and its command are the engine's, changed through
// the engine's settings. Running one now is not refused.
func refuseManagedJob(setting *entity.Setting) error {
	job, err := setting.AsSchedJob()
	if err != nil {
		return hperrors.Wrap(err)
	}
	if job.IsManaged() {
		return hperrors.Wrap(hperrors.ErrAppManagedFieldLocked).WithParam("Field", "schedJob").
			ExposeParams("Field")
	}
	return nil
}
