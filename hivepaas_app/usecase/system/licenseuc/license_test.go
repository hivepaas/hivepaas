package licenseuc

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/basedto"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/auditservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
	"github.com/hivepaas/hivepaas/hivepaas_app/usecase/system/licenseuc/licensedto"
)

const (
	thisCluster   = "cluster-this"
	keyText       = "eyJrZXkiOiJ0aGUga2V5In0"
	customerName  = "Acme Industries"
	customerEmail = "ops@acme.example"
)

// txDB runs a transaction's function on a transaction nothing reads: the fakes
// below stand in for everything that would.
type txDB struct {
	database.IDB
}

func (txDB) RunInTx(ctx context.Context, _ *sql.TxOptions, fn func(context.Context, bun.Tx) error) error {
	return fn(ctx, bun.Tx{})
}

// fakeLicenseService embeds the interface, so a method a test does not expect
// panics rather than passes.
type fakeLicenseService struct {
	licenseservice.Service

	current       *license.State
	installed     *licenseservice.InstallResp
	installErr    error
	installedKeys []string
	idErr         error
	idAsked       int

	refreshInfo *licenseservice.RefreshInfo
	fetchReqs   []*licenseservice.FetchRefreshReq
	fetched     *licenseservice.FetchRefreshResp
	applied     []*licenseservice.FetchRefreshResp
	apply       *licenseservice.ApplyRefreshResp
}

func (f *fakeLicenseService) RefreshInfo(context.Context) (*licenseservice.RefreshInfo, error) {
	if f.refreshInfo == nil {
		return &licenseservice.RefreshInfo{Enabled: true}, nil
	}
	return f.refreshInfo, nil
}

func (f *fakeLicenseService) FetchRefresh(_ context.Context, req *licenseservice.FetchRefreshReq) (
	*licenseservice.FetchRefreshResp, error) {
	f.fetchReqs = append(f.fetchReqs, req)
	return f.fetched, nil
}

func (f *fakeLicenseService) ApplyRefresh(_ context.Context, _ database.Tx, fetched *licenseservice.FetchRefreshResp) (
	*licenseservice.ApplyRefreshResp, error) {
	f.applied = append(f.applied, fetched)
	return f.apply, nil
}

func (f *fakeLicenseService) Current(context.Context) (*license.State, error) {
	return f.current, nil
}

func (f *fakeLicenseService) Usage(context.Context, database.IDB) (*licenseservice.Usage, error) {
	return &licenseservice.Usage{Used: map[base.LicenseLimit]int{base.LicenseLimitApps: 7}}, nil
}

func (f *fakeLicenseService) Install(_ context.Context, _ database.Tx,
	req *licenseservice.InstallReq) (*licenseservice.InstallResp, error) {
	f.installedKeys = append(f.installedKeys, req.Key)
	if f.installErr != nil {
		return nil, f.installErr
	}
	return f.installed, nil
}

func (f *fakeLicenseService) InstallationID(context.Context) (string, error) {
	f.idAsked++
	if f.idErr != nil {
		return "", f.idErr
	}
	return thisCluster, nil
}

type spyAuditService struct {
	entries []*auditservice.Entry
	err     error
}

func (s *spyAuditService) Record(_ context.Context, _ database.IDB, entry *auditservice.Entry) error {
	s.entries = append(s.entries, entry)
	return s.err
}

func newTestUC() (*UC, *fakeLicenseService, *spyAuditService) {
	licenseService := &fakeLicenseService{}
	audit := &spyAuditService{}
	cfg := &config.Config{License: config.License{PortalURL: "https://portal.example"}}
	return &UC{db: txDB{}, licenseService: licenseService, auditService: audit,
		cfg: func() *config.Config { return cfg }}, licenseService, audit
}

func adminAuth() *basedto.Auth {
	return &basedto.Auth{User: &basedto.User{User: &entity.User{ID: "admin-user", Username: "admin"}}}
}

// installed is what the service answers for a Colony trial key it has just
// stored.
func installed(created bool) *licenseservice.InstallResp {
	key := &license.Payload{
		Version:       license.PayloadVersion,
		LicenseID:     "lic_e2e",
		Customer:      license.Customer{ID: "acc_1", Name: customerName, Email: customerEmail},
		Plan:          base.LicensePlanColony,
		Installations: []string{thisCluster},
		IssuedAt:      time.Date(2026, 10, 1, 0, 0, 0, 0, time.UTC),
		ExpiresAt:     time.Date(2026, 10, 31, 12, 0, 0, 0, time.FixedZone("+07", 7*3600)),
		Trial:         true,
	}
	return &licenseservice.InstallResp{
		Setting: &entity.Setting{ID: "setting-license", Name: "License", Type: base.SettingTypeLicense},
		Created: created,
		State: &license.State{
			Status:         base.LicenseStatusTrial,
			Plan:           key.Plan,
			Limits:         license.LimitsOf(key),
			Key:            key,
			InstallationID: thisCluster,
			GraceEndsAt:    key.GraceEndsAt(),
		},
	}
}

