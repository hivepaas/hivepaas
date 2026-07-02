package emailservice

import "github.com/hivepaas/hivepaas/hivepaas_app/entity"

type TemplateName string

const (
	TemplateNamePasswordReset TemplateName = "password-reset"
	TemplateNameUserInvite    TemplateName = "user-invite"
)

type TemplateData interface {
	GetEmail() *entity.Email
	GetRecipients() []string
	GetSubject() string
}

type BaseTemplateData struct {
	Email      *entity.Email
	Recipients []string
	Subject    string
}

func (d *BaseTemplateData) GetEmail() *entity.Email {
	return d.Email
}

func (d *BaseTemplateData) GetRecipients() []string {
	return d.Recipients
}

func (d *BaseTemplateData) GetSubject() string {
	return d.Subject
}

type EmailDataPasswordReset struct {
	BaseTemplateData
	ResetPasswordLink string
}

type EmailDataUserInvite struct {
	BaseTemplateData
	InviterName    string
	UserSignupLink string
}
