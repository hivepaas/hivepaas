package cacheentity

import "github.com/hivepaas/hivepaas/hivepaas_app/pkg/failbackoff"

// AppSecretAttempt counts one user's consecutive wrong app secrets.
//
// It gets its own key space rather than sharing the login counter, because the
// two must not spill into each other: a mistyped app secret should not lock
// somebody out of logging in, and a run of failed logins should not decide
// whether they can reach the security settings.
type AppSecretAttempt = failbackoff.Attempt
