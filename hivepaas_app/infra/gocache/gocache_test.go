package gocache

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/apperrors"
)

func TestNewCache(t *testing.T) {
	c := NewCache()
	assert.NotNil(t, c)
	assert.NotNil(t, c.client)
}

func TestCache_GetSetDel(t *testing.T) {
	c := NewCache()
	key := "test-key"
	val := "test-value"

	// Get non-existent
	_, err := c.Get(key)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrNotFound))

	// Set
	err = c.Set(key, val, NoExpiration)
	assert.NoError(t, err)

	// Get existent
	got, err := c.Get(key)
	assert.NoError(t, err)
	assert.Equal(t, val, got)

	// Delete
	err = c.Del(key)
	assert.NoError(t, err)

	// Get deleted
	_, err = c.Get(key)
	assert.Error(t, err)
	assert.True(t, errors.Is(err, apperrors.ErrNotFound))
}

func TestCache_TypeSpecificGetters(t *testing.T) {
	c := NewCache()

	t.Run("String", func(t *testing.T) {
		key := "str-key"
		val := "hello"
		_ = c.Set(key, val, NoExpiration)

		v, err := c.GetStr(key)
		assert.NoError(t, err)
		assert.Equal(t, val, v)

		// Wrong type
		_ = c.Set(key, 123, NoExpiration)
		_, err = c.GetStr(key)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrMismatch))
	})

	t.Run("Int", func(t *testing.T) {
		key := "int-key"
		val := 42
		_ = c.Set(key, val, NoExpiration)

		v, err := c.GetInt(key)
		assert.NoError(t, err)
		assert.Equal(t, val, v)

		// Wrong type
		_ = c.Set(key, "not-an-int", NoExpiration)
		_, err = c.GetInt(key)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrMismatch))
	})

	t.Run("Bool", func(t *testing.T) {
		key := "bool-key"
		val := true
		_ = c.Set(key, val, NoExpiration)

		v, err := c.GetBool(key)
		assert.NoError(t, err)
		assert.Equal(t, val, v)

		// Wrong type
		_ = c.Set(key, "not-a-bool", NoExpiration)
		_, err = c.GetBool(key)
		assert.Error(t, err)
		assert.True(t, errors.Is(err, apperrors.ErrMismatch))
	})
}

func TestCache_TTL(t *testing.T) {
	c := NewCache()
	key := "ttl-key"

	// Non-existent
	ttl, err := c.TTL(key)
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), ttl)

	// No expiration
	_ = c.Set(key, "val", NoExpiration)
	ttl, err = c.TTL(key)
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(-1), ttl)

	// With expiration
	expiration := 10 * time.Second
	_ = c.Set(key, "val", expiration)
	ttl, err = c.TTL(key)
	assert.NoError(t, err)
	assert.True(t, ttl > 0)
	assert.True(t, ttl <= expiration)
}

func TestGlobal(t *testing.T) {
	assert.NotNil(t, Global)
}
