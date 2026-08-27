package sslcertuc

import (
	"archive/zip"
	"bytes"
	"context"
	"io"
	"net/url"
	"strings"

	"github.com/tiendc/gofn"

	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/fileutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/settings/sslcertuc/sslcertdto"
)

const (
	certEntryName   = "certificate.crt"
	keyEntryName    = "private.key"
	caCertEntryName = "ca.crt"
)

func (uc *UC) DownloadBundle(
	ctx context.Context,
	auth *basedto.Auth,
	req *sslcertdto.DownloadBundleReq,
) (*sslcertdto.DownloadBundleResp, error) {
	req.Type = currentSettingType
	resp, err := uc.GetSetting(ctx, auth, &req.GetSettingReq, &settings.GetSettingData{})
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	setting := resp.Data
	sslCert := setting.MustAsSSLCert()
	if err := sslCert.Decrypt(); err != nil {
		return nil, hperrors.Wrap(err)
	}

	privateKey, err := sslCert.PrivateKey.GetPlain()
	if err != nil {
		return nil, hperrors.Wrap(err)
	}

	buf := new(bytes.Buffer)
	zipWriter := zip.NewWriter(buf)

	// 1. Add certificate.crt
	certWriter, err := zipWriter.Create(certEntryName)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to create zip entry for certificate")
	}
	if _, err := certWriter.Write([]byte(sslCert.Certificate)); err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to write certificate to zip")
	}

	// 2. Add private.key
	keyWriter, err := zipWriter.Create(keyEntryName)
	if err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to create zip entry for private key")
	}
	if _, err := keyWriter.Write([]byte(privateKey)); err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to write private key to zip")
	}

	// 3. Add ca.crt if CA certificate is present
	if strings.TrimSpace(sslCert.CACertificate) != "" {
		caWriter, err := zipWriter.Create(caCertEntryName)
		if err != nil {
			return nil, hperrors.Wrap(err).WithMsgLog("failed to create zip entry for CA certificate")
		}
		if _, err := caWriter.Write([]byte(sslCert.CACertificate)); err != nil {
			return nil, hperrors.Wrap(err).WithMsgLog("failed to write CA certificate to zip")
		}
	}

	if err := zipWriter.Close(); err != nil {
		return nil, hperrors.Wrap(err).WithMsgLog("failed to finalize zip archive")
	}

	rawBaseName := gofn.Coalesce(setting.Name, sslCert.Domain, setting.ID, "ssl-certificate")
	baseName := fileutil.SanitizeFileName(rawBaseName)
	if baseName == "" {
		baseName = gofn.Coalesce(setting.ID, "ssl-certificate")
	}
	zipFileName := baseName + ".zip"

	zipBytes := buf.Bytes()
	extraHeaders := map[string]string{
		"Content-Disposition": `attachment; filename*=UTF-8''` + url.QueryEscape(zipFileName),
	}

	return &sslcertdto.DownloadBundleResp{
		Data: &sslcertdto.DownloadBundleDataResp{
			BaseDownloadDataResp: &settings.BaseDownloadDataResp{
				ContentType:   "application/zip",
				ContentLength: int64(len(zipBytes)),
				ExtraHeaders:  extraHeaders,
				Content:       io.NopCloser(bytes.NewReader(zipBytes)),
			},
		},
	}, nil
}
