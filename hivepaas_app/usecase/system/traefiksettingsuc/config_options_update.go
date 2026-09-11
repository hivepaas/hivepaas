package traefiksettingsuc

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/auditdetail"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/transaction"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/settingsprobationservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/traefiksettingsuc/traefiksettingsdto"
	"github.com/hivepaas/hivepaas/services/traefik/traefikhelper"
)

// UpdateConfigOptions rewrites traefik's startup command, on trial.
//
// The command is applied immediately and undone unless somebody comes back and
// confirms it. That is the only guard there is for the failure swarm cannot see:
// an argument traefik parses and starts under, but which leaves it serving
// nothing - a provider constraint matching no service, a default middleware that
// does not resolve. Its /ping keeps answering 200 through all of that, so the
// healthcheck and failure_action: rollback on the service never fire, and every
// app behind traefik stays unreachable with nobody able to reach the dashboard
// that would fix it.
//
// Confirming has to travel through the new traefik, which is what makes it proof.
// The undo does not: it runs from this process over the docker socket, and a
// command change replaces traefik's task without touching the HivePaaS app's - so
// the process holding the deadline is still alive to enforce it.
//
// What is left over when even that cannot run - the database unreachable, the
// node down - is the manual path in docs/recovery.md.
func (uc *UC) UpdateConfigOptions(
	ctx context.Context,
	auth *basedto.Auth,
	req *traefiksettingsdto.UpdateConfigOptionsReq,
) (*traefiksettingsdto.UpdateConfigOptionsResp, error) {
	var data *updateConfigOptionsData
	err := transaction.Execute(ctx, uc.db, func(db database.Tx) error {
		data = &updateConfigOptionsData{}
		if err := uc.loadConfigOptionsForUpdate(ctx, db, req, data); err != nil {
			return hperrors.Wrap(err)
		}

		// Nothing traefik is not already running. Applying it would replace the
		// task, taking every ingress down for the length of a restart, and then
		// hold the operator on a countdown to confirm a change they did not make.
		//
		// Still recorded. Somebody with admin wrote traefik's command; that it
		// resolved to the command already running is something the entry says, in
		// commandChanged, rather than a reason for the trail to be silent about
		// the write having been attempted at all.
		if !data.CommandChanged {
			return uc.recordTraefikSettingsUpdate(ctx, db, auth, auditSectionConfigOptions,
				auditdetail.New().Set("commandChanged", false))
		}

		uc.prepareUpdatingConfigOptions(data)

		// Inside the same transaction as the change, so a committed change always
		// has a committed deadline. There is no window in which one exists
		// without the other.
		if err := uc.probationService.Arm(ctx, db, auth,
			&settingsprobationservice.ArmReq{
				AppID:       data.AppID,
				Setting:     data.Setting,
				Snapshot:    data.Snapshot,
				Window:      data.ProbationWindow,
				SettleDelay: data.SettleDelay,
			},
			&data.ArmResult,
			func(task *entity.Task) {
				data.UpsertingTasks = append(data.UpsertingTasks, task)
			},
		); err != nil {
			return hperrors.Wrap(err)
		}

		if err := uc.persistConfigOptions(ctx, db, data); err != nil {
			return hperrors.Wrap(err)
		}

		// Before the apply for the same reason the apply is last: from there on
		// the connection carrying this request can be cut at any moment, and a
		// record that cannot be written has to be able to stop the change rather
		// than arrive after traefik is already running it.
		detail := auditdetail.New().
			Set("commandChanged", true).
			Set("onProbation", data.Probation != nil)
		if changed := changedCommandArgs(data.LiveArgs, data.NewArgs); len(changed) > 0 {
			detail.Set("changedArgs", changed)
		}
		if err := uc.recordTraefikSettingsUpdate(ctx, db, auth,
			auditSectionConfigOptions, detail); err != nil {
			return hperrors.Wrap(err)
		}

		// Last, because it is the step that takes traefik down: everything above
		// has to be committed-shaped before the connection carrying this request
		// is cut.
		return hperrors.Wrap(uc.applyConfigOptionsToTraefikService(ctx, data))
	})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	uc.probationService.Schedule(ctx, probationOf(data))

	return &traefiksettingsdto.UpdateConfigOptionsResp{
		Data: traefiksettingsdto.TransformPendingChange(data.Probation),
	}, nil
}

