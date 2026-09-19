package domainserviceimpl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

func TestPlanCertsWithoutADNSProviderAsksForEachNameItself(t *testing.T) {
	plans, skipped := planCerts([]string{"app.example.com", "api.example.com"}, false, true)

	if !assert.Len(t, plans, 2) {
		t.FailNow()
	}
	assert.Empty(t, skipped)
	assert.Equal(t, "app.example.com", plans[0].Name)
	assert.False(t, plans[0].Wildcard)
	assert.Equal(t, []string{"app.example.com"}, plans[0].Domains)
	assert.Equal(t, "api.example.com", plans[1].Name)
}

// With a DNS provider one wildcard covers every app under the same parent, so
// two domains become one certificate rather than two issuances.
func TestPlanCertsWithADNSProviderGroupsIntoOneWildcard(t *testing.T) {
	plans, skipped := planCerts([]string{"app.example.com", "api.example.com"}, true, true)

	if !assert.Len(t, plans, 1) {
		t.FailNow()
	}
	assert.Empty(t, skipped)
	assert.Equal(t, "*.example.com", plans[0].Name)
	assert.True(t, plans[0].Wildcard)
	assert.Equal(t, []string{"app.example.com", "api.example.com"}, plans[0].Domains)
}

// A name with nothing to the left of the registrable part has no wildcard to
// ask for: *.com is not a certificate anybody issues.
func TestPlanCertsKeepsAnApexNameEvenWithADNSProvider(t *testing.T) {
	plans, _ := planCerts([]string{"example.com"}, true, true)

	if !assert.Len(t, plans, 1) {
		t.FailNow()
	}
	assert.Equal(t, "example.com", plans[0].Name)
	assert.False(t, plans[0].Wildcard)
}

func TestPlanCertsSkipsWhatNoAuthorityIssuesFor(t *testing.T) {
	domains := []string{"10.1.2.3", "localhost", "app.localhost", "dev.test", "app.example.com"}

	plans, skipped := planCerts(domains, false, true)

	if !assert.Len(t, plans, 1) {
		t.FailNow()
	}
	assert.Equal(t, "app.example.com", plans[0].Name)
	assert.Len(t, skipped, 4)
	for _, domain := range []string{"10.1.2.3", "localhost", "app.localhost", "dev.test"} {
		assert.NotEmpty(t, skipped[domain], domain)
	}
}

// An installation signing its own certificates can serve app.localhost over TLS,
// which is what a developer install is: there is no authority to refuse it.
func TestPlanCertsKeepsLocalNamesWhenNothingHasToValidateThem(t *testing.T) {
	plans, skipped := planCerts([]string{"app.localhost", "app.example.com"}, false, false)

	if !assert.Len(t, plans, 2) {
		t.FailNow()
	}
	assert.Empty(t, skipped)
	assert.Equal(t, "app.localhost", plans[0].Name)
	assert.Equal(t, "app.example.com", plans[1].Name)
}

func TestPlanCertsNormalizesAndDeduplicates(t *testing.T) {
	plans, _ := planCerts([]string{" APP.Example.com ", "app.example.com"}, false, true)

	if !assert.Len(t, plans, 1) {
		t.FailNow()
	}
	assert.Equal(t, []string{"app.example.com", "app.example.com"}, plans[0].Domains)
}

// A wildcard serves apps in projects that do not exist yet, so it belongs to the
// installation; a name issued for one app belongs to the project that asked.
func TestCertSettingForPutsAWildcardAtTheInstallationScope(t *testing.T) {
	policy := &entity.DomainCertSettings{Email: "admin@example.com"}
	timeNow := time.Now()

	setting, cert := certSettingFor(&certPlan{Name: "*.example.com", Wildcard: true},
		policy, "project-1", entity.ObjectID{}, entity.ObjectID{ID: "acme-1"}, "setting-1", timeNow)

	assert.Equal(t, base.ObjectScopeGlobal, setting.Scope)
	assert.Empty(t, setting.ObjectID)
	assert.True(t, setting.Inheritable)
	assert.Equal(t, "*.example.com", setting.Name)
	assert.Equal(t, base.SettingStatusActive, setting.Status)
	assert.Equal(t, "*.example.com", cert.Domain)
	assert.Equal(t, base.SSLCertTypeLetsEncrypt, cert.CertType)
	assert.Equal(t, base.SSLKeyTypeDefault, cert.KeyType)
	assert.Equal(t, "acme-1", cert.AcmeProvider.ID)
	assert.True(t, cert.AutoRenew)
}

func TestCertSettingForPutsASingleNameInTheProject(t *testing.T) {
	policy := &entity.DomainCertSettings{
		CertType:    base.SSLCertTypeZeroSSL,
		KeyType:     base.SSLKeyTypeRSA4096,
		Email:       "admin@example.com",
		ValidPeriod: timeutil.Duration(90 * 24 * time.Hour),
	}

	setting, cert := certSettingFor(&certPlan{Name: "app.example.com"},
		policy, "project-1", entity.ObjectID{ID: "provider-1"}, entity.ObjectID{}, "setting-1", time.Now())

	assert.Equal(t, base.ObjectScopeProject, setting.Scope)
	assert.Equal(t, "project-1", setting.ObjectID)
	assert.True(t, setting.Inheritable)
	assert.Equal(t, string(base.SSLCertTypeZeroSSL), setting.Kind)
	assert.Equal(t, base.SSLCertTypeZeroSSL, cert.CertType)
	assert.Equal(t, base.SSLKeyTypeRSA4096, cert.KeyType)
	assert.Equal(t, "provider-1", cert.Provider.ID)
	assert.Equal(t, "admin@example.com", cert.Email)
	assert.Equal(t, timeutil.Duration(90*24*time.Hour), cert.ValidPeriod)
}

func TestRetryableWaitsOutAFailureAndAsksAgainAfterIt(t *testing.T) {
	timeNow := time.Now()
	settingWith := func(cert *entity.SSLCert) *entity.Setting {
		setting := &entity.Setting{Type: base.SettingTypeSSLCert, Name: "app.example.com"}
		assert.NoError(t, setting.SetData(cert))
		return setting
	}

	cases := map[string]struct {
		cert      *entity.SSLCert
		wantRetry bool
	}{
		"an attempt still in flight": {
			&entity.SSLCert{Domain: "app.example.com"}, false},
		"a failure still waiting": {
			&entity.SSLCert{Domain: "app.example.com", LastError: "dns", RetryAfter: timeNow.Add(time.Hour)}, false},
		"a failure that has waited": {
			&entity.SSLCert{Domain: "app.example.com", LastError: "dns", RetryAfter: timeNow.Add(-time.Hour)}, true},
		"one that holds a certificate": {
			&entity.SSLCert{Domain: "app.example.com", Certificate: "-----BEGIN CERTIFICATE-----"}, false},
	}
	for name, tc := range cases {
		t.Run(name, func(t *testing.T) {
			retry, reason := retryable(settingWith(tc.cert), timeNow)
			assert.Equal(t, tc.wantRetry, retry)
			assert.Equal(t, tc.wantRetry, reason == "")
		})
	}
}
