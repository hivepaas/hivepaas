package mcp

import (
	"bytes"
	"context"
	"encoding/json"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// memPlans is the plan store in memory.
type memPlans struct {
	mu    sync.Mutex
	plans map[string][]byte
}

func newMemPlans() *memPlans { return &memPlans{plans: map[string][]byte{}} }

func (m *memPlans) Set(_ context.Context, planID string, sealed []byte, _ time.Duration) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.plans[planID] = sealed
	return nil
}

func (m *memPlans) GetDel(_ context.Context, planID string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	sealed := m.plans[planID]
	delete(m.plans, planID)
	return sealed, nil
}

func testPlan() *storedPlan {
	return &storedPlan{Tool: "plan_install_app", Method: "POST", Path: "/projects/p1/prod/apps/from-template",
		Body: json.RawMessage(`{"params":{"password":"hunter2-s3cr3t"}}`), Summary: "install postgres as db"}
}

func callerOf(userID, keyID string) *caller {
	auth := testAuth()
	auth.User.ID = userID
	return &caller{auth: auth, keyID: keyID}
}

func TestAPlanRoundTrips(t *testing.T) {
	repo := newMemPlans()
	token, _, err := savePlan(context.Background(), repo, callerOf("u1", "key1"), testPlan())
	if !assert.NoError(t, err) {
		t.FailNow()
	}
	assert.True(t, strings.HasPrefix(token, "mcpp_"))

	plan, _, err := takePlan(context.Background(), repo, callerOf("u1", "key1"), token)
	if assert.NoError(t, err) {
		assert.Equal(t, "POST", plan.Method)
		assert.JSONEq(t, `{"params":{"password":"hunter2-s3cr3t"}}`, string(plan.Body))
	}
}

func TestAPlanIsUsedOnce(t *testing.T) {
	repo := newMemPlans()
	c := callerOf("u1", "key1")
	token, _, _ := savePlan(context.Background(), repo, c, testPlan())
	_, _, err := takePlan(context.Background(), repo, c, token)
	assert.NoError(t, err)
	_, _, err = takePlan(context.Background(), repo, c, token)
	assert.Equal(t, errNoSuchPlan, err)
}

func TestAPlanIsBoundToItsCaller(t *testing.T) {
	for name, other := range map[string]*caller{
		"another user":           callerOf("u2", "key1"),
		"another key, same user": callerOf("u1", "key2"),
	} {
		repo := newMemPlans()
		token, _, _ := savePlan(context.Background(), repo, callerOf("u1", "key1"), testPlan())
		_, _, err := takePlan(context.Background(), repo, other, token)
		assert.Equal(t, errNoSuchPlan, err, name)
	}
}

func TestATokenCannotBeForged(t *testing.T) {
	repo := newMemPlans()
	c := callerOf("u1", "key1")
	token, _, _ := savePlan(context.Background(), repo, c, testPlan())
	planID, _, _ := strings.Cut(strings.TrimPrefix(token, "mcpp_"), ".")

	for name, forged := range map[string]string{
		"no prefix":       strings.TrimPrefix(token, "mcpp_"),
		"no key":          "mcpp_" + planID,
		"another key":     "mcpp_" + planID + ".AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
		"a short key":     "mcpp_" + planID + ".AAAA",
		"expired or none": "mcpp_01ARZ3NDEKTSV4RRFFQ69G5FAV.AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA",
	} {
		_, _, err := takePlan(context.Background(), repo, c, forged)
		assert.Equal(t, errNoSuchPlan, err, name)
	}
}

// Redis holds the plan sealed, with a key that exists only in the token: an
// install's password is not in it.
func TestAPlanIsUnreadableAtRest(t *testing.T) {
	repo := newMemPlans()
	_, _, err := savePlan(context.Background(), repo, callerOf("u1", "key1"), testPlan())
	if !assert.NoError(t, err) || !assert.Len(t, repo.plans, 1) {
		t.FailNow()
	}
	for _, sealed := range repo.plans {
		for _, clear := range []string{"hunter2", "s3cr3t", "from-template", "install postgres", "u1"} {
			assert.False(t, bytes.Contains(sealed, []byte(clear)), clear)
		}
	}
}
