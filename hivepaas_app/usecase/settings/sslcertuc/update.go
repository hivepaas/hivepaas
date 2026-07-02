package sslcertuc

import (
	"context"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
)

func (uc *UC) UpdateSSLCert(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslcertdto.UpdateSSLCertReq,
) (*sslcertdto.UpdateSSLCertResp, error) {
	req.Type = currentSettingType
	newCert := req.ToEntity()
	reObtainCert := false
	_, err := uc.UpdateSetting(ctx, &req.UpdateSettingReq, &settings.UpdateSettingData{
		VerifyingName:   req.Domain,
		VerifyingRefIDs: newCert.GetRefObjectIDs(),
		AfterLoading: func(ctx context.Context, db database.Tx, data *settings.UpdateSettingData) error {
			if err := uc.verifyDomainInProject(ctx, db, req.Scope, newCert); err != nil {
				return apperrors.New(err)
			}

			currCert, err := data.Setting.AsSSLCert()
			if err != nil {
				return apperrors.New(err)
			}
			switch newCert.CertType {
			case base.SSLCertTypeLetsEncrypt, base.SSLCertTypeZeroSSL, base.SSLCertTypeGoogleTrust:
				reObtainCert = newCert.Domain != currCert.Domain || newCert.KeyType != currCert.KeyType ||
					newCert.Email != currCert.Email
			case base.SSLCertTypeSelfSigned:
				reObtainCert = newCert.Domain != currCert.Domain || newCert.KeyType != currCert.KeyType ||
					newCert.Email != currCert.Email || newCert.ValidPeriod != currCert.ValidPeriod
			case base.SSLCertTypeCustom:
				// Do nothing
			}
			return nil
		},
		PrepareUpdate: func(
			ctx context.Context,
			db database.Tx,
			data *settings.UpdateSettingData,
			pData *settings.PersistingSettingData,
		) error {
			err := pData.Setting.SetData(newCert)
			if err != nil {
				return apperrors.New(err)
			}

			if reObtainCert {
				refObjects, err := uc.SettingService.LoadReferenceObjects(ctx, db, req.Scope,
					true, true, pData.Setting)
				if err != nil {
					return apperrors.New(err)
				}

				_, err = uc.sslService.ObtainCert(ctx, pData.Setting, refObjects, false)
				if err != nil {
					return apperrors.New(err)
				}
			}

			// Save SSL cert/key files in a directory with forceRecreate=true
			err = uc.sslService.WriteCertFiles(true, pData.Setting)
			if err != nil {
				return apperrors.New(err)
			}

			return nil
		},
	})
	if err != nil {
		return nil, apperrors.New(err)
	}

	return &sslcertdto.UpdateSSLCertResp{}, nil
}
