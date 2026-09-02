// Package dblock provides a mutex for work that is too long to hold a database transaction open
// for - pruning a backup repository, syncing snapshots, and the like.
//
// It is built on Postgres advisory locks rather than a locked row, because an advisory lock is
// held by the session instead of the transaction. That is what lets the caller commit, run the
// slow work with no transaction open, and then commit again while still holding the lock. A row
// lock cannot do that: it lives and dies with its transaction.
//
// The lock is per-database, so it also serializes across app instances, and Postgres releases it
// on its own if the process dies.
package dblock

import (
	"context"
	"hash/fnv"

	"github.com/uptrace/bun"

	"github.com/hivepaas/hivepaas/hivepaas_app/hperrors"
	"github.com/hivepaas/hivepaas/hivepaas_app/infra/database"
)

// Lock is a held advisory lock. Release it with Release when the work is done.
type Lock struct {
	conn bun.Conn
	key  int64
	held bool
}

// Key turns a name into an advisory lock key. Two different names can collide, which only makes
// them wait for each other - it never lets two holders of the same name run at once.
func Key(name string) int64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(name))
	return int64(h.Sum64()) //nolint:gosec // the full 64-bit pattern is the key, overflow is fine
}

// TryAcquire takes the lock named by name without waiting. It reports whether the lock was taken;
// when it was not, somebody else is holding it and the caller should give up rather than queue.
//
// The returned Lock pins one connection from the pool for as long as it is held, so always
// Release it - a deferred call right after a successful acquire is the safe shape.
func TryAcquire(ctx context.Context, db *database.DB, name string) (*Lock, bool, error) {
	// The lock belongs to a session, so it has to run on a connection that stays put. Anything
	// taken from the pool per query could unlock on a different connection than it locked on.
	conn, err := db.Conn(ctx)
	if err != nil {
		return nil, false, hperrors.Wrap(err)
	}

	key := Key(name)
	var acquired bool
	err = conn.QueryRowContext(ctx, "SELECT pg_try_advisory_lock(?)", key).Scan(&acquired)
	if err != nil {
		_ = conn.Close()
		return nil, false, hperrors.Wrap(err)
	}
	if !acquired {
		_ = conn.Close()
		return nil, false, nil
	}

	return &Lock{conn: conn, key: key, held: true}, true, nil
}

// Release gives the lock back and returns the connection to the pool. It is safe to call twice.
func (l *Lock) Release(ctx context.Context) error {
	if l == nil || !l.held {
		return nil
	}
	l.held = false

	// Unlock explicitly rather than relying on the connection being reset when it goes back to
	// the pool: a pooled connection may be reused before that ever happens.
	_, unlockErr := l.conn.ExecContext(ctx, "SELECT pg_advisory_unlock(?)", l.key)
	closeErr := l.conn.Close()

	if unlockErr != nil {
		return hperrors.Wrap(unlockErr)
	}
	if closeErr != nil {
		return hperrors.Wrap(closeErr)
	}
	return nil
}
