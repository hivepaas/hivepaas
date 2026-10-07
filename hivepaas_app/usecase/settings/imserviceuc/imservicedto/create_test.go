package imservicedto

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

// A service takes a name and a kind: an empty body was stored as a nameless
// setting of no kind.
func TestAnIMServiceWithoutANameOrKindIsRefused(t *testing.T) {
	create := NewCreateIMServiceReq()
	create.IMServiceBaseReq = &IMServiceBaseReq{}
	assert.ElementsMatch(t, []string{"name", "kind"}, pathsOf(create.Validate()))

	update := NewUpdateIMServiceReq()
	update.IMServiceBaseReq = &IMServiceBaseReq{}
	assert.Subset(t, pathsOf(update.Validate()), []string{"name", "kind"})
}

func TestASlackServiceIsAccepted(t *testing.T) {
	req := NewCreateIMServiceReq()
	req.IMServiceBaseReq = &IMServiceBaseReq{Name: "team", Kind: base.IMServiceKindSlack,
		Slack: &IMSlackReq{Webhook: "https://hooks.slack.com/services/x"}}
	assert.Empty(t, pathsOf(req.Validate()))
}

// A test message is sent before the service has a name.
func TestATestMessageTakesNoName(t *testing.T) {
	req := NewTestSendInstantMsgReq()
	req.IMServiceBaseReq = &IMServiceBaseReq{Kind: base.IMServiceKindSlack,
		Slack: &IMSlackReq{Webhook: "https://hooks.slack.com/services/x"}}
	assert.NoError(t, req.ModifyRequest())
	assert.Empty(t, pathsOf(req.Validate()))
}