type updateConfigOptionsData struct {
	AppID   string
	Setting *entity.Setting

	TraefikService *swarm.Service
	NewArgs        []string
	CommandChanged bool

	// LiveArgs is the command traefik is running as this request starts, kept so
	// the record can name the flags that moved.
	LiveArgs []string

	// Snapshot is the command as it was before this request touched it: the state
	// a revert restores. Taken from the live service spec rather than from the
	// setting row, so an operator who edited the service by hand is restored to
	// what was actually running.
	Snapshot        entity.SettingSnapshot
	SettleDelay     time.Duration
	ProbationWindow time.Duration

	UpsertingTasks []*entity.Task

	settingsprobationservice.ArmResult
}

// probationOf survives a transaction that never got as far as building one.
func probationOf(data *updateConfigOptionsData) *settingsprobationservice.ArmResult {
	if data == nil {
		return nil
	}
	return &data.ArmResult
}

func (uc *UC) loadConfigOptionsForUpdate(
	ctx context.Context,
	db database.Tx,
	req *traefiksettingsdto.UpdateConfigOptionsReq,
	data *updateConfigOptionsData,
) error {
	// Locked, and locked here rather than anywhere else, because the setting row
	// this update is about may not exist yet: on the first change an install ever
	// makes there is nothing to take FOR UPDATE, and two callers arriving together
	// would each find nothing and each insert one. The app row is the one thing
	// certain to exist, so it is what serializes them.
	appID, err := uc.traefikAppID(ctx, db, bunex.SelectFor("UPDATE OF app"))
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.AppID = appID

	traefikSvc, err := uc.traefikService.GetTraefikSwarmService(ctx)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.TraefikService = traefikSvc

	var liveArgs []string
	if traefikSvc.Spec.TaskTemplate.ContainerSpec != nil {
		liveArgs = traefikSvc.Spec.TaskTemplate.ContainerSpec.Args
	}

	setting, err := uc.loadOrSeedConfigSetting(ctx, db, appID, liveArgs)
	if err != nil {
		return hperrors.Wrap(err)
	}
	data.Setting = setting
	data.Snapshot = entity.SettingSnapshotOf(setting)

	data.LiveArgs = liveArgs
	data.NewArgs = uc.buildStartupCommand(req, traefikSvc)
	data.CommandChanged = !(&entity.TraefikConfig{Args: liveArgs}).SameArgsAs(data.NewArgs)
	// The shared floor. Only traefik's task is replaced - the HivePaaS app keeps
	// running throughout - and that replacement is measured in seconds.
	data.SettleDelay = entity.SettingsProbationSettleDelay
	data.ProbationWindow = settingsprobationservice.ResolveWindow(
		req.ConfirmWindow.ToDuration(), data.SettleDelay)

	return nil
}

// loadOrSeedConfigSetting returns the row the trial hangs off, creating it from
// what is running if this install has never been through here.
//
// The seed is the live command, not the new one. That is what makes the first
// change on an existing install revertible: without a row there would be no
// snapshot, and the very first edit - the one most likely to be exploratory -
// would be the one with no way back.
func (uc *UC) loadOrSeedConfigSetting(
	ctx context.Context,
	db database.Tx,
	appID string,
	liveArgs []string,
) (*entity.Setting, error) {
	setting, err := uc.settingRepo.GetSingle(ctx, db, nil, base.SettingTypeTraefikConfig, true,
		bunex.SelectFor("UPDATE"),
	)
	// Not found is the normal case exactly once per install: nothing writes this
	// row until the first config options change goes through here. It is a reason
	// to seed, not to fail.
	if err != nil && !errors.Is(err, hperrors.ErrNotFound) {
		return nil, hperrors.Wrap(err)
	}
	if setting != nil {
		// Re-seed from what is live. The row is a record of what HivePaaS applied,
		// and docker is the record of what is running; when they disagree, the
		// snapshot has to be the one a revert can actually return to.
		setting.MustSetData(&entity.TraefikConfig{Args: liveArgs})
		return setting, nil
	}

	timeNow := timeutil.NowUTC()
	setting = &entity.Setting{
		ID:        gofn.Must(ulid.NewStringULID()),
		Scope:     base.ObjectScopeApp,
		ObjectID:  appID,
		Type:      base.SettingTypeTraefikConfig,
		Status:    base.SettingStatusActive,
		Version:   entity.CurrentTraefikConfigVersion,
		CreatedAt: timeNow,
		UpdatedAt: timeNow,
	}
	setting.MustSetData(&entity.TraefikConfig{Args: liveArgs})
	if err = uc.settingRepo.Insert(ctx, db, setting); err != nil {
		return nil, hperrors.Wrap(err)
	}
	return setting, nil
}

