package licenseserviceimpl

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/license"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
	"github.com/hivepaas/hivepaas/hivepaas_app/service/licenseservice"
)

// fakePortal is the license portal's refresh: it answers what a test sets, and
// keeps what it was sent.
type fakePortal struct {
	server *httptest.Server

	mu       sync.Mutex
	requests []*http.Request
	bodies   []string
	status   int
	answer   string
	delay    time.Duration
}

// newPortal starts a portal over https and points the harness at it.
func newPortal(t *testing.T, h *harness) *fakePortal {
	t.Helper()
	p := &fakePortal{status: http.StatusOK}
	p.server = httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		p.mu.Lock()
		p.requests = append(p.requests, r)
		p.bodies = append(p.bodies, string(body))
		status, answer, delay := p.status, p.answer, p.delay
		p.mu.Unlock()
		if delay > 0 {
			select {
			case <-time.After(delay):
			case <-r.Context().Done():
				return
			}
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = io.WriteString(w, answer)
	}))
	t.Cleanup(p.server.Close)
	h.config.License.PortalURL = p.server.URL
	h.httpClient.Transport = p.server.Client().Transport
	return p
}

func (p *fakePortal) answers(status int, answer string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.status, p.answer = status, answer
}

func (p *fakePortal) asked() int {
	p.mu.Lock()
	defer p.mu.Unlock()
	return len(p.requests)
}

func newerAnswer(key string) string {
	data, _ := json.Marshal(map[string]any{"data": map[string]any{"result": "newer", "key": key}})
	return string(data)
}

// refresh fetches and applies, as the daily job and the API do.
func (h *harness) refresh(t *testing.T, manual bool) *licenseservice.ApplyRefreshResp {
	t.Helper()
	fetched, err := h.FetchRefresh(context.Background(), &licenseservice.FetchRefreshReq{Manual: manual})
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	applied, err := h.ApplyRefresh(context.Background(), tx, fetched)
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return applied
}

func (h *harness) stored(t *testing.T) *entity.LicenseSettings {
	t.Helper()
	settings, err := h.settings.setting.AsLicenseSettings()
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	return settings
}

// storeWith makes a production key for p the stored license, moved at movedAt,
// with the refresh's schedule.
func (h *harness) storeWith(t *testing.T, p *license.Payload, movedAt time.Time) (string, entity.SchedJobSchedule) {
	t.Helper()
	key := h.prod.keyText(t, p)
	schedule := entity.SchedJobSchedule{Interval: timeutil.Duration(timeutil.Day),
		InitialTime: time.Date(2026, 11, 1, 3, 30, 0, 0, time.UTC)}
	h.settings.store(t, key)
	h.settings.setting.MustSetData(&entity.LicenseSettings{Key: key, MovedAt: movedAt,
		Refresh: entity.LicenseRefresh{Schedule: schedule}})
	return key, schedule
}

// A refresh sends the key and the installation's ID, and nothing else: not the
// version, not the plan, not the usage.
func TestRefreshSendsOnlyTheKeyAndTheInstallationID(t *testing.T) {
	h := newHarness(t)
	portal := newPortal(t, h)
	portal.answers(http.StatusOK, `{"data":{"result":"current"}}`)
	key, _ := h.storeWith(t, payload(), time.Time{})

	h.refresh(t, false)

	if !assert.Equal(t, 1, portal.asked()) {
		return
	}
	req := portal.requests[0]
	assert.Equal(t, http.MethodPost, req.Method)
	assert.Equal(t, "/api/licenses/refresh", req.URL.Path)
	assert.Empty(t, req.URL.RawQuery)
	assert.Equal(t, "HivePaaS", req.Header.Get("User-Agent"), "no version")
	assert.Equal(t, "application/json", req.Header.Get("Content-Type"))
	assert.Empty(t, req.Header.Get("Cookie"))
	assert.Empty(t, req.Header.Get("Authorization"))

	var sent map[string]any
	assert.NoError(t, json.Unmarshal([]byte(portal.bodies[0]), &sent))
	assert.Equal(t, map[string]any{"key": key, "installationId": thisCluster}, sent)
}

