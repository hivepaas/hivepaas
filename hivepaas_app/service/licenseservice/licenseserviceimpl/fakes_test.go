package licenseserviceimpl

import (
	"context"
	"crypto/ed25519"
	"crypto/mldsa"
	"crypto/rand"
	"crypto/sha256"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"testing"
	"time"

	"github.com/moby/moby/api/types/swarm"
	"github.com/moby/moby/client"
	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/bunex"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/releasesig"
	"github.com/hivepaas/hivepaas/hivepaas_app/repository"
	"github.com/hivepaas/hivepaas/services/docker"
)

const (
	thisCluster  = "cluster-this"
	otherCluster = "cluster-other"
	day          = 24 * time.Hour
)

var testNow = time.Date(2026, 11, 10, 12, 0, 0, 0, time.UTC)

// signer is a license key pair, the way the portal holds one: an ed25519 key and
// an ML-DSA-65 key, both of which sign every key.
type signer struct {
	keys releasesig.PublicKeys
	sign func(data []byte) []releasesig.Entry
}

func newSigner(t *testing.T, prefix string) *signer {
	t.Helper()
	edPub, edPriv, err := ed25519.GenerateKey(rand.Reader)
	assert.NoError(t, err)
	mlPriv, err := mldsa.GenerateKey(mldsa.MLDSA65())
	assert.NoError(t, err)
	keys, err := license.Purpose.ParsePublicKeys(map[string][]byte{
		prefix + "-ed": pemOf(t, edPub),
		prefix + "-ml": pemOf(t, mlPriv.PublicKey()),
	})
	assert.NoError(t, err)
	return &signer{keys: keys, sign: func(data []byte) []releasesig.Entry {
		edSig, err := edPriv.Sign(nil, data, &ed25519.Options{Context: license.Context})
		assert.NoError(t, err)
		mlSig, err := mlPriv.SignDeterministic(data, &mldsa.Options{Context: license.Context})
		assert.NoError(t, err)
		return []releasesig.Entry{
			{KeyID: prefix + "-ed", Algorithm: releasesig.AlgEd25519, Sig: base64.StdEncoding.EncodeToString(edSig)},
			{KeyID: prefix + "-ml", Algorithm: releasesig.AlgMLDSA65, Sig: base64.StdEncoding.EncodeToString(mlSig)},
		}
	}}
}

func pemOf(t *testing.T, pub any) []byte {
	t.Helper()
	der, err := x509.MarshalPKIXPublicKey(pub)
	assert.NoError(t, err)
	return pem.EncodeToMemory(&pem.Block{Type: "PUBLIC KEY", Bytes: der})
}

// keyText is a key's text as it is issued: the envelope's JSON, base64.
func (s *signer) keyText(t *testing.T, payload *license.Payload) string {
	t.Helper()
	data, err := json.Marshal(payload)
	assert.NoError(t, err)
	sum := sha256.Sum256(data)
	envelope, err := json.Marshal(releasesig.Envelope{
		Payload:    base64.StdEncoding.EncodeToString(data),
		SHA256:     hex.EncodeToString(sum[:]),
		Signatures: s.sign(data),
	})
	assert.NoError(t, err)
	return base64.StdEncoding.EncodeToString(envelope)
}

// payload is a key for this installation, in force at testNow: Hive's limits.
func payload(change ...func(p *license.Payload)) *license.Payload {
	p := &license.Payload{
		Version:       license.PayloadVersion,
		LicenseID:     "lic_01",
		Customer:      license.Customer{ID: "acc_01", Name: "Acme", Email: "ops@acme.example"},
		Plan:          base.LicensePlanHive,
		Limits:        map[string]int{"nodes": 10, "users": 10},
		Installations: []string{thisCluster},
		IssuedAt:      testNow.Add(-day),
		ExpiresAt:     testNow.Add(29 * day),
		GraceDays:     14,
	}
	for _, f := range change {
		f(p)
	}
	return p
}

// fakeSettings holds the license setting, if any.
type fakeSettings struct {
	repository.SettingRepo
	setting  *entity.Setting
	reads    int
	upserted []*entity.Setting
	lockedTo []int // how many select options each read passed
}

func (f *fakeSettings) GetSingle(_ context.Context, _ database.IDB, scope *entity.ObjectScope,
	typ base.SettingType, _ bool, opts ...bunex.SelectQueryOption) (*entity.Setting, error) {
	f.reads++
	f.lockedTo = append(f.lockedTo, len(opts))
	if !scope.IsGlobalScope() || typ != base.SettingTypeLicense {
		return nil, fmt.Errorf("read %s in scope %q", typ, scope.ScopeType)
	}
	if f.setting == nil {
		return nil, hperrors.NewNotFound("Setting")
	}
	stored := *f.setting
	return &stored, nil
}

func (f *fakeSettings) Upsert(_ context.Context, _ database.IDB, setting *entity.Setting, _, _ []string,
	_ ...bunex.InsertQueryOption) error {
	stored := *setting
	f.setting = &stored
	f.upserted = append(f.upserted, &stored)
	return nil
}