func (uc *UC) prepareUpdatingConfigOptions(data *updateConfigOptionsData) {
	setting := data.Setting
	setting.MustSetData(&entity.TraefikConfig{Args: data.NewArgs})
	setting.UpdateVer++
	setting.UpdatedAt = timeutil.NowUTC()
}

func (uc *UC) persistConfigOptions(
	ctx context.Context,
	db database.IDB,
	data *updateConfigOptionsData,
) error {
	if err := uc.settingRepo.Update(ctx, db, data.Setting); err != nil {
		return hperrors.Wrap(err)
	}
	if err := uc.taskRepo.UpsertMulti(ctx, db, data.UpsertingTasks,
		entity.TaskUpsertingConflictCols, entity.TaskUpsertingUpdateCols); err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

func (uc *UC) applyConfigOptionsToTraefikService(
	ctx context.Context,
	data *updateConfigOptionsData,
) error {
	err := uc.dockerManager.ServiceUpdateFunc(ctx, data.TraefikService.ID, data.TraefikService,
		func(i int, svc *swarm.Service) (bool, error) {
			if svc.Spec.TaskTemplate.ContainerSpec == nil {
				svc.Spec.TaskTemplate.ContainerSpec = &swarm.ContainerSpec{}
			}
			svc.Spec.TaskTemplate.ContainerSpec.Args = data.NewArgs
			return true, nil
		}, serviceUpdateMaxRetry, 0)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}

//nolint:gocognit
func (uc *UC) buildStartupCommand(
	req *traefiksettingsdto.UpdateConfigOptionsReq,
	traefikSvc *swarm.Service,
) []string {
	newArgs := make([]string, 0, 20) //nolint:mnd

	// 1. Preserve binary executable name ("traefik") and any non-settable system args
	if traefikSvc.Spec.TaskTemplate.ContainerSpec != nil {
		for _, arg := range traefikSvc.Spec.TaskTemplate.ContainerSpec.Args {
			arg = strings.TrimSpace(arg)
			key, _, valid := traefikhelper.ParseCommandArg(arg)
			if key == "traefik" || (valid && !base.IsTraefikCmdArgSettable(key)) {
				newArgs = append(newArgs, arg)
			}
		}
	}

	// 2. Append updated user-configured settings and args from StartupCommand
	if req.StartupCommand != nil {
		cmd := req.StartupCommand
		if cmd.LogLevel != "" {
			newArgs = append(newArgs, "--log=true")
			if strings.ToLower(cmd.LogLevel) != "default" {
				newArgs = append(newArgs, fmt.Sprintf("--log.level=%s", cmd.LogLevel))
			}
		}
		if cmd.AccessLog {
			newArgs = append(newArgs, "--accesslog=true")
		}
		if cmd.HTTP3 {
			newArgs = append(newArgs, "--entrypoints.websecure.http3=true")
		}
		if cmd.FastProxy {
			newArgs = append(newArgs, "--experimental.fastproxy=true")
		}

		for _, kv := range cmd.ParsedArgs {
			switch kv[0] {
			case "log", "log.level", "accesslog", "entrypoints.websecure.http3", "experimental.fastproxy":
				continue
			}
			switch {
			case len(kv) >= 3 && kv[2] != "":
				newArgs = append(newArgs, kv[2])
			case len(kv) >= 2 && kv[1] != "":
				newArgs = append(newArgs, fmt.Sprintf("--%s=%s", kv[0], kv[1]))
			case len(kv) >= 1 && kv[0] != "":
				newArgs = append(newArgs, fmt.Sprintf("--%s", kv[0]))
			}
		}
	}

	return newArgs
}