// A newer key is installed as Install installs one: a move is cleared, and the
// refresh's schedule is kept beside the outcome.
func TestRefreshInstallsANewerKey(t *testing.T) {
	h := newHarness(t)
	portal := newPortal(t, h)
	_, schedule := h.storeWith(t, payload(), testNow.Add(-day))
	renewed := h.prod.keyText(t, payload(func(p *license.Payload) {
		p.LicenseID, p.Plan, p.ExpiresAt = "lic_02", base.LicensePlanColony, testNow.Add(60*day)
	}))
	// As an email client would wrap it: whitespace means nothing in a key.
	portal.answers(http.StatusOK, newerAnswer(renewed[:50]+"\n"+renewed[50:]))
	_, err := h.Current(context.Background())
	assert.NoError(t, err)
	assert.NotNil(t, h.verified)

	applied := h.refresh(t, true)

	assert.Equal(t, base.LicenseRefreshNewer, applied.Result)
	assert.True(t, applied.Changed)
	assert.NoError(t, applied.Err)
	if assert.NotNil(t, applied.Setting) {
		assert.Equal(t, "LICENSE", applied.Setting.ID)
	}
	if assert.NotNil(t, applied.Installed) {
		assert.Equal(t, "LICENSE", applied.Installed.Setting.ID)
		assert.False(t, applied.Installed.Created)
	}
	if assert.NotNil(t, applied.State) {
		assert.Equal(t, base.LicenseStatusValid, applied.State.Status)
		assert.Equal(t, base.LicensePlanColony, applied.State.Plan)
	}
	assert.Equal(t, &entity.LicenseSettings{Key: renewed, Refresh: entity.LicenseRefresh{
		Schedule: schedule, LastCheckedAt: testNow, LastResult: base.LicenseRefreshNewer,
	}}, h.stored(t))
	assert.Nil(t, h.verified, "the old key's verification is dropped")
	assert.Equal(t, []int{0, 1}, h.settings.lockedTo[len(h.settings.lockedTo)-2:],
		"read without a lock to ask, and for update to apply")

	state, err := h.Current(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, "lic_02", state.Key.LicenseID)
}

// The portal is trusted with which key is newest, never with whether a key
// verifies or is for this installation: one Install would refuse is not
// installed, and the check failed.
func TestRefreshRefusesANewerKeyInstallWouldRefuse(t *testing.T) {
	for name, tc := range map[string]struct {
		key func(h *harness) string
		err error
	}{
		"for another installation": {err: hperrors.ErrLicenseOtherInstallation, key: func(h *harness) string {
			return h.prod.keyText(t, payload(func(p *license.Payload) { p.Installations = []string{otherCluster} }))
		}},
		"signed by a development key": {err: hperrors.ErrLicenseInvalid, key: func(h *harness) string {
			return h.dev.keyText(t, payload())
		}},
		"past its grace": {err: hperrors.ErrLicenseExpired, key: func(h *harness) string {
			return h.prod.keyText(t, payload(func(p *license.Payload) {
				p.IssuedAt, p.ExpiresAt = testNow.Add(-60*day), testNow.Add(-30*day)
			}))
		}},
		"not a key": {err: hperrors.ErrLicenseInvalid, key: func(*harness) string { return "not-a-key" }},
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			portal := newPortal(t, h)
			stored, schedule := h.storeWith(t, payload(), time.Time{})
			portal.answers(http.StatusOK, newerAnswer(tc.key(h)))

			applied := h.refresh(t, false)

			assert.Equal(t, base.LicenseRefreshFailed, applied.Result)
			assert.False(t, applied.Changed)
			assert.Nil(t, applied.Installed)
			assert.ErrorIs(t, applied.Err, tc.err)
			settings := h.stored(t)
			assert.Equal(t, stored, settings.Key, "the stored key is kept")
			assert.Equal(t, schedule, settings.Refresh.Schedule)
			assert.Equal(t, base.LicenseRefreshFailed, settings.Refresh.LastResult)
			assert.Equal(t, testNow, settings.Refresh.LastCheckedAt)
			assert.Contains(t, settings.Refresh.LastError, "The portal sent this newer key")
			assert.NotContains(t, settings.Refresh.LastError, "ops@acme.example")
		})
	}
}