func (f *fakeSettings) EnsureUnique(context.Context, database.IDB, *entity.ObjectScope, base.SettingType,
	...bunex.SelectQueryOption) error {
	return nil
}

// store makes key the stored license.
func (f *fakeSettings) store(t *testing.T, key string) {
	t.Helper()
	f.setting = &entity.Setting{ID: "LICENSE", Type: base.SettingTypeLicense, Status: base.SettingStatusActive,
		Version: entity.CurrentLicenseSettingsVersion}
	assert.NoError(t, f.setting.SetData(&entity.LicenseSettings{Key: key}))
}

// fakeStatus is system_statuses, as far as the latest time seen goes.
type fakeStatus struct {
	repository.SystemStatusRepo
	seenAt   time.Time
	gets     int
	recorded []time.Time
}

func (f *fakeStatus) Get(context.Context, database.IDB, ...bunex.SelectQueryOption) (*entity.SystemStatus, error) {
	f.gets++
	return &entity.SystemStatus{ID: "1", LicenseSeenAt: f.seenAt}, nil
}

func (f *fakeStatus) RecordLicenseSeenAt(_ context.Context, _ database.IDB, t time.Time) (time.Time, error) {
	f.recorded = append(f.recorded, t)
	if t.After(f.seenAt) {
		f.seenAt = t
	}
	return f.seenAt, nil
}

// fakeUsage counts what it is told, and records what it is asked in order:
// "lock apps", "count apps".
type fakeUsage struct {
	repository.PlanUsageRepo
	counts map[base.LicenseLimit]int
	oauth  int
	asked  []string
}

func (f *fakeUsage) LockLimit(_ context.Context, _ database.IDB, limit base.LicenseLimit) error {
	f.asked = append(f.asked, "lock "+string(limit))
	return nil
}

func (f *fakeUsage) CountApps(_ context.Context, _ database.IDB, functions bool) (int, error) {
	if functions {
		return f.count(base.LicenseLimitFunctions), nil
	}
	return f.count(base.LicenseLimitApps), nil
}

func (f *fakeUsage) CountPreviews(context.Context, database.IDB) (int, error) {
	return f.count(base.LicenseLimitPreviews), nil
}

func (f *fakeUsage) CountUsers(context.Context, database.IDB) (int, error) {
	return f.count(base.LicenseLimitUsers), nil
}

func (f *fakeUsage) CountActiveOAuth(_ context.Context, _ database.IDB, kinds []base.OAuthKind) (int, error) {
	f.asked = append(f.asked, fmt.Sprintf("count oauth %v", kinds))
	return f.oauth, nil
}

func (f *fakeUsage) count(limit base.LicenseLimit) int {
	f.asked = append(f.asked, "count "+string(limit))
	return f.counts[limit]
}

// fakeDocker is the swarm: its ID and its nodes.
type fakeDocker struct {
	docker.Manager
	clusterID    string
	clusterErr   error
	clusterCalls int
	nodes        int
	nodeCalls    int
}

func (f *fakeDocker) SwarmClusterID(context.Context) (string, error) {
	f.clusterCalls++
	return f.clusterID, f.clusterErr
}

func (f *fakeDocker) NodeList(context.Context, ...docker.NodeListOption) (*client.NodeListResult, error) {
	f.nodeCalls++
	return &client.NodeListResult{Items: make([]swarm.Node, f.nodes)}, nil
}

// harness is a service on fakes, at testNow, trusting prod's keys and, in
// development, dev's.
type harness struct {
	*service
	settings *fakeSettings
	status   *fakeStatus
	usage    *fakeUsage
	docker   *fakeDocker
	prod     *signer
	dev      *signer

	now        time.Time
	devEnv     bool
	trustCalls int
	trustErr   error
	// config is the config in force: production, with a portal no test reaches
	// unless it starts one (newPortal).
	config *config.Config
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{
		settings: &fakeSettings{},
		status:   &fakeStatus{},
		usage:    &fakeUsage{counts: map[base.LicenseLimit]int{}},
		docker:   &fakeDocker{clusterID: thisCluster},
		prod:     newSigner(t, "lic-test"),
		dev:      newSigner(t, "dev-test"),
		now:      testNow,
		config: &config.Config{Env: config.EnvProd,
			License: config.License{PortalURL: "https://portal.invalid"}},
	}
	h.service = &service{
		settingRepo:      h.settings,
		systemStatusRepo: h.status,
		planUsageRepo:    h.usage,
		dockerManager:    h.docker,
		timeNow:          func() time.Time { return h.now },
		allowDev:         func() bool { return h.devEnv },
		trust: func(allowDev bool) (*license.Trust, error) {
			h.trustCalls++
			if h.trustErr != nil {
				return nil, h.trustErr
			}
			return &license.Trust{Prod: h.prod.keys, Dev: h.dev.keys, AllowDev: allowDev}, nil
		},
		cfg:        func() *config.Config { return h.config },
		httpClient: newPortalClient(),
	}
	return h
}

// storeKey makes a production key for p the stored license.
func (h *harness) storeKey(t *testing.T, p *license.Payload) {
	t.Helper()
	h.settings.store(t, h.prod.keyText(t, p))
}
