package appuc

import (
	"context"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/entity"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/pkg/bunex"
	"github.com/localpaas/localpaas/localpaas_app/pkg/reflectutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/shellutil"
	"github.com/localpaas/localpaas/localpaas_app/pkg/transaction"
	"github.com/localpaas/localpaas/localpaas_app/usecase/appuc/appdto"
)

func (uc *UC) UpdateAppContainerSettings(
	ctx context.Context,
	auth *basedto.Auth,
	req *appdto.UpdateAppContainerSettingsReq,
) (*appdto.UpdateAppContainerSettingsResp, error) {
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data := &updateAppContainerSettingsData{}
		err := uc.loadAppContainerSettingsForUpdate(ctx, db, req, data)
		if err != nil {
			return apperrors.Wrap(err)
		}

		persistingData := &persistingAppData{}
		uc.prepareUpdatingAppContainerSettings(req, data)

		err = uc.persistData(ctx, db, persistingData)
		if err != nil {
			return apperrors.Wrap(err)
		}

		err = uc.applyAppContainerSettings(ctx, data)
		if err != nil {
			return apperrors.Wrap(err)
		}
		return nil
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &appdto.UpdateAppContainerSettingsResp{}, nil
}

type updateAppContainerSettingsData struct {
	App     *entity.App
	Service *swarm.Service
}

func (uc *UC) loadAppContainerSettingsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *appdto.UpdateAppContainerSettingsReq,
	data *updateAppContainerSettingsData,
) error {
	app, err := uc.appService.LoadApp(ctx, db, req.ProjectID, req.AppID, true, true,
		bunex.SelectExcludeColumns(entity.AppDefaultExcludeColumns...),
		bunex.SelectFor("UPDATE OF app"),
		bunex.SelectRelation("Project",
			bunex.SelectExcludeColumns(entity.ProjectDefaultExcludeColumns...),
		),
	)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.App = app

	service, err := uc.appService.ServiceInspect(ctx, app.ServiceID, false)
	if err != nil {
		return apperrors.Wrap(err)
	}
	data.Service = service

	if data.Service == nil || data.Service.Version.Index != uint64(req.UpdateVer) { //nolint:gosec
		return apperrors.Wrap(apperrors.ErrUpdateVerMismatched)
	}

	return nil
}

func (uc *UC) prepareUpdatingAppContainerSettings(
	req *appdto.UpdateAppContainerSettingsReq,
	data *updateAppContainerSettingsData,
) {
	uc.prepareUpdatingAppContainerSpec(req, data)
}

func (uc *UC) prepareUpdatingAppContainerSpec(
	req *appdto.UpdateAppContainerSettingsReq,
	data *updateAppContainerSettingsData,
) {
	service := data.Service
	containerSpec := service.Spec.TaskTemplate.ContainerSpec

	containerSpec.Labels = req.Labels
	containerSpec.Image = req.Image
	containerSpec.Command = gofn.If(req.Command == "", nil, gofn.Must(shellutil.CmdSplit(req.Command)))
	containerSpec.Dir = req.WorkingDir
	containerSpec.Hostname = req.Hostname
	containerSpec.User = req.User
	containerSpec.Groups = req.Groups
	containerSpec.StopSignal = req.StopSignal
	containerSpec.TTY = req.TTY
	containerSpec.OpenStdin = req.OpenStdin
	containerSpec.ReadOnly = req.ReadOnly
	if req.StopGracePeriod != nil {
		containerSpec.StopGracePeriod = new(time.Duration(*req.StopGracePeriod))
	}

	uc.prepareUpdatingAppContainerHealthcheck(req, data)
	uc.prepareUpdatingAppContainerPrivileges(req, data)
	uc.prepareUpdatingAppContainerRestartPolicy(req, data)
}

