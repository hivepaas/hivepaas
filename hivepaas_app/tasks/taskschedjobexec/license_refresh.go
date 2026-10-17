package taskschedjobexec

import (
	"context"
	"fmt"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

// refreshLicense asks the license portal for a newer key than the one stored,
// and stores what it answers. The portal is asked before the license setting is
// locked, and the answer applied in the run's transaction.
//
// A check that failed fails the run, so its one retry asks again. What it
// learned is kept all the same: the run's transaction is committed whatever the
// run returns. It sends no notification either way: a check that failed says so
// on System > License.
func (e *Executor) refreshLicense(ctx context.Context, db database.Tx, run *jobRun,
	schedJob *entity.SchedJob) (*jobResult, error) {
	if run.refObjects.RefSettings[schedJob.TargetSetting.ID] == nil {
		return nil, hperrors.NewNotFound("License settings")
	}
	fetched, err := e.licenseService.FetchRefresh(ctx, &licenseservice.FetchRefreshReq{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	applied, err := e.licenseService.ApplyRefresh(ctx, db, fetched)
	if err != nil {
		return nil, hperrors.Wrap(err)
	}
	_ = run.execData.LogStore.Add(ctx, tasklog.NewOutFrame(licenseRefreshLog(applied), tasklog.TsNow))
	if applied.Result == base.LicenseRefreshFailed {
		return nil, hperrors.Wrap(applied.Err)
	}
	return &jobResult{skipNotification: true}, nil
}

// licenseRefreshLog is the line a run writes: what the portal answered, and what
// changed. It names the key by plan and dates only: a key carries the
// customer's email.
func licenseRefreshLog(applied *licenseservice.ApplyRefreshResp) string {
	switch applied.Result {
	case base.LicenseRefreshNewer:
		if state := applied.State; state != nil && state.Key != nil {
			return fmt.Sprintf("The portal had a newer license key, now installed: %s, valid until %s.\n",
				state.Key.Plan.DisplayName(), state.Key.ExpiresAt.UTC().Format(time.DateOnly))
		}
		return "The portal had a newer license key, now installed.\n"
	case base.LicenseRefreshMoved:
		if state := applied.State; state != nil {
			return fmt.Sprintf("The license key was moved to another installation on %s. It holds here until %s.\n",
				state.MovedAt.UTC().Format(time.DateOnly), state.GraceEndsAt.UTC().Format(time.DateOnly))
		}
		return "The license key was moved to another installation, as was already known.\n"
	case base.LicenseRefreshCurrent:
		return "The license key stored is the newest.\n"
	case base.LicenseRefreshUnknown:
		return "The portal did not issue the license key stored, and has no newer one for it.\n"
	case base.LicenseRefreshNone:
		return "No license key is stored, so there is nothing to ask the portal with.\n"
	case base.LicenseRefreshDisabled:
		return "The daily license refresh is switched off in the config: refresh_disabled under [license] " +
			"in hivepaas.toml, or HP_LICENSE_REFRESH_DISABLED.\n"
	case base.LicenseRefreshFailed:
		return "The license refresh failed: " + hperrors.GetErrorDetail(applied.Err, "") + "\n"
	}
	return fmt.Sprintf("The license refresh answered %q.\n", applied.Result)
}
