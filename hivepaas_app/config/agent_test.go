package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The agent's config file carries no [agent] section - config.production.toml has
// none - so the node's secret reaches the agent from the environment alone, and
// the loaded config is the only place it survives: the load clears HP_* from the
// process, which TestLoadConfigClearsEnvAfterwards pins. An agent that read the
// secret from the environment per call would find nothing there to read.
func TestLoadConfigTakesTheAgentSecretFromTheEnvironment(t *testing.T) {
	resetLoadState()
	t.Cleanup(resetLoadState)
	appPath := t.TempDir()
	assert.NoError(t, os.WriteFile(filepath.Join(appPath, configFileName),
		[]byte("env = \"myenv\"\n"), 0o600))

	t.Setenv("HP_APP_PATH", appPath)
	t.Setenv("HP_RUN_MODE", RunModeAgent)
	t.Setenv("HP_AGENT_SECRET_TOKEN", "the-node-secret")

	cfg, err := LoadConfig()
	if assert.NoError(t, err) {
		assert.Equal(t, "the-node-secret", cfg.Agent.SecretToken)
	}
	assert.Empty(t, os.Getenv("HP_AGENT_SECRET_TOKEN"), "HP_* must not survive the load")
}

// An agent node has no config file of its own: the app data directory, where the
// installation's config.toml lives, is on the control plane and is not mounted
// into the agent, which takes only the docker socket and the host filesystem. So
// the agent loads the config file baked into its own image, which HP_CONFIG_FILE
// names and every node that pulled the image therefore has. This is the path an
// agent takes on every node, the control plane's included.
func TestAnAgentLoadsTheConfigFileFromItsImageNotFromTheAppDirectory(t *testing.T) {
	resetLoadState()
	t.Cleanup(resetLoadState)

	// The app data directory as an agent sees it: present, and without the
	// installation's config.toml, which is the control plane's alone.
	appPath := t.TempDir()

	// The config file the agent image ships, as deployment/release/Dockerfile.agent
	// copies config.production.toml into the image beside the binary.
	imageDir := t.TempDir()
	imageConfig := filepath.Join(imageDir, "config.production.toml")
	assert.NoError(t, os.WriteFile(imageConfig, []byte("env = \"production\"\n"), 0o600))

	t.Setenv("HP_APP_PATH", appPath)
	t.Setenv("HP_CONFIG_FILE", imageConfig)
	t.Setenv("HP_RUN_MODE", RunModeAgent)
	t.Setenv("HP_AGENT_PORT", "10001")
	t.Setenv("HP_AGENT_SECRET_TOKEN", "the-node-secret")

	cfg, err := LoadConfig()
	if !assert.NoError(t, err) {
		return
	}
	// The port it listens on and the secret it checks come from the same load: an
	// agent that could not read one could not read the other either.
	assert.Equal(t, 10001, cfg.Agent.Port)
	assert.Equal(t, "the-node-secret", cfg.Agent.SecretToken)
	assert.False(t, cfg.DevMode.Enabled, "production is not dev mode")
}