// A newer key that is the one stored is no change.
func TestRefreshAnsweringTheStoredKeyIsCurrent(t *testing.T) {
	h := newHarness(t)
	portal := newPortal(t, h)
	key, _ := h.storeWith(t, payload(), time.Time{})
	portal.answers(http.StatusOK, newerAnswer(key))

	applied := h.refresh(t, false)

	assert.Equal(t, base.LicenseRefreshCurrent, applied.Result)
	assert.False(t, applied.Changed)
	assert.Equal(t, base.LicenseRefreshCurrent, h.stored(t).Refresh.LastResult)
}

// The first answer that the key moved says when; the key then holds for its
// grace from then. A later one changes nothing.
func TestRefreshLearnsOfAMove(t *testing.T) {
	h := newHarness(t)
	portal := newPortal(t, h)
	key, schedule := h.storeWith(t, payload(), time.Time{})
	portal.answers(http.StatusOK, `{"data":{"result":"moved","movedAt":"2026-11-08T10:00:00+02:00"}}`)

	applied := h.refresh(t, false)

	assert.Equal(t, base.LicenseRefreshMoved, applied.Result)
	assert.True(t, applied.Changed)
	assert.Nil(t, applied.Installed)
	assert.Equal(t, "LICENSE", applied.Setting.ID)
	movedAt := time.Date(2026, 11, 8, 8, 0, 0, 0, time.UTC)
	if assert.NotNil(t, applied.State) {
		assert.Equal(t, base.LicenseStatusMoved, applied.State.Status)
		assert.Equal(t, movedAt.Add(14*day), applied.State.GraceEndsAt)
	}
	assert.Equal(t, &entity.LicenseSettings{Key: key, MovedAt: movedAt, Refresh: entity.LicenseRefresh{
		Schedule: schedule, LastCheckedAt: testNow, LastResult: base.LicenseRefreshMoved,
	}}, h.stored(t))

	h.now = testNow.Add(day)
	portal.answers(http.StatusOK, `{"data":{"result":"moved","movedAt":"2026-11-09T00:00:00Z"}}`)
	applied = h.refresh(t, false)

	assert.Equal(t, base.LicenseRefreshMoved, applied.Result)
	assert.False(t, applied.Changed, "already known")
	assert.Equal(t, movedAt, h.stored(t).MovedAt)
	assert.Equal(t, testNow.Add(day), h.stored(t).Refresh.LastCheckedAt)
}

// A move the portal does not date is dated when it was learned of, and one dated
// after that - the portal's clock ahead of this one - is too.
func TestRefreshDatesAMoveNoLaterThanItWasLearnedOf(t *testing.T) {
	for name, answer := range map[string]string{
		"undated":       `{"data":{"result":"moved"}}`,
		"in the future": `{"data":{"result":"moved","movedAt":"2026-11-10T13:00:00Z"}}`,
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			portal := newPortal(t, h)
			h.storeWith(t, payload(), time.Time{})
			portal.answers(http.StatusOK, answer)

			h.refresh(t, false)

			assert.Equal(t, testNow, h.stored(t).MovedAt)
		})
	}
}

// Moved away and back: the portal answers the key for this installation as
// newer, and installing it clears the move.
func TestRefreshAfterAMoveBack(t *testing.T) {
	h := newHarness(t)
	portal := newPortal(t, h)
	h.storeWith(t, payload(), time.Time{})
	portal.answers(http.StatusOK, `{"data":{"result":"moved","movedAt":"2026-11-09T00:00:00Z"}}`)
	h.refresh(t, false)
	state, err := h.Current(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, base.LicenseStatusMoved, state.Status)

	back := h.prod.keyText(t, payload(func(p *license.Payload) { p.IssuedAt = testNow }))
	portal.answers(http.StatusOK, newerAnswer(back))
	applied := h.refresh(t, false)

	assert.Equal(t, base.LicenseRefreshNewer, applied.Result)
	assert.True(t, h.stored(t).MovedAt.IsZero())
	state, err = h.Current(context.Background())
	assert.NoError(t, err)
	assert.Equal(t, base.LicenseStatusValid, state.Status)
}