// Installing a key is recorded as a write of the license setting, naming the key
// by its license ID, plan, trial and expiry, and never by the key itself or by
// who it was issued to.
func TestUpdateLicenseRecordsTheInstallWithoutTheKey(t *testing.T) {
	for name, tc := range map[string]struct {
		created bool
		logType base.AuditLogType
	}{
		"the first key":   {created: true, logType: base.AuditLogTypeSettingCreate},
		"a key replacing": {created: false, logType: base.AuditLogTypeSettingUpdate},
	} {
		t.Run(name, func(t *testing.T) {
			uc, licenseService, audit := newTestUC()
			licenseService.installed = installed(tc.created)
			auth := adminAuth()

			resp, err := uc.UpdateLicense(context.Background(), auth, &licensedto.UpdateLicenseReq{Key: keyText})

			assert.NoError(t, err)
			assert.Equal(t, []string{keyText}, licenseService.installedKeys)
			if !assert.Len(t, audit.entries, 1) {
				return
			}
			entry := audit.entries[0]
			assert.Equal(t, tc.logType, entry.Type)
			assert.Equal(t, base.ObjectScopeGlobal, entry.Scope)
			assert.Empty(t, entry.ObjectID)
			assert.Equal(t, base.AuditLogSourceAPIUpdate, entry.Source)
			assert.Equal(t, base.AuditLogResultAllowed, entry.Result)
			assert.Same(t, auth, entry.Auth)
			assert.Equal(t, base.ResourceType(base.SettingTypeLicense), entry.ResType)
			assert.Equal(t, "setting-license", entry.ResID)
			assert.Equal(t, "License", entry.ResName)
			assert.JSONEq(t,
				`{"licenseId":"lic_e2e","plan":"colony","trial":true,"expiresAt":"2026-10-31T05:00:00Z"}`,
				entry.Detail)

			recorded, _ := json.Marshal(entry)
			for _, secret := range []string{keyText, customerName, customerEmail, "acc_1"} {
				assert.NotContains(t, string(recorded), secret)
			}

			assert.Equal(t, base.LicenseStatusTrial, resp.Data.Status)
			assert.Equal(t, base.LicensePlanColony, resp.Data.Plan)
			assert.Equal(t, thisCluster, resp.Data.InstallationID)
			assert.Zero(t, licenseService.idAsked, "the state the key put in force has the ID")
			assert.NotNil(t, resp.Meta)
		})
	}
}

// A key the service refuses is not recorded, and the refusal is the answer.
func TestUpdateLicenseRefusedRecordsNothing(t *testing.T) {
	uc, licenseService, audit := newTestUC()
	licenseService.installErr = hperrors.Wrap(hperrors.ErrLicenseExpired)

	resp, err := uc.UpdateLicense(context.Background(), adminAuth(), &licensedto.UpdateLicenseReq{Key: keyText})

	assert.ErrorIs(t, err, hperrors.ErrLicenseExpired)
	assert.Nil(t, resp)
	assert.Empty(t, audit.entries)
}

// An install that cannot be recorded fails, which rolls the key back with the
// transaction.
func TestUpdateLicenseFailsWhenItCannotBeRecorded(t *testing.T) {
	uc, licenseService, audit := newTestUC()
	licenseService.installed = installed(true)
	audit.err = errors.New("audit store down")

	resp, err := uc.UpdateLicense(context.Background(), adminAuth(), &licensedto.UpdateLicenseReq{Key: keyText})

	assert.ErrorContains(t, err, "audit store down")
	assert.Nil(t, resp)
}

// With no key stored, Current does not ask docker for the ID; the page still
// shows it, as a key is bought for it.
func TestGetLicenseWithNoKeyShowsTheInstallationID(t *testing.T) {
	uc, licenseService, _ := newTestUC()
	licenseService.current = license.Evaluate(&license.EvalInput{Now: time.Now()})

	resp, err := uc.GetLicense(context.Background(), adminAuth(), licensedto.NewGetLicenseReq())

	assert.NoError(t, err)
	assert.Equal(t, base.LicenseStatusNone, resp.Data.Status)
	assert.Equal(t, thisCluster, resp.Data.InstallationID)
	assert.Empty(t, licenseService.current.InstallationID, "the service's state is not changed")
	for _, item := range resp.Data.Limits {
		if item.Limit == base.LicenseLimitApps {
			assert.Equal(t, 7, item.Used)
		}
	}
}

// The page must load to take a key, so an ID docker cannot give is left empty
// rather than failing it.
func TestGetLicenseWithoutAnInstallationIDStillAnswers(t *testing.T) {
	uc, licenseService, _ := newTestUC()
	licenseService.current = license.Evaluate(&license.EvalInput{Now: time.Now()})
	licenseService.idErr = hperrors.NewNotFound("Swarm cluster")

	resp, err := uc.GetLicense(context.Background(), adminAuth(), licensedto.NewGetLicenseReq())

	assert.NoError(t, err)
	assert.Empty(t, resp.Data.InstallationID)
	assert.Len(t, resp.Data.Limits, len(base.AllLicenseLimits))
}

// A state that has the ID already is answered as it is, with no second ask.
func TestGetLicenseWithAKeyUsesItsInstallationID(t *testing.T) {
	uc, licenseService, _ := newTestUC()
	licenseService.current = installed(false).State

	resp, err := uc.GetLicense(context.Background(), adminAuth(), licensedto.NewGetLicenseReq())

	assert.NoError(t, err)
	assert.Equal(t, thisCluster, resp.Data.InstallationID)
	assert.Equal(t, customerName, resp.Data.CustomerName)
	assert.Zero(t, licenseService.idAsked)
}
