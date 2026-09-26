package getstartedserviceimpl

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/domainservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/getstartedservice"
)

func TestDashboardDomainIsTheFirstEnabledHTTPOne(t *testing.T) {
	routing := &entity.AppRoutingSettings{
		ExposePublicly: true,
		Domains: []*entity.AppDomain{
			{Enabled: false, Domain: "off.example.com"},
			{Enabled: true, Domain: "tcp.example.com", Protocol: base.NetworkProtocolTCP},
			{Enabled: true, Domain: "pass.example.com", TLSPassthrough: true},
			{Enabled: true, Domain: "dash.example.com"},
			{Enabled: true, Domain: "second.example.com"},
		},
	}

	domain := dashboardDomain(routing)

	if assert.NotNil(t, domain) {
		assert.Equal(t, "dash.example.com", domain.Domain)
	}
}

func TestDashboardDomainIsNoneWithoutRouting(t *testing.T) {
	assert.Nil(t, dashboardDomain(nil))
	assert.Nil(t, dashboardDomain(&entity.AppRoutingSettings{
		ExposePublicly: false,
		Domains:        []*entity.AppDomain{{Enabled: true, Domain: "dash.example.com"}},
	}))
}

func TestDashboardCertItem(t *testing.T) {
	timeNow := time.Date(2026, 9, 26, 12, 0, 0, 0, time.UTC)
	letsEncrypt := func(expireAt time.Time) *entity.SSLCert {
		return &entity.SSLCert{
			CertType:    base.SSLCertTypeLetsEncrypt,
			Certificate: "-----BEGIN CERTIFICATE-----",
			ExpireAt:    expireAt,
		}
	}
	selfSigned := &entity.SSLCert{
		CertType:    base.SSLCertTypeSelfSigned,
		Certificate: "-----BEGIN CERTIFICATE-----",
		ExpireAt:    timeNow.AddDate(10, 0, 0),
	}
	failed := &entity.SSLCert{CertType: base.SSLCertTypeLetsEncrypt, LastError: "acme: NXDOMAIN"}

	tests := []struct {
		name      string
		attached  *entity.SSLCert
		pending   *entity.SSLCert
		obtaining bool
		want      getstartedservice.Item
	}{
		{
			name: "nothing yet",
			want: getstartedservice.Item{Status: getstartedservice.ItemStatusTodo, Domain: "dash.example.com"},
		},
		{
			name:     "a trusted certificate attached",
			attached: letsEncrypt(timeNow.AddDate(0, 2, 0)),
			want:     getstartedservice.Item{Status: getstartedservice.ItemStatusDone, Domain: "dash.example.com"},
		},
		{
			name:     "a trusted certificate with no expiry recorded",
			attached: letsEncrypt(time.Time{}),
			want:     getstartedservice.Item{Status: getstartedservice.ItemStatusDone, Domain: "dash.example.com"},
		},
		{
			name:     "the self-signed certificate is not done",
			attached: selfSigned,
			want:     getstartedservice.Item{Status: getstartedservice.ItemStatusTodo, Domain: "dash.example.com"},
		},
		{
			name:     "an expired certificate is not done",
			attached: letsEncrypt(timeNow.Add(-time.Hour)),
			want:     getstartedservice.Item{Status: getstartedservice.ItemStatusTodo, Domain: "dash.example.com"},
		},
		{
			name:     "an attached setting with no content is not done",
			attached: &entity.SSLCert{CertType: base.SSLCertTypeLetsEncrypt},
			want:     getstartedservice.Item{Status: getstartedservice.ItemStatusTodo, Domain: "dash.example.com"},
		},
		{
			name:      "a task on it",
			pending:   &entity.SSLCert{CertType: base.SSLCertTypeLetsEncrypt},
			obtaining: true,
			want:      getstartedservice.Item{Status: getstartedservice.ItemStatusObtaining, Domain: "dash.example.com"},
		},
		{
			name:      "a task on it again after a failure",
			pending:   failed,
			obtaining: true,
			want:      getstartedservice.Item{Status: getstartedservice.ItemStatusObtaining, Domain: "dash.example.com"},
		},
		{
			name:    "the last attempt failed",
			pending: failed,
			want: getstartedservice.Item{Status: getstartedservice.ItemStatusFailed, Domain: "dash.example.com",
				Error: "acme: NXDOMAIN"},
		},
		{
			name:     "an expired certificate being renewed",
			attached: letsEncrypt(timeNow.Add(-time.Hour)), pending: failed, obtaining: true,
			want: getstartedservice.Item{Status: getstartedservice.ItemStatusObtaining, Domain: "dash.example.com"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := dashboardCertItem("dash.example.com", tt.attached, tt.pending, tt.obtaining, timeNow)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNotAskedReasonSaysWhyNothingWasAskedFor(t *testing.T) {
	skipped := &domainservice.EnsureCertsResp{
		Skipped: map[string]string{"dash.example.com": "automatic certificates are off for this project"},
	}
	assert.Equal(t, "no certificate was asked for dash.example.com: automatic certificates are off for this project",
		notAskedReason("dash.example.com", skipped))

	matched := &domainservice.EnsureCertsResp{
		Skipped: map[string]string{},
		Matched: map[string]*entity.Setting{"dash.example.com": {Name: "*.example.com"}},
	}
	assert.Equal(t, `the certificate "*.example.com" covers dash.example.com already: attach it to the domain `+
		"in the HivePaaS routing settings", notAskedReason("dash.example.com", matched))

	assert.Equal(t, "no certificate was asked for dash.example.com",
		notAskedReason("dash.example.com", &domainservice.EnsureCertsResp{}))
}

func TestAnyStillObtainingCountsATaskWaitingToRetry(t *testing.T) {
	task := func(status base.TaskStatus, retry, maxRetry int) *entity.Task {
		return &entity.Task{Status: status, Config: entity.TaskConfig{Retry: retry, MaxRetry: maxRetry}}
	}

	assert.True(t, anyStillObtaining([]*entity.Task{task(base.TaskStatusNotStarted, 0, 2)}))
	assert.True(t, anyStillObtaining([]*entity.Task{task(base.TaskStatusInProgress, 0, 2)}))
	assert.True(t, anyStillObtaining([]*entity.Task{task(base.TaskStatusFailed, 1, 2)}), "failed, a retry to come")
	assert.False(t, anyStillObtaining([]*entity.Task{task(base.TaskStatusFailed, 2, 2)}), "failed, no retry left")
	assert.False(t, anyStillObtaining([]*entity.Task{task(base.TaskStatusDone, 0, 2)}))
	assert.False(t, anyStillObtaining([]*entity.Task{task(base.TaskStatusCanceled, 0, 2)}))
	assert.False(t, anyStillObtaining(nil))
	assert.True(t, anyStillObtaining([]*entity.Task{
		task(base.TaskStatusFailed, 2, 2), task(base.TaskStatusFailed, 0, 2),
	}), "an earlier attempt gave up, a later one will retry")
}

func TestFirstObtainableSkipsTheSelfSignedCertificate(t *testing.T) {
	selfSigned := &entity.Setting{Name: "example.com", Kind: string(base.SSLCertTypeSelfSigned)}
	obtaining := &entity.Setting{Name: "example.com", Kind: string(base.SSLCertTypeLetsEncrypt)}

	assert.Same(t, obtaining, firstObtainable([]*entity.Setting{selfSigned, obtaining}))
	assert.Nil(t, firstObtainable([]*entity.Setting{selfSigned}))
	assert.Nil(t, firstObtainable(nil))
}