func (uc *UC) prepareUpdatingAppContainerHealthcheck(
	req *appdto.UpdateAppContainerSettingsReq,
	data *updateAppContainerSettingsData,
) {
	service := data.Service
	containerSpec := service.Spec.TaskTemplate.ContainerSpec

	if req.Healthcheck == nil {
		containerSpec.Healthcheck = nil
		return
	}
	if containerSpec.Healthcheck == nil {
		containerSpec.Healthcheck = &container.HealthConfig{}
	}
	cmd := gofn.Must(shellutil.CmdSplit(req.Healthcheck.Command))
	containerSpec.Healthcheck.Test = gofn.Concat([]string{string(req.Healthcheck.Mode)}, cmd)
	containerSpec.Healthcheck.Interval = time.Duration(req.Healthcheck.Interval)
	containerSpec.Healthcheck.Timeout = time.Duration(req.Healthcheck.Timeout)
	containerSpec.Healthcheck.StartPeriod = time.Duration(req.Healthcheck.StartPeriod)
	containerSpec.Healthcheck.StartInterval = time.Duration(req.Healthcheck.StartInterval)
	containerSpec.Healthcheck.Retries = req.Healthcheck.Retries
}

func (uc *UC) prepareUpdatingAppContainerPrivileges(
	req *appdto.UpdateAppContainerSettingsReq,
	data *updateAppContainerSettingsData,
) {
	service := data.Service
	containerSpec := service.Spec.TaskTemplate.ContainerSpec

	if req.Privileges == nil {
		containerSpec.Privileges = nil
		return
	}
	if containerSpec.Privileges == nil {
		containerSpec.Privileges = &swarm.Privileges{}
	}
	containerSpec.Privileges.NoNewPrivileges = req.Privileges.NoNewPrivileges

	// SELinux
	if req.Privileges.SELinuxContext != nil {
		if containerSpec.Privileges.SELinuxContext == nil {
			containerSpec.Privileges.SELinuxContext = &swarm.SELinuxContext{}
		}
		containerSpec.Privileges.SELinuxContext.Disable = req.Privileges.SELinuxContext.Disable
		containerSpec.Privileges.SELinuxContext.User = req.Privileges.SELinuxContext.User
		containerSpec.Privileges.SELinuxContext.Role = req.Privileges.SELinuxContext.Role
		containerSpec.Privileges.SELinuxContext.Type = req.Privileges.SELinuxContext.Type
		containerSpec.Privileges.SELinuxContext.Level = req.Privileges.SELinuxContext.Level
	} else {
		containerSpec.Privileges.SELinuxContext = nil
	}

	// Seccomp
	if req.Privileges.Seccomp != nil {
		if containerSpec.Privileges.Seccomp == nil {
			containerSpec.Privileges.Seccomp = &swarm.SeccompOpts{}
		}
		containerSpec.Privileges.Seccomp.Mode = req.Privileges.Seccomp.Mode
		containerSpec.Privileges.Seccomp.Profile = reflectutil.UnsafeStrToBytes(req.Privileges.Seccomp.Profile)
	} else {
		containerSpec.Privileges.Seccomp = nil
	}

	// AppArmor
	if req.Privileges.AppArmor != nil {
		if containerSpec.Privileges.AppArmor == nil {
			containerSpec.Privileges.AppArmor = &swarm.AppArmorOpts{}
		}
		containerSpec.Privileges.AppArmor.Mode = req.Privileges.AppArmor.Mode
	} else {
		containerSpec.Privileges.AppArmor = nil
	}
}

func (uc *UC) prepareUpdatingAppContainerRestartPolicy(
	req *appdto.UpdateAppContainerSettingsReq,
	data *updateAppContainerSettingsData,
) {
	service := data.Service
	taskSpec := &service.Spec.TaskTemplate

	if req.RestartPolicy == nil {
		taskSpec.RestartPolicy = nil
		return
	}
	if taskSpec.RestartPolicy == nil {
		taskSpec.RestartPolicy = &swarm.RestartPolicy{}
	}
	taskSpec.RestartPolicy.Condition = req.RestartPolicy.Condition
	taskSpec.RestartPolicy.MaxAttempts = req.RestartPolicy.MaxAttempts
	if req.RestartPolicy.Delay != nil {
		taskSpec.RestartPolicy.Delay = new(time.Duration(*req.RestartPolicy.Delay))
	}
	if req.RestartPolicy.Window != nil {
		taskSpec.RestartPolicy.Window = new(time.Duration(*req.RestartPolicy.Window))
	}
}

func (uc *UC) applyAppContainerSettings(
	ctx context.Context,
	data *updateAppContainerSettingsData,
) error {
	service := data.Service

	_, err := uc.dockerManager.ServiceUpdate(ctx, service.ID, &service.Version, &service.Spec)
	if err != nil {
		return apperrors.Wrap(err)
	}

	return nil
}
