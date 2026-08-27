package notificationserviceimpl

import (
	"context"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bbpool"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/notificationservice"
	"github.com/hivepaas/hivepaas/services/im/imapi"
)

const (
	imSendRetryMax   = 2
	imSendRetryDelay = 2 * time.Second
)

func (s *service) imSendMsg(
	ctx context.Context,
	db database.IDB,
	imSetting *entity.Setting,
	templateType notificationservice.TemplateType,
	templateName notificationservice.TemplateName,
	templateData any,
) error {
	if imSetting == nil {
		return hperrors.NewMissing("IM service setting")
	}

	template, err := s.GetTemplate(ctx, db, templateType, templateName)
	if err != nil {
		return hperrors.Wrap(err)
	}

	buf, bufDefer := bbpool.Small()
	defer bufDefer(buf)
	err = template.Execute(buf, templateData)
	if err != nil {
		return hperrors.Wrap(err)
	}

	err = imapi.SendMessageWithRetry(ctx, imSetting, buf.String(), imSendRetryMax, imSendRetryDelay)
	if err != nil {
		return hperrors.Wrap(err)
	}
	return nil
}
