package emailserviceimpl

import (
	"context"
	"html/template"
	"sync"

	"github.com/localpaas/localpaas/localpaas_app/apperrors"
	"github.com/localpaas/localpaas/localpaas_app/infra/database"
	"github.com/localpaas/localpaas/localpaas_app/service/emailservice"
)

var (
	templateMap = map[emailservice.TemplateName]*template.Template{}
	mu          sync.Mutex
)

func (s *service) GetTemplate(
	_ context.Context,
	_ database.IDB,
	name emailservice.TemplateName,
) (tpl *template.Template, err error) {
	mu.Lock()
	defer mu.Unlock()

	if tpl, exists := templateMap[name]; exists {
		return tpl, nil
	}

	switch name { //nolint
	case emailservice.TemplateNamePasswordReset:
		tpl, err = template.ParseFiles("config/email/templates/password_reset.html")
	case emailservice.TemplateNameUserInvite:
		tpl, err = template.ParseFiles("config/email/templates/user_invite.html")
	}
	if err != nil {
		return nil, apperrors.Wrap(err)
	}
	templateMap[name] = tpl

	return tpl, nil
}
