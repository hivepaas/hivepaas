package apppreviewserviceimpl

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
)

// freeDomains is a domain service whose every domain is free.
type freeDomains struct {
	domainservice.Service
	asked []string
}

func (f *freeDomains) VerifyDomainsAvailable(_ context.Context, _ database.IDB, domains, _ []string) error {
	f.asked = append(f.asked, domains...)
	return nil
}

// A preview's domains are beside the app's, and carry none of the app's
// certificates: the copy's routing, applied, gives them the ones that cover them.
func TestAPreviewsDomainsHaveNoCertificateOfTheApps(t *testing.T) {
	domains := &freeDomains{}
	s := &service{domainService: domains}
	setting := &entity.Setting{Type: base.SettingTypeAppRouting}
	assert.NoError(t, setting.SetData(&entity.AppRoutingSettings{Port: 8080, ExposePublicly: true,
		Domains: []*entity.AppDomain{
			{Enabled: true, Domain: "shop.example.com", SSLCert: entity.ObjectID{ID: "cert-of-the-app"}},
			{Enabled: false, Domain: "old.example.com"},
		}}))

	data := &createPreviewData{CalcSubdomain: "pr-42", PreviewApp: &entity.App{ID: "preview-1"}}
	out, err := s.onCloneRoutingSetting(context.Background(), nil, setting, data)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	routing := out.MustAsAppRoutingSettings()
	if assert.Len(t, routing.Domains, 1, "a disabled domain is not copied") {
		assert.Equal(t, "pr-42-shop.example.com", routing.Domains[0].Domain)
		assert.Empty(t, routing.Domains[0].SSLCert.ID)
	}
	assert.Equal(t, []string{"pr-42-shop.example.com"}, domains.asked)
}

// A preview made with no database cloned keeps its app's variables as they
// are: there is no app's name in them to change.
func TestAPreviewWithNoDatabaseClonedKeepsItsVariables(t *testing.T) {
	s := &service{}
	setting := &entity.Setting{Type: base.SettingTypeEnvVar}
	assert.NoError(t, setting.SetData(&entity.EnvVars{Data: []*entity.EnvVar{
		{Key: "GREETING", Value: "hello"},
		{Key: "DB_URL", Value: "${db.HIVEPAAS_URL}"},
	}}))

	out, err := s.onCloneEnvVars(setting, &createPreviewData{})

	assert.NoError(t, err)
	assert.Nil(t, out, "nothing to change")
}
