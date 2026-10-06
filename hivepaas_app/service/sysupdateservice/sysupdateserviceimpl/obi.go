package sysupdateserviceimpl

import (
	"context"
	"errors"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/obi"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/systemappservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/sysupdateservice"
)

// obiChange is what an update does to OBI. OBI is no service: each node's
// agent runs it, the image of the agent's own release, so the update moves it
// by moving the agent - and what runs now is this release's, the release of
// the binary the update runs in.
func (s *service) obiChange(
	ctx context.Context,
	db database.IDB,
	target *base.ReleaseInfo,
) (*sysupdateservice.ComponentChange, error) {
	deployed, err := s.obiDeployed(ctx, db)
	if err != nil {
		return nil, err
	}
	return planComponent(base.HivepaasOBIKey, currentOBIImage(), deployed, target.OBIImage, false,
		target.BlockMajorUpgrade), nil
}

// currentOBIImage is the OBI this release's agents run.
func currentOBIImage() string {
	if image := systemappservice.CurrentRelease().OBIImage; image != "" {
		return image
	}
	return obi.DefaultImage
}

// obiDeployed is whether the nodes run OBI: apps' routes and calls on, the
// logs stored, a node chosen.
func (s *service) obiDeployed(ctx context.Context, db database.IDB) (bool, error) {
	setting, err := s.settingRepo.GetSingle(ctx, db, entity.NewObjectScopeGlobal(), base.SettingTypeLogging, true)
	if errors.Is(err, hperrors.ErrNotFound) || (err == nil && setting == nil) {
		return false, nil
	}
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	cfg, err := setting.AsLoggingSettings()
	if err != nil {
		return false, hperrors.Wrap(err)
	}
	perf := cfg.Performance
	return cfg.Enabled && (cfg.Sources.Apps || cfg.Sources.HivePaaS) && perf != nil && perf.Enabled &&
		len(perf.Nodes) > 0, nil
}

// noteOBIMove says in the update's log what becomes of OBI, once the agent has
// moved: each node's agent pulls its release's OBI while the old one still
// runs, swaps them, and removes the old image. Nothing is done here - the
// agents do it as they start - and a failure to tell is no failure of the
// update.
func (s *service) noteOBIMove(ctx context.Context, db database.IDB, data *sysUpdateData,
	target *base.ReleaseInfo) {
	change, err := s.obiChange(ctx, db, target)
	if err != nil || (change.Change != sysupdateservice.ChangeUpdate &&
		change.Change != sysupdateservice.ChangeMajor) {
		return
	}
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame("obi: each node's agent moves OBI from "+
		change.CurrentImage+" to "+change.TargetImage+" as it starts on the new version", tasklog.TsNow))
}
