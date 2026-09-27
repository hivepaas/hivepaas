package reposerveragentuc

import (
	"os"
	"os/exec"
	"path/filepath"
	"time"

	"github.com/hivepaas/hivepaas/hivepaas_app/config"
	"github.com/hivepaas/hivepaas/hivepaas_app/pkg/logging"
)

const (
	// sessionDirName, under the temp dir, holds a directory per server session:
	// its config, certificate, cache and logs, removed when the session ends.
	sessionDirName = "hivepaas/repo-servers"
	// stopGraceDefault is how long the server has to stop once asked, before it
	// is killed.
	stopGraceDefault = 5 * time.Second
)

// UC runs kopia repository servers on the agent's node, one a session.
type UC struct {
	logger logging.Logger
	// command builds a kopia command with its arguments and environment; tests
	// replace it.
	command     func(args []string, env []string) *exec.Cmd
	sessionRoot string
	// memLimit and startTimeout are the configuration's when not set.
	memLimit     string
	startTimeout time.Duration
	stopGrace    time.Duration
}

func New(logger logging.Logger) *UC {
	return &UC{
		logger:      logger,
		command:     kopiaCommand,
		sessionRoot: filepath.Join(os.TempDir(), sessionDirName),
		stopGrace:   stopGraceDefault,
	}
}

// settings are the server's memory limit and start timeout, read when a
// session starts: the configuration is loaded after the use case is built.
func (uc *UC) settings() (memLimit string, startTimeout time.Duration) {
	memLimit, startTimeout = uc.memLimit, uc.startTimeout
	if memLimit != "" && startTimeout != 0 {
		return memLimit, startTimeout
	}
	cfg := config.Current().Agent.RepoServer
	if memLimit == "" {
		memLimit = cfg.MemLimit
	}
	if startTimeout == 0 {
		startTimeout = cfg.StartTimeout
	}
	return memLimit, startTimeout
}

func kopiaCommand(args []string, env []string) *exec.Cmd {
	cmd := exec.Command("kopia", args...)
	cmd.Env = env
	return cmd
}
