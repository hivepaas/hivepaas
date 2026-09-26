package mcp

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/ulid"
)

// A change is made in two calls: a plan tool answers what would happen and a
// token, and apply_plan carries out exactly what was planned. The plan is kept
// here between the two, bound to the caller, for planTTL, and taken once.
// See docs/superpowers/specs/2026-09-26-mcp-server-phase2-design.md §3.

const (
	planTTL         = 10 * time.Minute
	planTokenPrefix = "mcpp_"
	planKeyLen      = 32
)

// planRepo is where sealed plans wait: cacherepository.MCPPlanRepo.
type planRepo interface {
	Set(ctx context.Context, planID string, sealed []byte, exp time.Duration) error
	GetDel(ctx context.Context, planID string) ([]byte, error)
}

// storedPlan is a planned change: the request that makes it, and who planned it.
type storedPlan struct {
	Tool   string `json:"tool"`
	UserID string `json:"userId"`
	KeyID  string `json:"keyId"`

	Method string          `json:"method"`
	Path   string          `json:"path"`
	Body   json.RawMessage `json:"body,omitempty"`

	// Summary is what the person was shown, in one line: the audit log's.
	Summary string `json:"summary"`
	// Check is what apply compares before sending, the tool's own; empty for a
	// change that depends on nothing that could have moved.
	Check json.RawMessage `json:"check,omitempty"`
}

// errNoSuchPlan is every reason a token is not a plan: expired, used, another
// caller's, or not a token at all. They are one answer, so that a token
// someone else holds says nothing about theirs.
var errNoSuchPlan = &InputError{Message: "no such plan: a plan lasts ten minutes, is used once, " +
	"and only by the key that made it. Plan again."}

var errSealing = errors.New("mcp: sealing a plan")

// savePlan keeps a plan for its caller and answers its token.
func savePlan(ctx context.Context, repo planRepo, c *caller, plan *storedPlan) (token, planID string, err error) {
	plan.UserID, plan.KeyID = c.auth.User.ID, c.keyID
	raw, err := json.Marshal(plan)
	if err != nil {
		return "", "", fmt.Errorf("%w: %w", errSealing, err)
	}
	planID, err = ulid.NewStringULID()
	if err != nil {
		return "", "", fmt.Errorf("%w: %w", errSealing, err)
	}
	key := make([]byte, planKeyLen)
	if _, err = rand.Read(key); err != nil {
		return "", "", fmt.Errorf("%w: %w", errSealing, err)
	}
	sealed, err := seal(key, planID, raw)
	if err != nil {
		return "", "", err
	}
	if err = repo.Set(ctx, planID, sealed, planTTL); err != nil {
		return "", "", fmt.Errorf("%w: %w", errSealing, err)
	}
	return planTokenPrefix + planID + "." + base64.RawURLEncoding.EncodeToString(key), planID, nil
}

// takePlan takes the plan a token stands for, if its caller made it. The plan
// is gone after this, whatever the caller does with it.
func takePlan(ctx context.Context, repo planRepo, c *caller, token string) (*storedPlan, string, error) {
	planID, encodedKey, found := strings.Cut(strings.TrimPrefix(strings.TrimSpace(token), planTokenPrefix), ".")
	if !found || !strings.HasPrefix(strings.TrimSpace(token), planTokenPrefix) {
		return nil, "", errNoSuchPlan
	}
	key, err := base64.RawURLEncoding.DecodeString(encodedKey)
	if err != nil || len(key) != planKeyLen {
		return nil, "", errNoSuchPlan
	}
	sealed, err := repo.GetDel(ctx, planID)
	if err != nil {
		return nil, "", fmt.Errorf("mcp: reading a plan: %w", err)
	}
	if sealed == nil {
		return nil, "", errNoSuchPlan
	}
	raw, err := open(key, planID, sealed)
	if err != nil {
		return nil, "", errNoSuchPlan
	}
	var plan storedPlan
	if err = json.Unmarshal(raw, &plan); err != nil {
		return nil, "", errNoSuchPlan
	}
	if plan.UserID != c.auth.User.ID || plan.KeyID != c.keyID {
		return nil, "", errNoSuchPlan
	}
	return &plan, planID, nil
}

// seal encrypts a plan with AES-256-GCM, the plan's ID bound in as additional
// data so that a sealed plan cannot be moved under another ID. The key is random,
// so it is used as it is: no derivation.
func seal(key []byte, planID string, plaintext []byte) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err = rand.Read(nonce); err != nil {
		return nil, fmt.Errorf("%w: %w", errSealing, err)
	}
	return gcm.Seal(nonce, nonce, plaintext, []byte(planID)), nil
}

func open(key []byte, planID string, sealed []byte) ([]byte, error) {
	gcm, err := newGCM(key)
	if err != nil {
		return nil, err
	}
	if len(sealed) < gcm.NonceSize() {
		return nil, errSealing
	}
	nonce, ciphertext := sealed[:gcm.NonceSize()], sealed[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, ciphertext, []byte(planID))
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errSealing, err)
	}
	return plaintext, nil
}

func newGCM(key []byte) (cipher.AEAD, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errSealing, err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("%w: %w", errSealing, err)
	}
	return gcm, nil
}
