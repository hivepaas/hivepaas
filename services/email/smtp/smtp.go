package smtp

import (
	"context"

	"github.com/tiendc/gofn"
	"gopkg.in/gomail.v2"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type SendMailOption func(*gomail.Message)

func SendMail(
	_ context.Context,
	conf *entity.EmailSMTP,
	recipients []string,
	subject string,
	content string,
	options ...SendMailOption,
) error {
	fromAddress := conf.Username
	fromName := gofn.Coalesce(conf.DisplayName, fromAddress)

	message := gomail.NewMessage()
	message.SetAddressHeader("From", fromAddress, fromName)
	message.SetHeader("To", recipients...)
	message.SetHeader("Subject", subject)
	message.SetBody("text/html", content)

	for _, option := range options {
		option(message)
	}

	password, err := conf.Password.GetPlain()
	if err != nil {
		return apperrors.New(err)
	}

	dialer := gomail.NewDialer(conf.Host, conf.Port, conf.Username, password)
	dialer.SSL = conf.SSL

	err = dialer.DialAndSend(message)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
