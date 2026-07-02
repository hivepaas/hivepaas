package smtp

import (
	"context"

	"github.com/darkrockmountain/gomail"
	"github.com/darkrockmountain/gomail/providers/smtp"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
)

type SendMailOption2 func(*gomail.EmailMessage)

func SendMail2(
	_ context.Context,
	conf *entity.EmailSMTP,
	recipients []string,
	subject string,
	content string,
	options ...SendMailOption2,
) error {
	fromAddress := conf.Username
	if conf.DisplayName != "" {
		fromAddress = formatAddress(fromAddress, conf.DisplayName)
	}

	message := gomail.NewEmailMessage(fromAddress, recipients, subject, content)
	for _, option := range options {
		option(message)
	}

	password, err := conf.Password.GetPlain()
	if err != nil {
		return apperrors.New(err)
	}

	sender, err := smtp.NewSmtpEmailSender(conf.Host, conf.Port, conf.Username, password, "")
	if err != nil {
		return apperrors.New(err)
	}

	err = sender.SendEmail(message)
	if err != nil {
		return apperrors.New(err)
	}
	return nil
}