// Current changes nothing but the record of the check: not the key, not a move.
func TestRefreshCurrentOnlyRecordsTheCheck(t *testing.T) {
	h := newHarness(t)
	portal := newPortal(t, h)
	key, schedule := h.storeWith(t, payload(), testNow.Add(-day))
	h.settings.setting.MustSetData(&entity.LicenseSettings{Key: key, MovedAt: testNow.Add(-day),
		Refresh: entity.LicenseRefresh{Schedule: schedule, LastResult: base.LicenseRefreshFailed, LastError: "x"}})
	portal.answers(http.StatusOK, `{"data":{"result":"current"}}`)

	applied := h.refresh(t, false)

	assert.Equal(t, base.LicenseRefreshCurrent, applied.Result)
	assert.False(t, applied.Changed)
	assert.Nil(t, applied.State)
	assert.Equal(t, &entity.LicenseSettings{Key: key, MovedAt: testNow.Add(-day), Refresh: entity.LicenseRefresh{
		Schedule: schedule, LastCheckedAt: testNow, LastResult: base.LicenseRefreshCurrent,
	}}, h.stored(t), "the last error is cleared")
}

// A key the portal did not issue is an answer, not a failure. A 404 without
// the portal's code is a URL with no portal behind it.
func TestRefreshOfAKeyThePortalDidNotIssue(t *testing.T) {
	h := newHarness(t)
	portal := newPortal(t, h)
	key, _ := h.storeWith(t, payload(), time.Time{})
	portal.answers(http.StatusNotFound, `{"status":404,"code":"ERR_LICENSE_KEY_UNKNOWN","detail":"Unknown key"}`)

	applied := h.refresh(t, true)

	assert.Equal(t, base.LicenseRefreshUnknown, applied.Result)
	assert.NoError(t, applied.Err)
	assert.Equal(t, key, h.stored(t).Key)
	assert.Equal(t, base.LicenseRefreshUnknown, h.stored(t).Refresh.LastResult)
	assert.Empty(t, h.stored(t).Refresh.LastError)

	portal.answers(http.StatusNotFound, `<html>Not found</html>`)
	applied = h.refresh(t, true)

	assert.Equal(t, base.LicenseRefreshFailed, applied.Result)
	assert.ErrorIs(t, applied.Err, hperrors.ErrLicensePortalRefused)
}

// A portal that fails to answer is recorded, never fatal: the stored key stays
// in force, and the next check tries again.
func TestRefreshFailuresAreRecorded(t *testing.T) {
	for name, tc := range map[string]struct {
		status int
		answer string
		err    error
		detail string
	}{
		"down": {status: http.StatusBadGateway, answer: `<html>Bad gateway</html>`,
			err: hperrors.ErrLicensePortalUnavailable, detail: "502 Bad Gateway"},
		"busy": {status: http.StatusTooManyRequests,
			answer: `{"status":429,"code":"ERR_TOO_MANY_REQUESTS","detail":"Slow down"}`,
			err:    hperrors.ErrLicensePortalUnavailable, detail: "429 ERR_TOO_MANY_REQUESTS: Slow down"},
		"another installation's key": {status: http.StatusPreconditionFailed,
			answer: `{"status":412,"code":"ERR_LICENSE_KEY_OTHER_INSTALLATION","detail":"Not for this one"}`,
			err:    hperrors.ErrLicensePortalRefused, detail: "412 ERR_LICENSE_KEY_OTHER_INSTALLATION: Not for this one"},
		"moved elsewhere": {status: http.StatusTemporaryRedirect, err: hperrors.ErrLicensePortalRefused,
			detail: "307 Temporary Redirect"},
		"not an answer": {status: http.StatusOK, answer: `<html>Welcome</html>`,
			err: hperrors.ErrLicensePortalUnavailable, detail: "the portal's answer could not be read"},
		"no data": {status: http.StatusOK, answer: `{}`, err: hperrors.ErrLicensePortalUnavailable},
		"newer with no key": {status: http.StatusOK, answer: `{"data":{"result":"newer"}}`,
			err: hperrors.ErrLicensePortalUnavailable},
		"a result HivePaaS gives": {status: http.StatusOK, answer: `{"data":{"result":"failed"}}`,
			err: hperrors.ErrLicensePortalUnavailable},
		"a result nobody knows of": {status: http.StatusOK, answer: `{"data":{"result":"later"}}`,
			err: hperrors.ErrLicensePortalUnavailable},
	} {
		t.Run(name, func(t *testing.T) {
			h := newHarness(t)
			portal := newPortal(t, h)
			key, _ := h.storeWith(t, payload(), time.Time{})
			portal.answers(tc.status, tc.answer)

			applied := h.refresh(t, false)

			assert.Equal(t, base.LicenseRefreshFailed, applied.Result)
			assert.False(t, applied.Changed)
			assert.ErrorIs(t, applied.Err, tc.err)
			settings := h.stored(t)
			assert.Equal(t, key, settings.Key)
			assert.True(t, settings.MovedAt.IsZero())
			assert.Equal(t, base.LicenseRefreshFailed, settings.Refresh.LastResult)
			assert.Equal(t, testNow, settings.Refresh.LastCheckedAt)
			assert.Contains(t, settings.Refresh.LastError, tc.detail)
			assert.NotContains(t, settings.Refresh.LastError, key)
		})
	}
}

