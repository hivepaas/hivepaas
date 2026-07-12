package sslcertuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
)

func (uc *UC) CreateSSLCert(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslcertdto.CreateSSLCertReq,
) (*sslcertdto.CreateSSLCertResp, error) {
	req.Type = currentSettingType
	sslCert := req.ToEntity()
	resp, err := uc.CreateSetting(ctx, &req.CreateSettingReq, &settings.CreateSettingData{
		VerifyingName:   req.Domain,
		VerifyingRefIDs: sslCert.GetRefObjectIDs(),
		Version:         currentSettingVersion,
		AfterLoading: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
		) error {
			if err := uc.verifyDomainInProject(ctx, db, req.Scope, sslCert); err != nil {
				return apperrors.Wrap(err)
			}
			return nil
		},
		PrepareCreation: func(
			ctx context.Context,
			db database.Tx,
			data *settings.CreateSettingData,
			pData *settings.PersistingSettingCreationData,
		) error {
			pData.Setting.Kind = string(req.CertType)
			err := pData.Setting.SetData(sslCert)
			if err != nil {
				return apperrors.Wrap(err)
			}

			refObjects, err := uc.SettingService.LoadReferenceObjects(ctx, db, req.Scope,
				true, true, pData.Setting)
			if err != nil {
				return apperrors.Wrap(err)
			}

			_, err = uc.sslService.ObtainCert(ctx, pData.Setting, refObjects, false)
			if err != nil {
				return apperrors.Wrap(err)
			}

			// Save SSL cert/key files in a directory with forceRecreate=true (for using later)
			err = uc.sslService.WriteCertFiles(true, pData.Setting)
			if err != nil {
				return apperrors.Wrap(err)
			}

			return nil
		},
	})
	if err != nil {
		return nil, apperrors.Wrap(err)
	}

	return &sslcertdto.CreateSSLCertResp{
		Data: resp.Data,
	}, nil
}

func (uc *UC) verifyDomainInProject(
	ctx context.Context,
	db database.IDB,
	scope *base.ObjectScope,
	sslCert *entity.SSLCert,
) error {
	if scope.ProjectID == "" {
		return nil
	}
	err := uc.domainService.VerifyProjectDomains(ctx, db, scope.ProjectID, []string{sslCert.Domain})
	if err != nil {
		return apperrors.Wrap(err)
	}
	return nil
}
