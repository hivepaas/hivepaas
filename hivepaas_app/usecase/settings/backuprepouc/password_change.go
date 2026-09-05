package backuprepouc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/cryptoutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/secrethelper"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/backupreposervice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/backuprepouc/backuprepodto"
)

// changeRepoPasswordData carries what the rollback needs out of the transaction.
type changeRepoPasswordData struct {
	// EngineChanged records that the repository itself was re-encrypted. Anything that fails
	// afterwards leaves the stored password pointing at a repository that no longer accepts it.
	EngineChanged bool
	Req           *backupreposervice.ChangeRepoPasswordReq
}

func (uc *UC) ChangeRepoPassword(
	ctx context.Context,
	auth *basedto.Auth,
	req *backuprepodto.ChangeRepoPasswordReq,
) (*backuprepodto.ChangeRepoPasswordResp, error) {
	req.Type = currentSettingType
	data := &changeRepoPasswordData{}
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			settingData *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			repoSetting := pData.Setting
			// If the setting is not the scope, need to check the user has access on the setting
			if repoSetting.ObjectID != req.Scope.ScopeObjectID() {
				err := uc.CheckAccessOnSetting(ctx, db, auth, repoSetting, base.ActionTypeWrite)
				if err != nil {
					return hperrors.Wrap(err)
				}
			}

			backupRepo := repoSetting.MustAsBackupRepo()

			currentPassword, err := backupRepo.Password.GetPlain()
			if err != nil {
				return hperrors.Wrap(err)
			}
			// The repository is what actually holds the password, so a mismatch here would only
			// surface later as a repository nobody can open.
			if !cryptoutil.SecureCompare(req.CurrentPassword, currentPassword) {
				return hperrors.Wrap(hperrors.ErrPasswordCurrentMismatched)
			}
			// Validate password strength
			strengthReqs := backupreposervice.PasswordRequirements
			strengthReqs.PrevSecrets = []string{currentPassword}
			if err := secrethelper.ValidateStrength(req.NewPassword, &strengthReqs); err != nil {
				return hperrors.Wrap(err)
			}

			data.Req = &backupreposervice.ChangeRepoPasswordReq{
				Scope:       req.Scope,
				Repo:        backupRepo,
				RepoID:      repoSetting.ID,
				OldPassword: currentPassword,
				NewPassword: req.NewPassword,
			}
			// NOTE: this re-encrypts the repository from inside the transaction. It cannot be part
			// of it, hence the rollback below.
			if err := uc.backupRepoService.ChangeRepoPassword(ctx, db, data.Req); err != nil {
				return hperrors.Wrap(err)
			}
			data.EngineChanged = true

			backupRepo.Password.Set(req.NewPassword)
			if err := repoSetting.SetData(backupRepo); err != nil {
				return hperrors.Wrap(err)
			}
			return nil
		},
	})
	if err != nil {
		return nil, hperrors.Wrap(uc.revertRepoPassword(ctx, data, err))
	}

	return &backuprepodto.ChangeRepoPasswordResp{}, nil
}

// revertRepoPassword puts the repository back on its old password after the surrounding
// transaction was rolled back, so the password stored in the DB keeps opening the repository.
// It returns the error the caller should report.
func (uc *UC) revertRepoPassword(
	ctx context.Context,
	data *changeRepoPasswordData,
	cause error,
) error {
	if !data.EngineChanged {
		return cause
	}

	revertReq := *data.Req
	revertReq.OldPassword, revertReq.NewPassword = data.Req.NewPassword, data.Req.OldPassword
	if err := uc.backupRepoService.ChangeRepoPassword(ctx, uc.DB, &revertReq); err != nil {
		// The repository now only opens with a password that was never stored. Reporting the
		// original failure would hide that, and the repository would look merely unchanged.
		return hperrors.Wrap(hperrors.ErrBackupRepoPasswordOutOfSync).
			WithCause(cause).
			WithMsgLog("failed to restore the previous password of backup repo %s: %v", data.Req.RepoID, err)
	}
	return cause
}