// A redirect is not followed: the key goes to the configured portal and nowhere
// else.
func TestRefreshFollowsNoRedirect(t *testing.T) {
	h := newHarness(t)
	elsewhere := newPortal(t, newHarness(t))
	portal := httptest.NewTLSServer(http.RedirectHandler(elsewhere.server.URL+"/api/licenses/refresh",
		http.StatusPermanentRedirect))
	t.Cleanup(portal.Close)
	h.config.License.PortalURL = portal.URL
	h.httpClient.Transport = portal.Client().Transport
	h.storeWith(t, payload(), time.Time{})

	applied := h.refresh(t, false)

	assert.Equal(t, base.LicenseRefreshFailed, applied.Result)
	assert.Zero(t, elsewhere.asked())
}

func TestRefreshTimesOut(t *testing.T) {
	h := newHarness(t)
	portal := newPortal(t, h)
	h.storeWith(t, payload(), time.Time{})
	portal.mu.Lock()
	portal.delay = time.Minute
	portal.mu.Unlock()
	h.httpClient.Timeout = 50 * time.Millisecond

	applied := h.refresh(t, false)

	assert.Equal(t, base.LicenseRefreshFailed, applied.Result)
	assert.ErrorIs(t, applied.Err, hperrors.ErrLicensePortalUnavailable)
	assert.NotEmpty(t, h.stored(t).Refresh.LastError)
}

// A portal over plain http is asked only in development: its answer is not
// signed as a whole, and a "moved" must not cross the network in the clear.
func TestRefreshOverHTTPOnlyInDevelopment(t *testing.T) {
	h := newHarness(t)
	asked := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		asked++
		_, _ = io.WriteString(w, `{"data":{"result":"current"}}`)
	}))
	t.Cleanup(server.Close)
	h.config.License.PortalURL = server.URL
	h.storeWith(t, payload(), time.Time{})

	applied := h.refresh(t, true)

	assert.Equal(t, base.LicenseRefreshFailed, applied.Result)
	assert.ErrorIs(t, applied.Err, hperrors.ErrLicensePortalUnavailable)
	assert.Contains(t, h.stored(t).Refresh.LastError, "not https")
	assert.Zero(t, asked)

	h.config.Env = config.EnvDev
	applied = h.refresh(t, true)

	assert.Equal(t, base.LicenseRefreshCurrent, applied.Result)
	assert.Equal(t, 1, asked)
}

