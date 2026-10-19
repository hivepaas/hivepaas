package appdeploymentserviceimpl

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/tasklog"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/managedservice"
)

const (
	stepRestorePoint = "restore-point"

	// deployPointPrefix and deployPointTime make a deployment's point's name:
	// deploy-<app key>-<when the deployment was made>.
	deployPointPrefix = "deploy-"
	deployPointTime   = "20060102150405"
	// pointNameMax is the longest restore point name every engine takes
	// (Engine.PointNameMax): PostgreSQL's 63 bytes, which the others keep too so
	// that names read the same.
	pointNameMax = 63
)

// deployStepRestorePoint marks a restore point on the managed database the app
// names, before its pre-deploy command, so that a migration that breaks data
// has its way back already marked (design §4, "Restore points"). The point
// cannot be made: the deployment fails before anything changes, as the option
// asks for the point.
func (s *service) deployStepRestorePoint(ctx context.Context, data *appDeploymentData) (err error) {
	databaseID := data.Deployment.Settings.RestorePointBeforeDeploy.ID
	if databaseID == "" {
		return nil
	}
	data.Step = stepRestorePoint
	name := deployPointName(data.App.Key, data.Deployment.CreatedAt)
	s.addStepStartLog(ctx, data, fmt.Sprintf("Marking restore point %s on the app's database...", name))
	start := timeutil.NowUTC()
	defer func() { s.addStepEndLog(ctx, data, start, err) }()

	marked, err := s.managedService.MarkRestorePoint(ctx, &managedservice.MarkPointReq{
		ProjectID: data.App.ProjectID,
		AppID:     databaseID,
		Name:      name,
		Note:      fmt.Sprintf("Before deployment %s of %s", data.Deployment.ID, data.App.Name),
		Origin:    entity.ManagedPointOriginDeploy,
		// A deployment run again asks for the point it made the first time.
		ExistingOK: true,
	})
	if err != nil {
		return hperrors.Wrap(hperrors.ErrAppDeployRestorePointFailed).
			WithParam("Database", s.appName(ctx, data.App.ProjectID, databaseID)).
			WithParam("Reason", reasonOf(err)).
			WithCause(err)
	}
	_ = data.LogStore.Add(ctx, tasklog.NewOutFrame(fmt.Sprintf("Restore point %s is on %s, at %s.",
		marked.Point.Name, marked.App.Name, marked.Point.Position), tasklog.TsNow))
	return nil
}

// reasonOf is what an error says, for a sentence: its message without its code.
func reasonOf(err error) string {
	info, _ := hperrors.ParseError(err, translation.GetDefaultLang())
	if info == nil || info.Detail == "" {
		return err.Error()
	}
	return info.Detail
}

// appName is an app's name for a message, its id when it cannot be read.
func (s *service) appName(ctx context.Context, projectID, appID string) string {
	app, err := s.appService.LoadApp(ctx, s.db, projectID, appID, false, false)
	if err != nil || app == nil {
		return appID
	}
	return app.Name
}

// deployPointName is a deployment's point's name, from when the deployment was
// made, so that a run of it again names the same point. The app's key is cut
// to fit the 63 bytes every engine takes.
func deployPointName(appKey string, made time.Time) string {
	suffix := "-" + made.UTC().Format(deployPointTime)
	key := appKey
	if room := pointNameMax - len(deployPointPrefix) - len(suffix); len(key) > room {
		key = strings.TrimRight(key[:room], "-_.")
	}
	return deployPointPrefix + key + suffix
}
