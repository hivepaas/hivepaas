package sslcertuc

import (
	"context"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/base"
	"github.com/localpaas/localpaas/localpaas_app/basedto"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings"
	"github.com/localpaas/localpaas/localpaas_app/usecase/settings/sslcertuc/sslcertdto"
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
			currCert, err := data.Setting.AsSSLCert()
			if err != nil {
				return apperrors.Wrap(err)
			}
			// Not allow to change cert type
			if currCert.CertType != newCert.CertType {
				return apperrors.NewNonEditable("Certificate type")
			}
			// Not allow to change provider
			if currCert.Provider.ID != newCert.Provider.ID {
				return apperrors.NewNonEditable("Certificate provider")
			}
			switch newCert.CertType {
			case base.SSLCertTypeLetsEncrypt, base.SSLCertTypeZeroSSL, base.SSLCertTypeGoogleTS:
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
				return apperrors.Wrap(err)
			}

			if reObtainCert {
				refObjects, err := uc.SettingService.LoadReferenceObjects(ctx, db, req.Scope,
					true, true, pData.Setting)
				if err != nil {
					return apperrors.Wrap(err)
				}

				_, err = uc.sslService.ObtainCert(ctx, pData.Setting, refObjects, false)
				if err != nil {
					return apperrors.Wrap(err)
				}
			}

			// Save SSL cert/key in files
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

	return &sslcertdto.UpdateSSLCertResp{}, nil
}