// With no key there is nothing to ask with; with the daily check off, the daily
// job asks nothing, though a person asking now is answered. Neither is recorded.
func TestRefreshAsksNothingWithoutAKeyOrWhenSwitchedOff(t *testing.T) {
	h := newHarness(t)
	portal := newPortal(t, h)
	portal.answers(http.StatusOK, `{"data":{"result":"current"}}`)

	assert.Equal(t, base.LicenseRefreshNone, h.refresh(t, true).Result, "no setting")
	h.settings.store(t, "")
	assert.Equal(t, base.LicenseRefreshNone, h.refresh(t, false).Result, "no key")
	assert.Zero(t, portal.asked())
	assert.Empty(t, h.settings.upserted)

	h.storeWith(t, payload(), time.Time{})
	h.config.License.RefreshDisabled = true
	applied := h.refresh(t, false)
	assert.Equal(t, base.LicenseRefreshDisabled, applied.Result)
	assert.False(t, applied.Changed)
	assert.Zero(t, portal.asked())
	assert.Empty(t, h.settings.upserted)
	assert.Zero(t, h.stored(t).Refresh.LastCheckedAt)

	assert.Equal(t, base.LicenseRefreshCurrent, h.refresh(t, true).Result, "asked now")
	assert.Equal(t, 1, portal.asked())
}

// An answer about a key replaced while the portal was asked is dropped: the key
// installed meanwhile is newer than anything it could say.
func TestRefreshDropsAnAnswerForAKeyNoLongerStored(t *testing.T) {
	h := newHarness(t)
	portal := newPortal(t, h)
	h.storeWith(t, payload(), time.Time{})
	portal.answers(http.StatusOK, `{"data":{"result":"moved","movedAt":"2026-11-09T00:00:00Z"}}`)
	fetched, err := h.FetchRefresh(context.Background(), &licenseservice.FetchRefreshReq{})
	assert.NoError(t, err)

	pasted := h.prod.keyText(t, payload(func(p *license.Payload) { p.LicenseID = "lic_pasted" }))
	_, err = h.Install(context.Background(), tx, &licenseservice.InstallReq{Key: pasted})
	assert.NoError(t, err)
	upserts := len(h.settings.upserted)

	applied, err := h.ApplyRefresh(context.Background(), tx, fetched)

	assert.NoError(t, err)
	assert.Equal(t, base.LicenseRefreshCurrent, applied.Result)
	assert.False(t, applied.Changed)
	assert.Len(t, h.settings.upserted, upserts, "nothing written")
	assert.Equal(t, pasted, h.stored(t).Key)
	assert.True(t, h.stored(t).MovedAt.IsZero())
}

// Docker that cannot say the installation's ID fails the check, as the portal
// cannot be asked without it.
func TestRefreshWithoutTheInstallationID(t *testing.T) {
	h := newHarness(t)
	portal := newPortal(t, h)
	h.storeWith(t, payload(), time.Time{})
	h.docker.clusterErr = hperrors.NewNotFound("Swarm cluster")

	applied := h.refresh(t, false)

	assert.Equal(t, base.LicenseRefreshFailed, applied.Result)
	assert.Zero(t, portal.asked())
	assert.NotEmpty(t, h.stored(t).Refresh.LastError)
}

func TestRefreshInfo(t *testing.T) {
	h := newHarness(t)
	_, schedule := h.storeWith(t, payload(), time.Time{})

	info, err := h.RefreshInfo(context.Background())
	assert.NoError(t, err)
	assert.True(t, info.Enabled)
	assert.Equal(t, schedule, info.Refresh.Schedule)

	h.config.License.RefreshDisabled = true
	info, err = h.RefreshInfo(context.Background())
	assert.NoError(t, err)
	assert.False(t, info.Enabled)
}

// A refresh's error never quotes the key, whatever the portal says back.
func TestRefreshLastErrorIsBounded(t *testing.T) {
	h := newHarness(t)
	portal := newPortal(t, h)
	h.storeWith(t, payload(), time.Time{})
	portal.answers(http.StatusBadRequest, `{"code":"ERR_VLD","detail":"`+strings.Repeat("x", 5000)+`"}`)

	h.refresh(t, false)

	assert.Less(t, len(h.stored(t).Refresh.LastError), 600)
}
