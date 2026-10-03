package entity

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
)

// Every request now asks whether its user is the demo user: with no
// configuration loaded, or no demo user configured, nobody is.
func TestNobodyIsTheDemoUserWithoutOne(t *testing.T) {
	previous := config.Current()
	t.Cleanup(func() { config.SetCurrent(previous) })

	config.SetCurrent(nil)
	assert.False(t, (&User{ID: "u1"}).IsDemoUser())

	config.SetCurrent(&config.Config{})
	assert.False(t, (&User{ID: ""}).IsDemoUser(), "an empty demo id names nobody")

	cfg := &config.Config{}
	cfg.Users.Demo.UserID = "demo"
	config.SetCurrent(cfg)
	assert.True(t, (&User{ID: "demo"}).IsDemoUser())
	assert.False(t, (&User{ID: "u1"}).IsDemoUser())
}
