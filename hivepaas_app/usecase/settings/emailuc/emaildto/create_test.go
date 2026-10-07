package emaildto

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/translation"
)

func pathsOf(errs hperrors.ValidationErrors) []string {
	var paths []string
	for _, inner := range errs.Build(translation.LangEn).InnerErrors {
		paths = append(paths, inner.Path)
	}
	return paths
}

// An account takes a name and a kind: an empty body was stored as a nameless
// setting of no kind.
func TestAnEmailAccountWithoutANameOrKindIsRefused(t *testing.T) {
	create := NewCreateEmailReq()
	create.EmailBaseReq = &EmailBaseReq{}
	assert.ElementsMatch(t, []string{"name", "kind"}, pathsOf(create.Validate()))

	update := NewUpdateEmailReq()
	update.EmailBaseReq = &EmailBaseReq{}
	assert.Subset(t, pathsOf(update.Validate()), []string{"name", "kind"})
}

func TestAnSMTPAccountIsAccepted(t *testing.T) {
	req := NewCreateEmailReq()
	req.EmailBaseReq = &EmailBaseReq{Name: "mail", Kind: base.EmailKindSMTP,
		SMTP: &EmailSMTP{Host: "smtp.example.com", Port: 587, Username: "u", Password: "p"}}
	assert.Empty(t, pathsOf(req.Validate()))
}

// A test mail is sent before the account has a name.
func TestATestMailTakesNoName(t *testing.T) {
	req := NewTestSendMailReq()
	req.EmailBaseReq = &EmailBaseReq{Kind: base.EmailKindSMTP,
		SMTP: &EmailSMTP{Host: "smtp.example.com", Port: 587, Username: "u", Password: "p"}}
	req.TestRecipient = "a@example.com"
	assert.NoError(t, req.ModifyRequest())
	assert.Empty(t, pathsOf(req.Validate()))
}
