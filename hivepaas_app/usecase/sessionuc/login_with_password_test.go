package sessionuc

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/base"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity"
	"github.com/hivepaas/hivepaas/hivepaas_app/entity/cacheentity"
	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/timeutil"
)

// fakeLoginAttemptRepo stands in for redis, answering ErrNotFound for a user with
// no stored attempt the way the real one does.
type fakeLoginAttemptRepo struct {
	attempt *cacheentity.LoginAttempt
	dels    int
}

func (f *fakeLoginAttemptRepo) Get(_ context.Context, _ string) (*cacheentity.LoginAttempt, error) {
	if f.attempt == nil {
		return nil, hperrors.ErrNotFound
	}
	return f.attempt, nil
}

func (f *fakeLoginAttemptRepo) Set(
	_ context.Context, _ string, attempt *cacheentity.LoginAttempt, _ time.Duration,
) error {
	f.attempt = attempt
	return nil
}

func (f *fakeLoginAttemptRepo) Del(_ context.Context, _ string) error {
	f.dels++
	f.attempt = nil
	return nil
}

func newLoginUCTest(attempt *cacheentity.LoginAttempt) (*UC, *fakeLoginAttemptRepo) {
	repo := &fakeLoginAttemptRepo{attempt: attempt}
	return &UC{cacheLoginAttemptRepo: repo}, repo
}

func testUser() *entity.User {
	return &entity.User{ID: "user-1"}
}

func TestAllowPasswordLoginAtTheMoment(t *testing.T) {
	ctx := context.Background()
	now := timeutil.NowUTC()

	t.Run("no failures yet", func(t *testing.T) {
		uc, _ := newLoginUCTest(nil)
		attempt, err := uc.allowPasswordLoginAtTheMoment(ctx, testUser())
		assert.NoError(t, err)
		assert.Nil(t, attempt)
	})

	t.Run("below the threshold", func(t *testing.T) {
		uc, _ := newLoginUCTest(&cacheentity.LoginAttempt{
			Fails: maxPasswordFailsInARow - 1, LastFailAt: now,
		})
		_, err := uc.allowPasswordLoginAtTheMoment(ctx, testUser())
		assert.NoError(t, err)
	})

	t.Run("at the threshold, just failed", func(t *testing.T) {
		uc, _ := newLoginUCTest(&cacheentity.LoginAttempt{
			Fails: maxPasswordFailsInARow, LastFailAt: now,
		})
		_, err := uc.allowPasswordLoginAtTheMoment(ctx, testUser())
		assert.ErrorIs(t, err, hperrors.ErrTooManyLoginFailures)
	})

	t.Run("at the threshold, waited it out", func(t *testing.T) {
		uc, _ := newLoginUCTest(&cacheentity.LoginAttempt{
			Fails: maxPasswordFailsInARow, LastFailAt: now.Add(-time.Hour),
		})
		_, err := uc.allowPasswordLoginAtTheMoment(ctx, testUser())
		assert.NoError(t, err)
	})

	// The behavior that changed. Measured from the first failure, a run spread
	// over an hour had already outlived its window before it reached the
	// threshold, so a patient guesser was never slowed down at all.
	t.Run("guessing slowly no longer avoids the wait", func(t *testing.T) {
		uc, _ := newLoginUCTest(&cacheentity.LoginAttempt{
			Fails:       maxPasswordFailsInARow,
			FirstFailAt: now.Add(-time.Hour),
			LastFailAt:  now,
		})
		_, err := uc.allowPasswordLoginAtTheMoment(ctx, testUser())
		assert.ErrorIs(t, err, hperrors.ErrTooManyLoginFailures)
	})

	// Entries written before LastFailAt existed must not read as an elapsed wait,
	// or the deploy that added the field would release every lockout at once.
	t.Run("an entry without the last timestamp still counts", func(t *testing.T) {
		uc, _ := newLoginUCTest(&cacheentity.LoginAttempt{
			Fails: maxPasswordFailsInARow, FirstFailAt: now,
		})
		_, err := uc.allowPasswordLoginAtTheMoment(ctx, testUser())
		assert.ErrorIs(t, err, hperrors.ErrTooManyLoginFailures)
	})

	t.Run("SSO-only accounts cannot use a password at all", func(t *testing.T) {
		uc, _ := newLoginUCTest(nil)
		user := testUser()
		user.SecurityOption = base.UserSecurityEnforceSSO

		_, err := uc.allowPasswordLoginAtTheMoment(ctx, user)
		assert.ErrorIs(t, err, hperrors.ErrSSORequired)
	})
}

func TestSavePasswordCheckingStatus(t *testing.T) {
	ctx := context.Background()

	t.Run("a failure is counted and timestamped", func(t *testing.T) {
		uc, repo := newLoginUCTest(nil)
		assert.NoError(t, uc.savePasswordCheckingStatus(ctx, testUser(), nil, false))

		assert.Equal(t, 1, repo.attempt.Fails)
		assert.False(t, repo.attempt.FirstFailAt.IsZero())
		assert.False(t, repo.attempt.LastFailAt.IsZero())
	})

	t.Run("a further failure moves only the last timestamp", func(t *testing.T) {
		firstFail := timeutil.NowUTC().Add(-time.Hour)
		attempt := &cacheentity.LoginAttempt{Fails: 1, FirstFailAt: firstFail, LastFailAt: firstFail}
		uc, repo := newLoginUCTest(attempt)

		assert.NoError(t, uc.savePasswordCheckingStatus(ctx, testUser(), attempt, false))
		assert.Equal(t, 2, repo.attempt.Fails)
		assert.Equal(t, firstFail, repo.attempt.FirstFailAt)
		assert.True(t, repo.attempt.LastFailAt.After(firstFail))
	})

	t.Run("a success ends the run", func(t *testing.T) {
		attempt := &cacheentity.LoginAttempt{Fails: 3, LastFailAt: timeutil.NowUTC()}
		uc, repo := newLoginUCTest(attempt)

		assert.NoError(t, uc.savePasswordCheckingStatus(ctx, testUser(), attempt, true))
		assert.Equal(t, 1, repo.dels)
		assert.Nil(t, repo.attempt)
	})
}
